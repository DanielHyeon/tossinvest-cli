package verifylive_test

// reconcile_marshal_test.go — a121 GREEN: Artifact.MarshalJSON(대사된 artifact 만 확정형으로 씀)이 대사되지 않은 artifact 의
// 직렬화 바이트를 바꾸지 않음을 구현 base 의 Artifact 사본(reconcile_rollback_test.go — 토큰 sha 로 핀)과 대조한다.
// 기록 바이트·Digest 가 이 변경으로 움직이면 다른 프로세스·구 바이너리와 지문이 갈린다.

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

func TestUnreconciledArtifactsSerialiseExactlyAsTheBase(t *testing.T) {
	at := time.Date(2026, 7, 30, 1, 2, 3, 4, time.UTC)
	cases := []Artifact{
		{},
		{Kind: KindConditional, ID: "CO-1", Symbol: "005930", CreatedAt: at, Deliberate: true, ChainID: "c", HeldUntil: StepConditionalCancel},
		{Kind: KindOrder, ID: "O-1", Symbol: "AAPL", CreatedAt: at, CancelledAt: at.Add(time.Minute), Cancelled: true},
		{Kind: KindOrder, ID: "O-2", Symbol: "MWG", CreatedAt: at, Filled: true, FilledAt: at.Add(time.Hour), Note: "<a&b> \"q\""},
	}
	for i, old := range cases {
		cur := verifylive.Artifact{Kind: old.Kind, ID: old.ID, Symbol: old.Symbol, CreatedAt: old.CreatedAt,
			CancelledAt: old.CancelledAt, Cancelled: old.Cancelled, Filled: old.Filled, FilledAt: old.FilledAt,
			Deliberate: old.Deliberate, HeldUntil: old.HeldUntil, ChainID: old.ChainID, Note: old.Note}
		want, err := json.Marshal(old)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(cur)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("case %d: an unreconciled artifact now serialises differently\n got %s\nwant %s", i, got, want)
		}
		gotEntry, _ := json.Marshal(verifylive.Entry{Artifacts: []verifylive.Artifact{cur}})
		wantEntry, _ := json.Marshal(struct {
			FormatVersion int                `json:"format_version"`
			Kind          string             `json:"kind"`
			RunID         string             `json:"run_id"`
			Process       verifylive.Process `json:"process"`
			StepID        StepID             `json:"step_id"`
			Title         string             `json:"title"`
			StartedAt     time.Time          `json:"started_at"`
			FinishedAt    time.Time          `json:"finished_at"`
			AccountRef    string             `json:"account_ref"`
			Mutating      bool               `json:"mutating"`
			Verdict       Verdict            `json:"verdict"`
			Artifacts     []Artifact         `json:"artifacts,omitempty"`
		}{Artifacts: []Artifact{old}})
		if !bytes.Equal(gotEntry, wantEntry) {
			t.Fatalf("case %d: an entry carrying an unreconciled artifact serialises differently\n got %s\nwant %s", i, gotEntry, wantEntry)
		}
	}
}
