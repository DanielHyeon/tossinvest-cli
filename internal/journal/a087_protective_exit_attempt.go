package journal

import (
	"context"
	"fmt"
	"strings"
)

// HasProtectiveExitAttempt 는 같은 계좌·시장·거래일·종목의 보호 청산 의도 존재만 읽음.
// 상태와 가격은 읽지 않아 첫 시도 금지의 단조성을 해치지 않음.
func (j *Journal) HasProtectiveExitAttempt(ctx context.Context, accountRef, market, tradingDay, symbol string,
	protectiveActions []string) (bool, error) {
	if len(protectiveActions) == 0 {
		return false, fmt.Errorf("journal: protective exit actions are empty")
	}
	placeholders := make([]string, len(protectiveActions))
	args := make([]any, 0, 5+len(protectiveActions))
	args = append(args, strings.TrimSpace(accountRef), normaliseMarket(market), strings.TrimSpace(tradingDay), normaliseSymbol(symbol))
	for i, action := range protectiveActions {
		placeholders[i] = "?"
		args = append(args, strings.TrimSpace(action))
	}
	query := `SELECT EXISTS(
		SELECT 1
		  FROM intents i
		  JOIN exit_events e ON e.proposed_intent_id = i.id
		 WHERE i.account_ref = ? AND i.market = ? AND i.trading_day = ?
		   AND upper(trim(i.symbol)) = ? AND i.side = 'SELL'
		   AND e.action IN (` + strings.Join(placeholders, ",") + `)
	)`
	var exists int
	if err := j.db.QueryRowContext(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("journal: reading protective exit attempt: %w", err)
	}
	return exists != 0, nil
}
