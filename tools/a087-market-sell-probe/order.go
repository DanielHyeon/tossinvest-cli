package main

// order.go 는 보낼 수 있는 주문의 모양을 하나로 고정함 — KR · 비분수 · MARKET · SELL · 1~2주.
//
// 계약 출처(지어낸 필드 없음):
//   - docs/migration/openapi.latest.json `POST /api/v1/orders` → OrderCreateRequest.oneOf[0]
//     (title OrderCreateQuantityBased): required symbol·side·orderType·quantity,
//     price 는 "MARKET: 전달 불가", timeInForce 는 CLS 가 US+LIMIT 전용이라 MARKET 에 싣지 않음,
//     clientOrderId 는 멱등성 키(최대 36자, `^[a-zA-Z0-9\-_]+$`, 10분 유효, 서버 자동 생성 없음).
//   - internal/official/orders_write.go `orderCreateV0`(:29) 필드 순서·태그, `buildOrderCreate`(:100)
//     의 비분수 MARKET 갈래(:136-152 — price·timeInForce 없음, ConfirmHighValueOrder false).
//     둘의 바이트 일치는 contract_test.go 가 실제 official.Client.PlaceOrder 를 mock 에 쏴 대조함.

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 코드에 고정한 하드 가드 값 — 플래그로 바꿀 수 없음.
const (
	fixedMarket    = "KR"
	fixedSide      = "sell"
	fixedOrderType = "market"
	maxQuantity    = 2
)

// confirmWindow 는 미리보기에서 발급한 키가 전송에 쓰일 수 있는 시간임.
// 브로커 멱등성 창(10분)보다 짧게 둬서, 같은 실행 명령을 다시 쳐도 그 재요청은
// 언제나 창 안에 떨어지고 "이전 주문 결과를 그대로 재반환" 받는 쪽이 되게 함(새 주문 아님).
const confirmWindow = 5 * time.Minute

// futureSkew 는 키 발급 시각이 현재보다 앞설 때 허용하는 시계 오차임.
const futureSkew = time.Minute

// orderPath 는 주문 생성 endpoint 임(openapi `POST /api/v1/orders`).
const orderPath = "/api/v1/orders"

// krSymbol 은 openapi `symbol` 설명 "KRX: 6자리 숫자" 를 그대로 옮긴 형식임.
var krSymbol = regexp.MustCompile(`^[0-9]{6}$`)

// clientOrderIDShape 는 이 도구가 발급하는 키의 모양임 — 접두사·KST 발급 시각·난수 8자.
// 길이 31 이라 openapi maxLength 36 과 pattern `^[a-zA-Z0-9\-_]+$` 안에 듦.
var clientOrderIDShape = regexp.MustCompile(`^a087ms-([0-9]{8}T[0-9]{6})-[0-9a-f]{8}$`)

const clientOrderIDTimeLayout = "20060102T150405"

// kst 는 영수증·키의 시간대임.
var kst = time.FixedZone("KST", 9*3600)

// orderSpec 은 조립 입력임. CLI 는 Symbol·Quantity 만 받고 나머지는 고정값을 넣음 —
// 가드는 그래도 모든 칸을 다시 검사함(조립 함수가 이 도구에서 본문을 만드는 유일한 자리라서).
type orderSpec struct {
	Market        string
	Side          string
	OrderType     string
	Symbol        string
	Quantity      int
	ClientOrderID string
}

// wireOrder 는 OrderCreateQuantityBased 의 MARKET 부분집합임. 필드 순서는 orderCreateV0 와 같아야
// 바이트가 같음(price·timeInForce 는 MARKET 에서 omitempty 로 빠지므로 아예 두지 않음).
type wireOrder struct {
	Symbol                string `json:"symbol"`
	Side                  string `json:"side"`
	OrderType             string `json:"orderType"`
	Quantity              string `json:"quantity"`
	ClientOrderID         string `json:"clientOrderId,omitempty"`
	ConfirmHighValueOrder bool   `json:"confirmHighValueOrder"`
}

// assembledOrder 는 가드를 통과한 결과임 — 전송은 이 값만 받음.
type assembledOrder struct {
	Spec     orderSpec
	WireBody string
	IssuedAt time.Time
}

// assembleOrder 는 하드 가드를 모두 통과한 경우에만 wire body 를 만듦.
// 가드 순서는 사람이 읽는 거절 문구가 원인 하나를 가리키도록 칸 순서대로 둠.
func assembleOrder(spec orderSpec) (assembledOrder, error) {
	if spec.Market != fixedMarket {
		return assembledOrder{}, fmt.Errorf("market %q refused: this probe sends KR orders only", spec.Market)
	}
	if spec.Side != fixedSide {
		return assembledOrder{}, fmt.Errorf("side %q refused: this probe sends sell orders only", spec.Side)
	}
	if spec.OrderType != fixedOrderType {
		return assembledOrder{}, fmt.Errorf("order type %q refused: this probe sends market orders only", spec.OrderType)
	}
	if !krSymbol.MatchString(spec.Symbol) {
		return assembledOrder{}, fmt.Errorf("symbol %q refused: a KR symbol is exactly six digits", spec.Symbol)
	}
	if spec.Quantity < 1 || spec.Quantity > maxQuantity {
		return assembledOrder{}, fmt.Errorf("quantity %d refused: this probe sends 1..%d shares", spec.Quantity, maxQuantity)
	}
	issued, err := parseClientOrderID(spec.ClientOrderID)
	if err != nil {
		return assembledOrder{}, err
	}
	encoded, err := json.Marshal(wireOrder{
		Symbol:        spec.Symbol,
		Side:          strings.ToUpper(spec.Side),
		OrderType:     strings.ToUpper(spec.OrderType),
		Quantity:      strconv.Itoa(spec.Quantity),
		ClientOrderID: spec.ClientOrderID,
		// 1억원 이상 확인 플래그는 buildOrderCreate 와 같이 false 고정 — 2주 이하 KR 주문에서
		// 넘을 일이 없고, 넘으면 브로커가 400 confirm-high-value-required 로 거절하는 쪽이 안전함.
		ConfirmHighValueOrder: false,
	})
	if err != nil {
		return assembledOrder{}, err
	}
	return assembledOrder{Spec: spec, WireBody: string(encoded), IssuedAt: issued}, nil
}

// mintClientOrderID 는 발급 시각을 담은 새 멱등성 키를 만듦. 시각을 키 안에 두는 이유는
// 실행 단계가 별도 상태 파일 없이 키 하나로 유효 시간을 판정하게 하려는 것임.
func mintClientOrderID(now time.Time, random io.Reader) (string, error) {
	var nonce [4]byte
	if _, err := io.ReadFull(random, nonce[:]); err != nil {
		return "", fmt.Errorf("minting the idempotency key: %w", err)
	}
	return "a087ms-" + now.In(kst).Format(clientOrderIDTimeLayout) + "-" + hex.EncodeToString(nonce[:]), nil
}

// parseClientOrderID 는 이 도구가 발급한 모양의 키만 받고 발급 시각을 돌려줌.
func parseClientOrderID(id string) (time.Time, error) {
	match := clientOrderIDShape.FindStringSubmatch(id)
	if match == nil {
		return time.Time{}, fmt.Errorf("client order id %q refused: not a key this probe issued", id)
	}
	issued, err := time.ParseInLocation(clientOrderIDTimeLayout, match[1], kst)
	if err != nil {
		return time.Time{}, fmt.Errorf("client order id %q refused: %w", id, err)
	}
	return issued, nil
}

// confirmToken 은 wire body 바이트 전체(키 포함)에서 유도함 — 칸 하나만 바뀌어도 값이 바뀜.
// 형식은 `<clientOrderId>.<sha256 앞 16자>` 라 실행 단계가 토큰 하나로 키를 되찾음.
func confirmToken(order assembledOrder) string {
	sum := sha256.Sum256([]byte("a087-market-sell-probe/confirm/v1\x00" + order.WireBody))
	return order.Spec.ClientOrderID + "." + hex.EncodeToString(sum[:])[:16]
}

// splitConfirmToken 은 토큰에서 키를 꺼냄. 검증은 verifyConfirm 이 함.
func splitConfirmToken(token string) (string, error) {
	id, digest, found := strings.Cut(token, ".")
	if !found || id == "" || len(digest) != 16 {
		return "", errors.New("confirm token refused: expected <clientOrderId>.<16 hex> as printed by the preview")
	}
	return id, nil
}

// verifyConfirm 은 실행 직전 마지막 로컬 판정임: 토큰이 지금 다시 조립한 본문과 일치하고,
// 키가 유효 창 안에 있어야 함. 둘 중 하나라도 아니면 전송하지 않음.
func verifyConfirm(order assembledOrder, token string, now time.Time) error {
	want := confirmToken(order)
	if subtle.ConstantTimeCompare([]byte(want), []byte(token)) != 1 {
		return errors.New("confirm token refused: it does not match this exact wire body — run the preview again")
	}
	age := now.Sub(order.IssuedAt)
	if age > confirmWindow {
		return fmt.Errorf("confirm token refused: issued %s ago, the window is %s — run the preview again",
			age.Round(time.Second), confirmWindow)
	}
	if age < -futureSkew {
		return fmt.Errorf("confirm token refused: issued %s in the future — check the clock", (-age).Round(time.Second))
	}
	return nil
}

// newRandom 은 운영 난수원임(시험은 고정 바이트를 넣음).
func newRandom() io.Reader { return rand.Reader }
