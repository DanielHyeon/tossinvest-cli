// Command a087-market-sell-probe 는 a087 §5.1·5.2 실측 도구임 — 브로커(토스 공식 Open API)가
// KR 비분수 MARKET 매도를 실제로 받는지 production 변경 **전에** 잼.
//
// ⚠ 실돈 주문을 냄(mutating). 에이전트는 실행하지 않음 — 사람이 터미널에서 돌림.
//
// 왜 tossctl 이 아닌가: internal/trading.placeIntentSupported 가 비지정가를 브로커에 닿기 전에
// 로컬 거부함(a087 review.md 「2차 proposal-freeze 재리뷰」 P1-1). 그 거부를 브로커 응답으로
// 기록하면 거짓 음성이 됨. 그래서 이 도구는 trading 을 거치지 않고 official 의 공유 토큰
// 관리자 헤더(AuthHeaders)만 빌려 POST /api/v1/orders 를 한 번 보냄(사용자 결정 2026-10-10 2항).
//
// 흐름(두 번 실행):
//
//	① 미리보기 — 전송 없음. wire body 원문 + 1회용 confirm token 을 출력함.
//	   go run ./tools/a087-market-sell-probe --symbol 005930 --qty 1
//	② 실행 — 사람이 토큰을 붙여 다시 침. 토큰은 body 바이트(멱등성 키 포함)에서 유도되고 5분 유효.
//	   go run ./tools/a087-market-sell-probe --symbol 005930 --qty 1 --execute --confirm <token>
//
// 하드 가드(코드 고정, 플래그 없음): KR · sell · market · 1~2주 · 6자리 숫자 심볼.
// 전송은 단 1회 — 재시도 없음. 결과는 성공·거절·불명 전부 영수증 JSON 으로 남김.
// 5.2(세션 밖 응답 코드) 는 같은 명령을 장 밖에 돌리면 됨 — 영수증의 session_context 가 시각을 남김.
//
// 종료 코드: 0 = 미리보기 완료 또는 접수됨, 2 = 브로커 거절, 3 = 결과 불명, 1 = 도구 오류(미전송).
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
	"syscall"
	"time"

	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"golang.org/x/term"
)

const defaultReceiptDir = "openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/market-sell-receipts"

// sendTimeout 은 전송 한 번의 상한임. 넘기면 결과 불명으로 기록함(재전송 없음).
const sendTimeout = 15 * time.Second

const (
	exitOK      = 0
	exitError   = 1
	exitRefused = 2
	exitUnknown = 3
)

// deps 는 시험이 바꿔 끼우는 바깥 세계임 — 시계·난수·출력·터미널 판정·official client·전송 시한.
type deps struct {
	now             func() time.Time
	random          io.Reader
	stdout          io.Writer
	stdinIsTerminal func() bool
	openClient      func(configDir string) (headerSource, string, error)
	sendTimeout     time.Duration
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	code, err := run(ctx, os.Args[1:], deps{
		now:             time.Now,
		random:          newRandom(),
		stdout:          os.Stdout,
		stdinIsTerminal: func() bool { return term.IsTerminal(int(os.Stdin.Fd())) },
		openClient:      openOfficial,
		sendTimeout:     sendTimeout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "a087 market sell probe: "+err.Error())
	}
	os.Exit(code)
}

// openOfficial 은 a128 프로브와 같은 방식으로 생산 client 를 만듦 — 같은 토큰 캐시 파일을 공유해야
// 엔진·콘솔의 토큰을 무효화하는 토큰 전쟁(a082)이 나지 않음. 별도 토큰 발급 경로는 없음.
func openOfficial(configDir string) (headerSource, string, error) {
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
	client := official.New(*credentials, tokenFile)
	return client, client.BaseURL(), nil
}

type options struct {
	symbol    string
	quantity  int
	execute   bool
	confirm   string
	out       string
	configDir string
	note      string
}

func parseOptions(args []string) (options, error) {
	var o options
	flags := flag.NewFlagSet("a087-market-sell-probe", flag.ContinueOnError)
	flags.StringVar(&o.symbol, "symbol", "", "required: KR symbol, six digits (e.g. 005930)")
	flags.IntVar(&o.quantity, "qty", 0, fmt.Sprintf("required: whole shares to sell, 1..%d", maxQuantity))
	flags.BoolVar(&o.execute, "execute", false, "send the order (needs --confirm from the preview)")
	flags.StringVar(&o.confirm, "confirm", "", "the confirm token the preview printed")
	flags.StringVar(&o.out, "out", defaultReceiptDir, "receipt directory")
	flags.StringVar(&o.configDir, "config-dir", "", "optional: the tossctl config directory")
	flags.StringVar(&o.note, "note", "", "optional: free text stored in the receipt (no account details)")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments %v", flags.Args())
	}
	if o.confirm != "" && !o.execute {
		return options{}, errors.New("--confirm is only meaningful with --execute")
	}
	if o.execute && o.confirm == "" {
		return options{}, errors.New("--execute needs --confirm <token>; run without --execute first to get one")
	}
	return o, nil
}

// specFrom 은 CLI 가 고를 수 있는 칸(심볼·수량)만 사람에게서 받고 나머지는 고정값으로 채움.
func specFrom(o options, clientOrderID string) orderSpec {
	return orderSpec{
		Market:        fixedMarket,
		Side:          fixedSide,
		OrderType:     fixedOrderType,
		Symbol:        o.symbol,
		Quantity:      o.quantity,
		ClientOrderID: clientOrderID,
	}
}

func run(ctx context.Context, args []string, d deps) (int, error) {
	o, err := parseOptions(args)
	if err != nil {
		return exitError, err
	}
	if !o.execute {
		return preview(o, d)
	}
	return execute(ctx, o, d)
}

// preview 는 네트워크·자격증명을 전혀 건드리지 않음. 키를 새로 발급하고 보낼 바이트를 그대로 보여 줌.
func preview(o options, d deps) (int, error) {
	now := d.now()
	id, err := mintClientOrderID(now, d.random)
	if err != nil {
		return exitError, err
	}
	order, err := assembleOrder(specFrom(o, id))
	if err != nil {
		return exitError, err
	}
	token := confirmToken(order)
	fmt.Fprintf(d.stdout, "a087 KR MARKET sell probe — PREVIEW (nothing was sent)\n\n")
	fmt.Fprintf(d.stdout, "POST %s\n%s\n\n", orderPath, order.WireBody)
	fmt.Fprintf(d.stdout, "clientOrderId (idempotency key): %s\n", order.Spec.ClientOrderID)
	fmt.Fprintf(d.stdout, "confirm token: %s\n", token)
	fmt.Fprintf(d.stdout, "valid until:   %s (%s)\n\n", kstStamp(order.IssuedAt.Add(confirmWindow)), confirmWindow)
	fmt.Fprintf(d.stdout, "This places a REAL order. To send it once, a human runs in a terminal:\n")
	fmt.Fprintf(d.stdout, "  go run ./tools/a087-market-sell-probe --symbol %s --qty %d --execute --confirm %s\n",
		o.symbol, o.quantity, token)
	fmt.Fprintf(d.stdout, "(repeat --out / --config-dir / --note if you use them)\n")
	return exitOK, nil
}

// execute 의 순서가 안전 계약임: 로컬 판정 전부 → 자격증명·헤더 → 배타 영수증 → 단 1회 전송 → 최종 영수증.
// 전송 앞의 어느 단계든 실패하면 아무것도 보내지 않음.
func execute(ctx context.Context, o options, d deps) (int, error) {
	// TTY 검사는 우발 실행 방지임 — pty 래퍼로 통과되므로 "사람만 실행" 의 증명이 아님(3차 재리뷰 B-P3-1).
	// 에이전트 차단은 권한 체계·안전 불변식 2 가 맡음.
	if !d.stdinIsTerminal() {
		return exitError, errors.New("--execute refused: stdin is not a terminal — " +
			"this is an accident guard, not proof of a human; agents are barred by the permission system and invariant 2")
	}
	id, err := splitConfirmToken(o.confirm)
	if err != nil {
		return exitError, err
	}
	order, err := assembleOrder(specFrom(o, id))
	if err != nil {
		return exitError, err
	}
	if err := verifyConfirm(order, o.confirm, d.now()); err != nil {
		return exitError, err
	}
	dir, err := filepath.Abs(o.out)
	if err != nil {
		return exitError, err
	}
	path := receiptPath(dir, id)
	if _, err := os.Lstat(path); err == nil {
		return exitError, errAlreadySent
	} else if !errors.Is(err, os.ErrNotExist) {
		return exitError, fmt.Errorf("checking the receipt path (nothing was sent): %w", err)
	}

	source, baseURL, err := d.openClient(o.configDir)
	if err != nil {
		return exitError, err
	}
	headers, err := source.AuthHeaders(ctx)
	if err != nil {
		return exitError, fmt.Errorf("getting headers from the shared token manager (nothing was sent): %w", err)
	}
	if headers[official.HeaderAuthorization] == "" || headers[official.HeaderAccount] == "" {
		return exitError, errNoHeaders
	}
	req, err := buildRequest(ctx, baseURL, order.WireBody, headers)
	if err != nil {
		return exitError, err
	}

	pending := newPendingReceipt(order, o.confirm, d.now(), headerNames(req), o.note)
	if err := createPendingReceipt(path, pending); err != nil {
		return exitError, fmt.Errorf("%w (nothing was sent)", err)
	}

	result := sendOnce(newSendClient(d.sendTimeout), req, d.now)

	final := completeReceipt(pending, result, id)
	writeErr := finalizeReceipt(path, final)
	printOutcome(d.stdout, final, path, writeErr)
	if writeErr != nil {
		return exitUnknown, fmt.Errorf("the order WAS handed over but the final receipt could not be written: %w", writeErr)
	}
	switch final.Outcome {
	case outcomeAccepted:
		return exitOK, nil
	case outcomeRefused:
		return exitRefused, nil
	default:
		return exitUnknown, nil
	}
}

func printOutcome(w io.Writer, r receipt, path string, writeErr error) {
	fmt.Fprintf(w, "a087 KR MARKET sell probe — SENT ONCE (no retry)\n\n")
	fmt.Fprintf(w, "outcome: %s\n", r.Outcome)
	if r.Response != nil {
		fmt.Fprintf(w, "http status: %d\n", r.Response.HTTPStatus)
		if r.Response.ErrorCode != "" {
			fmt.Fprintf(w, "error code: %s\n", r.Response.ErrorCode)
		}
		if r.Response.OrderID != "" {
			fmt.Fprintf(w, "order id: %s\n", r.Response.OrderID)
		}
	}
	if r.TransportError != "" {
		fmt.Fprintf(w, "transport error: %s\n", r.TransportError)
	}
	fmt.Fprintf(w, "\n%s\n", r.Guidance)
	if writeErr != nil {
		// 최종 영수증을 못 쓰면 사람이 화면에서 옮겨 적을 수 있게 전문을 찍음.
		encoded, _ := marshalReceipt(r)
		fmt.Fprintf(w, "\nRECEIPT NOT WRITTEN (%v) — copy it from here:\n%s", writeErr, encoded)
		return
	}
	fmt.Fprintf(w, "receipt: %s\n", path)
}
