package engine

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestA087DormantBoardDoesNotAccumulateTokensOrAttempts(t *testing.T) {
	now := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	board := newExitLowerLimitBoard(clock.NewFake(now), nil, nil, nil, "acct", nil)
	board.cycleComplete(5 * time.Second)
	board.markProtectiveAttempt("kr", "005930", now)
	if got := len(board.tokens); got != 0 {
		t.Errorf("dormant tokens = %d, want 0", got)
	}
	if got := len(board.attempted); got != 0 {
		t.Errorf("dormant attempts = %d, want 0", got)
	}
}
