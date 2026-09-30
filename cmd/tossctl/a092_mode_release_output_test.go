package main

// a092 26라운드 codex 재확인 R2 · 보이스 B #6: 재조회 실패를 먼저 가르고, 읽지 못한 것은 추정하지 않음 — 「PENDING 목록에 없음」은
// 「이미 전달됨」이 아님(승인됐을 수도 있음). 통지 목록만 못 읽었으면 이미 읽은 모드 · 사유는 그대로 말함.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
)

func a092Render(r engine.ModeReleaseResult) string {
	var b strings.Builder
	_ = writeModeReleaseResult(&b, output.FormatTable, r)
	return b.String()
}

func a092NoticeLine(out string) string {
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "통지:") {
			return l
		}
	}
	return ""
}

func TestA092ModeReleaseOutputDoesNotGuessWhatItCouldNotRead(t *testing.T) {
	t.Run("mode-reread-failed", func(t *testing.T) {
		out := a092Render(engine.ModeReleaseResult{Changed: true, TransitionID: "t1", Notified: true,
			ReReadError: "re-reading the operating mode after the release failed"})
		if notice := a092NoticeLine(out); !strings.Contains(notice, "확인하지 못") {
			t.Errorf("the notice state was guessed while nothing was re-read: %q\n%s", notice, out)
		}
		if !strings.Contains(out, "재조회 실패") {
			t.Errorf("output does not say the state is unread:\n%s", out)
		}
	})
	t.Run("notice-list-failed", func(t *testing.T) {
		out := a092Render(engine.ModeReleaseResult{Changed: true, TransitionID: "t1", Mode: "NORMAL", Notified: true,
			EntryBlocks: []string{"ALERT_UNDELIVERED"}, NoticeReadError: "re-reading the release notice failed"})
		if !strings.Contains(out, "현재 모드 NORMAL") || !strings.Contains(out, "ALERT_UNDELIVERED") {
			t.Errorf("the mode and reasons that were read are missing:\n%s", out)
		}
		if strings.Contains(out, "현재 모드와 남은 사유를 읽지 못했다") {
			t.Errorf("the read mode was reported as unread:\n%s", out)
		}
		if notice := a092NoticeLine(out); !strings.Contains(notice, "확인하지 못") {
			t.Errorf("the notice state was not marked unconfirmed: %q\n%s", notice, out)
		}
	})
	// 26라운드 codex 2차 재확인 P2: 통지 기록 실패와 통지 목록 재조회 실패가 함께면 둘 다 보임.
	t.Run("notify-and-notice-list-failed", func(t *testing.T) {
		out := a092Render(engine.ModeReleaseResult{Changed: true, TransitionID: "t1", Mode: "NORMAL",
			NotifyError: "the release notice could not be recorded", NoticeReadError: "re-reading the release notice failed"})
		if !strings.Contains(out, "the release notice could not be recorded") || !strings.Contains(out, "re-reading the release notice failed") {
			t.Errorf("one of two failures disappeared:\n%s", out)
		}
	})
	t.Run("not-pending", func(t *testing.T) {
		out := a092Render(engine.ModeReleaseResult{Changed: true, TransitionID: "t1", Mode: "NORMAL", Notified: true})
		notice := a092NoticeLine(out)
		if strings.Contains(notice, "이미 전달 처리됨") || !strings.Contains(notice, "승인") {
			t.Errorf("not pending must say it may have been acknowledged, not that it was delivered: %q", notice)
		}
	})
}
