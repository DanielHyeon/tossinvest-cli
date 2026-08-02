package journal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// schemaV12 adds a generation-scoped control record beside, never inside, the
// immutable exit snapshot. That separation is what lets an override be desired
// state without silently reinterpreting a position already protected under a
// frozen a042 snapshot.
const schemaV12 = `
ALTER TABLE exit_states ADD COLUMN lifecycle_generation INTEGER;
ALTER TABLE exit_events ADD COLUMN lifecycle_generation INTEGER;

CREATE TABLE position_policy_lifecycles (
	position_id         TEXT NOT NULL REFERENCES positions(id),
	adoption_generation INTEGER NOT NULL CHECK (adoption_generation >= 1),
	version             INTEGER NOT NULL CHECK (version >= 1),
	status              TEXT NOT NULL CHECK (status IN ('MANAGED','RELEASED')),
	desired_policy_id   TEXT,
	observed_at         TEXT NOT NULL,
	updated_at          TEXT NOT NULL,
	PRIMARY KEY(position_id, adoption_generation)
) STRICT;

CREATE INDEX idx_position_policy_current
	ON position_policy_lifecycles(position_id, adoption_generation DESC);

CREATE TABLE position_policy_events (
	id                  INTEGER PRIMARY KEY AUTOINCREMENT,
	position_id         TEXT NOT NULL,
	adoption_generation INTEGER NOT NULL,
	action              TEXT NOT NULL CHECK (action IN ('OVERRIDE','INHERIT','RELEASE','READOPT')),
	before_json         TEXT NOT NULL,
	after_json          TEXT NOT NULL,
	actor               TEXT NOT NULL CHECK (actor = 'LOCAL_OPERATOR'),
	reason              TEXT NOT NULL CHECK (reason IN
		('OPERATOR_POLICY_OVERRIDE','OPERATOR_POLICY_INHERIT','OPERATOR_RELEASE','OPERATOR_READOPT')),
	created_at          TEXT NOT NULL,
	FOREIGN KEY(position_id, adoption_generation)
		REFERENCES position_policy_lifecycles(position_id, adoption_generation)
) STRICT;

CREATE INDEX idx_position_policy_events
	ON position_policy_events(position_id, adoption_generation, id);
`

// PositionPolicy returns the latest generation. Version zero is an explicit
// virtual row: the position exists and no lifecycle command has committed yet.
func (j *Journal) PositionPolicy(ctx context.Context, positionID string) (positionpolicy.State, error) {
	tx, err := j.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return positionpolicy.State{}, err
	}
	defer tx.Rollback()
	return positionPolicyTx(ctx, tx, positionID)
}

// PositionPolicies lists one current row per live projected position.
func (j *Journal) PositionPolicies(ctx context.Context) ([]positionpolicy.State, error) {
	return queryPositionPolicies(ctx, j.db)
}

// PositionPolicies exposes the same lifecycle SELECT through the compile-time
// read-only handle used by API processes. It cannot preview or apply a command.
func (r *ReadOnly) PositionPolicies(ctx context.Context) ([]positionpolicy.State, error) {
	return queryPositionPolicies(ctx, r.db)
}

type positionPolicyQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func queryPositionPolicies(ctx context.Context, queryer positionPolicyQueryer) ([]positionpolicy.State, error) {
	rows, err := queryer.QueryContext(ctx, positionPolicySelect+` WHERE p.state <> 'CLOSED'
		ORDER BY p.account_ref,p.market,p.symbol,p.instance_seq`)
	if err != nil {
		return nil, fmt.Errorf("journal: listing position policy scopes: %w", err)
	}
	defer rows.Close()
	var out []positionpolicy.State
	for rows.Next() {
		state, err := scanPositionPolicyScope(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, state)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// PreviewPositionPolicy validates the exact CAS precondition without writing.
func (j *Journal) PreviewPositionPolicy(ctx context.Context, req positionpolicy.Request) (positionpolicy.Preview, error) {
	if err := req.Validate(); err != nil {
		return positionpolicy.Preview{}, err
	}
	current, err := j.PositionPolicy(ctx, req.PositionID)
	if err != nil {
		return positionpolicy.Preview{}, err
	}
	return previewPositionPolicy(current, req)
}

// ApplyPositionPolicy compares, appends the lifecycle/audit, and commits them as
// one FULL/WAL transaction. A stale generation is the same refusal as a stale
// version; neither is retried or rebased automatically.
func (j *Journal) ApplyPositionPolicy(ctx context.Context, req positionpolicy.Request) (positionpolicy.State, error) {
	if err := req.Validate(); err != nil {
		return positionpolicy.State{}, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return positionpolicy.State{}, fmt.Errorf("journal: beginning position policy command: %w", err)
	}
	defer tx.Rollback()

	current, err := positionPolicyTx(ctx, tx, req.PositionID)
	if err != nil {
		return positionpolicy.State{}, err
	}
	if current.AdoptionGeneration != req.ExpectedGeneration || current.Version != req.ExpectedVersion {
		return positionpolicy.State{}, fmt.Errorf("%w: current generation=%d version=%d",
			positionpolicy.ErrVersionMismatch, current.AdoptionGeneration, current.Version)
	}
	preview, err := previewPositionPolicy(current, req)
	if err != nil {
		return positionpolicy.State{}, err
	}
	after := preview.After

	if req.Action == positionpolicy.ActionReadopt {
		if err := resetExitStateForReadoptTx(ctx, tx, req.PositionID, after.AdoptionGeneration,
			*req.ReAdoption, after.UpdatedAt); err != nil {
			return positionpolicy.State{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO position_policy_lifecycles
			(position_id,adoption_generation,version,status,desired_policy_id,observed_at,updated_at)
			VALUES (?,?,?,?,?,?,?)`, after.PositionID, after.AdoptionGeneration, after.Version,
			after.Status, nullableString(after.DesiredPolicyID), after.ObservedAt, after.UpdatedAt)
	} else if current.Version == 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO position_policy_lifecycles
			(position_id,adoption_generation,version,status,desired_policy_id,observed_at,updated_at)
			VALUES (?,?,?,?,?,?,?)`, after.PositionID, after.AdoptionGeneration, after.Version,
			after.Status, nullableString(after.DesiredPolicyID), after.ObservedAt, after.UpdatedAt)
	} else {
		var result sql.Result
		result, err = tx.ExecContext(ctx, `UPDATE position_policy_lifecycles
			SET version=?,status=?,desired_policy_id=?,updated_at=?
			WHERE position_id=? AND adoption_generation=? AND version=?`, after.Version, after.Status,
			nullableString(after.DesiredPolicyID), after.UpdatedAt, after.PositionID,
			after.AdoptionGeneration, current.Version)
		if err == nil {
			var affected int64
			affected, err = result.RowsAffected()
			if err == nil && affected != 1 {
				return positionpolicy.State{}, positionpolicy.ErrVersionMismatch
			}
		}
	}
	if err != nil {
		if isUniqueViolation(err) {
			return positionpolicy.State{}, positionpolicy.ErrVersionMismatch
		}
		return positionpolicy.State{}, fmt.Errorf("journal: applying position policy command: %w", err)
	}
	if err := j.runExitWriteHook("position_policy_after_state"); err != nil {
		return positionpolicy.State{}, err
	}

	beforeJSON, err := json.Marshal(current)
	if err != nil {
		return positionpolicy.State{}, err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return positionpolicy.State{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO position_policy_events
		(position_id,adoption_generation,action,before_json,after_json,actor,reason,created_at)
		VALUES (?,?,?,?,?,?,?,?)`, after.PositionID, after.AdoptionGeneration, req.Action,
		string(beforeJSON), string(afterJSON), req.Actor, req.Reason, after.UpdatedAt); err != nil {
		return positionpolicy.State{}, fmt.Errorf("journal: auditing position policy command: %w", err)
	}
	if err := j.runExitWriteHook("position_policy_after_audit"); err != nil {
		return positionpolicy.State{}, err
	}
	if err := tx.Commit(); err != nil {
		return positionpolicy.State{}, fmt.Errorf("journal: committing position policy command: %w", err)
	}
	return after, nil
}

func previewPositionPolicy(current positionpolicy.State, req positionpolicy.Request) (positionpolicy.Preview, error) {
	if current.AdoptionGeneration != req.ExpectedGeneration || current.Version != req.ExpectedVersion {
		return positionpolicy.Preview{}, positionpolicy.ErrVersionMismatch
	}
	if current.PositionState == PositionClosing || current.HasPendingExit {
		if req.Action == positionpolicy.ActionRelease || req.Action == positionpolicy.ActionReadopt {
			return positionpolicy.Preview{}, positionpolicy.ErrExitConflict
		}
	}
	switch current.PositionState {
	case PositionOpening, PositionOpen, PositionScaling, PositionClosing:
	default:
		return positionpolicy.Preview{}, fmt.Errorf("%w: position state %q", positionpolicy.ErrUnknownState,
			current.PositionState)
	}
	after := current
	after.Version++
	after.UpdatedAt = req.At.UTC().Format(time.RFC3339Nano)
	if current.Version == 0 {
		// The virtual generation becomes durable at its first command. Give that
		// lifecycle an explicit observation boundary instead of persisting an
		// ambiguous empty t0; re-adopt will replace it with a new boundary below.
		after.ObservedAt = after.UpdatedAt
	}
	switch req.Action {
	case positionpolicy.ActionOverride:
		if current.Status != positionpolicy.StatusManaged {
			return positionpolicy.Preview{}, positionpolicy.ErrUnknownState
		}
		after.DesiredPolicyID = strings.TrimSpace(req.PolicyID)
	case positionpolicy.ActionInherit:
		if current.Status != positionpolicy.StatusManaged {
			return positionpolicy.Preview{}, positionpolicy.ErrUnknownState
		}
		after.DesiredPolicyID = ""
	case positionpolicy.ActionRelease:
		if !current.ExternalLifecycleEligible() {
			return positionpolicy.Preview{}, positionpolicy.ErrIneligible
		}
		if current.Status != positionpolicy.StatusManaged {
			return positionpolicy.Preview{}, positionpolicy.ErrUnknownState
		}
		after.Status = positionpolicy.StatusReleased
		after.EffectivePolicyID = ""
	case positionpolicy.ActionReadopt:
		if !current.ExternalLifecycleEligible() {
			return positionpolicy.Preview{}, positionpolicy.ErrIneligible
		}
		if current.Status != positionpolicy.StatusReleased {
			return positionpolicy.Preview{}, positionpolicy.ErrUnknownState
		}
		after.AdoptionGeneration++
		after.Version = 1
		after.Status = positionpolicy.StatusManaged
		after.ObservedAt = after.UpdatedAt
		after.EffectivePolicyID = strings.TrimSpace(req.ReAdoption.PolicyID)
		if after.EffectivePolicyID == "" {
			after.EffectivePolicyID = exitpolicy.RatchetPolicyID
		}
	default:
		return positionpolicy.Preview{}, positionpolicy.ErrInvalidRequest
	}
	return positionpolicy.Preview{
		Before: current, After: after, Action: req.Action, Reason: req.Reason,
		ProtectionGap: req.Action == positionpolicy.ActionRelease,
	}, nil
}

func positionPolicyTx(ctx context.Context, tx *sql.Tx, positionID string) (positionpolicy.State, error) {
	id := strings.TrimSpace(positionID)
	state, err := scanPositionPolicyScope(tx.QueryRowContext(ctx, positionPolicySelect+` WHERE p.id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return positionpolicy.State{}, positionpolicy.ErrPositionNotFound
	}
	if err != nil {
		return positionpolicy.State{}, fmt.Errorf("journal: reading position policy scope: %w", err)
	}
	return state, nil
}

const positionPolicySelect = `SELECT p.id,p.account_ref,p.market,p.symbol,p.state,
	coalesce(e.policy_id,''),CASE WHEN e.position_id IS NULL THEN 0 ELSE 1 END,
	coalesce(e.pending_action,''),p.entry_decision_id,p.adoption_id,
	l.adoption_generation,l.version,l.status,l.desired_policy_id,l.observed_at,l.updated_at
	FROM positions p
	LEFT JOIN exit_states e ON e.position_id=p.id
	LEFT JOIN position_policy_lifecycles l ON l.position_id=p.id
	 AND l.adoption_generation=(SELECT max(latest.adoption_generation)
		FROM position_policy_lifecycles latest WHERE latest.position_id=p.id)`

func scanPositionPolicyScope(row rowScanner) (positionpolicy.State, error) {
	var state positionpolicy.State
	var exitPolicy, pending, entryDecision, adoptionID, status, desired, observedAt, updatedAt sql.NullString
	var lifecycleGeneration, version sql.NullInt64
	var hasExit int
	if err := row.Scan(&state.PositionID, &state.AccountRef, &state.Market, &state.Symbol,
		&state.PositionState, &exitPolicy, &hasExit, &pending, &entryDecision, &adoptionID,
		&lifecycleGeneration, &version, &status, &desired, &observedAt, &updatedAt); err != nil {
		return positionpolicy.State{}, err
	}
	state.EffectivePolicyID = strings.TrimSpace(exitPolicy.String)
	state.HasPendingExit = strings.TrimSpace(pending.String) != ""
	state.Provenance, state.Eligibility = positionpolicy.ClassifyProvenance(
		entryDecision.String, adoptionID.String)
	state.ExitEligible = position.ExitEligible(entryDecision.String, adoptionID.String)
	state.AdoptionGeneration = 1
	state.Status = positionpolicy.StatusReleased
	if hasExit == 1 {
		state.Status = positionpolicy.StatusManaged
	}
	if lifecycleGeneration.Valid {
		state.AdoptionGeneration = lifecycleGeneration.Int64
		state.Version = version.Int64
		state.Status = positionpolicy.Status(status.String)
		state.DesiredPolicyID = strings.TrimSpace(desired.String)
		state.ObservedAt, state.UpdatedAt = observedAt.String, updatedAt.String
	}
	if state.Status == positionpolicy.StatusReleased {
		state.EffectivePolicyID = ""
	}
	return state, nil
}

func (j *Journal) PositionPolicyAudit(ctx context.Context, positionID string) ([]positionpolicy.AuditEvent, error) {
	rows, err := j.db.QueryContext(ctx, `SELECT id,position_id,adoption_generation,action,before_json,
		after_json,actor,reason,created_at FROM position_policy_events WHERE position_id=? ORDER BY id`,
		strings.TrimSpace(positionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []positionpolicy.AuditEvent
	for rows.Next() {
		var event positionpolicy.AuditEvent
		if err := rows.Scan(&event.ID, &event.PositionID, &event.AdoptionGeneration, &event.Action,
			&event.BeforeJSON, &event.AfterJSON, &event.Actor, &event.Reason, &event.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}
