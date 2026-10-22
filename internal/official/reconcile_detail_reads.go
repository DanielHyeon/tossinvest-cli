package official

// reconcile_detail_reads.go 는 a164(verify-reconcile-closes-broker-expired-conditional) 만료 대사 전용 by-id 상세 읽기임.
//
// 같은 endpoint(GET /api/v1/conditional-orders/{id})·같은 getAcct 경로를 쓰되 result 원문을 받아 스키마를 따로 검증하고,
// 404 중 본문 code 가 정확히 conditional-order-not-found 인 것만 「silent」 로 가름(design D3 「by-id 응답 분류」).
// 기존 ConditionalOrder·ConditionalOrderRaw 는 null 을 제로값으로 접고(F2) 404 를 오류로만 돌려주므로 쓰지 않음.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ReconcileConditionalDetail 은 by-id 상세 GET 한 번의 분류 결과임.
// Silent 면 404 conditional-order-not-found(진술 아님 — 근거로 세지 않음) 이고 나머지 판정 필드는 비어 있음.
type ReconcileConditionalDetail struct {
	Silent bool
	// RequestID 는 silent 응답 본문의 requestId 임(계좌 정보 없음) — 거절 문구에만 씀.
	RequestID        string
	ID               string
	Type             string
	Status           string
	Symbol           string
	Market           string
	FirstStatus      string
	TriggeredOrderID string
	HasSecond        bool
	// ResultDigest 는 statement 의 result 원문 지문임(sha256 16바이트 절단). 성공 응답 requestId 를 클라이언트가
	// 노출하지 않아서 그 대신 적음(design R3).
	ResultDigest string
}

// reconcileDetailNotFoundCode 는 silent 로 가르는 유일한 404 본문 code 임(M40 실측 문자열).
const reconcileDetailNotFoundCode = "conditional-order-not-found"

// DecodeReconcileConditionalDetail 은 by-id 상세의 result 원문을 statement 로 엄격 해독함(design D3).
//
// 계약 required 다섯(conditionalOrderId · status · symbol · market · first.status)은 non-null 문자열이어야 하고, first 는 객체여야 함.
// nullable 둘(first.triggeredOrderId · second)은 키 없음과 null 이 같게 「없음」 이고, 값이 있으면 형만 봄(D3 정정 2026-10-22).
// type 은 판정(A4 type == SINGLE)에 쓰므로 있으면 문자열이어야 함 — 없거나 null 이면 빈 값이라 A4 가 거절함.
// null 이 Go 제로값으로 접혀 「빈 status」 가 되는 일을 막으려고 원문에서 먼저 가름(a121 F2 와 같은 이유).
func DecodeReconcileConditionalDetail(result []byte) (ReconcileConditionalDetail, error) {
	trimmed := bytes.TrimSpace(result)
	fields, err := reconcileObject(trimmed, "result")
	if err != nil {
		return ReconcileConditionalDetail{}, err
	}
	var d ReconcileConditionalDetail
	for key, dst := range map[string]*string{"conditionalOrderId": &d.ID, "status": &d.Status, "symbol": &d.Symbol, "market": &d.Market} {
		if *dst, err = reconcileRequiredString(fields, key); err != nil {
			return ReconcileConditionalDetail{}, err
		}
	}
	if d.Type, err = reconcileOptionalString(fields, "type"); err != nil {
		return ReconcileConditionalDetail{}, err
	}
	rawFirst, ok := fields["first"]
	if !ok {
		return ReconcileConditionalDetail{}, reconcileSchema("first is missing")
	}
	first, err := reconcileObject(bytes.TrimSpace(rawFirst), "first")
	if err != nil {
		return ReconcileConditionalDetail{}, err
	}
	if d.FirstStatus, err = reconcileRequiredString(first, "status"); err != nil {
		return ReconcileConditionalDetail{}, err
	}
	if d.TriggeredOrderID, err = reconcileOptionalString(first, "triggeredOrderId"); err != nil {
		return ReconcileConditionalDetail{}, err
	}
	// second: 없음·null = 단일 leg, 객체 = 둘째 leg 있음, 그 밖의 형은 거절(빈 값으로 접지 않음).
	if rawSecond, ok := fields["second"]; ok {
		switch t := bytes.TrimSpace(rawSecond); {
		case string(t) == "null":
		case len(t) > 0 && t[0] == '{':
			d.HasSecond = true
		default:
			return ReconcileConditionalDetail{}, reconcileSchema("second is neither null nor an object")
		}
	}
	sum := sha256.Sum256(trimmed)
	d.ResultDigest = "sha256:" + hex.EncodeToString(sum[:16])
	return d, nil
}

// reconcileObject 는 원문이 JSON 객체인지 보고 필드 원문을 돌려줌. 같은 키가 두 번 나오면 거절함 — map 으로 접으면
// 마지막 값이 앞 값을 덮어(예: triggeredOrderId "ORD" 뒤의 null) 반대 증거가 조용히 사라지기 때문(gstack 리뷰 Codex P1).
func reconcileObject(raw []byte, name string) (map[string]json.RawMessage, error) {
	if len(raw) == 0 || raw[0] != '{' {
		return nil, reconcileSchema("%s is not an object (%.20s)", name, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if _, err := dec.Token(); err != nil { // 여는 '{'
		return nil, reconcileSchema("%s does not decode: %v", name, err)
	}
	fields := map[string]json.RawMessage{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, reconcileSchema("%s does not decode: %v", name, err)
		}
		key := fmt.Sprint(tok) // 객체 키 자리의 토큰은 문자열 — 봉인 census 가 type assertion 을 금하므로 fmt 로 꺼냄
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, reconcileSchema("%s.%s does not decode: %v", name, key, err)
		}
		if _, dup := fields[key]; dup {
			return nil, reconcileSchema("%s repeats the key %q — a later value would hide an earlier one", name, key)
		}
		fields[key] = value
	}
	if _, err := dec.Token(); err != nil { // 닫는 '}'
		return nil, reconcileSchema("%s does not decode: %v", name, err)
	}
	if dec.More() {
		return nil, reconcileSchema("%s has trailing data", name)
	}
	return fields, nil
}

// reconcileRequiredString 은 키가 있고 값이 null 아닌 JSON 문자열임을 요구함.
func reconcileRequiredString(fields map[string]json.RawMessage, key string) (string, error) {
	raw, ok := fields[key]
	if !ok {
		return "", reconcileSchema("%s is missing", key)
	}
	if string(bytes.TrimSpace(raw)) == "null" {
		return "", reconcileSchema("%s is null", key)
	}
	return reconcileOptionalString(fields, key)
}

// reconcileOptionalString 은 키 없음·null 을 빈 값으로, 그 밖에는 JSON 문자열만 받음.
func reconcileOptionalString(fields map[string]json.RawMessage, key string) (string, error) {
	raw, ok := fields[key]
	t := bytes.TrimSpace(raw)
	if !ok || string(t) == "null" {
		return "", nil
	}
	var v string
	if len(t) == 0 || t[0] != '"' || json.Unmarshal(t, &v) != nil {
		return "", reconcileSchema("%s is not a string", key)
	}
	return v, nil
}

// reconcileDetailSilent 는 by-id 오류가 silent(404 + 본문 error.code == conditional-order-not-found)인지 가름.
// 그 밖의 404(다른 code·code 없음·해독 불능 본문)와 400·401/403·429·5xx·전송 오류는 silent 가 아님 — 호출자가 거절함.
func reconcileDetailSilent(err error) (requestID string, silent bool) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusNotFound {
		return "", false
	}
	var body struct {
		Error *struct {
			Code      string `json:"code"`
			RequestID string `json:"requestId"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(apiErr.Body), &body) != nil || body.Error == nil || body.Error.Code != reconcileDetailNotFoundCode {
		return "", false
	}
	return body.Error.RequestID, true
}

// ReconcileConditionalOrder 는 GET /api/v1/conditional-orders/{id} 를 대사용으로 읽음(계좌 헤더 포함, 자동 재시도 없음 — D6).
// 결과는 statement(엄격 해독 통과) 또는 silent 이고, 그 밖은 오류임.
func (c *Client) ReconcileConditionalOrder(ctx context.Context, id string) (ReconcileConditionalDetail, error) {
	if strings.TrimSpace(id) == "" {
		return ReconcileConditionalDetail{}, reconcileSchema("the by-id read names no identifier")
	}
	var raw json.RawMessage
	if err := c.getAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil, &raw); err != nil {
		if requestID, silent := reconcileDetailSilent(err); silent {
			return ReconcileConditionalDetail{Silent: true, RequestID: requestID}, nil
		}
		return ReconcileConditionalDetail{}, err
	}
	return DecodeReconcileConditionalDetail(raw)
}
