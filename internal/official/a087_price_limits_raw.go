package official

import (
	"context"
	"net/url"
)

// ProtectivePriceLimits 는 보호 청산 전용으로 하한가 원문과 시각을 보존한 읽기 결과임.
// 기존 PriceLimits 적응 결과는 CLI 계약을 지키기 위해 바꾸지 않음.
type ProtectivePriceLimits struct {
	LowerLimitPrice *string
	Timestamp       *string
}

type protectivePriceLimitsResponse struct {
	LowerLimitPrice *string `json:"lowerLimitPrice"`
	Timestamp       *string `json:"timestamp"`
}

// ProtectivePriceLimits 는 보호 청산 판 갱신기가 쓰는 원문 상하한가 읽기임.
func (c *Client) ProtectivePriceLimits(ctx context.Context, symbol string) (ProtectivePriceLimits, error) {
	q := url.Values{}
	q.Set("symbol", symbol)
	var raw protectivePriceLimitsResponse
	if err := c.get(ctx, "/api/v1/price-limits", q, &raw); err != nil {
		return ProtectivePriceLimits{}, err
	}
	return ProtectivePriceLimits{LowerLimitPrice: raw.LowerLimitPrice, Timestamp: raw.Timestamp}, nil
}
