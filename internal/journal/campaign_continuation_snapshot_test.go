package journal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func a160Fixture(t *testing.T) (*Journal, *ReadOnly, ContinuationSnapshotRequest) {
	t.Helper()
	j := openTestJournal(t)
	ctx := context.Background()
	c := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{CampaignID: c.ID, ExpectedVersion: c.Version, CommandKey: "plan-a160", Sequence: 1, PlanID: "leg-a160", RequestedQuantity: "10"})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, c.DecisionID, "a160-intent", "a160-attempt", "a160-order", "", "10")
	_, err = j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{CampaignID: c.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion, CommandKey: "link-a160", OrderID: "a160-order", RequestedCap: "10", IntentID: "a160-intent", AttemptID: "a160-attempt"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES('a160-pos','acct','kr','005930',1,'decision','OPEN','3','70000')`)
	if err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "a160-order", AccountRef: "acct", Market: "kr", Symbol: "005930", Delta: "3", CumulativeQuantity: "3", CommittedAt: "2026-03-30T00:31:00Z"})
	return j, openTestReadOnly(t, j.path), ContinuationSnapshotRequest{AccountRef: "acct", Market: "KR", Symbol: "005930", CampaignID: "campaign", ExpectedGeneration: 1}
}
func a160Rows(t *testing.T, j *Journal) []string {
	t.Helper()
	var result, tables []string
	catalog, err := j.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	for catalog.Next() {
		var name string
		if err := catalog.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err := catalog.Err(); err != nil {
		t.Fatal(err)
	}
	catalog.Close()
	for _, table := range tables {
		rows, err := j.db.Query(`SELECT * FROM "` + table + `"`)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			ptr := make([]any, len(columns))
			for i := range values {
				ptr[i] = &values[i]
			}
			if err := rows.Scan(ptr...); err != nil {
				t.Fatal(err)
			}
			result = append(result, fmt.Sprint(table, values))
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
	}
	sort.Strings(result)
	return result
}
func TestA160SnapshotAndReadOnlyPersistence(t *testing.T) {
	j, ro, req := a160Fixture(t)
	ctx := context.Background()
	before, err := ro.ReadCampaignContinuationSnapshot(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if before.Position.Quantity != "3" || before.Campaign.ActualPositionGeneration != 1 || before.ProjectionVersion < 1 || len(before.Legs) != 1 || len(before.Watermarks) != 1 || before.Watermarks[0].CumulativeFilled != "3" || before.PlanAvailability != FullPlanUnavailable {
		t.Fatalf("snapshot %+v", before)
	}
	rowsBefore := a160Rows(t, j)
	after, err := ro.ReadCampaignContinuationSnapshot(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if err := ro.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := openTestReadOnly(t, j.path)
	again, err := reopened.ReadCampaignContinuationSnapshot(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(before, again) || !reflect.DeepEqual(rowsBefore, a160Rows(t, j)) {
		t.Fatal("read changed stored state")
	}
}
func TestA160UnavailableNeverReturnsPartial(t *testing.T) {
	j, ro, req := a160Fixture(t)
	ctx := context.Background()
	for _, mut := range []func(*ContinuationSnapshotRequest){func(r *ContinuationSnapshotRequest) { r.AccountRef = "peer" }, func(r *ContinuationSnapshotRequest) { r.Symbol = "000660" }, func(r *ContinuationSnapshotRequest) { r.CampaignID = "other" }, func(r *ContinuationSnapshotRequest) { r.ExpectedGeneration = 2 }} {
		x := req
		mut(&x)
		got, err := ro.ReadCampaignContinuationSnapshot(ctx, x)
		if !errors.Is(err, ErrContinuationUnavailable) || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	}
	if _, err := j.db.Exec(`UPDATE positions SET entry_decision_id=NULL WHERE id='a160-pos'`); err != nil {
		t.Fatal(err)
	}
	got, err := ro.ReadCampaignContinuationSnapshot(ctx, req)
	if !errors.Is(err, ErrContinuationUnavailable) || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
		t.Fatalf("decision missing got=%+v err=%v", got, err)
	}
}
func TestA160ReadTransactionDoesNotMixConcurrentWrite(t *testing.T) {
	j, ro, req := a160Fixture(t)
	ctx := context.Background()
	req.Market = "kr"
	tx, err := ro.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := positionCampaignQuery(ctx, tx, req.CampaignID); err != nil {
		t.Fatal(err)
	}
	writer, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = writer.Exec(`UPDATE positions SET quantity='4' WHERE id='a160-pos'`); err != nil {
		t.Fatal(err)
	}
	handle := &ApplyTx{tx: writer, now: "2026-03-30T00:32:00Z"}
	err = ApplyPositionCampaignFill(ctx, handle, AppliedFill{OrderID: "a160-order", AccountRef: "acct", Market: "kr", Symbol: "005930", TradingDay: "2026-03-30", Side: "BUY", Delta: "1", CumulativeQuantity: "4", CommittedAt: "2026-03-30T00:32:00Z"})
	handle.invalidate()
	if err != nil {
		t.Fatal(err)
	}
	if err = writer.Commit(); err != nil {
		t.Fatal(err)
	}
	old, err := readCampaignContinuationSnapshot(ctx, tx, req)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	now, err := ro.ReadCampaignContinuationSnapshot(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if old.Position.Quantity != "3" || old.Watermarks[0].CumulativeFilled != "3" || now.Position.Quantity != "4" || now.Watermarks[0].CumulativeFilled != "4" || now.Campaign.Version <= old.Campaign.Version || now.ProjectionVersion <= old.ProjectionVersion {
		t.Fatalf("mixed old=%+v new=%+v", old, now)
	}
}

func TestA160StoredScopeAndFailures(t *testing.T) {
	cases := []struct {
		name, sql string
		available bool
	}{
		{"blocked", `UPDATE position_campaigns SET entry_blocked=1,state='RECONCILE' WHERE id='campaign'`, true},
		{"closed campaign", `UPDATE position_campaigns SET state='CLOSED' WHERE id='campaign'`, false},
		{"prospective", `UPDATE position_campaigns SET actual_position_generation=NULL WHERE id='campaign'`, false},
		{"prospective contaminated", `UPDATE position_campaigns SET state='PLANNED' WHERE id='campaign'`, false},
		{"missing projection", `DELETE FROM position_projection_versions WHERE position_id='a160-pos'`, false},
		{"missing leg", `DELETE FROM campaign_order_watermarks; DELETE FROM campaign_legs`, false},
		{"closed position", `UPDATE positions SET state='CLOSED' WHERE id='a160-pos'`, false},
		{"latest generation changed", `INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES('new-pos','acct','kr','005930',2,'decision','OPEN','1','70000')`, false},
		{"different decision", `INSERT INTO decisions SELECT 'other',account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,client_order_id,limits_json,nonce,issued_at,expires_at FROM decisions WHERE id='decision'`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j, ro, req := a160Fixture(t)
			if tc.name == "different decision" {
				insertCampaignDecision(t, j, "other", "acct")
				tc.sql = `UPDATE positions SET entry_decision_id='other' WHERE id='a160-pos'`
			}
			if _, err := j.db.Exec(tc.sql); err != nil {
				t.Fatal(err)
			}
			got, err := ro.ReadCampaignContinuationSnapshot(context.Background(), req)
			if tc.available {
				if err != nil || !got.Campaign.EntryBlocked || string(got.Campaign.State) != "RECONCILE" {
					t.Fatalf("got=%+v err=%v", got, err)
				}
			} else if err == nil || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
}

func TestA160InvalidAndCancelledContext(t *testing.T) {
	_, ro, req := a160Fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, x := range []struct {
		ctx context.Context
		req ContinuationSnapshotRequest
	}{{nil, req}, {ctx, req}, {context.Background(), ContinuationSnapshotRequest{}}} {
		got, err := ro.ReadCampaignContinuationSnapshot(x.ctx, x.req)
		if err == nil || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	}
}

func TestA160USAndPlannedLegWithoutWatermark(t *testing.T) {
	j, ro, req := a160Fixture(t)
	c, err := j.PositionCampaign(context.Background(), req.CampaignID)
	if err != nil {
		t.Fatal(err)
	}
	planned, err := j.PlanCampaignLeg(context.Background(), PlanCampaignLegRequest{CampaignID: c.ID, ExpectedVersion: c.Version, CommandKey: "plan-second", Sequence: 2, PlanID: "second", RequestedQuantity: "4"})
	if err != nil {
		t.Fatal(err)
	}
	if string(planned.State) != "PLANNED" {
		t.Fatalf("leg=%+v", planned)
	}
	for _, query := range []string{
		`UPDATE position_campaigns SET market='us',symbol='AAPL' WHERE id='campaign'`,
		`UPDATE positions SET market='us',symbol='AAPL' WHERE id='a160-pos'`,
		`UPDATE campaign_order_watermarks SET market='us',symbol='AAPL' WHERE campaign_id='campaign'`,
	} {
		if _, err := j.db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	req.Market, req.Symbol = "US", "aapl"
	got, err := ro.ReadCampaignContinuationSnapshot(context.Background(), req)
	if err != nil || got.Position.Market != "us" || len(got.Legs) != 2 || got.Legs[1].Sequence != 2 || string(got.Legs[1].State) != "PLANNED" || len(got.Watermarks) != 1 || got.Watermarks[0].LegSequence != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestA160OrphanWatermarkFailsClosed(t *testing.T) {
	j, ro, req := a160Fixture(t)
	conn, err := j.db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err = conn.ExecContext(context.Background(), `PRAGMA foreign_keys=OFF`); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.ExecContext(context.Background(), `UPDATE campaign_order_watermarks SET leg_sequence=2 WHERE campaign_id='campaign'`); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.ExecContext(context.Background(), `PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	got, err := ro.ReadCampaignContinuationSnapshot(context.Background(), req)
	if !errors.Is(err, ErrContinuationUnavailable) || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestA160ForeignWatermarkScopeFailsClosed(t *testing.T) {
	for _, change := range []string{`account_ref='foreign'`, `market='us'`, `symbol='000660'`} {
		t.Run(change, func(t *testing.T) {
			j, ro, req := a160Fixture(t)
			if _, err := j.db.Exec(`UPDATE campaign_order_watermarks SET ` + change + ` WHERE campaign_id='campaign'`); err != nil {
				t.Fatal(err)
			}
			got, err := ro.ReadCampaignContinuationSnapshot(context.Background(), req)
			if !errors.Is(err, ErrContinuationUnavailable) || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
}
func TestA160ClosedAndNilHandleFails(t *testing.T) {
	_, ro, req := a160Fixture(t)
	if err := ro.Close(); err != nil {
		t.Fatal(err)
	}
	var nilHandle *ReadOnly
	for _, h := range []*ReadOnly{ro, nilHandle} {
		got, err := h.ReadCampaignContinuationSnapshot(context.Background(), req)
		if err == nil || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	}
}

func TestA160PeerScopeAndMissingTable(t *testing.T) {
	j, ro, req := a160Fixture(t)
	statements := []string{
		`INSERT INTO position_campaigns(id,account_ref,market,symbol,lane_id,lane_version,decision_id,evidence_digest,expected_position_generation,expected_position_version,prospective_token,actual_position_generation,state,version,created_at,updated_at) SELECT 'peer',account_ref,market,'000660',lane_id,lane_version,decision_id,evidence_digest,expected_position_generation,expected_position_version,'peer-token',actual_position_generation,state,version,created_at,updated_at FROM position_campaigns WHERE id='campaign'`,
		`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES('peer-pos','acct','kr','000660',1,'decision','OPEN','8','70000')`,
		`INSERT INTO campaign_legs(campaign_id,sequence,plan_id,requested_quantity,filled_quantity,residual_quantity,state,version,created_at,updated_at) SELECT 'peer',sequence,plan_id,requested_quantity,filled_quantity,residual_quantity,state,version,created_at,updated_at FROM campaign_legs WHERE campaign_id='campaign'`,
	}
	for _, q := range statements {
		if _, err := j.db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	peer := req
	peer.CampaignID = "peer"
	peer.Symbol = "000660"
	one, err := ro.ReadCampaignContinuationSnapshot(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	two, err := ro.ReadCampaignContinuationSnapshot(context.Background(), peer)
	if err != nil {
		t.Fatal(err)
	}
	if one.Position.ID != "a160-pos" || one.Position.Quantity != "3" || two.Position.ID != "peer-pos" || two.Position.Quantity != "8" || len(one.Watermarks) != 1 || len(two.Watermarks) != 0 {
		t.Fatalf("one=%+v two=%+v", one, two)
	}
	if _, err := j.db.Exec(`DROP TABLE campaign_order_watermarks`); err != nil {
		t.Fatal(err)
	}
	got, err := ro.ReadCampaignContinuationSnapshot(context.Background(), req)
	if err == nil || !reflect.DeepEqual(got, ContinuationSnapshot{}) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
