# Function Logic Map: `TestTheLedgerCanSayWhichBrokerOrdersTheEngineIssued`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L276-297). `ReadOnly.BrokerOrderIDs`의 세 가지 계약을 한 번에 잰다:
정렬, 중복 제거, 빈 id 제외. 픽스처는 attempt 4건 — `ord-b`, `ord-a`, `ord-a`(재시도),
`''`(ack 없음) — 이고 기대값은 `["ord-a", "ord-b"]`다.

이 테스트가 없으면 화면은 "엔진 발주" 열을 빼거나 지어내고, 지어낸 "수동" 라벨은
운영자가 **엔진이 놀고 있다**고 결론 내리는 근거가 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal 파일 | `t.TempDir()/DBFileName` | `openTestJournalAt`(최신 스키마) | `t.Fatalf` |
| intent 1건 | `insertIntent(t, j, "intent-1")` | 테스트 헬퍼 | FK 위반 시 실패 |
| attempt 4건 | `insertAttemptWithBrokerOrder` ×4 | 위 헬퍼 | 동상 |
| 읽기 handle | `openTestReadOnly(t, path)` (`mode=ro`) | `readonly.go` | 동상 |

불변식: 기대값을 `reflect.DeepEqual`로 **순서까지** 비교한다 — `ORDER BY broker_order_id`가
빠지면 실패한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L290) | `BrokerOrderIDs`가 오류 | 없음 | `t.Fatalf("BrokerOrderIDs: %v")` | 자체 실행 |
| B2 (if, L294) | 결과 ≠ `["ord-a","ord-b"]` | 없음 | `t.Errorf` — 정렬·중복 제거·빈 id 제외 중 하나가 깨졌다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournalAt` / `openTestReadOnly` | 쓰기·읽기 두 handle | 실패 시 `t.Fatalf` | ast.json calls |
| `insertIntent` / `insertAttemptWithBrokerOrder` | 픽스처 | 동상 | ast.json calls |
| `ReadOnly.BrokerOrderIDs` | 측정 대상 | 오류 그대로 단언 | ast.json calls |
| `reflect.DeepEqual` | 순서 포함 비교 | — | ast.json calls |

## State mutations and fallbacks

- `t.TempDir()` 안의 journal만 만든다. 실계좌·브로커·네트워크 무접촉.

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산. 기존 테스트 무변경.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉) — 다만 이 테스트가 **없어지거나
  약해지는 것**은 High-risk다. 이것이 화면의 발주 주체 귀속을 재는 유일한 측정이다.
