package journal

// a112 5.3.3: v31 저널이 앞으로 넘어오고, 그 순간 레인 잠금 테이블이 **비어**
// 있다.
//
// 이 마이그레이션의 위험은 잃을 행이 없다는 데 있지 않다(새 테이블 둘뿐이다).
// 위험은 반대쪽이다: 올라오자마자 잠긴 레인이 하나라도 있으면, 아무도 잠근 적
// 없는 레인의 신규 진입이 서명된 매니페스트를 새로 낼 때까지 막힌다. 업그레이드가
// 조용히 거래를 멈추는 모양이다.

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationV31ToV32StartsWithNoLaneLatched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 31)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 32)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 32 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	open, err := j.OpenStrategyLaneLatches(context.Background(), "acct-lane")
	if err != nil {
		t.Fatalf("open latches on a migrated journal: %v", err)
	}
	if len(open) != 0 {
		t.Fatalf("업그레이드 직후 잠긴 레인 %d 개 — 아무도 잠근 적이 없는데 진입이 막힌다", len(open))
	}

	// 그리고 그 저널은 잠금을 받을 수 있어야 한다. 테이블만 생기고 못 쓰면
	// 다음 로트가 그것을 발견한다.
	if _, err := j.RecordStrategyLaneLatch(context.Background(), laneLatchFixture(t)); err != nil {
		t.Fatalf("마이그레이션한 저널이 잠금을 받지 못했다: %v", err)
	}
}
