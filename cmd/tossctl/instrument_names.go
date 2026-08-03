package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/ratebudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

func newConsoleInstrumentNames(shared *consoleBroker) consoleInstrumentNameReader {
	return &lazyInstrumentNames{shared: shared}
}

// newConsoleInstrumentNamesWithBudget is the test seam for an isolated lease
// path. Production always derives the path from shared.root and cannot disable
// cross-process verification exclusion with a constructor argument.
func newConsoleInstrumentNamesWithBudget(shared *consoleBroker, budgetPath string) consoleInstrumentNameReader {
	return &lazyInstrumentNames{shared: shared, budgetPath: budgetPath, budgetConfigured: true}
}

type lazyInstrumentNames struct {
	shared           *consoleBroker
	budgetPath       string
	budgetConfigured bool
}

type consoleInstrumentMetadata interface {
	Stocks(context.Context, []string) ([]domain.Quote, error)
}

const consoleInstrumentBatchLimit = 200

func (l *lazyInstrumentNames) Names(ctx context.Context, refs []consoleInstrumentRef) ([]consoleInstrumentName, error) {
	budgetPath := strings.TrimSpace(l.budgetPath)
	if !l.budgetConfigured {
		var err error
		budgetPath, err = verifyRateBudgetPath(l.shared.root)
		if err != nil {
			return nil, err
		}
	}
	if budgetPath != "" {
		lease, ok, err := ratebudget.TryAcquire(budgetPath)
		if err != nil {
			return nil, fmt.Errorf("console: reserving Open API metadata budget: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("console: live verification owns the Open API rate budget")
		}
		defer lease.Release()
		runLockPath := filepath.Join(filepath.Dir(budgetPath), runlock.FileName)
		if fresh, _ := runlock.Fresh(runLockPath, time.Now(), runlock.StaleAfter); fresh {
			return nil, fmt.Errorf("console: live verification started before the metadata lease; symbol-only history retained")
		}
	}
	reader, err := l.shared.instrumentMetadata(ctx)
	if err != nil {
		return nil, err
	}

	requested := make(map[string]bool, len(refs))
	symbols := make([]string, 0, len(refs))
	seenSymbols := make(map[string]bool, len(refs))
	for _, ref := range refs {
		market := strings.ToLower(strings.TrimSpace(ref.Market))
		symbol := strings.ToUpper(strings.TrimSpace(ref.Symbol))
		if (market != "kr" && market != "us") || symbol == "" {
			continue
		}
		requested[market+"|"+symbol] = true
		if !seenSymbols[symbol] {
			seenSymbols[symbol] = true
			symbols = append(symbols, symbol)
		}
	}

	out := make([]consoleInstrumentName, 0, len(symbols))
	for start := 0; start < len(symbols); start += consoleInstrumentBatchLimit {
		end := start + consoleInstrumentBatchLimit
		if end > len(symbols) {
			end = len(symbols)
		}
		quotes, err := reader.Stocks(ctx, symbols[start:end])
		if err != nil {
			return nil, err
		}
		for _, quote := range quotes {
			market := consoleQuoteMarket(quote)
			symbol := strings.ToUpper(strings.TrimSpace(quote.Symbol))
			if !requested[market+"|"+symbol] {
				continue
			}
			out = append(out, consoleInstrumentName{Market: market, Symbol: symbol, Name: quote.Name})
		}
	}
	return out, nil
}

func buildConsoleAccountBroker(ctx context.Context, root *rootOptions) (verifylive.Broker, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return nil, "", err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return nil, "", fmt.Errorf("console: reading Open API credentials for instrument names: %w", err)
	}
	if creds == nil {
		return nil, "", fmt.Errorf("console: no Open API credentials for instrument names")
	}
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	client := official.New(*creds, tokenFile)
	accountRef, _, err := resolveVerifyAccount(ctx, client, sleepFor)
	if err != nil {
		return nil, "", err
	}
	return client, accountRef, nil
}

func consoleQuoteMarket(quote domain.Quote) string {
	marketCode := ""
	switch strings.ToUpper(strings.TrimSpace(quote.MarketCode)) {
	case "KOSPI", "KOSDAQ", "KR_ETC":
		marketCode = "kr"
	case "NYSE", "NASDAQ", "AMEX", "US_ETC":
		marketCode = "us"
	}
	currency := ""
	switch strings.ToUpper(strings.TrimSpace(quote.Currency)) {
	case "KRW":
		currency = "kr"
	case "USD":
		currency = "us"
	}
	if marketCode != "" && currency != "" && marketCode != currency {
		return ""
	}
	if marketCode != "" {
		return marketCode
	}
	return currency
}
