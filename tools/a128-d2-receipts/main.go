// Command a128-d2-receipts 는 a128 사전등록 D2 의 공식 1분봉 영수증을 읽기 전용으로 만듦.
//
//	go run ./tools/a128-d2-receipts --market KR --date 2026-10-23 --config-dir ~/.config/tossctl
//	go run ./tools/a128-d2-receipts --market KR --date 2026-10-23 --dry-run   # 목록·요청 계획만, 네트워크 없음
//
// 출력: <out>/minute-bars/<MARKET>-<YYYY-MM-DD>.json (기본 out = a128 change 의 analysis/d2-receipts).
// --config-dir 는 필수이고 엔진·a128 타이머와 **같은** ~/.config/tossctl 을 줌 — 같은 토큰 캐시를 함께 써야 함.
// 빠지면 기본 캐시(~/.cache/tossctl)로 새 토큰을 따로 받아 엔진 토큰을 무효화함(a082 토큰 전쟁, 2026-10-19 실측).
// 토큰이 만료돼 있으면 공식 클라이언트가 POST /token 으로 새로 받음 — 이 도구의 유일한 비-GET 요청(주문 아님).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	changeAnalysis = "openspec/changes/a128-jev-shadow-judgment-probe/analysis"
	preregDoc      = changeAnalysis + "/prereg-2026-10-23-d2-twice-daily-index-correlation.md"
	runBudget      = 5 * time.Minute
)

func main() {
	market := flag.String("market", "", `required: "KR" or "US"`)
	date := flag.String("date", "", "required: session date YYYY-MM-DD (UTC date of the judgments = KST date for KR, open day for US)")
	configDir := flag.String("config-dir", "", "required unless --dry-run: the tossctl config directory")
	ledger := flag.String("ledger", changeAnalysis+"/shadow-ledger/ledger.jsonl", "a128 ledger (read only)")
	out := flag.String("out", changeAnalysis+"/d2-receipts", "receipt root")
	gap := flag.Duration("gap", 300*time.Millisecond, "pause between calls")
	maxSymbols := flag.Int("max-symbols", 0, "only the first N listed symbols (0 = all; for retention checks)")
	dryRun := flag.Bool("dry-run", false, "print the list and the call plan, no network")
	flag.Parse()
	if err := run(*market, *date, *configDir, *ledger, *out, *gap, *maxSymbols, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, "a128-d2-receipts:", err)
		if errors.Is(err, errRateLimited) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(market, date, configDir, ledgerPath, out string, gap time.Duration, maxSymbols int, dryRun bool) error {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("--date: %w", err)
	}
	plans, err := planFor(market, day)
	if err != nil {
		return err
	}
	f, err := os.Open(ledgerPath)
	if err != nil {
		return err
	}
	symbols, digest, err := listSymbols(f, market, date)
	f.Close()
	if err != nil {
		return err
	}
	if maxSymbols > 0 && len(symbols) > maxSymbols {
		symbols = symbols[:maxSymbols]
	}
	if len(symbols) == 0 {
		return fmt.Errorf("no %s 10:00-round J2 judgments on %s in the ledger; nothing to fetch", market, date)
	}
	if dryRun {
		view := map[string]any{"market": market, "date": date, "symbols": symbols, "calls": len(symbols) * len(plans), "plans": plans}
		body, _ := json.MarshalIndent(view, "", "  ")
		fmt.Println(string(body))
		return nil
	}
	if configDir == "" {
		return errors.New("--config-dir is required (the default token cache would invalidate the engine's token)")
	}
	credentialFile, tokenFile, err := apppaths.OpenAPI(configDir)
	if err != nil {
		return fmt.Errorf("resolving the Open API paths: %w", err)
	}
	credentials, err := official.LoadCredentials(os.Getenv, credentialFile)
	if err != nil {
		return fmt.Errorf("reading the Open API credentials: %w", err)
	}
	if credentials == nil {
		return errors.New("no Open API credentials; run `tossctl openapi login` first")
	}
	client := official.New(*credentials, tokenFile)
	ctx, cancel := context.WithTimeout(context.Background(), runBudget)
	defer cancel()
	rc, err := collect(ctx, client, market, date, symbols, plans, gap, sleepCtx, time.Now)
	if err != nil {
		return err
	}
	rc.Prereg, rc.LedgerDigest = preregDoc, digest
	rc.ListRule = "J2 judgments with session date " + date + " and local time 10:00:00 <= at < 10:10:00"
	path, err := writeReceipt(out, rc)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s: symbols %d, calls %d, elapsed %.1fs, symbols with bars %d\n",
		path, len(rc.Symbols), rc.Calls, rc.ElapsedSeconds, len(rc.Bars))
	return nil
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
