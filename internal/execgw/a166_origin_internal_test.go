package execgw

import (
	"context"
	"database/sql"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestA166ReplayOriginClassification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j, err := journal.Open(context.Background(), journal.Options{Path: path, FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price) VALUES ('position','acct','kr','005930',1,'OPEN','1','1')`); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, tc := range []struct {
		name, source, side, class, action string
		want                              accountbudget.Priority
		ok                                bool
	}{
		{"entry", "engine", "BUY", journal.SafetyClassExposureRaising, "", accountbudget.Normal, true},
		{"entry-contradiction", "engine", "BUY", journal.SafetyClassExposureRaising, "STOP_LOSS_LADDER", accountbudget.Normal, false},
		{"flatten", "flatten-all", "SELL", journal.SafetyClassRiskReducing, "", accountbudget.Critical, true},
		{"stop", "engine", "SELL", journal.SafetyClassRiskReducing, "STOP_LOSS_LADDER", accountbudget.Critical, true},
		{"baseline", "engine", "SELL", journal.SafetyClassRiskReducing, "BASELINE_BREACH_EXIT", accountbudget.Critical, true},
		{"ratchet", "engine", "SELL", journal.SafetyClassRiskReducing, "RATCHET_PARTIAL", accountbudget.Normal, true},
		{"ladder", "engine", "SELL", journal.SafetyClassRiskReducing, "LADDER_PARTIAL", accountbudget.Normal, true},
		{"profit", "engine", "SELL", journal.SafetyClassRiskReducing, "LADDER_TAKE_PROFIT", accountbudget.Normal, true},
		{"unknown", "engine", "SELL", journal.SafetyClassRiskReducing, "UNKNOWN", accountbudget.Normal, false},
		{"other-source", "manual", "SELL", journal.SafetyClassRiskReducing, "STOP_LOSS_LADDER", accountbudget.Normal, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var pre journal.Preimage = journal.ReductionIntent{AccountRef: "acct", Market: "kr", Symbol: "005930", Side: tc.side, MaxQuantity: "1", Reason: "fixture"}
			limits := ""
			if tc.class == journal.SafetyClassExposureRaising {
				pre = journal.RiskIntent{AccountRef: "acct", Market: "kr", Symbol: "005930", Side: tc.side, Quantity: "1", EntryPrice: "2", StopPrice: "1", PolicyVersion: "fixture"}
				limits = "{}"
			}
			dec, e := j.RecordDecision(context.Background(), journal.DecisionRequest{ID: tc.name, AccountRef: "acct", SafetyClass: tc.class, Kind: journal.KindPlace, Preimage: pre, LimitsJSON: limits, Nonce: tc.name, IssuedAt: now, ExpiresAt: now.Add(time.Hour)})
			if e != nil {
				t.Fatal(e)
			}
			intent := journal.Intent{ID: tc.name, AccountRef: "acct", Market: "kr", Symbol: "005930", Side: tc.side, Source: tc.source}
			if tc.action != "" {
				if _, e = db.Exec(`INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at) VALUES ('position',?,?,?)`, tc.action, tc.name, now.Format(time.RFC3339)); e != nil {
					t.Fatal(e)
				}
			}
			rec := journal.AttemptRecord{IntentID: tc.name, AccountRef: "acct", DecisionID: tc.name, SafetyClass: tc.class, ClientOrderID: dec.ClientOrderID}
			g := Gateway{journal: j, accountRef: "acct"}
			priority, ok, e := g.replayOrigin(context.Background(), rec, intent)
			if e != nil || ok != tc.ok || (ok && priority != tc.want) {
				t.Fatalf("priority=%v ok=%v err=%v", priority, ok, e)
			}
			rec.AccountRef = "other"
			if _, ok, e = g.replayOrigin(context.Background(), rec, intent); e != nil || ok {
				t.Fatalf("foreign binding accepted: %v %v", ok, e)
			}
			if tc.ok {
				rec.AccountRef = "acct"
				g.accountRef = "other"
				if _, ok, e = g.replayOrigin(context.Background(), rec, intent); e != nil || ok {
					t.Fatalf("foreign gateway accepted: %v %v", ok, e)
				}
			}
		})
	}
}

func TestA166OrderableActionCoverage(t *testing.T) {
	parse := func(path, fn string) map[string]bool {
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]bool{}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name.Name != fn {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				switch v := n.(type) {
				case *ast.Ident:
					if strings.HasPrefix(v.Name, "Action") && v.Name != "Action" {
						out[v.Name] = true
					}
				case *ast.SelectorExpr:
					if strings.HasPrefix(v.Sel.Name, "Action") && v.Sel.Name != "Action" {
						out[v.Sel.Name] = true
					}
				}
				return true
			})
		}
		return out
	}
	expected := parse("../exitpolicy/ratchet.go", "Orderable")
	got := parse("a166_replay_origin.go", "replayOrigin")
	if len(expected) == 0 {
		t.Fatal("orderable action source not found")
	}
	for action := range expected {
		if !got[action] {
			t.Errorf("unclassified orderable action %s", action)
		}
	}
}

func TestA166ReplayOriginDBErrorPropagates(t *testing.T) {
	j, err := journal.Open(context.Background(), journal.Options{Path: filepath.Join(t.TempDir(), "journal.db"), FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	if err = j.Close(); err != nil {
		t.Fatal(err)
	}
	g := Gateway{journal: j, accountRef: "acct"}
	rec := journal.AttemptRecord{IntentID: "id", AccountRef: "acct", DecisionID: "decision", ClientOrderID: "key"}
	_, ok, err := g.replayOrigin(context.Background(), rec, journal.Intent{ID: "id", AccountRef: "acct", Market: "kr", Symbol: "005930", Side: "SELL", Source: "engine"})
	if err == nil || ok {
		t.Fatalf("closed DB hidden: ok=%v err=%v", ok, err)
	}
}

type a166AckReader struct {
	priority accountbudget.Priority
	deadline time.Time
}

func (s *a166AckReader) OrderRaw(ctx context.Context, _ string) (json.RawMessage, error) {
	s.priority = accountbudget.PriorityFromContext(ctx)
	s.deadline, _ = ctx.Deadline()
	return json.RawMessage(`{"orderId":"fixture","symbol":"005930"}`), nil
}
func TestA166AckReadInheritsPriorityAndDeadline(t *testing.T) {
	reader := &a166AckReader{}
	g := Gateway{orders: reader}
	check := g.roundTripFor(mutationPlan{kind: journal.KindPlace, symbol: "005930"})
	ctx, cancel := context.WithTimeout(accountbudget.WithPriority(context.Background(), accountbudget.Critical), time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	if err := check(ctx, "fixture"); err != nil {
		t.Fatal(err)
	}
	if reader.priority != accountbudget.Critical || reader.deadline.After(deadline) || reader.deadline.IsZero() {
		t.Fatalf("priority=%v deadline=%v", reader.priority, reader.deadline)
	}
}
