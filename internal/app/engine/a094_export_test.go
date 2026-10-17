package engine

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// CatchUpExitProposalsForTest 는 기동 따라잡기를 부름(critical 은 nil 가능). TESTS ONLY.
func CatchUpExitProposalsForTest(ctx context.Context, j *journal.Journal, accountRef string, critical CriticalRecorder) int {
	return catchUpExitProposals(ctx, j, accountRef, critical, nil)
}
