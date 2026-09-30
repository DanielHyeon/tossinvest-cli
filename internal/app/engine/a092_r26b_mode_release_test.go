package engine

// a092 26라운드 2차 수리(Manager 판정 2026-09-30) — 모드 완화 표면:
//   - 보이스 A #1: 커밋 뒤의 통지 · 재읽기는 요청 ctx 에 볼모잡히지 않음(a066 notifyRelaxation 선례).
//   - codex 재확인 R2: 통지 목록 재읽기 실패는 모드 재읽기 실패와 따로 — 이미 읽은 모드 · 사유는 버리지 않음.
//   - 보이스 B #2(불변식 8): 결과 · 오류 본문에 원문 오류(계좌 담긴 키 · 원장 문구)를 싣지 않음 — 고정 문구.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a092CancelingAnnouncer 는 커밋 뒤 통지가 시작되는 순간 요청 ctx 를 끝냄 — 클라이언트 Ctrl-C · 타임아웃이 그 창에 떨어진 경우.
type a092CancelingAnnouncer struct {
	cancel context.CancelFunc
	inner  journal.ModeAnnouncer
}

func (a a092CancelingAnnouncer) AnnounceOperatingMode(ctx context.Context, previous string, rec journal.OperatingModeRecord) error {
	a.cancel()
	return a.inner.AnnounceOperatingMode(ctx, previous, rec)
}

// A#1: 요청 ctx 가 커밋 뒤에 끝나도 통지 행이 남고, 재읽기가 성공하고, 전달 실패 래치가 서지 않음.
func TestA092ACommittedReleaseIsNotHostageToTheRequestContext(t *testing.T) {
	fx := a092ReleaseEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fx.ops.announcer = a092CancelingAnnouncer{cancel: cancel, inner: obs.RecordOnly{N: fx.n}}
	res, err := fx.ops.Release(ctx, a092ReleaseRequest())
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if !res.Changed || !res.Notified || res.NotifyError != "" || res.ReReadError != "" || res.NoticeReadError != "" {
		t.Fatalf("result = %+v, want the notice recorded and the state re-read despite the ended request", res)
	}
	if res.Mode != journal.ModeNormal || !res.NoticePending {
		t.Errorf("mode=%s pending=%v, want NORMAL and a pending notice row", res.Mode, res.NoticePending)
	}
	if _, latched := fx.gate.Blocks()[execgw.ReasonAlertUndelivered]; latched {
		t.Error("the release notice failed on the request's dead context and latched — an ack would erase its only trace")
	}
}

// R2: 통지 목록만 못 읽으면 그 사실만 — 모드 · 남은 사유는 읽은 대로 둠.
func TestA092ANoticeListFailureKeepsTheModeThatWasRead(t *testing.T) {
	fx := a092ReleaseEngine(t)
	fx.ops.pending = func(context.Context, int) ([]journal.Alert, error) {
		return nil, errors.New("outbox read failed")
	}
	res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if res.ReReadError != "" {
		t.Errorf("ReReadError = %q — the mode was read; only the notice list failed", res.ReReadError)
	}
	if res.NoticeReadError == "" || res.NoticePending {
		t.Errorf("notice read error = %q pending = %v, want the notice list failure named and no pending guess",
			res.NoticeReadError, res.NoticePending)
	}
	if res.Mode != journal.ModeNormal || res.Seq <= 0 {
		t.Errorf("mode = %q seq = %d, want the re-read NORMAL kept", res.Mode, res.Seq)
	}
}

// B#2: 통지 기록 실패 · 재읽기 실패의 결과 칸에 계좌 · 원문 오류가 없음.
func TestA092ReleaseResultsCarryNoRawLedgerText(t *testing.T) {
	t.Run("notify", func(t *testing.T) {
		fx := a092ReleaseEngine(t)
		broken, err := journal.Open(context.Background(), journal.Options{
			Path:     filepath.Join(t.TempDir(), journal.DBFileName),
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
		})
		if err != nil {
			t.Fatal(err)
		}
		_ = broken.Close()
		fx.ops.announcer = obs.RecordOnly{N: &obs.Notifier{Journal: broken, Gate: fx.gate}}
		res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
		if err != nil {
			t.Fatal(err)
		}
		if res.NotifyError == "" {
			t.Fatal("arrangement: the notice must fail")
		}
		if strings.Contains(res.NotifyError, a092Account) || strings.Contains(res.NotifyError, "database is closed") {
			t.Errorf("NotifyError carries raw ledger text: %q", res.NotifyError)
		}
		if d := fx.gate.Blocks()[execgw.ReasonAlertUndelivered]; strings.Contains(d, a092Account) {
			t.Errorf("gate description carries the account: %q", d)
		}
	})
	t.Run("reread", func(t *testing.T) {
		fx := a092ReleaseEngine(t)
		fx.ops.current = func(context.Context, string) (journal.ModeSnapshot, error) {
			return journal.ModeSnapshot{}, errors.New("reading the mode for " + a092Account + " failed")
		}
		res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
		if err != nil {
			t.Fatal(err)
		}
		if res.ReReadError == "" || strings.Contains(res.ReReadError, a092Account) {
			t.Errorf("ReReadError = %q, want a fixed wording without the account", res.ReReadError)
		}
	})
	t.Run("notice-list", func(t *testing.T) {
		fx := a092ReleaseEngine(t)
		fx.ops.pending = func(context.Context, int) ([]journal.Alert, error) {
			return nil, errors.New("reading alerts for " + a092Account + " failed")
		}
		res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
		if err != nil {
			t.Fatal(err)
		}
		if res.NoticeReadError == "" || strings.Contains(res.NoticeReadError, a092Account) {
			t.Errorf("NoticeReadError = %q, want a fixed wording without the account", res.NoticeReadError)
		}
	})
}
