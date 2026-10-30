package journal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// ExitProposalForIntent는 정확한 intent의 모든 후보를 검사하여 유일한 주문 제안만 반환함.
func (j *Journal) ExitProposalForIntent(ctx context.Context, intent Intent) (string, bool, error) {
	if strings.TrimSpace(intent.ID) == "" || strings.TrimSpace(intent.AccountRef) == "" || normaliseMarket(intent.Market) == "" || normaliseSymbol(intent.Symbol) == "" {
		return "", false, nil
	}
	rows, err := j.db.QueryContext(ctx, `SELECT e.action, p.account_ref, p.market, p.symbol
 FROM exit_events e LEFT JOIN positions p ON p.id=e.position_id
 WHERE e.proposed_intent_id=? LIMIT 65`, intent.ID)
	if err != nil {
		return "", false, fmt.Errorf("journal: exit origin query: %w", err)
	}
	defer rows.Close()
	count := 0
	proposal := ""
	invalid := false
	for rows.Next() {
		count++
		var action, account, market, symbol sql.NullString
		if err = rows.Scan(&action, &account, &market, &symbol); err != nil {
			return "", false, fmt.Errorf("journal: exit origin scan: %w", err)
		}
		if count > 64 || !action.Valid || !account.Valid || !market.Valid || !symbol.Valid || account.String != intent.AccountRef || normaliseMarket(market.String) != normaliseMarket(intent.Market) || normaliseSymbol(symbol.String) != normaliseSymbol(intent.Symbol) {
			invalid = true
			continue
		}
		switch action.String {
		case ExitEventOpened, ExitEventReadopted, ExitEventProposalRefused, ExitEventProposalCancelled, ExitEventProposalFilled, ExitEventCompleted, ExitEventAdjustmentClosed:
			continue
		}
		if !exitpolicy.Action(action.String).Orderable() || proposal != "" {
			invalid = true
			continue
		}
		proposal = action.String
	}
	if err = rows.Err(); err != nil {
		return "", false, fmt.Errorf("journal: exit origin iteration: %w", err)
	}
	if invalid || proposal == "" {
		return "", false, nil
	}
	return proposal, true, nil
}

// HasExitEventForIntent는 진입·flatten 출처와 충돌하는 exit 원장 존재를 확인함.
func (j *Journal) HasExitEventForIntent(ctx context.Context, intentID string) (bool, error) {
	var exists int
	if err := j.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM exit_events WHERE proposed_intent_id=?)`, intentID).Scan(&exists); err != nil {
		return false, fmt.Errorf("journal: exit origin existence: %w", err)
	}
	return exists != 0, nil
}
