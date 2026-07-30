package journal

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestLadderExitStateSnapshotsItsPolicyID(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-policy", attemptID: "a-policy", orderID: "o-policy", decisionID: "d-policy"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	state, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50,
		EntryPrice: "70000", InitialStop: "68000",
	})
	if err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	if state.PolicyID != exitpolicy.CommonLadderHybrid50 {
		t.Fatalf("policy id = %q", state.PolicyID)
	}
}

func TestLegacyLadderNullPolicyReadsAsDefaultV1(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-legacy", attemptID: "a-legacy", orderID: "o-legacy", decisionID: "d-legacy"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE exit_states SET policy_id=NULL WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	state, err := j.ExitState(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.PolicyID != "default_v1" {
		t.Fatalf("legacy policy id = %q, want default_v1", state.PolicyID)
	}
}

func TestAdoptionPolicySnapshotSurvivesUntilExitStateRecovery(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-policy")
	req := sampleAdoption("p-policy")
	req.ExitPolicyID = exitpolicy.CommonLadderRunner
	adoption, err := j.AdoptPosition(ctx, req)
	if err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	if adoption.ExitPolicyID != exitpolicy.CommonLadderRunner {
		t.Fatalf("adoption policy id = %q", adoption.ExitPolicyID)
	}
	state, err := j.OpenAdoptedExitState(ctx, "p-policy")
	if err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	if state.PolicyKind != ExitPolicyLadder || state.PolicyID != exitpolicy.CommonLadderRunner {
		t.Fatalf("recovered exit state = kind:%s id:%s", state.PolicyKind, state.PolicyID)
	}
}

func TestSchemaV9AddsOnlyNullablePolicySnapshotColumns(t *testing.T) {
	j := openTestJournal(t)
	for table, column := range map[string]string{
		"exit_states": "policy_id", "position_adoptions": "exit_policy_id",
	} {
		rows, err := j.db.Query(`PRAGMA table_info(` + table + `)`)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for rows.Next() {
			var cid, notNull, primaryKey int
			var name, typ string
			var defaultValue any
			if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
				t.Fatal(err)
			}
			if name == column {
				found = true
				if notNull != 0 || defaultValue != nil {
					t.Fatalf("%s.%s must be nullable with no default", table, column)
				}
			}
		}
		rows.Close()
		if !found {
			t.Fatalf("%s.%s missing", table, column)
		}
	}
}
