package journal

import (
	"context"
	"testing"
)

func TestA166ExactExitOrigin(t *testing.T) {
	for _, tc := range []struct {
		name    string
		actions []string
		account string
		want    string
		ok      bool
	}{
		{"protective", []string{"BASELINE_BREACH_EXIT"}, "acct-1", "BASELINE_BREACH_EXIT", true},
		{"profit-lifecycle", []string{"PROPOSAL_FILLED", "LADDER_TAKE_PROFIT"}, "acct-1", "LADDER_TAKE_PROFIT", true},
		{"duplicate", []string{"BASELINE_BREACH_EXIT", "BASELINE_BREACH_EXIT"}, "acct-1", "", false},
		{"unknown", []string{"UNKNOWN", "BASELINE_BREACH_EXIT"}, "acct-1", "", false},
		{"state-only", []string{"LADDER_HOLD_STOP_PROMOTED"}, "acct-1", "", false},
		{"foreign", []string{"BASELINE_BREACH_EXIT"}, "other", "", false},
		{"missing", nil, "acct-1", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			insertPosition(t, j, "position", nil)
			for _, action := range tc.actions {
				a087ExitEvent(t, j, "position", nil, action, "exact")
			}
			a087ExitEvent(t, j, "position", nil, "BASELINE_BREACH_EXIT", "other-intent")
			action, ok, err := j.ExitProposalForIntent(context.Background(), Intent{ID: "exact", AccountRef: tc.account, Market: "kr", Symbol: "005930"})
			if err != nil || action != tc.want || ok != tc.ok {
				t.Fatalf("action=%q ok=%v err=%v", action, ok, err)
			}
		})
	}
}

func TestA166OriginRejectsForeignRowsAndBound(t *testing.T) {
	j := openTestJournal(t)
	insertPosition(t, j, "p1", nil)
	if _, err := j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price) VALUES ('p2','other','kr','005930',1,'OPEN','1','1')`); err != nil {
		t.Fatal(err)
	}
	a087ExitEvent(t, j, "p1", nil, "STOP_LOSS_LADDER", "mixed")
	a087ExitEvent(t, j, "p2", nil, ExitEventProposalFilled, "mixed")
	intent := Intent{ID: "mixed", AccountRef: "acct-1", Market: "kr", Symbol: "005930"}
	if _, ok, err := j.ExitProposalForIntent(context.Background(), intent); err != nil || ok {
		t.Fatalf("foreign lifecycle accepted: %v %v", ok, err)
	}
	for i := 0; i < 65; i++ {
		a087ExitEvent(t, j, "p1", nil, ExitEventProposalFilled, "bounded")
	}
	a087ExitEvent(t, j, "p1", nil, "STOP_LOSS_LADDER", "bounded")
	intent.ID = "bounded"
	if _, ok, err := j.ExitProposalForIntent(context.Background(), intent); err != nil || ok {
		t.Fatalf("over limit accepted: %v %v", ok, err)
	}
}

func TestA166OriginDBErrorReturned(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.db.Exec(`DROP TABLE exit_events`); err != nil {
		t.Fatal(err)
	}
	_, ok, err := j.ExitProposalForIntent(context.Background(), Intent{ID: "x", AccountRef: "acct", Market: "kr", Symbol: "X"})
	if err == nil || ok {
		t.Fatalf("DB failure hidden: ok=%v err=%v", ok, err)
	}
}
