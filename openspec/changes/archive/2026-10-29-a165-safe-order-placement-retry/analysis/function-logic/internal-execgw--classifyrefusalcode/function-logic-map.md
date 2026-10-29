# Function Logic Map: `classifyRefusalCode`

- Source: `internal/execgw/refusal_code.go` (53-78)
- Qualified function: `classifyRefusalCode`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 4e885e54a90205d2…)
- AST evidence: `ast.json` — AST branches 6, 반환 6, 호출 4
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

확정 거절 code 목록(`opposite-pending-order-exists`)만 확정으로. `already-processing` 은 목록에 없어 B6(74) 판정 없음 → 종전 경로(409 AMBIGUOUS). B3(63) 두 자리 모순이면 모호 강제. 주문 종류를 가리지 않으므로 a165 는 409 AP 를 이 목록에 **넣지 않고** 취소에만 별도 판정을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| err | official 오류 | classifyMutation | APIError 아니면 B1 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 56:3, 60:3, 64:3, 71:3, 75:3, 77:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 55:2` | if !errors.As(err, &apiErr) { | 아래 「State mutations」 | 소스 줄 55 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 59:2` | if !ok { | 아래 「State mutations」 | 소스 줄 59 이하 | 실행됨 — 패키지 스위트 count=1 |
| B3 | `if at 63:2` | if top.present && nested.present && !strings.EqualFold(top.value, nested.value) { | 아래 「State mutations」 | 소스 줄 63 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `if at 67:2` | if !code.present { | 아래 「State mutations」 | 소스 줄 67 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 70:2` | if !code.present { | 아래 「State mutations」 | 소스 줄 70 이하 | 실행됨 — 패키지 스위트 count=1 |
| B6 | `if at 74:2` | if !listed { | 아래 「State mutations」 | 소스 줄 74 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`refusalCodes`.

AST 호출 이름 전체: `errors.As`, `refusalCodes`, `strings.EqualFold`, `strings.ToLower`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
