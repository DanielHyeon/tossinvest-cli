package official

// reconcile_response_test.go — a168 GREEN 로트: 응답 판정 함수(reconcile_response.go)를 직접 잼.
// 클라이언트 경로 시험은 reconcile_ambiguity_test.go 에 있음.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestReconcileAcceptedNeedsAnObservedAttempt 는 M8 — 관측 0 은 통과가 아니라 거절임(도달 불가 방어의 고정).
func TestReconcileAcceptedNeedsAnObservedAttempt(t *testing.T) {
	if err := reconcileAccepted(nil); err == nil {
		t.Fatal("no observed attempt was accepted")
	}
}

// TestReconcileAcceptedJudgesOnlyTheLastAttempt 는 앞 시도의 상태·본문이 판정에 들어가지 않음을 잼(M2 의 직접 증인).
func TestReconcileAcceptedJudgesOnlyTheLastAttempt(t *testing.T) {
	clean := []byte(`{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`)
	dup := []byte(`{"error":{"code":"a","code":"b"}}`)
	if err := reconcileAccepted([]AttemptTrace{{StatusCode: http.StatusUnauthorized, Body: dup}, {StatusCode: http.StatusOK, Body: clean}}); err != nil {
		t.Fatalf("a refused first attempt leaked into the judgement: %v", err)
	}
	if err := reconcileAccepted([]AttemptTrace{{StatusCode: http.StatusOK, Body: clean}, {StatusCode: http.StatusCreated, Body: clean}}); err == nil {
		t.Fatal("a last attempt answered 201 was accepted")
	}
}

// TestReconcileUnambiguousAcceptsTheDecoderFixtures 는 C6 — 기존 직접 해독 시험의 수락 픽스처(봉투로 감쌈)와 by-id 골든이 새 판정을
// 통과함. 직접 해독 시험은 클라이언트 경로를 지나지 않으므로 여기서 따로 셈(design §6).
func TestReconcileUnambiguousAcceptsTheDecoderFixtures(t *testing.T) {
	for _, result := range []string{
		`{"conditionalOrders":[],"nextCursor":null,"hasNext":false}`,
		`{"conditionalOrders":[],"nextCursor":"c2","hasNext":true}`,
		`{"orders":[],"nextCursor":null,"hasNext":false}`,
		`{"orders":[],"nextCursor":"c2","hasNext":true}`,
		`{"conditionalOrders":[` +
			`{"conditionalOrderId":"CO-1","clientOrderId":"c1","type":"SINGLE","status":"COMPLETED","symbol":"005930","market":"KR",` +
			`"quantity":"1","orderType":"MARKET","expireDate":"2026-10-30","createdAt":"2026-10-01T09:00:00+09:00",` +
			`"first":{"orderSide":"SELL","type":"STOP","status":"COMPLETED","triggerPrice":"70000","targetProfitRate":null,` +
			`"orderPrice":null,"triggeredOrderId":"ORD-9"},"second":null},` +
			`{"conditionalOrderId":"CO-2","clientOrderId":"c2","type":"OCO","status":"WATCHING","symbol":"005930","market":"KR",` +
			`"quantity":"1","orderType":"LIMIT","expireDate":"2026-10-30","createdAt":"2026-10-01T09:00:00+09:00",` +
			`"first":{"orderSide":"SELL","type":"STOP","status":"WATCHING","triggerPrice":"70000","targetProfitRate":null,` +
			`"orderPrice":"69900","triggeredOrderId":null},` +
			`"second":{"orderSide":"SELL","type":"PROFIT_RATE","status":"WATCHING","triggerPrice":null,"targetProfitRate":"10.5",` +
			`"orderPrice":"80000","triggeredOrderId":null}}` +
			`],"nextCursor":null,"hasNext":false}`,
		detailGoldenResult,
		`[{"symbol":"005930","name":"x","market":"KOSPI"}]`,
	} {
		if err := reconcileUnambiguous([]byte(wrapResult(result)), "response"); err != nil {
			t.Fatalf("a well-formed decoder fixture was refused: %v\n%s", err, result)
		}
	}
}

// TestReconcileUnambiguousNamesTheRuleThatRefused 는 세 규칙(정확한 중복 · 계약 키 변형 · 계약 밖 접힘 짝)이 각자 자기 문구로 거절함을 잼.
func TestReconcileUnambiguousNamesTheRuleThatRefused(t *testing.T) {
	for body, want := range map[string]string{
		// 경로의 키는 인용됨 — 제어 문자 키가 거절 문구에 날것으로 나가지 않음(리뷰 1회차 보안 지적)
		`{"result":{"x\u001b[2J":{"a":1,"a":2}}}`:   `response."result"."x\x1b[2J"`,
		`{"result":{"a":1,"a":2}}`:                  "repeats the key",
		`{"result":{"Symbol":"x"}}`:                 "a case variant of the key \"symbol\"",
		`{"result":{"datetime":1,"dateTime":2}}`:    "folds into one field",
		`{"result":[{"meta":{"size":1,"ſize":2}}]}`: `response."result"[0]."meta"`,
	} {
		err := reconcileUnambiguous([]byte(body), "response")
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("%s: want a refusal naming %q, got %v", body, want, err)
		}
	}
}

// TestReconcileContractKeysAreTheKeysTheReadersRead 는 계약 키 목록의 핀임(Manager 결정 4 — 이것만 허용된 구조 시험).
// 목록이 리더가 실제로 읽는 키와 어긋나면(새 필드를 읽는데 목록에 없음, 또는 낡은 키가 남음) 실패함.
// 읽는 키 = 이름 있는 구조체 셋의 json 태그(reflect) ∪ 두 리더 파일의 **모든 함수** 본문의 키 자리(go/ast — 인덱스 리터럴 ·
// reconcileRequiredString/reconcileOptionalString/decodeReconcileEnvelope 의 둘째 인자 · map 리터럴 키 · 함수 안 익명 구조체 태그).
// 200 아닌 본문만 읽는 함수(reconcileDetailSilent — 404 의 error·code·requestId)만 이름으로 뺌 — 새 키 읽기 함수가 생겨도 빠지지 않음.
func TestReconcileContractKeysAreTheKeysTheReadersRead(t *testing.T) {
	read := map[string]bool{}
	for _, typ := range []reflect.Type{reflect.TypeOf(apiEnvelope{}), reflect.TypeOf(apiConditionalOrder{}), reflect.TypeOf(apiConditionalCondition{})} {
		for i := 0; i < typ.NumField(); i++ {
			read[strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]] = true
		}
	}
	non200 := map[string]bool{"reconcileDetailSilent": true}
	argCalls := map[string]bool{"reconcileRequiredString": true, "reconcileOptionalString": true, "decodeReconcileEnvelope": true}
	literal := func(e ast.Expr) {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if v, err := strconv.Unquote(lit.Value); err == nil {
				read[v] = true
			}
		}
	}
	seen := 0
	for _, file := range []string{"reconcile_reads.go", "reconcile_detail_reads.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if non200[fn.Name.Name] {
				seen++
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.IndexExpr:
					literal(n.Index)
				case *ast.CallExpr:
					if id, ok := n.Fun.(*ast.Ident); ok && argCalls[id.Name] && len(n.Args) == 2 {
						literal(n.Args[1])
					}
				case *ast.CompositeLit:
					if _, ok := n.Type.(*ast.MapType); ok {
						for _, elt := range n.Elts {
							if kv, ok := elt.(*ast.KeyValueExpr); ok {
								literal(kv.Key)
							}
						}
					}
				case *ast.StructType:
					for _, field := range n.Fields.List {
						if field.Tag != nil {
							tag, _ := strconv.Unquote(field.Tag.Value)
							read[strings.Split(reflect.StructTag(tag).Get("json"), ",")[0]] = true
						}
					}
				}
				return true
			})
		}
	}
	if seen != len(non200) {
		t.Fatalf("found %d of %d excluded non-200 parsers — the exclusion names a function that no longer exists", seen, len(non200))
	}
	var got, want []string
	for k := range read {
		got = append(got, k)
	}
	seenKey := map[string]bool{}
	for _, k := range reconcileContractKeys {
		if seenKey[k] {
			t.Fatalf("reconcileContractKeys repeats %q", k)
		}
		seenKey[k] = true
		want = append(want, k)
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reconcileContractKeys %v\ndiffers from the keys the readers read %v", want, got)
	}
}
