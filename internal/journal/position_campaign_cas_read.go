package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type PositionCampaignCASRead struct {
	Generation int64
	Version    int64
	State      string
	Claimed    bool
}

// CurrentPositionCampaignCAS reads the exact generation/version and active
// prospective-claim facts that the atomic first-leg transaction will recheck.
// It is read-only and grants no campaign, token, owner, lease or order authority.
func (j *Journal) CurrentPositionCampaignCAS(ctx context.Context, accountRef, market, symbol string) (PositionCampaignCASRead, error) {
	if j == nil || ctx == nil || strings.TrimSpace(accountRef) == "" || strings.TrimSpace(market) == "" || strings.TrimSpace(symbol) == "" {
		return PositionCampaignCASRead{}, ErrInvalidRequest
	}
	read := PositionCampaignCASRead{State: "FLAT"}
	var version sql.NullInt64
	err := j.db.QueryRowContext(ctx, `SELECT p.instance_seq,p.state,v.version
		FROM positions p LEFT JOIN position_projection_versions v ON v.position_id=p.id
		WHERE p.account_ref=? AND p.market=? AND p.symbol=? ORDER BY p.instance_seq DESC LIMIT 1`,
		strings.TrimSpace(accountRef), normaliseMarket(market), normaliseSymbol(symbol)).Scan(&read.Generation, &read.State, &version)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		read.Generation, read.Version, read.State = 0, 0, "FLAT"
	case err != nil:
		return PositionCampaignCASRead{}, fmt.Errorf("journal: reading first-leg position CAS: %w", err)
	case !version.Valid:
		return PositionCampaignCASRead{}, fmt.Errorf("%w: position generation predates authoritative versioning", ErrGenerationConflict)
	default:
		read.Version = version.Int64
	}
	var claim string
	err = j.db.QueryRowContext(ctx, `SELECT campaign_id FROM position_campaign_claims WHERE account_ref=? AND market=? AND symbol=?`,
		strings.TrimSpace(accountRef), normaliseMarket(market), normaliseSymbol(symbol)).Scan(&claim)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PositionCampaignCASRead{}, fmt.Errorf("journal: reading first-leg campaign claim: %w", err)
	}
	read.Claimed = err == nil
	return read, nil
}
