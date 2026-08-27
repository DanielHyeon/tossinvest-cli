// Command a112-l1c-quote-probe runs one human-approved, read-only a112 L1c
// acceptance observation: two GETs for one symbol in one market, reported as
// shape only. It is deliberately not installed into tossctl.
//
// 무엇을 하는가 (a112 결정 42·45)
//
// 새 엄격 리더 둘(StrictOrderbookTop, StrictLastPrice)을 실제 브로커에 한 번씩 태우고,
// 받은 것의 **모양**만 보고한다. 가격도 잔량도 원문도 찍지 않고 파일도 쓰지 않으며
// 증거를 적재하지도 않는다. 시장이 열려 있을 때 돌려야 사다리가 비어 있지 않다
// (KR 09:00–15:30 KST, US 22:30–05:00 KST).
//
// 사람이 직접 돌린다. 에이전트가 대신 돌리지 않는다. 실행마다 새 승인이 필요하다.
//
//	go run ./tools/a112-l1c-quote-probe --market US --symbol AAPL
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// 한 번의 프로브가 쓸 수 있는 시간 전부. 두 GET 이상은 구조적으로 불가능하다.
const probeBudget = 30 * time.Second

func main() {
	market := flag.String("market", "", `required: "KR" or "US"`)
	symbol := flag.String("symbol", "", "required: the exact broker symbol (KR six digits, US ticker)")
	configDir := flag.String("config-dir", "", "optional: the tossctl config directory")
	flag.Parse()

	if err := run(*market, *symbol, *configDir); err != nil {
		fmt.Fprintln(os.Stderr, "a112 l1c quote probe: "+err.Error())
		os.Exit(1)
	}
}

func run(market, symbol, configDir string) error {
	// 기본값을 두지 않는다. 시장과 종목을 손으로 적어야만 요청이 나간다.
	if market == "" || symbol == "" {
		return errors.New("--market and --symbol are required; this probe has no defaults")
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

	ctx, cancel := context.WithTimeout(context.Background(), probeBudget)
	defer cancel()

	// 순서는 생산자와 같다: 호가가 먼저다. 실패한 절반은 다시 부르지 않는다.
	top, err := client.StrictOrderbookTop(ctx, market, symbol)
	if err != nil {
		return fmt.Errorf("reading the top of book: %w", err)
	}
	last, err := client.StrictLastPrice(ctx, market, symbol)
	if err != nil {
		return fmt.Errorf("reading the last price: %w", err)
	}
	fmt.Print(renderReport(observation{Market: market, Symbol: symbol, Top: top, Last: last}))
	return nil
}
