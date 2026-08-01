package journal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

func strategyDecisionRecordFixture(t *testing.T, suffix string) (strategyengine.DecisionRecord, string, string) {
	t.Helper()
	record := strategyengine.DecisionRecord{
		CandidateLifeID: "candidate-life:v1:sha256:" + strings.Repeat("a", 64),
		Market:          "KR", Symbol: "005930", LaneID: "krx_parker_vwap_conservative_v1", LaneVersion: "1",
		SourceDigest: strings.Repeat("d", 64), ConstantsDigest: "sha256:" + strings.Repeat("e", 64),
		ThresholdVersion: "threshold-v1", ThresholdSetDigest: "sha256:" + strings.Repeat("b", 64),
		EvidenceDigest: "sha256:" + strings.Repeat("c", 64), CalendarVersion: "calendar-" + suffix,
		EntryPrice: "100.1", StopPrice: "99.3993", TargetPrice: "102.2021",
	}
	identityInput, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	identityHash := sha256.Sum256(identityInput)
	record.Identity = "strategy-decision:v1:sha256:" + hex.EncodeToString(identityHash[:])
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	payloadHash := sha256.Sum256(payload)
	return record, string(payload), "sha256:" + hex.EncodeToString(payloadHash[:])
}

func strategyPlanFixture(t *testing.T, suffix, account string) StrategyAtomicPlan {
	t.Helper()
	issued := testIssued(t)
	request := DecisionRequest{
		ID: "risk-" + suffix, AccountRef: account, SafetyClass: SafetyClassExposureRaising, Kind: KindPlace,
		Preimage:   RiskIntent{AccountRef: account, Market: "KR", Symbol: "005930", Side: "BUY", Quantity: "1", EntryPrice: "100.1", StopPrice: "99.3993", TargetPrice: "102.2021", PolicyVersion: "strategy/v1"},
		LimitsJSON: `{"max_quantity":"1"}`, Nonce: "nonce-" + suffix, IssuedAt: issued, ExpiresAt: issued.Add(time.Minute),
	}
	record, payload, payloadDigest := strategyDecisionRecordFixture(t, suffix)
	return StrategyAtomicPlan{
		RiskDecision: request,
		Lineage: StrategyDecisionLineage{
			DecisionIdentity: record.Identity, CandidateLifeID: record.CandidateLifeID,
			Market: "KR", Symbol: "005930", ThresholdVersion: "threshold-v1",
			ThresholdSetDigest: "sha256:" + strings.Repeat("b", 64), EvidenceDigest: "sha256:" + strings.Repeat("c", 64),
			LaneID: "krx_parker_vwap_conservative_v1", LaneVersion: "1",
			LaneSourceDigest: strings.Repeat("d", 64), LaneConstantsDigest: "sha256:" + strings.Repeat("e", 64),
			EntryPrice: "100.1", StopPrice: "99.3993", TargetPrice: "102.2021", Quantity: "1", PolicyVersion: "strategy/v1", SettingsDigest: "sha256:" + strings.Repeat("9", 64),
			DecisionPayload: payload, DecisionPayloadDigest: payloadDigest, ActivationManifestDigest: "sha256:" + strings.Repeat("f", 64),
			CreatedAt: issued,
		},
		AttemptID: "attempt-" + suffix, GuardianDecisionID: "guardian-" + suffix,
		ActivationManifestDigest: "sha256:" + strings.Repeat("f", 64),
		Revision:                 1, CreatedAt: issued, ClientOrderID: DeriveClientOrderID(request.ID, 0),
	}
}

func TestStrategyDecisionPayloadSchemaMatchesDecisionRecordExactly(t *testing.T) {
	want := reflect.TypeFor[strategyengine.DecisionRecord]()
	got := reflect.TypeFor[strategyDecisionPayload]()
	if got.NumField() != want.NumField() {
		t.Fatalf("payload fields=%d DecisionRecord fields=%d", got.NumField(), want.NumField())
	}
	for index := range want.NumField() {
		if got.Field(index).Name != want.Field(index).Name || got.Field(index).Type != want.Field(index).Type || got.Field(index).Tag != want.Field(index).Tag {
			t.Fatalf("field %d payload=%s/%s/%q DecisionRecord=%s/%s/%q", index, got.Field(index).Name, got.Field(index).Type, got.Field(index).Tag, want.Field(index).Name, want.Field(index).Type, want.Field(index).Tag)
		}
	}
}

func TestStrategyPlanIsAtomicExactAndIdempotent(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	ctx := context.Background()
	plan := strategyPlanFixture(t, "1", "acct-1")
	first, err := j.planStrategyEntryForTest(ctx, plan)
	if err != nil || first.Idempotent {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	replay, err := j.planStrategyEntryForTest(ctx, plan)
	if err != nil || !replay.Idempotent || replay != (StrategyPlanReceipt{
		AttemptID: plan.AttemptID, AccountRef: plan.RiskDecision.AccountRef, DecisionIdentity: plan.Lineage.DecisionIdentity,
		RiskIntentID: plan.RiskDecision.ID, ClientOrderID: plan.ClientOrderID, Quantity: plan.Lineage.Quantity,
		Revision: 1, State: "PLANNED", Idempotent: true,
	}) {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	for table, want := range map[string]int{"decisions": 1, "strategy_decision_lineage": 1, "strategy_attempt_lineage": 1, "strategy_execution_lineage": 1} {
		var got int
		if err := j.db.QueryRow("SELECT count(*) FROM " + table).Scan(&got); err != nil || got != want {
			t.Fatalf("%s count=%d err=%v", table, got, err)
		}
	}
}

func TestStrategyPlanRejectsEveryDivergentReplay(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*StrategyAtomicPlan)
		stage  string
	}{
		{"risk", func(p *StrategyAtomicPlan) { p.RiskDecision.LimitsJSON = `{"max_quantity":"2"}` }, "RiskIntent"},
		{"lineage", func(p *StrategyAtomicPlan) { p.Lineage.DecisionPayload = `{"forged":true}` }, "decision lineage"},
		{"attempt", func(p *StrategyAtomicPlan) { p.GuardianDecisionID = "other" }, "attempt"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
			defer j.Close()
			plan := strategyPlanFixture(t, "2", "acct-1")
			if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&plan)
			_, err := j.planStrategyEntryForTest(context.Background(), plan)
			var collision *StrategyCollisionError
			if !errors.As(err, &collision) || collision.Stage != tc.stage {
				t.Fatalf("err=%v want stage=%s", err, tc.stage)
			}
		})
	}
}

func TestStrategyPlanRejectsDivergentExecutionReplay(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	plan := strategyPlanFixture(t, "8", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DROP TRIGGER strategy_execution_lineage_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE strategy_execution_lineage SET recorded_at='forged' WHERE attempt_id=?`, plan.AttemptID); err != nil {
		t.Fatal(err)
	}
	_, err := j.planStrategyEntryForTest(context.Background(), plan)
	var collision *StrategyCollisionError
	if !errors.As(err, &collision) || collision.Stage != "execution" {
		t.Fatalf("err=%v", err)
	}
}

func TestStrategyPlanRollbackLeavesNoPartialRows(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	if _, err := j.db.Exec(`CREATE TRIGGER fail_strategy_execution BEFORE INSERT ON strategy_execution_lineage BEGIN SELECT RAISE(ABORT,'synthetic crash'); END;`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "3", "acct-1")); err == nil {
		t.Fatal("failing execution insert committed")
	}
	for _, table := range []string{"decisions", "strategy_decision_lineage", "strategy_attempt_lineage", "strategy_execution_lineage"} {
		var count int
		if err := j.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s partial count=%d err=%v", table, count, err)
		}
	}
}

func strategyIssueFixture(t *testing.T, j *Journal, suffix, account string) StrategyIssueRequest {
	t.Helper()
	plan := strategyPlanFixture(t, suffix, account)
	issue := issueRequest(j, plan.RiskDecision.ID, account, plan.Lineage.EntryPrice, "0", "1000000", mustVersion(t, j, account))
	issue.Decision = plan.RiskDecision
	issue.Reserve.Reservations[0].ID = "res-" + plan.RiskDecision.ID
	return StrategyIssueRequest{
		Issue: issue,
		Plan: StrategyPlanRequest{
			Lineage: plan.Lineage, AttemptID: plan.AttemptID,
			ActivationManifestDigest: plan.ActivationManifestDigest, Revision: plan.Revision,
		},
	}
}

func TestStrategyProductionIssuanceCommitsAuthorityReservationAndLineageTogether(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	request := strategyIssueFixture(t, j, "f", "acct-1")
	result, err := j.RecordStrategyDecisionAndReserve(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Receipt.AccountRef != "acct-1" || result.Receipt.RiskIntentID != result.Issue.Decision.ID ||
		result.Receipt.ClientOrderID != result.Issue.Decision.ClientOrderID || result.Receipt.Quantity != request.Plan.Lineage.Quantity ||
		result.Receipt.State != "PLANNED" || len(result.Issue.Reservations) != 1 ||
		result.Issue.Reservations[0].DecisionID != result.Receipt.RiskIntentID {
		t.Fatalf("result=%+v", result)
	}
	for table, want := range map[string]int{"decisions": 1, "risk_reservations": 1, "strategy_decision_lineage": 1, "strategy_attempt_lineage": 1, "strategy_execution_lineage": 1} {
		var got int
		if err := j.db.QueryRow("SELECT count(*) FROM " + table).Scan(&got); err != nil || got != want {
			t.Fatalf("%s count=%d err=%v", table, got, err)
		}
	}
}

func rehashStrategyPayload(t *testing.T, request *StrategyIssueRequest) {
	t.Helper()
	hash := sha256.Sum256([]byte(request.Plan.Lineage.DecisionPayload))
	request.Plan.Lineage.DecisionPayloadDigest = "sha256:" + hex.EncodeToString(hash[:])
}

func TestStrategyProductionIssuanceRejectsEveryDivergentLineageProjection(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*StrategyDecisionLineage)
	}{
		{name: "decision identity", mutate: func(v *StrategyDecisionLineage) { v.DecisionIdentity += "-other" }},
		{name: "candidate life", mutate: func(v *StrategyDecisionLineage) { v.CandidateLifeID += "-other" }},
		{name: "market", mutate: func(v *StrategyDecisionLineage) { v.Market = "US" }},
		{name: "symbol", mutate: func(v *StrategyDecisionLineage) { v.Symbol = "000660" }},
		{name: "threshold version", mutate: func(v *StrategyDecisionLineage) { v.ThresholdVersion += "-other" }},
		{name: "threshold digest", mutate: func(v *StrategyDecisionLineage) { v.ThresholdSetDigest += "-other" }},
		{name: "evidence digest", mutate: func(v *StrategyDecisionLineage) { v.EvidenceDigest += "-other" }},
		{name: "lane id", mutate: func(v *StrategyDecisionLineage) { v.LaneID += "-other" }},
		{name: "lane version", mutate: func(v *StrategyDecisionLineage) { v.LaneVersion += "-other" }},
		{name: "source digest", mutate: func(v *StrategyDecisionLineage) { v.LaneSourceDigest += "-other" }},
		{name: "constants digest", mutate: func(v *StrategyDecisionLineage) { v.LaneConstantsDigest += "-other" }},
		{name: "entry price", mutate: func(v *StrategyDecisionLineage) { v.EntryPrice = "100.2" }},
		{name: "stop price", mutate: func(v *StrategyDecisionLineage) { v.StopPrice = "99" }},
		{name: "target price", mutate: func(v *StrategyDecisionLineage) { v.TargetPrice = "103" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j, _ := openReservationJournal(t)
			request := strategyIssueFixture(t, j, "projection", "acct-1")
			tc.mutate(&request.Plan.Lineage)
			if _, err := j.RecordStrategyDecisionAndReserve(context.Background(), request); err == nil || !strings.Contains(err.Error(), "exact binding mismatch") {
				t.Fatalf("divergent projection accepted: err=%v", err)
			}
			var count int
			if err := j.db.QueryRow(`SELECT count(*) FROM decisions`).Scan(&count); err != nil || count != 0 {
				t.Fatalf("partial authority count=%d err=%v", count, err)
			}
		})
	}
}

func TestStrategyProductionIssuanceBindsFullDecisionRecordIdentity(t *testing.T) {
	j, _ := openReservationJournal(t)
	request := strategyIssueFixture(t, j, "full-record", "acct-1")
	var record strategyengine.DecisionRecord
	if err := json.Unmarshal([]byte(request.Plan.Lineage.DecisionPayload), &record); err != nil {
		t.Fatal(err)
	}
	record.CalendarVersion += "-forged"
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	request.Plan.Lineage.DecisionPayload = string(payload)
	rehashStrategyPayload(t, &request)
	if _, err := j.RecordStrategyDecisionAndReserve(context.Background(), request); err == nil || !strings.Contains(err.Error(), "exact binding mismatch") {
		t.Fatalf("full-record mutation accepted: %v", err)
	}
}

func TestStrategyProductionIssuanceRejectsNonCanonicalDecisionPayload(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(string) string
	}{
		{name: "leading whitespace", mutate: func(raw string) string { return " " + raw }},
		{name: "unknown field", mutate: func(raw string) string { return strings.TrimSuffix(raw, "}") + `,"Unknown":true}` }},
		{name: "trailing value", mutate: func(raw string) string { return raw + `{}` }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j, _ := openReservationJournal(t)
			request := strategyIssueFixture(t, j, "canonical", "acct-1")
			request.Plan.Lineage.DecisionPayload = tc.mutate(request.Plan.Lineage.DecisionPayload)
			rehashStrategyPayload(t, &request)
			if _, err := j.RecordStrategyDecisionAndReserve(context.Background(), request); err == nil {
				t.Fatal("noncanonical payload accepted")
			}
		})
	}
}

func TestStrategyAttemptLineageAllowsOnlySanctionedStateCAS(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	plan := strategyPlanFixture(t, "immutable", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ column, value string }{
		{column: "attempt_id", value: "'other-attempt'"},
		{column: "account_ref", value: "'other-account'"},
		{column: "entry_decision_identity", value: "'other-decision'"},
		{column: "risk_intent_id", value: "'other-risk'"},
		{column: "guardian_decision_id", value: "'other-guardian'"},
		{column: "activation_manifest_digest", value: "'other-manifest'"},
		{column: "client_order_id", value: "'other-client'"},
		{column: "created_at", value: "'other-time'"},
	} {
		t.Run(tc.column, func(t *testing.T) {
			query := `UPDATE strategy_attempt_lineage SET ` + tc.column + `=` + tc.value + ` WHERE attempt_id=?`
			if _, err := j.db.Exec(query, plan.AttemptID); err == nil {
				t.Fatalf("immutable column %s changed", tc.column)
			}
		})
	}
	for _, query := range []string{
		`UPDATE strategy_attempt_lineage SET revision=revision+1 WHERE attempt_id=?`,
		`UPDATE strategy_attempt_lineage SET state='DISPATCHED' WHERE attempt_id=?`,
		`UPDATE strategy_attempt_lineage SET state='PLANNED',revision=revision+1 WHERE attempt_id=?`,
	} {
		if _, err := j.db.Exec(query, plan.AttemptID); err == nil {
			t.Fatalf("invalid transition accepted: %s", query)
		}
	}
	if _, err := j.db.Exec(`UPDATE strategy_attempt_lineage SET state='IN_DOUBT',revision=revision+1 WHERE attempt_id=?`, plan.AttemptID); err != nil {
		t.Fatalf("sanctioned CAS rejected: %v", err)
	}
	if _, err := j.db.Exec(`UPDATE strategy_attempt_lineage SET state='DISPATCHED',revision=revision+1 WHERE attempt_id=?`, plan.AttemptID); err != nil {
		t.Fatalf("sanctioned recovery CAS rejected: %v", err)
	}
}

func TestStrategyExecutionLinkReplayPreservesFirstTimestamp(t *testing.T) {
	j, fake := openReservationJournal(t)
	plan := strategyPlanFixture(t, "link-replay", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if err := j.AppendStrategyExecutionLink(context.Background(), "acct-1", plan.AttemptID, "FILL", "fill-1"); err != nil {
		t.Fatal(err)
	}
	var first string
	if err := j.db.QueryRow(`SELECT recorded_at FROM strategy_execution_lineage WHERE account_ref='acct-1' AND kind='FILL' AND external_ref='fill-1'`).Scan(&first); err != nil {
		t.Fatal(err)
	}
	fake.Advance(time.Hour)
	if err := j.AppendStrategyExecutionLink(context.Background(), "acct-1", plan.AttemptID, "FILL", "fill-1"); err != nil {
		t.Fatalf("exact replay: %v", err)
	}
	var count int
	var replayed string
	if err := j.db.QueryRow(`SELECT count(*),min(recorded_at) FROM strategy_execution_lineage WHERE account_ref='acct-1' AND kind='FILL' AND external_ref='fill-1'`).Scan(&count, &replayed); err != nil || count != 1 || replayed != first {
		t.Fatalf("count=%d first=%s replayed=%s err=%v", count, first, replayed, err)
	}
}

func TestStrategyProductionIssuanceFailureRollsBackReservationAndAllAuthority(t *testing.T) {
	j, _ := openReservationJournal(t)
	if _, err := j.db.Exec(`CREATE TRIGGER fail_strategy_attempt BEFORE INSERT ON strategy_attempt_lineage BEGIN SELECT RAISE(ABORT,'synthetic failure'); END;`); err != nil {
		t.Fatal(err)
	}
	request := strategyIssueFixture(t, j, "0", "acct-1")
	before := mustVersion(t, j, "acct-1")
	if _, err := j.RecordStrategyDecisionAndReserve(context.Background(), request); err == nil {
		t.Fatal("synthetic strategy failure committed")
	}
	for _, table := range []string{"decisions", "risk_reservations", "strategy_decision_lineage", "strategy_attempt_lineage", "strategy_execution_lineage"} {
		var count int
		if err := j.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s partial count=%d err=%v", table, count, err)
		}
	}
	if after := mustVersion(t, j, "acct-1"); after != before {
		t.Fatalf("reservation version changed %d -> %d", before, after)
	}
}

func TestStrategyPlanConcurrentDuplicateHasOneWriterAndExactReplays(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	plan := strategyPlanFixture(t, "4", "acct-1")
	const workers = 8
	results := make(chan StrategyPlanReceipt, workers)
	errs := make(chan error, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			receipt, err := j.planStrategyEntryForTest(context.Background(), plan)
			results <- receipt
			errs <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	nonIdempotent := 0
	for receipt := range results {
		if !receipt.Idempotent {
			nonIdempotent++
		}
	}
	if nonIdempotent != 1 {
		t.Fatalf("non-idempotent writers=%d", nonIdempotent)
	}
}

func TestStrategyLineageRestartReverseLookupAndAccountScope(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	ctx := context.Background()
	for _, fixture := range []struct{ suffix, account string }{{"5", "acct-1"}, {"6", "acct-2"}} {
		plan := strategyPlanFixture(t, fixture.suffix, fixture.account)
		if _, err := j.planStrategyEntryForTest(ctx, plan); err != nil {
			t.Fatal(err)
		}
		for _, kind := range []string{"BROKER_ORDER", "FILL", "POSITION", "CLOSE_OUTCOME"} {
			if err := j.AppendStrategyExecutionLink(ctx, fixture.account, plan.AttemptID, kind, "shared-ref"); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := openTestJournalAt(t, path)
	defer reopened.Close()
	trace, err := reopened.LookupStrategyTrace(ctx, "acct-2", "CLOSE_OUTCOME", "shared-ref")
	if err != nil || trace.AccountRef != "acct-2" || trace.AttemptID != "attempt-6" || trace.ExecutionKind != "CLOSE_OUTCOME" {
		t.Fatalf("trace=%+v err=%v", trace, err)
	}
	if _, err := reopened.LookupStrategyTrace(ctx, "acct-missing", "FILL", "shared-ref"); !errors.Is(err, ErrStrategyTraceNotFound) {
		t.Fatalf("not-found err=%v", err)
	}
}

func TestStrategyReverseLookupUsesBoundedCoveringIndex(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	rows, err := j.db.Query(`EXPLAIN QUERY PLAN SELECT e.account_ref,d.entry_decision_identity FROM strategy_execution_lineage e INDEXED BY idx_strategy_execution_reverse JOIN strategy_attempt_lineage a ON a.attempt_id=e.attempt_id JOIN strategy_decision_lineage d ON d.entry_decision_identity=a.entry_decision_identity WHERE e.account_ref=? AND e.kind=? AND e.external_ref=?`, "acct", "FILL", "fill")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var details []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		details = append(details, detail)
	}
	plan := strings.Join(details, " | ")
	if !strings.Contains(plan, "idx_strategy_execution_reverse") || strings.Contains(plan, "SCAN e") {
		t.Fatalf("query plan=%s", plan)
	}
}

func TestStrategyTerminalStateIsCASAndDurable(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	receipt, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "7", "acct-1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := j.RecordStrategyInDoubt(context.Background(), receipt, "official_timeout"); err != nil {
		t.Fatal(err)
	}
	if err := j.RecordStrategyRefusal(context.Background(), receipt, "late_refusal"); err == nil {
		t.Fatal("stale terminal transition accepted")
	}
	var state string
	var revision int
	if err := j.db.QueryRow(`SELECT state,revision FROM strategy_attempt_lineage WHERE attempt_id=?`, receipt.AttemptID).Scan(&state, &revision); err != nil || state != "IN_DOUBT" || revision != 2 {
		t.Fatalf("state/revision=%s/%d err=%v", state, revision, err)
	}
}

func prepareStrategyMutation(t *testing.T, j *Journal, receipt StrategyPlanReceipt, mutationID string, state AttemptState) *Attempt {
	t.Helper()
	req := testRequest()
	req.Intent.ID = receipt.AttemptID
	req.Intent.AccountRef = receipt.AccountRef
	req.Intent.Market = "kr"
	req.Intent.Symbol = "005930"
	req.Intent.Currency = "KRW"
	req.Intent.Fingerprint = "strategy-" + receipt.AttemptID
	req.AttemptID = mutationID
	attempt, err := j.Prepare(context.Background(), req)
	if err != nil {
		t.Fatalf("Prepare strategy mutation: %v", err)
	}
	switch state {
	case StateRecorded:
	case StateDispatchStarted:
		err = attempt.MarkDispatchStarted(context.Background())
	case StateAcked:
		if err = attempt.MarkDispatchStarted(context.Background()); err == nil {
			err = attempt.MarkAcked(context.Background(), "broker-"+mutationID)
		}
	case StateInDoubt:
		if err = attempt.MarkDispatchStarted(context.Background()); err == nil {
			err = attempt.MarkInDoubt(context.Background(), "test", "unknown")
		}
	case StateUnresolvedInDoubt:
		if err = attempt.MarkDispatchStarted(context.Background()); err == nil {
			err = attempt.MarkInDoubt(context.Background(), "test", "unknown")
		}
		if err == nil {
			err = attempt.Settle(context.Background(), StateUnresolvedInDoubt, "test", "unresolved")
		}
	case StateNotDispatched:
		err = attempt.Settle(context.Background(), state, "test", "terminal refusal")
	case StateFailedConfirmed:
		if err = attempt.MarkDispatchStarted(context.Background()); err == nil {
			err = attempt.Settle(context.Background(), state, "test", "terminal refusal")
		}
	case StateConfirmed:
		if err = attempt.MarkDispatchStarted(context.Background()); err == nil {
			err = attempt.MarkAcked(context.Background(), "broker-"+mutationID)
		}
		if err == nil {
			err = attempt.Settle(context.Background(), StateConfirmed, "test", "confirmed")
		}
	default:
		t.Fatalf("unsupported mutation state %s", state)
	}
	if err != nil {
		t.Fatalf("moving mutation to %s: %v", state, err)
	}
	return attempt
}

func assertStrategyAttemptState(t *testing.T, j *Journal, attemptID, state string, revision int, reason string) {
	t.Helper()
	var gotState string
	var gotRevision int
	if err := j.db.QueryRow(`SELECT state,revision FROM strategy_attempt_lineage WHERE attempt_id=?`, attemptID).Scan(&gotState, &gotRevision); err != nil {
		t.Fatal(err)
	}
	if gotState != state || gotRevision != revision {
		t.Fatalf("strategy state/revision=%s/%d want=%s/%d", gotState, gotRevision, state, revision)
	}
	if reason != "" {
		var gotReason string
		if err := j.db.QueryRow(`SELECT reason_code FROM strategy_attempt_refusals WHERE attempt_id=? AND revision=?`, attemptID, revision).Scan(&gotReason); err != nil || gotReason != reason {
			t.Fatalf("reason=%q err=%v want=%q", gotReason, err, reason)
		}
	}
}

func TestPendingStrategyPlansAreAccountScopedAndZeroAttemptRecoversToRefused(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	receipt, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "a", "acct-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "b", "acct-2")); err != nil {
		t.Fatal(err)
	}
	pending, err := j.PendingStrategyPlans(context.Background(), "acct-1")
	if err != nil || len(pending) != 1 || pending[0] != receipt {
		t.Fatalf("pending=%+v err=%v", pending, err)
	}
	if err := j.RecoverStrategyDispatch(context.Background(), pending[0], "acct-1"); err == nil || !strings.Contains(err.Error(), "no mutation attempt") {
		t.Fatalf("recovery err=%v", err)
	}
	assertStrategyAttemptState(t, j, receipt.AttemptID, "REFUSED", 2, "no_mutation_attempt")
	pending, err = j.PendingStrategyPlans(context.Background(), "acct-1")
	if err != nil || len(pending) != 0 {
		t.Fatalf("terminal pending=%+v err=%v", pending, err)
	}
}

func TestStrategyRecoveryClassifiesOneCoreAttemptExactly(t *testing.T) {
	for _, tc := range []struct {
		name       string
		core       AttemptState
		wantState  string
		wantReason string
	}{
		{name: "recorded nullable broker", core: StateRecorded, wantState: "IN_DOUBT", wantReason: "mutation_attempt_requires_recovery"},
		{name: "dispatch started", core: StateDispatchStarted, wantState: "IN_DOUBT", wantReason: "mutation_attempt_requires_recovery"},
		{name: "acked", core: StateAcked, wantState: "IN_DOUBT", wantReason: "mutation_attempt_requires_recovery"},
		{name: "in doubt", core: StateInDoubt, wantState: "IN_DOUBT", wantReason: "mutation_attempt_requires_recovery"},
		{name: "unresolved", core: StateUnresolvedInDoubt, wantState: "IN_DOUBT", wantReason: "mutation_attempt_requires_recovery"},
		{name: "not dispatched", core: StateNotDispatched, wantState: "REFUSED", wantReason: "mutation_attempt_refused"},
		{name: "failed confirmed", core: StateFailedConfirmed, wantState: "REFUSED", wantReason: "mutation_attempt_refused"},
		{name: "confirmed", core: StateConfirmed, wantState: "DISPATCHED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
			defer j.Close()
			receipt, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "c", "acct-1"))
			if err != nil {
				t.Fatal(err)
			}
			prepareStrategyMutation(t, j, receipt, "mutation-1", tc.core)
			err = j.RecoverStrategyDispatch(context.Background(), receipt, receipt.AccountRef)
			if tc.core == StateConfirmed {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("non-confirmed recovery reported success")
			}
			assertStrategyAttemptState(t, j, receipt.AttemptID, tc.wantState, 2, tc.wantReason)
			if tc.core == StateConfirmed {
				for _, link := range []struct{ kind, ref string }{{"MUTATION_ATTEMPT", "mutation-1"}, {"BROKER_ORDER", "broker-mutation-1"}} {
					if _, err := j.LookupStrategyTrace(context.Background(), receipt.AccountRef, link.kind, link.ref); err != nil {
						t.Fatalf("missing %s link: %v", link.kind, err)
					}
				}
			}
		})
	}
}

func TestStrategyRecoveryMultipleCoreAttemptsIsDurableInDoubt(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	receipt, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "d", "acct-1"))
	if err != nil {
		t.Fatal(err)
	}
	prepareStrategyMutation(t, j, receipt, "mutation-1", StateRecorded)
	prepareStrategyMutation(t, j, receipt, "mutation-2", StateRecorded)
	if err := j.RecoverStrategyDispatch(context.Background(), receipt, receipt.AccountRef); err == nil || !strings.Contains(err.Error(), "multiple mutation attempts") {
		t.Fatalf("recovery err=%v", err)
	}
	assertStrategyAttemptState(t, j, receipt.AttemptID, "IN_DOUBT", 2, "ambiguous_mutation_attempts")
}

func TestStrategyRecoveryCanPromoteCurrentInDoubtReceiptToDispatched(t *testing.T) {
	j := openTestJournalAt(t, filepath.Join(t.TempDir(), "journal.db"))
	defer j.Close()
	receipt, err := j.planStrategyEntryForTest(context.Background(), strategyPlanFixture(t, "e", "acct-1"))
	if err != nil {
		t.Fatal(err)
	}
	prepareStrategyMutation(t, j, receipt, "mutation-1", StateConfirmed)
	if err := j.RecordStrategyInDoubt(context.Background(), receipt, "response_lost"); err != nil {
		t.Fatal(err)
	}
	pending, err := j.PendingStrategyPlans(context.Background(), receipt.AccountRef)
	if err != nil || len(pending) != 1 || pending[0].State != "IN_DOUBT" || pending[0].Revision != 2 {
		t.Fatalf("pending=%+v err=%v", pending, err)
	}
	if err := j.RecoverStrategyDispatch(context.Background(), pending[0], receipt.AccountRef); err != nil {
		t.Fatal(err)
	}
	assertStrategyAttemptState(t, j, receipt.AttemptID, "DISPATCHED", 3, "")
}

func ExampleJournal_LookupStrategyTrace() {
	fmt.Println("account-scoped fill/position/close references map to one candidate life")
	// Output: account-scoped fill/position/close references map to one candidate life
}
