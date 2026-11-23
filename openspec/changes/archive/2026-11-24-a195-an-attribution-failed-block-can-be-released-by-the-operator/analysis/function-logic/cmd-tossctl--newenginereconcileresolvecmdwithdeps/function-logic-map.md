# Function Logic Map: `newEngineReconcileResolveCmdWithDeps` (편집 뒤 판)

- Source: `cmd/tossctl/engine_reconcile.go` (76-110), 작업 트리 sha `eec1cdcbd9ab…`. 편집 전 `../../pre-edit-ast/cmd-tossctl--newenginereconcileresolvecmdwithdeps.ast.json`(76-105). 재번호 `renumber.txt`(분기 0 → 0, 표 없음).
- AST evidence: `ast.json`(분기 0, return 2 — return@99 는 RunE 클로저, return@109 은 함수). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file cmd/tossctl/engine_reconcile.go --func newEngineReconcileResolveCmdWithDeps`.
- Risk scan: `risk-pattern-report.md`(일치 0). 커버리지: `../../coverage/pre-edit-runEngineReconcileResolve-and-cmd-blocks.txt`(76~105 전 블록 hit=1).
- 편집(Manager 결정 Q5): `Long` 도움말(:81-94)을 원인 중립으로 — 받는 원인(수량 불일치, 종목 범위 귀속 실패)·그 밖의 원인은 전체 거절·해제 조건이 「진입을 막는 차이 없음」(외부 보유·외부 주문·평단은 거절 사유 아님)임을 적음(구현 적대 리뷰 P2-4 반영). 플래그·Annotations·RunE 무변경. a166 census 키 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root`, `deps` | 그대로 RunE 로 전달 | 호출자 | — |
| 불변식 1 | `Annotations{"source":"official"}`, `mutating` 없음 | `TestEngineReconcileResolveIsRegisteredAsOfficialReadOnly` | — |
| 불변식 2 | `--confirm`·`--operator`·`--note` 필수 | 같은 시험 · `a098_the_operator_command_names_a_person_test.go:143` | — |
| 불변식 3 | 파일 텍스트에 금지 토큰(`PlaceOrder`·`CancelOrder`·`AmendPendingOrder`·`ModifyOrder` 등, 대소문자 구분) 없음 | `TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor` | 도움말에 그 토큰을 쓰면 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | — | cobra 명령 생성, 플래그 3·필수 표시 3 | return@109 cmd | 편집은 문자열 리터럴만 |
| (return@99, RunE) | — | `runEngineReconcileResolve` 호출 | 그 결과 | 편집 무관 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 도움말 정리 | — | AST |
| `cmd.Flags().BoolVar/StringVar`, `MarkFlagRequired` | 플래그 | 오류 무시(`_ =`) — 기존 | AST |
| live config | 없음 | — | — |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Edit boundary(실제): `Long` 리터럴 안의 문장만. 금지 토큰을 쓰지 않는다(불변식 3).
- High-risk impact: no(도움말 문장) — 단 대사 해제 명령의 운영자 설명이라 거짓이 되면 운영 판단을 그르친다.
