package strategyevidence

// snapshot_items 의 제약이 **실제로 거부하는지** 잰다.
//
// 원래 근거는 기본키 열이 2개이고 pragma_foreign_key_list 가 2행을 돌려준다는 것뿐이었다.
// 어떤 열인지, 어느 표를 가리키는지, 중복이나 고아 행이 실제로 거부되는지는 아무도 세지
// 않았다 — 열 이름을 바꾸거나 참조 대상을 바꿔도 개수는 그대로 2다. review.md 는 그
// 개수 세기를 "Runtime PRAGMA tests prove …"로 올려 적었다.

import (
	"context"
	"testing"
)

func TestSnapshotItemsSchemaNamesItsKeyAndReferences(t *testing.T) {
	t.Parallel()
	store := openTestStore(t)

	rows, err := store.db.Query(`SELECT name FROM pragma_table_info('snapshot_items') WHERE pk > 0 ORDER BY pk`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var primaryKey []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		primaryKey = append(primaryKey, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(primaryKey) != 2 || primaryKey[0] != "snapshot_id" || primaryKey[1] != "ordinal" {
		t.Fatalf("snapshot_items primary key = %v, want [snapshot_id ordinal]", primaryKey)
	}

	references := map[string]string{}
	foreignKeys, err := store.db.Query(`SELECT "table", "from", "to" FROM pragma_foreign_key_list('snapshot_items')`)
	if err != nil {
		t.Fatal(err)
	}
	defer foreignKeys.Close()
	for foreignKeys.Next() {
		var table, from, to string
		if err := foreignKeys.Scan(&table, &from, &to); err != nil {
			t.Fatal(err)
		}
		references[from] = table + "." + to
	}
	if err := foreignKeys.Err(); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"snapshot_id": "snapshots.snapshot_id",
		"evidence_id": "evidence_records.evidence_id",
	}
	for column, target := range want {
		if references[column] != target {
			t.Fatalf("snapshot_items.%s references %q, want %q", column, references[column], target)
		}
	}
	if len(references) != len(want) {
		t.Fatalf("snapshot_items foreign keys = %v, want exactly %v", references, want)
	}
}

func TestSnapshotItemsRejectDuplicateAndOrphanRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, snapshot := sealedSnapshotForTamper(t)
	evidenceID := snapshot.Items[0].Header().EvidenceID
	digest := snapshot.Items[0].PayloadDigest()

	for _, test := range []struct {
		name                            string
		snapshotID, ordinal, evidenceID string
		reason                          string
	}{
		{
			name: "duplicate-ordinal", snapshotID: snapshot.ID, ordinal: "0", evidenceID: evidenceID,
			reason: "the composite primary key must refuse a second row at the same ordinal",
		},
		{
			name: "duplicate-evidence", snapshotID: snapshot.ID, ordinal: "1", evidenceID: evidenceID,
			reason: "UNIQUE(snapshot_id, evidence_id) must refuse the same evidence twice in one snapshot",
		},
		{
			name: "orphan-evidence", snapshotID: snapshot.ID, ordinal: "2", evidenceID: "never-appended",
			reason: "the evidence_records foreign key must refuse an item with no immutable row behind it",
		},
		{
			name: "orphan-snapshot", snapshotID: "snapshot-" + goldenSnapshotDigest, ordinal: "0", evidenceID: evidenceID,
			reason: "the snapshots foreign key must refuse an item belonging to no sealed snapshot",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := store.db.ExecContext(ctx,
				`INSERT INTO snapshot_items(snapshot_id, ordinal, evidence_id, payload_digest) VALUES (?, ?, ?, ?)`,
				test.snapshotID, test.ordinal, test.evidenceID, digest)
			if err == nil {
				t.Fatalf("snapshot_items accepted %s: %s", test.name, test.reason)
			}
		})
	}

	// 양성 대조군: 같은 INSERT 문이 정당한 행은 통과시킨다. 이것이 없으면 "이 표에는
	// 아무것도 못 넣는다"가 위의 네 시험을 전부 통과시킨다.
	other := validHeader(snapshot.Market, KindUSParticipation, "second-item", "tamper-rev-2")
	other.SourceRecordID = "record-2"
	appended, err := store.Append(ctx, mustEnvelope(t, other, `{"value":9}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx,
		`INSERT INTO snapshot_items(snapshot_id, ordinal, evidence_id, payload_digest) VALUES (?, ?, ?, ?)`,
		snapshot.ID, 1, appended.Evidence.Header().EvidenceID, appended.Evidence.PayloadDigest()); err != nil {
		t.Fatalf("snapshot_items refused a legitimate second item: %v", err)
	}
}
