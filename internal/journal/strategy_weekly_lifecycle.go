package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func applyWeeklyReservationLifecycleInTx(ctx context.Context, tx *sql.Tx, fill AppliedFill) error {
	if tx == nil || strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY" {
		return nil
	}
	cumulative, err := campaignQuantity(fill.CumulativeQuantity)
	if err != nil {
		return err
	}
	positive, err := riskcalc.CompareDecimal(cumulative, "0")
	if err != nil {
		return err
	}
	toStatus := ""
	if positive > 0 {
		toStatus = WeeklyReservationConsumed
	} else if fill.Terminal {
		toStatus = WeeklyReservationReleased
	} else {
		return nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT w.reservation_id,w.campaign_id,w.market,r.status,s.version,s.positive_leg_count
		FROM strategy_weekly_first_leg_bindings w
		JOIN strategy_weekly_market_reservations r ON r.reservation_id=w.reservation_id
		JOIN strategy_weekly_reservation_scopes s ON s.campaign_id=w.campaign_id AND s.market=w.market
		JOIN strategy_dispatch_leases l ON l.guardian_decision_id=w.decision_id
		WHERE l.broker_order_id=? AND l.account_ref=? AND l.market=upper(?) AND l.symbol=upper(?)`,
		strings.TrimSpace(fill.OrderID), strings.TrimSpace(fill.AccountRef), strings.TrimSpace(fill.Market), strings.TrimSpace(fill.Symbol))
	if err != nil {
		return err
	}
	defer rows.Close()
	type binding struct {
		reservation, campaign, market, status string
		version, positive                     uint64
	}
	var found []binding
	for rows.Next() {
		var item binding
		if err := rows.Scan(&item.reservation, &item.campaign, &item.market, &item.status, &item.version, &item.positive); err != nil {
			return err
		}
		found = append(found, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(found) == 0 {
		return nil
	}
	if len(found) != 1 {
		return ErrWeeklyReservationConflict
	}
	item := found[0]
	if item.status != WeeklyReservationActive {
		if item.status == WeeklyReservationConsumed || item.status == WeeklyReservationReleased {
			return nil
		}
		return ErrWeeklyReservationConflict
	}
	if toStatus == WeeklyReservationConsumed && item.positive >= 7 {
		return ErrWeeklyReservationConflict
	}
	nextVersion := item.version + 1
	stamp := strings.TrimSpace(fill.CommittedAt)
	result, err := tx.ExecContext(ctx, `UPDATE strategy_weekly_market_reservations SET status=?,updated_at=?
		WHERE reservation_id=? AND status='ACTIVE'`, toStatus, stamp, item.reservation)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return ErrWeeklyReservationVersionConflict
	}
	positiveIncrement := uint64(0)
	if toStatus == WeeklyReservationConsumed {
		positiveIncrement = 1
	}
	result, err = tx.ExecContext(ctx, `UPDATE strategy_weekly_reservation_scopes
		SET version=?,positive_leg_count=positive_leg_count+?,updated_at=?
		WHERE campaign_id=? AND market=? AND version=? AND positive_leg_count=?`, nextVersion, positiveIncrement,
		stamp, item.campaign, item.market, item.version, item.positive)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return ErrWeeklyReservationVersionConflict
	}
	eventID := fillObservationID(fill)
	digestBytes := sha256.Sum256([]byte(strings.Join([]string{"strategy-weekly-reservation-lifecycle:v1", eventID,
		item.reservation, WeeklyReservationActive, toStatus, cumulative, stamp}, "\x00")))
	_, err = tx.ExecContext(ctx, `INSERT INTO strategy_weekly_reservation_lifecycle_receipts
		(event_id,reservation_id,from_status,to_status,scope_version,cumulative_quantity,record_digest,observed_at)
		VALUES(?,?,?,?,?,?,?,?)`, eventID, item.reservation, WeeklyReservationActive, toStatus, nextVersion,
		cumulative, hex.EncodeToString(digestBytes[:]), stamp)
	if err != nil {
		return err
	}
	return nil
}
