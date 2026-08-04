package engine

// exit_quarantine_command.go is the engine-owned command side of change a079:
// lifting an exit snapshot quarantine.
//
// # Why it lives on the position policy service
//
// Not because the two are the same domain — they are not, and design D1 is about
// exactly that — but because the capability the release needs is one the policy
// service already holds under the engine flock: the single writable journal. A
// second service would need a second wiring point in `engine run`, and every
// wiring point in that function is a place where a future reader has to work out
// which handle owns what.
//
// The repository is reached by asserting the handle the service already stores.
// That keeps positionPolicyRepository unwidened — widening it would break every
// test fake in the policy suite, none of which is about quarantines — and keeps
// NewPositionPolicyCommandService untouched.
//
// # Why the capability code is not shared with the policy pipeline
//
// It looks the same and it is deliberately separate. Extracting a common helper
// means editing position_policy_command.go, which is the ledger-facing approval
// path for policy changes, and it would put two expiry policies in one place —
// re-adoption's 15-second observation freshness has no meaning for a release,
// and a future edit to one must not silently move the other.

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const (
	exitQuarantineCapabilityDomain = "tossos/exit-quarantine/release/v1"
	// exitQuarantineCapabilityDelay is the danger delay. Releasing returns a
	// position to automated judgement, which is the same class of decision as a
	// lifecycle release, so it gets the same three seconds.
	exitQuarantineCapabilityDelay = 3 * time.Second
	exitQuarantineCapabilityTTL   = 5 * time.Minute
	exitQuarantineCapabilityLimit = 64
)

// exitQuarantineRepository is the narrow ledger capability a release needs: read
// the active rows, and close exactly one of them. It cannot write an exit state,
// a position, or a policy lifecycle, and that is the point — the whole safety
// claim of a079 is that a release moves one column.
type exitQuarantineRepository interface {
	ActiveExitSnapshotQuarantines(context.Context) ([]exitquarantine.Row, error)
	ReleaseExitSnapshotQuarantine(ctx context.Context, positionID string,
		generation, expectedVersion int64, kind, evidence string) error
}

type exitQuarantineCapability struct {
	digest     [sha256.Size]byte
	instanceID string
	domain     string
	request    exitquarantine.Request
	row        exitquarantine.Row
	issuedAt   time.Time
	notBefore  time.Time
	expiresAt  time.Time
}

// quarantineRepo reports the ledger capability, or that this build has none.
//
// ErrUnwired rather than a panic or a nil dereference: a console talking to an
// engine that predates a079 must get a screen that says the control plane does
// not offer this, not a 500.
func (s *PositionPolicyCommandService) quarantineRepo() (exitQuarantineRepository, error) {
	repo, ok := s.j.(exitQuarantineRepository)
	if !ok {
		return nil, exitquarantine.ErrUnwired
	}
	return repo, nil
}

// Quarantines lists the active quarantines an operator could act on.
func (s *PositionPolicyCommandService) Quarantines(ctx context.Context) ([]exitquarantine.Row, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repo, err := s.quarantineRepo()
	if err != nil {
		return nil, err
	}
	return repo.ActiveExitSnapshotQuarantines(ctx)
}

// PreviewQuarantineRelease binds one specific quarantine row to a one-time
// grant.
//
// The row is re-read here rather than trusted from the caller: the console's
// list may be seconds old, and the version bound into the grant has to be the
// one the ledger holds now, not the one a browser remembered.
func (s *PositionPolicyCommandService) PreviewQuarantineRelease(ctx context.Context,
	req exitquarantine.Request) (exitquarantine.Preview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := req.Validate(); err != nil {
		return exitquarantine.Preview{}, err
	}
	repo, err := s.quarantineRepo()
	if err != nil {
		return exitquarantine.Preview{}, err
	}
	row, err := s.findQuarantine(ctx, repo, req)
	if err != nil {
		return exitquarantine.Preview{}, err
	}
	token, err := s.issueQuarantineCapability(req, row)
	if err != nil {
		return exitquarantine.Preview{}, err
	}
	return exitquarantine.Preview{
		Row: row, Capability: token,
		WaitSeconds: int(exitQuarantineCapabilityDelay / time.Second),
	}, nil
}

// ReleaseQuarantine consumes a grant and closes the bound quarantine row.
func (s *PositionPolicyCommandService) ReleaseQuarantine(ctx context.Context,
	apply exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	grant, index, err := s.verifyQuarantineCapability(apply.Capability)
	if err != nil {
		return exitquarantine.Result{}, err
	}
	if !apply.Confirmed {
		// A missing danger acknowledgement is recoverable and therefore does not
		// consume the grant. The operator can tick the box and apply the same
		// preview.
		return exitquarantine.Result{}, exitquarantine.ErrConfirmationRequired
	}
	repo, err := s.quarantineRepo()
	if err != nil {
		return exitquarantine.Result{}, err
	}
	// Consumed before any ledger access. Restoring a partially attempted grant
	// would make replay safety depend on classifying every future storage error
	// correctly.
	s.quarantineGrants = append(s.quarantineGrants[:index], s.quarantineGrants[index+1:]...)

	// findQuarantine is the compare-and-swap: it resolves the row the grant names
	// and refuses unless the ledger still holds that exact version. A second
	// comparison against grant.row here would be unreachable — preview only issues
	// a grant whose row version equals its request version — and an unreachable
	// guard on a safety path reads as protection that is not being tested.
	current, err := s.findQuarantine(ctx, repo, grant.request)
	if err != nil {
		return exitquarantine.Result{}, err
	}

	at := s.clk.Now().UTC()
	evidence := exitquarantine.ComposeEvidence(grant.request.Actor, current, at)
	err = repo.ReleaseExitSnapshotQuarantine(ctx, current.PositionID, current.Generation,
		current.Version, journal.QuarantineReleaseHumanRepair, evidence)
	if err != nil {
		if errors.Is(err, journal.ErrExitSnapshotReleaseStale) {
			// Another writer closed this exact row between the read above and
			// this update. That is a version conflict in every sense that matters
			// to the operator, and nothing was written.
			return exitquarantine.Result{}, fmt.Errorf("%w: the quarantine was already released",
				exitquarantine.ErrVersionMismatch)
		}
		return exitquarantine.Result{}, err
	}
	return exitquarantine.Result{
		Row: current, ReleasedAt: journal.RFC3339(at), Evidence: evidence,
	}, nil
}

// findQuarantine resolves the one active row a request names.
func (s *PositionPolicyCommandService) findQuarantine(ctx context.Context,
	repo exitQuarantineRepository, req exitquarantine.Request) (exitquarantine.Row, error) {
	rows, err := repo.ActiveExitSnapshotQuarantines(ctx)
	if err != nil {
		return exitquarantine.Row{}, err
	}
	for _, row := range rows {
		if row.PositionID != strings.TrimSpace(req.PositionID) || row.Generation != req.Generation {
			continue
		}
		if row.Version != req.Version {
			return exitquarantine.Row{}, fmt.Errorf("%w: ledger holds v%d, the request named v%d",
				exitquarantine.ErrVersionMismatch, row.Version, req.Version)
		}
		return row, nil
	}
	return exitquarantine.Row{}, fmt.Errorf("%w: %s generation %d",
		exitquarantine.ErrNotQuarantined, req.PositionID, req.Generation)
}

func (s *PositionPolicyCommandService) issueQuarantineCapability(req exitquarantine.Request,
	row exitquarantine.Row) (string, error) {
	if err := s.ensureInstanceID(); err != nil {
		return "", err
	}
	now := s.clk.Now().UTC()
	active := s.quarantineGrants[:0]
	for _, grant := range s.quarantineGrants {
		if now.Before(grant.expiresAt) {
			active = append(active, grant)
		}
	}
	s.quarantineGrants = active
	if len(s.quarantineGrants) >= exitQuarantineCapabilityLimit {
		return "", errors.New("engine: too many outstanding exit quarantine capabilities")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("engine: generating exit quarantine capability: %w", err)
	}
	s.quarantineGrants = append(s.quarantineGrants, exitQuarantineCapability{
		digest: sha256.Sum256(raw), instanceID: s.instanceID,
		domain: exitQuarantineCapabilityDomain, request: req, row: row,
		issuedAt: now, notBefore: now.Add(exitQuarantineCapabilityDelay),
		expiresAt: now.Add(exitQuarantineCapabilityTTL),
	})
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *PositionPolicyCommandService) verifyQuarantineCapability(token string) (exitQuarantineCapability, int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil || len(raw) != 32 {
		return exitQuarantineCapability{}, -1, exitquarantine.ErrCapabilityInvalid
	}
	digest := sha256.Sum256(raw)
	match := -1
	for index := range s.quarantineGrants {
		if subtle.ConstantTimeCompare(digest[:], s.quarantineGrants[index].digest[:]) == 1 {
			match = index
		}
	}
	if match < 0 {
		return exitQuarantineCapability{}, -1, exitquarantine.ErrCapabilityInvalid
	}
	grant := s.quarantineGrants[match]
	if subtle.ConstantTimeCompare([]byte(grant.instanceID), []byte(s.instanceID)) != 1 ||
		subtle.ConstantTimeCompare([]byte(grant.domain), []byte(exitQuarantineCapabilityDomain)) != 1 {
		s.quarantineGrants = append(s.quarantineGrants[:match], s.quarantineGrants[match+1:]...)
		return exitQuarantineCapability{}, -1, exitquarantine.ErrCapabilityInvalid
	}
	now := s.clk.Now().UTC()
	if now.Before(grant.issuedAt) {
		// The clock moved backwards under an outstanding grant. Nothing that
		// depends on elapsed time can be trusted, so the grant goes.
		s.quarantineGrants = append(s.quarantineGrants[:match], s.quarantineGrants[match+1:]...)
		return exitQuarantineCapability{}, -1, exitquarantine.ErrCapabilityInvalid
	}
	if now.Before(grant.notBefore) {
		return exitQuarantineCapability{}, -1, exitquarantine.ErrCapabilityTooEarly
	}
	if !now.Before(grant.expiresAt) {
		s.quarantineGrants = append(s.quarantineGrants[:match], s.quarantineGrants[match+1:]...)
		return exitQuarantineCapability{}, -1, exitquarantine.ErrCapabilityExpired
	}
	return grant, match, nil
}
