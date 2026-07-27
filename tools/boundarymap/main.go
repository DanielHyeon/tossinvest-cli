package main

import (
	"fmt"
	"os"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/measure"
)

func main() {
	model := costs.DefaultModel()
	now := time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC)

	kr, err := measure.Grid(model, measure.DefaultKRGrid(), now)
	if err != nil {
		panic(err)
	}
	us, err := measure.Grid(model, measure.DefaultUSGrid(), now)
	if err != nil {
		panic(err)
	}
	entries, source := measure.StockOS058Entries()
	pop, err := measure.Population(model, entries, source, now)
	if err != nil {
		panic(err)
	}
	report := measure.Report{
		CostModelFingerprint: model.Fingerprint(),
		CostScope:            "FEE_TAX_ONLY",
		Markets:              []measure.MarketReport{kr, us},
		RealTradePopulation:  &pop,
	}
	if err := os.WriteFile(os.Args[1], []byte(report.Render()), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("wrote", os.Args[1])
}
