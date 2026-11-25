package official

// reconcile_response.go 는 a168(verify-read-decode-hardening) 대사 읽기 응답 판정임.
//
// 대사 리더 네 개(목록 둘 · 종목 · by-id)는 기존 해독이 끝난 뒤 이 판정을 거침: 결과를 정한 마지막 HTTP 시도가 정확히 200 이고,
// 그 본문 어느 객체에도 해독기가 한 필드로 접을 수 있는 키가 없어야 결과를 돌려줌(design D1 · D2 · D6).
// 공용 전송(send · doRequest · unwrapAndDecode)은 건드리지 않음 — 상태 코드와 본문 원문은 시도 관측자로만 받음(design §2).
// a121 봉인 census 대상 파일이라 type assertion · type switch 를 쓰지 않음(design D5).

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// reconcileContractKeys 는 대사 리더가 응답에서 읽는 키 전부임 — 봉투 · 목록 result · 조건주문 행과 다리 · 주문 행 · 종목 항목 · by-id.
// 목록은 여기 한 곳에만 둠. TestReconcileContractKeysAreTheKeysTheReadersRead 가 리더의 구조체 태그·키 자리와 대조함.
var reconcileContractKeys = []string{
	"result",
	"conditionalOrders", "orders", "hasNext", "nextCursor",
	"conditionalOrderId", "clientOrderId", "type", "status", "symbol", "market", "quantity", "orderType", "expireDate",
	"first", "second", "createdAt",
	"orderSide", "triggerPrice", "targetProfitRate", "orderPrice", "triggeredOrderId",
	"orderId",
}

// reconcileAttempts 는 read 를 시도 관측자와 함께 부르고, 관측된 시도 전부와 read 의 오류를 그대로 돌려줌.
// 바깥 관측자는 잇지 않음 — 이 호출 동안만 가려짐(design D5, 잔여 R4).
func reconcileAttempts(ctx context.Context, read func(context.Context) error) ([]AttemptTrace, error) {
	var attempts []AttemptTrace
	err := read(WithAttemptObserver(ctx, func(trace AttemptTrace) { attempts = append(attempts, trace) }))
	return attempts, err
}

// reconcileAccepted 는 결과를 정한 시도의 판정임. send 는 마지막 doRequest 의 (상태, 본문)으로 판정·해독하므로 마지막 시도만 봄.
func reconcileAccepted(attempts []AttemptTrace) error {
	if len(attempts) == 0 {
		// 도달 불가 방어: 읽기가 nil 을 돌려줬다면 doRequest 가 적어도 한 번 관측을 냈음.
		return reconcileSchema("no attempt was observed")
	}
	last := attempts[len(attempts)-1]
	if last.StatusCode != http.StatusOK {
		return reconcileSchema("the read was answered HTTP %d — the reconcile contract names 200 only", last.StatusCode)
	}
	return reconcileUnambiguous(last.Body, "response")
}

// reconcileUnambiguous 는 값 하나를 어느 깊이까지든 훑어 모호한 객체를 거절함.
//
// 한 객체 안 판정 순서: 정확히 같은 키(reconcileObject) → 계약 키의 대소문자 변형(D6) → 대소문자만 다른 두 키(D2).
// encoding/json 은 구조체 필드를 bytes.EqualFold 로 맞추고 뒤 값을 남기므로, 접히는 키가 있으면 해독기마다 다른 값을 읽게 됨.
// 뒤따르는 데이터는 여기서 보지 않음 — 본문은 이미 unwrapAndDecode 가 값 하나임을 요구했음(design P9).
func reconcileUnambiguous(raw []byte, name string) error {
	trimmed := bytes.TrimSpace(raw)
	switch {
	case len(trimmed) > 0 && trimmed[0] == '{':
		fields, err := reconcileObject(trimmed, name)
		if err != nil {
			return err
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys) // 거절 문구를 실행마다 같게 함
		for i, key := range keys {
			for _, contract := range reconcileContractKeys {
				if key != contract && strings.EqualFold(key, contract) {
					return reconcileSchema("%s carries %q, a case variant of the key %q the readers read", name, key, contract)
				}
			}
			for _, other := range keys[i+1:] {
				if strings.EqualFold(key, other) {
					return reconcileSchema("%s carries %q and %q, which the decoder folds into one field", name, key, other)
				}
			}
		}
		for _, key := range keys {
			if err := reconcileUnambiguous(fields[key], name+"."+key); err != nil {
				return err
			}
		}
	case len(trimmed) > 0 && trimmed[0] == '[':
		var items []json.RawMessage
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return reconcileSchema("%s does not decode: %v", name, err)
		}
		for i, item := range items {
			if err := reconcileUnambiguous(item, name+"["+strconv.Itoa(i)+"]"); err != nil {
				return err
			}
		}
	}
	return nil
}
