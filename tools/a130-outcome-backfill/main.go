// Command a130-outcome-backfill 는 a130 의 빠진 성과 행 복구 도구임 — 체결로 닫힌 편입 포지션 중 성과 행(trade_outcomes)이 없는 것을
// a130 판정(묶음별 편입가 = 브로커 공식 1분봉 종가)으로 다시 계산해, 사람 확인 뒤에만 채움
// (openspec/changes/a130-an-adopted-trade-keeps-its-outcome design D4).
//
// ⚠ 실행(--execute)은 원장 쓰기(mutating)임. 에이전트와 자동 시험은 운영 원장에 실행하지 않음 — 사람이 터미널에서 돌림.
//
// 흐름(두 번 실행):
//
//	① 미리보기 — 원장 쓰기 0. 읽기 전용 연결로 늘어난 묶음 목록을 읽고, 묶음마다 공식 1분봉 하나를 GET(읽기 전용, 공유 토큰 캐시 —
//	   별도 토큰 발급 없음, a082) 해서 항목마다 판정 · 계산 결과 · 쓴 분봉 · 거절 사유와 확인 토큰을 출력함.
//	   go run ./tools/a130-outcome-backfill --journal ~/.config/tossctl/journal.db --config-dir ~/.config/tossctl
//	② 실행 — 사람이 출력을 확인하고 토큰을 붙여 다시 침. 원장 스키마가 이 빌드와 같은지 읽기 전용으로 먼저 보고(다르면 쓰기 연결을 열지
//	   않음 — 도구는 원장을 이관하지 않음), 분봉을 다시 읽어 계획을 다시 계산하고, 토큰이 같을 때만 「fill」 항목마다 짧은 쓰기
//	   트랜잭션 하나에서 다시 계산한 값이 미리 본 값과 같을 때만 씀. 항목마다 쓰기 잠금을 쥔 시간을 출력함.
//	   go run ./tools/a130-outcome-backfill --journal ~/.config/tossctl/journal.db --config-dir ~/.config/tossctl --execute --confirm <token>
//
// --config-dir 는 늘어난 묶음이 있으면 필수임 — 엔진은 `--config-dir`(compose 의 TOSSOS_CONFIG_DIR)로 돌고 토큰 캐시가 그 디렉터리에 있음.
// 빼면 apppaths 기본값(사용자 캐시 디렉터리)의 다른 토큰 파일을 써서 새 토큰을 발급받고, 같은 앱 키의 엔진 토큰을 무효화하는 토큰 전쟁(a082)이
// 남(리뷰 4차 P1-1, 2026-10-13 실측: 기본값으로 돌린 미리보기 직후 두 토큰 파일이 2초 차이로 다시 쓰임).
//
// 분봉 읽기 예산: 공식 candles 는 초당 5회 묶음(internal/candidate veto.go 주석)인데, 이 도구는 1.2초에 한 번만 부름 — 엔진 · 탐색의 같은
// 묶음 사용에 여유를 남김. 429 · 오류는 재시도하지 않고 도구를 멈춤(쓰기 전).
//
// 원장 경로는 기본값이 없음 — 운영 원장은 ~/.config/tossctl 에 있고 journal.DefaultPath 는 다른 디렉터리를 가리킬 수 있음.
// 비용 모델은 엔진과 같은 costs.DefaultModel(). 출력에 계좌 참조 · 포지션 id 는 없음. 실행 창: 두 시장이 닫힌 시간(KST 기준 시각은 서머타임에 따라 달라짐 — 운영 절차가 정함).
//
// 종료 코드: 0 = 미리보기 완료 또는 실행 완료, 1 = 도구 오류 · 토큰 불일치 · 스키마 불일치(쓰기 0).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// candlePace 는 분봉 GET 사이의 간격임(초당 5회 묶음 한도 아래).
const candlePace = 1200 * time.Millisecond

// candleReader 는 이름이 label 인 1분봉의 종가와 **돌려받은 봉의 이름**을 읽음(그 분의 봉이 없으면 ok=false).
type candleReader func(ctx context.Context, market, symbol string, label time.Time) (
	closePrice string, barLabel time.Time, ok bool, err error)

type deps struct {
	now    func() time.Time
	stdout io.Writer
	// candles 는 config 디렉터리의 공유 토큰 캐시로 생산 client 를 열고, 읽기 함수와 그 토큰 파일 경로를 돌려줌.
	candles func(configDir string) (candleReader, string, error)
	pace    time.Duration
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	code, err := run(ctx, os.Args[1:], deps{now: time.Now, stdout: os.Stdout, candles: openCandles, pace: candlePace})
	if err != nil {
		fmt.Fprintln(os.Stderr, "a130 outcome backfill: "+err.Error())
	}
	os.Exit(code)
}

// openCandles 는 a087 · a128 도구와 같은 방식으로 생산 client 를 만듦 — 같은 토큰 캐시 파일을 공유해야 엔진 · 콘솔의 토큰을 무효화하는
// 토큰 전쟁(a082)이 나지 않음.
func openCandles(configDir string) (candleReader, string, error) {
	credentialFile, tokenFile, err := apppaths.OpenAPI(configDir)
	if err != nil {
		return nil, "", fmt.Errorf("resolving the Open API paths: %w", err)
	}
	credentials, err := official.LoadCredentials(os.Getenv, credentialFile)
	if err != nil {
		return nil, "", fmt.Errorf("reading the Open API credentials: %w", err)
	}
	if credentials == nil {
		return nil, "", errors.New("no Open API credentials; run `tossctl openapi login` first")
	}
	return official.New(*credentials, tokenFile).MinuteCandleClosingAt, tokenFile, nil
}

type options struct {
	journal   string
	configDir string
	execute   bool
	confirm   string
}

func parseOptions(args []string) (options, error) {
	var o options
	flags := flag.NewFlagSet("a130-outcome-backfill", flag.ContinueOnError)
	flags.StringVar(&o.journal, "journal", "", "required: the journal file (production: ~/.config/tossctl/journal.db)")
	flags.StringVar(&o.configDir, "config-dir", "", "optional: the tossctl config directory (Open API credentials · token cache)")
	flags.BoolVar(&o.execute, "execute", false, "write the planned rows (needs --confirm from the preview)")
	flags.StringVar(&o.confirm, "confirm", "", "the confirmation token the preview printed")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments %v", flags.Args())
	}
	if o.journal == "" {
		return options{}, errors.New("--journal <path> is required")
	}
	if o.confirm != "" && !o.execute {
		return options{}, errors.New("--confirm is only meaningful with --execute")
	}
	if o.execute && o.confirm == "" {
		return options{}, errors.New("--execute needs --confirm <token>; run without --execute first to get one")
	}
	return o, nil
}

func run(ctx context.Context, args []string, d deps) (int, error) {
	o, err := parseOptions(args)
	if err != nil {
		return 1, err
	}
	// 원장이 없는 경로에 새 원장을 만들지 않게 먼저 확인함.
	if info, err := os.Stat(o.journal); err != nil || info.IsDir() {
		return 1, fmt.Errorf("no journal file at %s", o.journal)
	}
	model := costs.DefaultModel()
	at := d.now().UTC()

	ro, err := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: o.journal})
	if err != nil {
		return 1, fmt.Errorf("opening the journal read-only: %w", err)
	}
	if o.execute {
		// journal.Open 은 스키마를 앞으로 옮김(migrate). 실행 중인 엔진보다 새 빌드가 운영 원장을 옮기면 그 엔진이 원장을 거부하므로,
		// 쓰기 연결은 원장의 스키마가 이 빌드와 **같을 때만** 엶 — 순서: 읽기 전용 열기 → 버전 비교 → (같으면) 쓰기 열기.
		if err := schemaGuard(ro.SchemaVersion(), journal.SchemaVersion); err != nil {
			_ = ro.Close()
			return 1, err
		}
	}
	lots, err := journal.IntakeLotsForBackfill(ctx, ro)
	if err != nil {
		_ = ro.Close()
		return 1, err
	}
	prices, err := fetchPrices(ctx, d, o.configDir, lots)
	if err != nil {
		_ = ro.Close()
		return 1, err
	}
	if !o.execute {
		defer ro.Close()
		plan, err := journal.PreviewTradeOutcomeBackfill(ctx, ro, model, at, prices)
		if err != nil {
			return 1, err
		}
		writePlan(d.stdout, plan, lots, prices)
		return 0, nil
	}
	_ = ro.Close()
	j, err := journal.Open(ctx, journal.Options{Path: o.journal, Clock: clock.System()})
	if err != nil {
		return 1, fmt.Errorf("opening the journal: %w", err)
	}
	defer j.Close()
	written, err := j.BackfillTradeOutcomes(ctx, model, at, prices, o.confirm)
	for _, w := range written {
		fmt.Fprintf(d.stdout, "written: %s %s #%d (write lock held %s)\n", w.Market, w.Symbol, w.InstanceSeq,
			w.LockHeld.Round(time.Microsecond))
	}
	fmt.Fprintf(d.stdout, "written: %d\n", len(written))
	if err != nil {
		return 1, err
	}
	return 0, nil
}

// fetchPrices 는 묶음마다 공식 1분봉 하나를 읽음(묶음이 없으면 부르지 않음). 그 분의 봉이 없으면 그 묶음은 빠짐(→ no_intake_candle).
func fetchPrices(ctx context.Context, d deps, configDir string, lots []journal.IntakeLot) (journal.IntakePrices, error) {
	prices := journal.IntakePrices{}
	if len(lots) == 0 {
		return prices, nil
	}
	if err := requireConfigDir(configDir); err != nil {
		return nil, err
	}
	read, tokenFile, err := d.candles(configDir)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(d.stdout, "token cache: %s\n", tokenFile)
	for i, lot := range lots {
		if i > 0 && d.pace > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(d.pace):
			}
		}
		closePrice, barLabel, ok, err := read(ctx, lot.Market, lot.Symbol, lot.BarLabel())
		if err != nil {
			return nil, fmt.Errorf("reading the %s 1m bar %s: %w", lot.Symbol, lot.BarLabel().Format(time.RFC3339), err)
		}
		if ok {
			// 돌려받은 이름을 그대로 적음 — 원장 판정(priceLots)이 그것을 묶음의 분과 다시 대조함(두 번째 가드).
			prices[lot.Key()] = journal.IntakeCandle{Label: barLabel, Close: closePrice,
				Source: journal.IntakeSourceBrokerCandle1m}
		}
	}
	return prices, nil
}

// requireConfigDir 는 분봉을 읽어야 할 때 엔진의 config 디렉터리(공유 토큰 캐시)를 요구함(a082 — 기본값은 다른 토큰 파일).
func requireConfigDir(configDir string) error {
	if strings.TrimSpace(configDir) == "" {
		return errors.New("--config-dir <the engine's config directory, e.g. ~/.config/tossctl> is required to read 1m bars: " +
			"the default resolves a different token cache and would mint a token that invalidates the engine's (a082)")
	}
	return nil
}

func schemaGuard(found, want int) error {
	if found != want {
		return fmt.Errorf("the journal is at schema %d and this build is %d; the tool never migrates a journal — "+
			"run it from the build the running engine uses (after deploying)", found, want)
	}
	return nil
}

// writePlan 은 계획을 사람이 읽는 표로 씀 — 계좌 참조 · 포지션 id 칸은 없음.
func writePlan(w io.Writer, plan journal.TradeOutcomeBackfillPlan, lots []journal.IntakeLot, prices journal.IntakePrices) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "ACTION\tMARKET\tSYMBOL\t#\tCLOSED_AT\tQTY\tREALIZED_R\tPNL_AFTER_COSTS\tINTAKE_BARS\tREASON")
	counts := map[string]int{}
	for _, it := range plan.Items {
		counts[it.Action]++
		qty, r, pnl := "-", "-", "-"
		if it.Action != journal.BackfillActionRefused {
			qty, r, pnl = it.Outcome.InitialQuantity, it.Outcome.RealizedR, it.Outcome.RealizedPnLAfterCosts
		}
		bars := "-"
		if len(it.IntakeCandles) > 0 {
			parts := make([]string, len(it.IntakeCandles))
			for i, c := range it.IntakeCandles {
				parts[i] = c.Label.UTC().Format(time.RFC3339) + "@" + c.Close
			}
			bars = strings.Join(parts, ",")
		}
		reason := string(it.Reason)
		if reason == "" {
			reason = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n", it.Action, it.Market, it.Symbol,
			it.InstanceSeq, it.ClosedAt, qty, r, pnl, bars, reason)
	}
	_ = tw.Flush()
	missing := 0
	for _, lot := range lots {
		if _, ok := prices[lot.Key()]; !ok {
			missing++
			fmt.Fprintf(w, "no 1m bar: %s %s bar %s (adjustment %s)\n", lot.Market, lot.Symbol,
				lot.BarLabel().Format(time.RFC3339), lot.CreatedAt.UTC().Format(time.RFC3339))
		}
	}
	fmt.Fprintf(w, "\npopulation (fill-closed adopted positions without a row): %d · fill: %d · refused: %d · outside retention: %d\n",
		len(plan.Items), counts[journal.BackfillActionFill], counts[journal.BackfillActionRefused],
		counts[journal.BackfillActionOutsideRetention])
	fmt.Fprintf(w, "intake lots: %d · 1m bars found: %d · missing: %d\n", len(lots), len(lots)-missing, missing)
	fmt.Fprintf(w, "confirmation token: %s\n", plan.Token)
	fmt.Fprintln(w, "nothing was written. To write the fill rows, re-run with --execute --confirm <token>.")
}
