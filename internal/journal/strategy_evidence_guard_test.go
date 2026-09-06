package journal

// a064 가 넣은 두 겹의 방어를 **각각** 잰다.
//
// 같은 여섯 입력을 Go 가드(insertExactStrategyDecision)와 v21 SQL 트리거가 둘 다 거부한다.
// 기존 시험은 `err != nil` 과 행 0개만 보았으므로, 어느 층이 막았는지 구분하지 못했다 —
// 그래서 a064 의 Go 가드를 통째로 지워도 스위트가 초록으로 남았다(한 규칙을 두 곳에서
// 판정하면 서로가 서로의 시험을 통과시킨다). 아래는 층을 갈라서 잰다.
//
//   - Go 층: 타입 있는 sentinel 로 거부되는가
//   - SQL 층: Go 를 우회한 직접 INSERT/UPDATE 에서 술어 **네 갈래가 각각** 거부되는가

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestStrategyEvidenceGoGuardRefusesBeforeSQL(t *testing.T) {
	validID, validDigest := canonicalSnapshotReference('d')
	for _, test := range []struct {
		name, id, digest string
	}{
		{name: "id-only", id: validID},
		{name: "digest-only", digest: validDigest},
		{name: "wrong-id", id: "snapshot-other", digest: validDigest},
		{name: "wrong-digest", id: validID, digest: strings.Repeat("z", 64)},
		{name: "uppercase-digest", id: "snapshot-" + strings.ToUpper(validDigest), digest: strings.ToUpper(validDigest)},
		{name: "whitespace", id: validID + " ", digest: validDigest},
		{name: "short-digest", id: "snapshot-" + strings.Repeat("a", 63), digest: strings.Repeat("a", 63)},
	} {
		t.Run(test.name, func(t *testing.T) {
			j := openTestJournal(t)
			plan := strategyPlanFixture(t, "go-guard-"+test.name, "acct-1")
			plan.Lineage.ConsumedEvidenceSnapshotID = test.id
			plan.Lineage.ConsumedEvidenceSnapshotDigest = test.digest
			_, err := j.planStrategyEntryForTest(context.Background(), plan)
			if !errors.Is(err, ErrStrategyEvidenceReferenceInvalid) {
				t.Fatalf("refusal came from another layer: %v", err)
			}
			for _, table := range []string{"decisions", "strategy_decision_lineage", "strategy_attempt_lineage"} {
				var count int
				if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s count=%d err=%v; invalid lineage partially committed", table, count, err)
				}
			}
		})
	}
}

// TestStrategyEvidenceGoGuardAcceptsACanonicalReference 는 위 가드가 정상 참조까지
// 막지 않는다는 양성 대조군이다.
func TestStrategyEvidenceGoGuardAcceptsACanonicalReference(t *testing.T) {
	j := openTestJournal(t)
	plan := strategyPlanFixture(t, "go-guard-canonical", "acct-1")
	plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('a')
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatalf("canonical consumed-evidence reference refused: %v", err)
	}
	legacy := strategyPlanFixture(t, "go-guard-legacy", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), legacy); err != nil {
		t.Fatalf("legacy NULL/NULL reference refused: %v", err)
	}
}

// insertLineageDirectly 는 Go 가드를 우회해 SQL 층에만 값을 들이민다. v21 트리거의
// 형식 술어들은 Go 가 먼저 거부하므로 이 길이 아니면 영원히 실행되지 않는다.
func insertLineageDirectly(t *testing.T, j *Journal, source, identity, id, digest string) error {
	t.Helper()
	_, err := j.db.Exec(`INSERT INTO strategy_decision_lineage(
entry_decision_identity,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at,consumed_evidence_snapshot_id,consumed_evidence_snapshot_digest)
SELECT ?,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at,NULLIF(?,'__null__'),NULLIF(?,'__null__')
FROM strategy_decision_lineage WHERE entry_decision_identity=?`, identity, id, digest, source)
	return err
}

func TestV21TriggerRefusesEveryMalformedReferenceShape(t *testing.T) {
	valid := strings.Repeat("a", 64)
	// 트리거 술어의 갈래를 하나씩 겨눈다. 다섯을 다 지워도 안 걸리던 자리다.
	//
	// 다섯 갈래 중 **결정적인 것은 넷**이다. `digest!=lower(digest)` 는 GLOB
	// '*[^0-9a-f]*' 에 논리적으로 포함된다 — 자기 소문자와 다르면서 동시에 [0-9a-f]
	// 안에 있는 ASCII 문자는 없다. 그래서 대문자 digest 는 언제나 GLOB 이 먼저 잡고,
	// lower 갈래를 지워도 어떤 시험도 빨개지지 않는다(2026-09-07 독립 리뷰 S2가 전체
	// 패키지 스위트로 실측). 아래 disjunct 표기는 그 사실을 그대로 적는다.
	//
	// 주의: Go 쪽 lower 검사(validConsumedEvidenceReference)는 이와 달리 결정적이다.
	// hex.DecodeString 이 대문자를 받아들이므로 그 검사를 지우면
	// TestStrategyEvidenceGoGuardRefusesBeforeSQL/uppercase-digest 가 빨개진다.
	cases := []struct {
		name, id, digest, disjunct string
	}{
		{name: "null-parity", id: "snapshot-" + valid, digest: "__null__", disjunct: "NULL parity"},
		{name: "prefix-mismatch", id: "snap-" + valid, digest: valid, disjunct: "'snapshot-'||digest"},
		{name: "short-digest", id: "snapshot-" + strings.Repeat("a", 63), digest: strings.Repeat("a", 63), disjunct: "length(digest)!=64"},
		{name: "uppercase-digest", id: "snapshot-" + strings.ToUpper(valid), digest: strings.ToUpper(valid), disjunct: "GLOB '*[^0-9a-f]*' (the lower() disjunct is implied by it and never decides)"},
		{name: "non-hex-digest", id: "snapshot-" + strings.Repeat("g", 64), digest: strings.Repeat("g", 64), disjunct: "GLOB '*[^0-9a-f]*'"},
	}
	for _, test := range cases {
		t.Run("insert/"+test.name, func(t *testing.T) {
			j := openTestJournal(t)
			source := strategyPlanFixture(t, "trigger-src-"+test.name, "acct-1")
			if _, err := j.planStrategyEntryForTest(context.Background(), source); err != nil {
				t.Fatal(err)
			}
			err := insertLineageDirectly(t, j, source.Lineage.DecisionIdentity, "trigger-"+test.name, test.id, test.digest)
			if err == nil {
				t.Fatalf("v21 insert trigger accepted %s; disjunct %s is unenforced", test.name, test.disjunct)
			}
			if !strings.Contains(err.Error(), "invalid consumed evidence snapshot reference") {
				t.Fatalf("%s was refused by something other than the v21 trigger: %v", test.name, err)
			}
		})
		t.Run("update/"+test.name, func(t *testing.T) {
			j := openTestJournal(t)
			source := strategyPlanFixture(t, "trigger-upd-"+test.name, "acct-1")
			if _, err := j.planStrategyEntryForTest(context.Background(), source); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`DROP TRIGGER strategy_decision_lineage_no_update`); err != nil {
				t.Fatal(err)
			}
			_, err := j.db.Exec(`UPDATE strategy_decision_lineage SET consumed_evidence_snapshot_id=NULLIF(?,'__null__'),consumed_evidence_snapshot_digest=NULLIF(?,'__null__') WHERE entry_decision_identity=?`,
				test.id, test.digest, source.Lineage.DecisionIdentity)
			if err == nil {
				t.Fatalf("v21 update trigger accepted %s; disjunct %s is unenforced", test.name, test.disjunct)
			}
			if !strings.Contains(err.Error(), "invalid consumed evidence snapshot reference") {
				t.Fatalf("%s update was refused by something other than the v21 trigger: %v", test.name, err)
			}
		})
	}

	// 양성 대조군: 같은 직접 INSERT 가 정당한 참조와 NULL/NULL 을 통과시킨다.
	// 이것이 없으면 "WHEN 1" 로 바꿔 전부 막는 트리거가 위 시험을 모두 통과한다.
	t.Run("accepts-canonical", func(t *testing.T) {
		j := openTestJournal(t)
		source := strategyPlanFixture(t, "trigger-src-ok", "acct-1")
		if _, err := j.planStrategyEntryForTest(context.Background(), source); err != nil {
			t.Fatal(err)
		}
		if err := insertLineageDirectly(t, j, source.Lineage.DecisionIdentity, "trigger-ok", "snapshot-"+valid, valid); err != nil {
			t.Fatalf("v21 trigger refused a canonical reference: %v", err)
		}
		if err := insertLineageDirectly(t, j, source.Lineage.DecisionIdentity, "trigger-null", "__null__", "__null__"); err != nil {
			t.Fatalf("v21 trigger refused a legacy NULL/NULL row: %v", err)
		}
	})
}

// journalContentDigest 는 저널 **전체**의 표 목록과 행 수를 한 문자열로 접는다.
// 세 표만 세던 원래 근거와 달리, 실행 경로가 어느 표에든 한 행을 쓰면 값이 달라진다.
func journalContentDigest(t *testing.T, j *Journal) string {
	t.Helper()
	rows, err := j.db.Query(`SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(tables) == 0 {
		t.Fatal("journal has no tables; the content digest would be vacuous")
	}
	sort.Strings(tables)
	var parts []string
	for _, table := range tables {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM "` + table + `"`).Scan(&count); err != nil {
			t.Fatalf("counting %s: %v", table, err)
		}
		parts = append(parts, fmt.Sprintf("%s=%d", table, count))
	}
	return strings.Join(parts, ",")
}

func TestDormantEvidenceReadWritesNothingAnywhereInTheJournal(t *testing.T) {
	j := openTestJournal(t)
	plan := strategyPlanFixture(t, "dormant-spy", "acct-1")
	plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('a')
	if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	before := journalContentDigest(t, j)

	readonly := openTestReadOnly(t, j.Path())
	boundary, err := NewStrategyEvidenceReadBoundary(readonly)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := boundary.ConsumedSnapshot(context.Background(), plan.Lineage.DecisionIdentity); err != nil {
		t.Fatal(err)
	}
	if _, err := boundary.ConsumedSnapshot(context.Background(), "absent"); !errors.Is(err, ErrStrategyTraceNotFound) {
		t.Fatalf("missing decision read error=%v", err)
	}
	if after := journalContentDigest(t, j); after != before {
		t.Fatalf("dormant evidence read changed the journal:\nbefore %s\nafter  %s", before, after)
	}

	// 읽기 핸들 **자체가** 쓸 수 없다는 것도 잰다. 행 수 비교만으로는 "이번에는 안 썼다"
	// 까지밖에 못 말한다.
	//
	// 대상은 불변 트리거가 지키지 않는 자리여야 한다. 앞선 판본은
	// `UPDATE strategy_decision_lineage` 를 썼는데, 그것은 strategy_decision_lineage_no_update
	// 트리거가 대신 막는다 — DSN 에서 mode=ro 를 빼도 그대로 통과했다(2026-09-07 독립
	// 리뷰 S5). 스키마 쓰기와 PRAGMA 는 어떤 트리거도 지키지 않으므로 읽기 전용 연결만이
	// 이것을 거부할 수 있다.
	for _, write := range []string{
		`CREATE TABLE dormant_read_probe(x TEXT)`,
		`PRAGMA user_version = 99`,
	} {
		if _, err := readonly.db.Exec(write); err == nil {
			t.Fatalf("the dormant read handle accepted a write: %s", write)
		}
	}

	// 양성 대조군: 같은 계측기가 실제 쓰기를 잡는다.
	second := strategyPlanFixture(t, "dormant-spy-2", "acct-1")
	if _, err := j.planStrategyEntryForTest(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	if after := journalContentDigest(t, j); after == before {
		t.Fatalf("journal content digest did not move after a real write: %s", after)
	}
}

// TestStrategyDecisionLineageInsertFailureSurfacesTheStatementError 는
// insertExactStrategyDecision 의 `if err != nil { return 0, err }` 갈래를 묶는다.
// 그 갈래를 `_ = err` 로 지워도 뒤따르는 SELECT 가 행을 못 찾아 StrategyCollisionError 로
// 끝나므로, "롤백됐다"만 보는 시험은 둘을 구별하지 못한다(2026-09-07 독립 리뷰 S6).
// 여기서는 **어떤 오류로 끝났는지**를 본다.
func TestStrategyDecisionLineageInsertFailureSurfacesTheStatementError(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.db.Exec(`CREATE TRIGGER fail_decision_lineage BEFORE INSERT ON strategy_decision_lineage
BEGIN SELECT RAISE(ABORT,'synthetic decision lineage failure'); END`); err != nil {
		t.Fatal(err)
	}
	plan := strategyPlanFixture(t, "insert-failure", "acct-1")
	plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('a')
	_, err := j.planStrategyEntryForTest(context.Background(), plan)
	if err == nil {
		t.Fatal("a failing lineage INSERT was reported as success")
	}
	var collision *StrategyCollisionError
	if errors.As(err, &collision) {
		t.Fatalf("the statement error was swallowed and re-read as a collision: %v", err)
	}
	if !strings.Contains(err.Error(), "synthetic decision lineage failure") {
		t.Fatalf("the statement error did not surface: %v", err)
	}
	for _, table := range []string{"decisions", "strategy_decision_lineage", "strategy_attempt_lineage"} {
		var count int
		if err := j.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s count=%d err=%v; the failed first leg did not roll back", table, count, err)
		}
	}
}

// TestStrategyEvidenceReadBoundaryRefusesAMalformedStoredReference 는 읽기 쪽
// validConsumedEvidenceReference 호출을 묶는다. 같은 규칙을 쓰기 가드와 v21 트리거도
// 판정하므로, 읽기 쪽 호출을 지워도 아무 시험이 안 깨졌다(2026-09-07 독립 리뷰 S3).
// 저장된 값이 이미 어긋난 상태를 만들려면 트리거를 먼저 떼야 한다 — 그것이 이
// 갈래가 실제로 지키는 상황이다(트리거가 없던 시절에 쓰인 행, 또는 손으로 고친 파일).
func TestStrategyEvidenceReadBoundaryRefusesAMalformedStoredReference(t *testing.T) {
	for _, test := range []struct{ name, id, digest string }{
		{name: "prefix-mismatch", id: "snap-" + strings.Repeat("a", 64), digest: strings.Repeat("a", 64)},
		{name: "short-digest", id: "snapshot-" + strings.Repeat("a", 63), digest: strings.Repeat("a", 63)},
		{name: "uppercase-digest", id: "snapshot-" + strings.Repeat("A", 64), digest: strings.Repeat("A", 64)},
		{name: "non-hex-digest", id: "snapshot-" + strings.Repeat("g", 64), digest: strings.Repeat("g", 64)},
	} {
		t.Run(test.name, func(t *testing.T) {
			j := openTestJournal(t)
			plan := strategyPlanFixture(t, "stored-malformed-"+test.name, "acct-1")
			if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
				t.Fatal(err)
			}
			for _, trigger := range []string{"strategy_decision_lineage_no_update", "strategy_evidence_reference_update_guard"} {
				if _, err := j.db.Exec(`DROP TRIGGER ` + trigger); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := j.db.Exec(`UPDATE strategy_decision_lineage SET consumed_evidence_snapshot_id=?,consumed_evidence_snapshot_digest=? WHERE entry_decision_identity=?`,
				test.id, test.digest, plan.Lineage.DecisionIdentity); err != nil {
				t.Fatal(err)
			}
			boundary, err := NewStrategyEvidenceReadBoundary(openTestReadOnly(t, j.Path()))
			if err != nil {
				t.Fatal(err)
			}
			got, err := boundary.ConsumedSnapshot(context.Background(), plan.Lineage.DecisionIdentity)
			if !errors.Is(err, ErrStrategyEvidenceSnapshotUnavailable) {
				t.Fatalf("a malformed stored reference was handed back as usable: %+v err=%v", got, err)
			}
		})
	}

	// 양성 대조군: 정상 참조는 그대로 읽힌다.
	t.Run("canonical", func(t *testing.T) {
		j := openTestJournal(t)
		plan := strategyPlanFixture(t, "stored-canonical", "acct-1")
		plan.Lineage.ConsumedEvidenceSnapshotID, plan.Lineage.ConsumedEvidenceSnapshotDigest = canonicalSnapshotReference('a')
		if _, err := j.planStrategyEntryForTest(context.Background(), plan); err != nil {
			t.Fatal(err)
		}
		boundary, err := NewStrategyEvidenceReadBoundary(openTestReadOnly(t, j.Path()))
		if err != nil {
			t.Fatal(err)
		}
		got, err := boundary.ConsumedSnapshot(context.Background(), plan.Lineage.DecisionIdentity)
		if err != nil || got.SnapshotID != plan.Lineage.ConsumedEvidenceSnapshotID {
			t.Fatalf("a canonical stored reference was refused: %+v err=%v", got, err)
		}
	})
}
