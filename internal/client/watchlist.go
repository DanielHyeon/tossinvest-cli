package client

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func (c *Client) ListWatchlist(ctx context.Context) ([]domain.WatchlistItem, error) {
	groups, err := c.ListWatchlistGroups(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]domain.WatchlistItem, 0)
	for _, group := range groups {
		items = append(items, group.Items...)
	}
	return items, nil
}
