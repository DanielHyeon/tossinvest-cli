package engine

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const (
	positionPolicyCapabilityDomain = "tossos/position-policy/apply/v1"
	positionPolicyCapabilityDelay  = 3 * time.Second
	positionPolicyCapabilityTTL    = 5 * time.Minute
	// Re-adopt's exact approved after-state contains a quote-derived t0. Keep its
	// authority deliberately short while leaving ordinary policy grants usable
	// for the broader capability TTL. The 15-second inclusive boundary leaves a
	// 12-second action window after the mandatory danger delay.
	positionPolicyReadoptFreshness = 15 * time.Second
	positionPolicyCapabilityLimit  = 256
)

type positionPolicyPrices interface {
	Prices(context.Context, []string) ([]domain.Quote, error)
}

type positionPolicyBlockReader interface {
	Blocks() []reconcile.Block
}

type positionPolicyRepository interface {
	PositionPolicies(context.Context) ([]positionpolicy.State, error)
	PositionPolicy(context.Context, string) (positionpolicy.State, error)
	PreviewPositionPolicy(context.Context, positionpolicy.Request) (positionpolicy.Preview, error)
	ApplyPositionPolicy(context.Context, positionpolicy.Request) (positionpolicy.State, error)
}

// PositionPolicyCommandService runs inside the engine process and reuses the
// journal handle already opened under the engine flock. It never opens, closes,
// or migrates a journal. The console reaches it only through the authenticated
// loopback transport in position_policy_transport.go.
type PositionPolicyCommandService struct {
	mu           sync.Mutex
	j            positionPolicyRepository
	prices       positionPolicyPrices
	retrier      *execgw.Retrier
	adoption     config.Adoption
	blocks       positionPolicyBlockReader
	commonPolicy string
	clk          clock.Clock
	instanceID   string
	capabilities []positionPolicyCapability
	// quarantineGrants is change a069's separate capability store. Its zero value
	// is the whole of its initialisation, which is why the constructor below is
	// untouched. See exit_quarantine_command.go for why it is not shared with
	// capabilities above.
	quarantineGrants []exitQuarantineCapability
}

type positionPolicyCapability struct {
	digest     [sha256.Size]byte
	instanceID string
	domain     string
	request    positionpolicy.Request
	before     positionpolicy.State
	after      positionpolicy.State
	issuedAt   time.Time
	notBefore  time.Time
	expiresAt  time.Time
}

func NewPositionPolicyCommandService(ectx *Context, clk clock.Clock) (*PositionPolicyCommandService, error) {
	if ectx == nil || ectx.Journal == nil {
		return nil, errors.New("engine: position policy command service requires the owned journal")
	}
	if clk == nil {
		clk = clock.System()
	}
	return &PositionPolicyCommandService{
		j: ectx.Journal, prices: ectx.Official, retrier: ectx.Retrier,
		adoption:     ectx.Config.Engine.Adoption,
		blocks:       ectx.Reconcile,
		commonPolicy: strings.TrimSpace(ectx.Config.Engine.ExitPolicy.CommonPolicy), clk: clk,
	}, nil
}

// Runtime returns the engine instance's immutable startup adoption settings
// and the same active tracker projection that gates external adoption. It is a
// read-only snapshot and deliberately carries no account identifier or release
// capability.
func (s *PositionPolicyCommandService) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	runtime := positionpolicy.ManagementRuntime{
		Effective: positionpolicy.NewAdoptionSettings(s.adoption.Enabled,
			s.adoption.DefaultStopPct, s.adoption.IncludeSymbols, s.adoption.ExcludeSymbols,
			s.adoption.Rejected),
		EffectiveKnown: true,
		BlockSource:    positionpolicy.AdoptionBlockingTrackerProjection,
		Blocks:         []positionpolicy.ReconcileBlock{},
	}
	if s.blocks == nil {
		return runtime, nil
	}
	for _, block := range s.blocks.Blocks() {
		runtime.Blocks = append(runtime.Blocks, positionpolicy.NewReconcileBlock(
			positionPolicyReconcileScope(block.Scope), block.Market, block.Symbol,
			string(block.Reason), block.Detail, block.Since, block.Permanent))
	}
	return runtime, nil
}

func positionPolicyReconcileScope(scope reconcile.Scope) positionpolicy.ReconcileScope {
	switch scope {
	case reconcile.ScopeAccount:
		return positionpolicy.ScopeAccount
	case reconcile.ScopeMarket:
		return positionpolicy.ScopeMarket
	case reconcile.ScopeSymbol:
		return positionpolicy.ScopeSymbol
	default:
		return positionpolicy.ReconcileScope(scope)
	}
}

func (s *PositionPolicyCommandService) List(ctx context.Context) ([]positionpolicy.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.j.PositionPolicies(ctx)
}

func (s *PositionPolicyCommandService) Preview(ctx context.Context, req positionpolicy.Request) (positionpolicy.Preview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prepared, err := s.prepare(ctx, req)
	if err != nil {
		return positionpolicy.Preview{}, err
	}
	preview, err := s.j.PreviewPositionPolicy(ctx, prepared)
	if err != nil {
		return positionpolicy.Preview{}, err
	}
	token, err := s.issueCapability(prepared, preview)
	if err != nil {
		return positionpolicy.Preview{}, err
	}
	preview.Capability = token
	return preview, nil
}

func (s *PositionPolicyCommandService) Apply(ctx context.Context,
	apply positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	grant, index, err := s.verifyCapability(apply.Capability)
	if err != nil {
		return positionpolicy.State{}, err
	}
	if (grant.request.Action == positionpolicy.ActionRelease ||
		grant.request.Action == positionpolicy.ActionReadopt) && !apply.Confirmed {
		// A missing UI danger acknowledgement is recoverable and therefore does
		// not consume the grant. The engine checks the bound action rather than
		// trusting a browser-provided action field.
		return positionpolicy.State{}, positionpolicy.ErrConfirmationRequired
	}
	// Once a capability reaches an authorized Apply attempt it is consumed
	// before any repository read/write. Failure is fail-closed: restoring a
	// stale or partially attempted authority would make replay safety depend on
	// classifying every future storage error correctly.
	s.capabilities = append(s.capabilities[:index], s.capabilities[index+1:]...)
	current, err := s.j.PositionPolicy(ctx, grant.request.PositionID)
	if err != nil {
		return positionpolicy.State{}, err
	}
	if current != grant.before {
		return positionpolicy.State{}, positionpolicy.ErrVersionMismatch
	}
	result, err := s.j.ApplyPositionPolicy(ctx, grant.request)
	if err != nil {
		return positionpolicy.State{}, err
	}
	if result != grant.after {
		return positionpolicy.State{}, fmt.Errorf("%w: committed result differs from bound preview",
			positionpolicy.ErrCapabilityInvalid)
	}
	return result, nil
}

func (s *PositionPolicyCommandService) issueCapability(req positionpolicy.Request,
	preview positionpolicy.Preview) (string, error) {
	if err := s.ensureInstanceID(); err != nil {
		return "", err
	}
	now := s.clk.Now().UTC()
	active := s.capabilities[:0]
	for _, grant := range s.capabilities {
		if now.Before(grant.expiresAt) {
			active = append(active, grant)
		}
	}
	s.capabilities = active
	if len(s.capabilities) >= positionPolicyCapabilityLimit {
		return "", errors.New("engine: too many outstanding position policy capabilities")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("engine: generating position policy capability: %w", err)
	}
	digest := sha256.Sum256(raw)
	notBefore := now
	if req.Action == positionpolicy.ActionRelease || req.Action == positionpolicy.ActionReadopt {
		notBefore = now.Add(positionPolicyCapabilityDelay)
	}
	s.capabilities = append(s.capabilities, positionPolicyCapability{
		digest: digest, instanceID: s.instanceID, domain: positionPolicyCapabilityDomain,
		request: req, before: preview.Before, after: preview.After,
		issuedAt: now, notBefore: notBefore, expiresAt: now.Add(positionPolicyCapabilityTTL),
	})
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *PositionPolicyCommandService) ensureInstanceID() error {
	if s.instanceID != "" {
		return nil
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Errorf("engine: generating position policy capability instance: %w", err)
	}
	s.instanceID = base64.RawURLEncoding.EncodeToString(raw)
	return nil
}

func (s *PositionPolicyCommandService) verifyCapability(token string) (positionPolicyCapability, int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil || len(raw) != 32 {
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityInvalid
	}
	digest := sha256.Sum256(raw)
	match := -1
	for index := range s.capabilities {
		if subtle.ConstantTimeCompare(digest[:], s.capabilities[index].digest[:]) == 1 {
			match = index
		}
	}
	if match < 0 {
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityInvalid
	}
	grant := s.capabilities[match]
	if subtle.ConstantTimeCompare([]byte(grant.instanceID), []byte(s.instanceID)) != 1 ||
		subtle.ConstantTimeCompare([]byte(grant.domain), []byte(positionPolicyCapabilityDomain)) != 1 {
		s.capabilities = append(s.capabilities[:match], s.capabilities[match+1:]...)
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityInvalid
	}
	now := s.clk.Now().UTC()
	if now.Before(grant.issuedAt) {
		s.capabilities = append(s.capabilities[:match], s.capabilities[match+1:]...)
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityStale
	}
	if grant.request.Action == positionpolicy.ActionReadopt &&
		now.After(grant.issuedAt.Add(positionPolicyReadoptFreshness)) {
		s.capabilities = append(s.capabilities[:match], s.capabilities[match+1:]...)
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityStale
	}
	if now.Before(grant.notBefore) {
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityTooEarly
	}
	if !now.Before(grant.expiresAt) {
		s.capabilities = append(s.capabilities[:match], s.capabilities[match+1:]...)
		return positionPolicyCapability{}, -1, positionpolicy.ErrCapabilityExpired
	}
	return grant, match, nil
}

// prepare discards client time/observation fields. Re-adopt t0 is an engine
// price observation taken immediately before the journal transaction; its stop
// and policy come from the engine's validated configuration/registry.
func (s *PositionPolicyCommandService) prepare(ctx context.Context,
	req positionpolicy.Request) (positionpolicy.Request, error) {
	req.At = s.clk.Now().UTC()
	req.ReAdoption = nil
	if req.Action != positionpolicy.ActionRelease && req.Action != positionpolicy.ActionReadopt {
		return req, nil
	}
	state, err := s.j.PositionPolicy(ctx, req.PositionID)
	if err != nil {
		return positionpolicy.Request{}, err
	}
	if !state.ExternalLifecycleEligible() {
		return positionpolicy.Request{}, positionpolicy.ErrIneligible
	}
	if req.Action == positionpolicy.ActionRelease {
		return req, nil
	}
	if s.prices == nil {
		return positionpolicy.Request{}, errors.New("engine: authoritative price reader is unavailable")
	}
	observedAt := s.clk.Now().UTC()
	var quotes []domain.Quote
	read := func(readCtx context.Context) error {
		var readErr error
		quotes, readErr = s.prices.Prices(readCtx, []string{state.Symbol})
		return readErr
	}
	if s.retrier != nil {
		err = s.retrier.Query(ctx, execgw.QueryPrice, read)
	} else {
		err = read(ctx)
	}
	if err != nil {
		return positionpolicy.Request{}, fmt.Errorf("engine: observing re-adopt t0 for %s: %w", state.Symbol, err)
	}
	observed := ""
	for _, quote := range quotes {
		if strings.EqualFold(strings.TrimSpace(quote.Symbol), strings.TrimSpace(state.Symbol)) && quote.Last > 0 {
			observed = decimalOf(quote.Last)
			break
		}
	}
	if observed == "" {
		return positionpolicy.Request{}, fmt.Errorf("engine: no positive authoritative price for %s", state.Symbol)
	}
	stopPct := s.adoption.DefaultStopPct
	if err := exitpolicy.ValidateStopPct(stopPct); err != nil {
		stopPct = 0.05
	}
	stop, err := exitpolicy.SyntheticStop(observed, stopPct)
	if err != nil {
		return positionpolicy.Request{}, fmt.Errorf("engine: deriving re-adopt stop: %w", err)
	}
	policyID := strings.TrimSpace(state.DesiredPolicyID)
	if policyID == "" {
		policyID = s.commonPolicy
	}
	req.ReAdoption = &positionpolicy.ReAdoptionObservation{
		ObservedPrice: observed, SyntheticStop: stop,
		ObservedAt: journal.RFC3339(observedAt), PolicyID: policyID,
	}
	return req, nil
}
