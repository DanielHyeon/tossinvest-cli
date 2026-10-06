// Command a128-jev-shadow-probe 는 a128 shadow 측정 도구임 — 생산 배선 0.
//
// 무엇을 하는가 (openspec/changes/a128-jev-shadow-judgment-probe)
//
// 정규장 동안 간격마다 official rankings 상위 K 종목의 공개 시장 데이터로 state 를 조립하고,
// TypeSafe Jev(Noul)에 J1(h분 뒤 상승?)·J2(매수 보류할 악재?)를 물어 p 를 JSONL 원장에 적음.
// h분 뒤 /prices 재조회로 실현 수익률을 같은 원장에 라벨로 적고, report 가 p 구간표를 만듦.
//
// 하지 않는 것: 주문·정정·취소·토글·설정 변경 전부. 브로커에는 공개 시장 데이터 GET 만
// 보냄(marketSource). 외부로 나가는 state 는 공개 시장 데이터뿐(PublicState).
//
//	TYPESAFE_API_KEY=… go run ./tools/a128-jev-shadow-probe run --market BOTH --wait-open
//	go run ./tools/a128-jev-shadow-probe report
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const defaultLedgerDir = "openspec/changes/a128-jev-shadow-judgment-probe/analysis/shadow-ledger"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := dispatch(ctx, os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "a128 jev shadow probe: "+err.Error())
		os.Exit(1)
	}
}

func dispatch(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New(`a subcommand is required: "run" or "report"`)
	}
	switch args[0] {
	case "run":
		return runCommand(ctx, args[1:], getenv)
	case "report":
		return reportCommand(args[1:], stdout)
	default:
		return fmt.Errorf("unknown subcommand %q; use \"run\" or \"report\"", args[0])
	}
}

// runCommand 는 키 검사를 가장 먼저 함 — 키가 없으면 브로커에 한 번도 묻지 않고 거부.
func runCommand(ctx context.Context, args []string, getenv func(string) string) error {
	judge, err := newJudgeClient(getenv(typeSafeAPIKeyEnv), defaultTypeSafeBase, "")
	if err != nil {
		return err
	}

	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	market := flags.String("market", "", `required: "KR", "US" or "BOTH"`)
	interval := flags.Duration("interval", 10*time.Minute, "sampling interval")
	top := flags.Int("top", 10, "symbols per market per cycle (official MARKET_TRADING_AMOUNT ranking)")
	horizon := flags.Int("horizon", 60, "label horizon in minutes")
	model := flags.String("model", defaultJudgeModel, "TypeSafe model id (pinned version by default)")
	out := flags.String("out", defaultLedgerDir, "ledger directory (ledger.jsonl + states/)")
	configDir := flags.String("config-dir", "", "optional: the tossctl config directory")
	waitOpen := flags.Bool("wait-open", false, "wait for the regular session to open instead of refusing")
	if err := flags.Parse(args); err != nil {
		return err
	}
	markets, err := parseMarkets(*market)
	if err != nil {
		return err
	}
	if *top < 1 || *top > 100 || *horizon < 1 || *interval < time.Minute {
		return errors.New("--top must be 1..100, --horizon ≥ 1 minute and --interval ≥ 1m")
	}
	judge.model = *model

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
	// 같은 토큰 캐시 파일을 공유하는 생산 client 를 그대로 씀 — 따로 토큰을 사면 엔진·콘솔의
	// 토큰을 무효화하는 토큰 전쟁이 남(a082). 이 도구는 이 값을 marketSource 로만 다룸.
	var src marketSource = official.New(*credentials, tokenFile)

	ledgerDir, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	l, err := openLedger(ledgerDir)
	if err != nil {
		return err
	}
	defer l.Close()

	p := newProbe(src, judge, l, config{
		Interval: *interval, TopK: *top, Horizon: time.Duration(*horizon) * time.Minute, WaitOpen: *waitOpen,
		LabelTick: 30 * time.Second, ReadGap: 250 * time.Millisecond,
		RetryDelays: []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second, 40 * time.Second},
	})
	existing, err := readLedger(filepath.Join(ledgerDir, ledgerFileName))
	if err != nil {
		return err
	}
	p.restorePending(existing)
	p.logf("ledger %s · markets %v · model %s · interval %s · top %d · horizon %dm · restored %d pending",
		ledgerDir, markets, judge.model, *interval, *top, *horizon, p.pendingCount())
	return p.run(ctx, markets)
}

func parseMarkets(value string) ([]string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "KR":
		return []string{"KR"}, nil
	case "US":
		return []string{"US"}, nil
	case "BOTH":
		return []string{"KR", "US"}, nil
	default:
		return nil, fmt.Errorf(`--market must be "KR", "US" or "BOTH", got %q`, value)
	}
}

func reportCommand(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("report", flag.ContinueOnError)
	ledgerPath := flags.String("ledger", filepath.Join(defaultLedgerDir, ledgerFileName), "ledger.jsonl path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	contents, err := readLedger(*ledgerPath)
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, renderCalibrationReport(contents, *ledgerPath))
	return err
}
