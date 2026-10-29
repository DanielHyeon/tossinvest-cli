# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go:1483-1596`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| proposal/action | 보호 또는 익절 발의 | journal exit state | 보호 가격 선택은 새 판만 읽고 발의를 거부하지 않음 |
| quantity | 양의 canonical 수량 | policy + confirmed floor | floor 오류/0은 기존 해제 규칙 유지 |
| observed/baseline | decimal 또는 빈 문자열 | quote/state | 기존 `sellIntent` B1~B4가 검증 |
| lower-limit board | nil 또는 immutable snapshot | a087 새 board | nil·무효·첫 시도 아님은 기존 지정가로 fallback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | floor 산술 오류 | 미발송 줄 기록 | 오류 반환 | floor 실패 가격 출처 baseline/observed |
| B2 | floor가 0 수량 | proposal release·미발송 줄 | release 결과 | 보호 가격 선택이 B2를 바꾸지 않음 |
| B3 | Guardian 거절 | alert·release·미발송 줄 | release 결과 | 하한가 판이어도 Guardian 경로 동일 |
| B4 | intent attach 실패 | 미발송 줄 | wrapped error | 제출 전 price 결정은 broker GET 0 |
| B5 | sellIntent 거절 | alert·release·미발송 줄 | release 결과 | empty observed/baseline 기존 실패 |
| B6 | Place 결과 switch | result별 처리 | 아래 case | limit type·수량 불변 |
| B7 | CONFIRMED | 확정 로그·줄 | nil | 하한가 price/intents.price |
| B8 | IN_DOUBT | unresolved 줄 | nil | 첫 시도 latch가 재발의 차단 |
| B9 | UNRESOLVED_IN_DOUBT | event 변경 | B8 계속 | 같은 latch |
| B10 | symbol in flight | delay·cancel release | release 결과 | broker 전 경로는 latch하지 않음 |
| B11 | attempt가 비수용 확정 아님 | unresolved 줄 | wrapped error | first attempt latch 유지 |
| B12 | err 없음 | synthetic error | B11 계속 | 오류 문구만 기존 유지 |
| B13 | default refusal | alert·release·줄 | release 결과 | FAILED_CONFIRMED 뒤 fallback |
| B14 | failed confirmed | rejected event | B13 계속 | 거부 줄의 price_source |
| B15 | outcome detail 없음 | err 문자열 선택 | B13 계속 | 로그 whitelist 유지 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `applyFloor` | 기존 RECONCILE 상한 | 오류는 B1 | AST B1 |
| `exitLimitPrice` | 한 번만 가격과 출처 선택 | total; broker·network 호출 없음 | a087 insertion point before line construction |
| `exitOrderLineFor` | 전송 가격과 원 관측가를 줄에 기록 | 오류 없음 | AST B1~B15 |
| `sellIntent` | 기존 LIMIT intent 조립 | B5 기존 계약 | AST B5 |
| `Submit.Place` | Guardian 보장 제출 | 결과 switch B6~B15 | AST B6 |

## State mutations and fallbacks

- 새 선택값은 `sellIntent`의 observed 인자로만 전달하여 기존 LIMIT 조립 본문을 수정하지 않음.
- `Place` 호출이 끝나는 즉시 보호 첫 시도 latch를 닫음. broker 전에 끝난 B1~B5는 latch하지 않음.
- 판·원장 재구성이 실패하면 관측가, 그마저 비면 기준선으로 fallback하여 현재 거절 사유만 남김.

## Safety conclusion

- Safe edit boundary: price 선택 1회와 line/intent 인자만 변경.
- High-risk impact: yes — 보호 청산 가격. order type·수량·순서·제출 전 broker GET은 변경 금지.
