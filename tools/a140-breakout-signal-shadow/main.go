// Command a140-breakout-signal-shadow 는 a140 그림자 측정 도구 — 생산 배선 0, 주문 0, Jev 0.
//
// 무엇을 하는가 (openspec/changes/a140-breakout-signals-are-measured-in-shadow)
//
// KR 정규장 동안 세션 시작의 거래대금 랭킹 상위 N(1 주 가격 ≤ 주문 상한)을 고정하고, 종목마다 1 분에 한 번 공식 닫힌 1분봉을
// **도구 전용** evidence 저장소에 적재해 breakout v1 동결 평가기로 단계를 굴림. ARMED 에 닿으면 최대 5 분 새 호가로 다시 판정해
// 「냈을 제안」(가상 진입 · 손절 · 목표 · 수량)을 원장에 적고, 같은 저장소의 뒤 봉으로 결과를 라벨함.
//
// 하지 않는 것: 주문 · 정정 · 취소 · 토글 · 활성화 · 설정 쓰기 · 엔진 evidence.db · 저널 쓰기 · 서명 매니페스트 · 외부 판단 API.
// 쓰는 것은 셋뿐 — 도구 전용 evidence 저장소, 원장, 공유 Open API 토큰 캐시 갱신(엔진과의 토큰 전쟁 방지 — a128 선례).
//
//	go run ./tools/a140-breakout-signal-shadow run --evidence-db ~/.local/state/tossos-a140/shadow-evidence.sqlite
//	go run ./tools/a140-breakout-signal-shadow report
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata" // 시장 현지 시각 해석이 tz DB 없는 환경에서도 맞도록

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	tossconfig "github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

const defaultLedgerDir = "openspec/changes/a140-breakout-signals-are-measured-in-shadow/analysis/shadow-ledger"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := dispatch(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "a140 breakout signal shadow: "+err.Error())
		os.Exit(1)
	}
}

func dispatch(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New(`a subcommand is required: "run" or "report"`)
	}
	switch args[0] {
	case "run":
		return runCommand(ctx, args[1:])
	case "report":
		return reportCommand(args[1:], stdout)
	default:
		return fmt.Errorf("unknown subcommand %q; use \"run\" or \"report\"", args[0])
	}
}

// runCommand 는 브로커를 부르기 전에 경로 가드 · 설정 · 플래그를 전부 검사함(어느 하나라도 틀리면 GET 0 으로 거부).
func runCommand(ctx context.Context, args []string) error {
	cfg := defaultProbeConfig()
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	market := flags.String("market", "KR", `only "KR" (design D1)`)
	evidenceDB := flags.String("evidence-db", "", "required: the tool's own evidence store (never the engine's evidence.db)")
	out := flags.String("out", defaultLedgerDir, "ledger directory")
	configDir := flags.String("config-dir", "", "optional: the tossctl config directory (Open API credentials + config.json, read only)")
	top := flags.Int("top", cfg.Top, "symbols per session: ranking top by traded value, priced within the order cap (D8)")
	retest := flags.Uint64("retest-tolerance-ppm", 250_000, "breakout v1 retest tolerance, 100,000..250,000 (D21)")
	quoteAge := flags.Uint64("max-quote-age-ms", 5_000, "breakout v1 max quote age (D21)")
	spread := flags.Uint64("max-spread-ppm", 3_000, "breakout v1 max spread (D21)")
	drift := flags.Uint64("max-entry-drift-ppm", 5_000, "breakout v1 max entry drift (D21)")
	riskBudget := flags.Uint64("risk-budget-krw", 0, "override: hypothetical risk budget (default floor(max_daily_loss_amount / 3))")
	notional := flags.Uint64("notional-cap-krw", 0, "override: hypothetical notional cap (default max_order_notional)")
	sessionWindow := flags.String("session", "", `optional: bypass the calendar with a KST window "HH:MM-HH:MM" (only when the calendar is wrong)`)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.ToUpper(strings.TrimSpace(*market)) != "KR" {
		return fmt.Errorf("--market %q: a140 measures KR only (design D1)", *market)
	}
	if *top < 1 || *top > rankingFetch {
		return fmt.Errorf("--top must be 1..%d", rankingFetch)
	}
	paths, err := apppaths.Resolve(*configDir, "")
	if err != nil {
		return fmt.Errorf("resolving the config path: %w", err)
	}
	if err := checkEvidencePath(*evidenceDB, productionEvidenceCandidates(paths.ConfigDir)); err != nil {
		return err
	}
	limits, source, err := loadLimits(ctx, paths.ConfigFile, *riskBudget, *notional)
	if err != nil {
		return err
	}
	costPPM, err := costRatePPM(costs.DefaultModel())
	if err != nil {
		return err
	}
	cfg.Top = *top
	cfg.Params = shadowParams{RetestTolerancePPM: *retest, MaxQuoteAgeMS: *quoteAge, MaxSpreadPPM: *spread, MaxEntryDriftPPM: *drift,
		RiskBudgetKRW: limits.risk, NotionalCapKRW: limits.notional, CostPPM: costPPM, MinRRPPM: 1_500_000}
	cfg.ParamsSource = source + "; costs=costs.DefaultModel"
	if *retest < 100_000 || *retest > 250_000 || *quoteAge == 0 || *spread == 0 || *drift == 0 {
		return errors.New("breakout v1 needs --retest-tolerance-ppm in 100,000..250,000 and positive quote age, spread and drift limits")
	}
	var manual *[2]int
	if strings.TrimSpace(*sessionWindow) != "" {
		if manual, err = parseManualWindow(*sessionWindow); err != nil {
			return err
		}
	}

	credentialFile, tokenFile, err := apppaths.OpenAPI(*configDir)
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
	// 엔진과 같은 토큰 캐시를 공유하는 생산 client — 이 도구는 이 값을 marketSource(읽기 GET 다섯)로만 다룸.
	var src marketSource = official.New(*credentials, tokenFile)

	store, err := strategyevidence.Open(ctx, strategyevidence.Options{Path: *evidenceDB})
	if err != nil {
		return fmt.Errorf("opening the shadow evidence store: %w", err)
	}
	defer store.Close()
	ledgerDir, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	led, err := openLedger(ledgerDir)
	if err != nil {
		return err
	}
	defer led.Close()

	p := newProbe(src, store, led, cfg, func() time.Time { return time.Now().UTC() })
	calendar, sessionSource, err := resolveCalendar(ctx, p, manual)
	if err != nil {
		return err
	}
	if sessionSource == sessionSourceManual {
		p.sess.ID = "KRX:" + calendar.Today.Date
		if err := led.write(ledgerRow{Kind: kindGap, At: time.Now().UTC().Format(time.RFC3339), Session: p.sess.ID, Stage: "session",
			Reason: "calendar bypassed by --session " + *sessionWindow + " (operator override)"}); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "a140: ledger %s · evidence %s · session %s (%s) · top %d · params %+v\n",
		ledgerDir, *evidenceDB, calendar.Today.Date, sessionSource, cfg.Top, cfg.Params)
	return p.run(ctx, calendar, sessionSource)
}

// limits 는 가상 사이징의 두 값(원).
type limits struct{ risk, notional uint64 }

// loadLimits 는 config 파일을 읽기만 함: 위험 예산 = floor(max_daily_loss_amount / 3)(a139 사용자 결정과 같은 식),
// 명목 상한 = max_order_notional. KRW 가 아니거나 양의 유한수가 아니면 거부. 플래그가 0 이 아니면 그 값을 씀.
func loadLimits(ctx context.Context, configFile string, riskOverride, notionalOverride uint64) (limits, string, error) {
	out := limits{risk: riskOverride, notional: notionalOverride}
	source := "flags"
	if out.risk == 0 || out.notional == 0 {
		status, err := tossconfig.NewService(configFile).Status(ctx)
		if err != nil {
			return limits{}, "", fmt.Errorf("reading the config %s: %w", configFile, err)
		}
		if !status.Exists {
			return limits{}, "", fmt.Errorf("config %s not found (or pass --risk-budget-krw and --notional-cap-krw)", configFile)
		}
		gate := status.Engine.AutomationGate
		if strings.ToUpper(strings.TrimSpace(gate.LimitCurrency)) != "KRW" {
			return limits{}, "", fmt.Errorf("config engine.automation_gate.limit_currency %q is not KRW", gate.LimitCurrency)
		}
		positive := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 1 && v < 1e15 }
		if out.risk == 0 {
			if !positive(gate.MaxDailyLossAmount) {
				return limits{}, "", errors.New("config engine.automation_gate.max_daily_loss_amount is not a positive KRW amount")
			}
			out.risk = uint64(math.Floor(gate.MaxDailyLossAmount / 3))
		}
		if out.notional == 0 {
			if !positive(gate.MaxOrderNotional) {
				return limits{}, "", errors.New("config engine.automation_gate.max_order_notional is not a positive KRW amount")
			}
			out.notional = uint64(math.Floor(gate.MaxOrderNotional))
		}
		source = "config " + configFile + " (risk = floor(max_daily_loss_amount/3) unless --risk-budget-krw; notional = max_order_notional unless --notional-cap-krw)"
	}
	if out.risk == 0 || out.notional == 0 {
		return limits{}, "", errors.New("the hypothetical risk budget and notional cap must be positive")
	}
	return out, source, nil
}

// resolveCalendar 는 공식 달력(기본) 또는 수동 창(--session)으로 오늘 KR 세션을 정함.
func resolveCalendar(ctx context.Context, p *probe, manual *[2]int) (scheduler.CalendarSnapshot, string, error) {
	seoul, err := marketclock.MarketKR.Location()
	if err != nil {
		return scheduler.CalendarSnapshot{}, "", err
	}
	now := p.now()
	if manual != nil {
		day := now.In(seoul)
		y, m, d := day.Date()
		open := time.Date(y, m, d, manual[0]/60, manual[0]%60, 0, 0, seoul)
		closeAt := time.Date(y, m, d, manual[1]/60, manual[1]%60, 0, 0, seoul)
		date := day.Format("2006-01-02")
		sum := sha256.Sum256([]byte("a140-manual-session\x00" + date + "\x00" + open.Format(time.RFC3339) + "\x00" + closeAt.Format(time.RFC3339)))
		return scheduler.CalendarSnapshot{Market: marketclock.MarketKR, Version: "sha256:" + hex.EncodeToString(sum[:]),
			Source: sessionSourceManual, FetchedAt: now,
			Today: scheduler.CalendarDay{Date: date, Regular: &scheduler.SessionWindow{Open: open.UTC(), Close: closeAt.UTC()}}}, sessionSourceManual, nil
	}
	payload, err := p.paced.TypedMarketCalendar(ctx, "KR", now.In(seoul).Format("2006-01-02"))
	if err != nil {
		return scheduler.CalendarSnapshot{}, "", fmt.Errorf("reading the KR market calendar: %w", err)
	}
	cal, err := scheduler.AdaptOfficialCalendar(marketclock.MarketKR, payload, now)
	if err != nil {
		return scheduler.CalendarSnapshot{}, "", err
	}
	if cal.Today.Regular == nil {
		return scheduler.CalendarSnapshot{}, "", fmt.Errorf("KR has no regular session on %s per the calendar (use --session only if the calendar is wrong)", cal.Today.Date)
	}
	return cal, sessionSourceCalendar, nil
}

// parseManualWindow 는 "HH:MM-HH:MM" 을 시작 · 끝 분(자정부터)으로. 역전창 · 형식 오류는 거부.
func parseManualWindow(raw string) (*[2]int, error) {
	start, end, ok := strings.Cut(raw, "-")
	if !ok {
		return nil, fmt.Errorf(`--session must be "HH:MM-HH:MM", got %q`, raw)
	}
	parse := func(value string) (int, error) {
		t, err := time.Parse("15:04", strings.TrimSpace(value))
		if err != nil {
			return 0, fmt.Errorf("invalid HH:MM %q", strings.TrimSpace(value))
		}
		return t.Hour()*60 + t.Minute(), nil
	}
	from, err := parse(start)
	if err != nil {
		return nil, err
	}
	to, err := parse(end)
	if err != nil {
		return nil, err
	}
	if from >= to {
		return nil, fmt.Errorf("--session window %q is not a forward interval", raw)
	}
	return &[2]int{from, to}, nil
}

func reportCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("report", flag.ContinueOnError)
	ledgerPath := flags.String("ledger", filepath.Join(defaultLedgerDir, ledgerFileName), "ledger.jsonl path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	rows, err := readLedger(*ledgerPath)
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, renderReport(rows))
	return err
}
