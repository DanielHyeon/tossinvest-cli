package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func A087EnginePreviewForTest(c *Context, intent orderintent.PlaceIntent) trading.Preview {
	return c.tradingService.PreviewPlace(intent)
}
