package main

// contract_test.go 는 wire body 를 두 영수증에 대조함 — 설계 산문이 아니라 계약 원본.
//   1. 생산 직렬화기: 실제 official.Client.PlaceOrder 를 mock 에 쏴 받은 바이트와 바이트 동일.
//      (official 의 buildOrderCreate 가 바뀌면 이 시험이 깨짐 — execgw wirebody 표류 가드와 같은 방식)
//   2. docs/migration/openapi.latest.json 의 POST /api/v1/orders 요청 스키마.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

func TestWireBodyIsByteIdenticalToTheOfficialSerializer(t *testing.T) {
	t.Parallel()
	for _, qty := range []int{1, 2} {
		mock := newOfficialMock(t)
		client := official.New(official.Credentials{APIKey: sentinelAPIKey, SecretKey: sentinelSecret},
			filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(mock.server.URL), official.WithAccountSeq(1))
		// 시험 안에서만 생산 PlaceOrder 를 부름 — mock 서버라 실주문 아님.
		if _, err := client.PlaceOrder(context.Background(), orderintent.PlaceIntent{
			Symbol: "005930", Market: "KR", Side: "sell", OrderType: "market",
			Quantity: float64(qty), CurrencyMode: "KRW", ClientOrderID: testKey,
		}); err != nil {
			t.Fatalf("official PlaceOrder against the mock: %v", err)
		}
		spec := goodSpec()
		spec.Quantity = qty
		order, err := assembleOrder(spec)
		if err != nil {
			t.Fatal(err)
		}
		sent, _ := mock.orderRequest(0)
		if string(sent) != order.WireBody {
			t.Fatalf("qty %d: official serializer sent\n %s\nthis probe builds\n %s", qty, sent, order.WireBody)
		}
	}
}

func openapiSpec(t *testing.T) map[string]any {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "docs", "migration", "openapi.latest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func dig(t *testing.T, v any, keys ...string) any {
	t.Helper()
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("openapi: %q is not under an object", k)
		}
		v, ok = m[k]
		if !ok {
			t.Fatalf("openapi: missing key %q", k)
		}
	}
	return v
}

// TestWireBodyFitsTheOpenAPIRequestSchema 는 본문 칸이 OrderCreateQuantityBased 의 정의 안에만 있고
// required 를 다 채우며 enum·pattern·maxLength 를 지키는지 봄. MARKET 에 price 를 실으면 400 이라는
// 계약 문장도 원문에서 읽어 대조함.
func TestWireBodyFitsTheOpenAPIRequestSchema(t *testing.T) {
	t.Parallel()
	doc := openapiSpec(t)
	ref := dig(t, doc, "paths", "/api/v1/orders", "post", "requestBody", "content", "application/json", "schema", "$ref").(string)
	if ref != "#/components/schemas/OrderCreateRequest" {
		t.Fatalf("request schema ref = %q", ref)
	}
	var variant map[string]any
	for _, candidate := range dig(t, doc, "components", "schemas", "OrderCreateRequest", "oneOf").([]any) {
		if m := candidate.(map[string]any); m["title"] == "OrderCreateQuantityBased" {
			variant = m
		}
	}
	if variant == nil {
		t.Fatal("openapi has no OrderCreateQuantityBased variant")
	}
	props := variant["properties"].(map[string]any)

	order, err := assembleOrder(goodSpec())
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(order.WireBody), &body); err != nil {
		t.Fatal(err)
	}
	for key := range body {
		if _, ok := props[key]; !ok {
			t.Errorf("body field %q is not in the openapi schema", key)
		}
	}
	for _, req := range variant["required"].([]any) {
		if _, ok := body[req.(string)]; !ok {
			t.Errorf("required field %q is missing", req)
		}
	}
	for _, field := range []string{"side", "orderType"} {
		enum := props[field].(map[string]any)["enum"].([]any)
		found := false
		for _, v := range enum {
			found = found || v == body[field]
		}
		if !found {
			t.Errorf("%s=%v is outside the enum %v", field, body[field], enum)
		}
	}
	for _, field := range []string{"clientOrderId", "quantity"} {
		schema := props[field].(map[string]any)
		value := body[field].(string)
		if !regexp.MustCompile(schema["pattern"].(string)).MatchString(value) {
			t.Errorf("%s=%q does not match %v", field, value, schema["pattern"])
		}
		if limit := int(schema["maxLength"].(float64)); len(value) > limit {
			t.Errorf("%s=%q is longer than %d", field, value, limit)
		}
	}
	if q, _ := strconv.Atoi(body["quantity"].(string)); strconv.Itoa(q) != body["quantity"] {
		t.Errorf("quantity %v is not a whole number (fractional quantity is US market sell only)", body["quantity"])
	}
	if _, ok := body["price"]; ok {
		t.Error("a MARKET body carries price")
	}
	if desc := props["price"].(map[string]any)["description"].(string); !strings.Contains(desc, "`MARKET`: 전달 불가") {
		t.Errorf("the openapi price rule changed: %q", desc)
	}
	if _, ok := body["timeInForce"]; ok {
		t.Error("a MARKET body carries timeInForce")
	}
	if desc := props["symbol"].(map[string]any)["description"].(string); !strings.Contains(desc, "KRX: 6자리 숫자") {
		t.Errorf("the openapi symbol rule changed: %q", desc)
	}
	if desc := props["clientOrderId"].(map[string]any)["description"].(string); !strings.Contains(desc, "10분간 유효") {
		t.Errorf("the idempotency window wording changed (confirmWindow depends on it): %q", desc)
	}
	if body["confirmHighValueOrder"] != false {
		t.Errorf("confirmHighValueOrder = %v", body["confirmHighValueOrder"])
	}
}
