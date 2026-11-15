# Function Logic Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` (cmd/tossctl)

- Source: `cmd/tossctl/console_test.go`(편집 뒤 720~819행). 편집 전 logic-map 산출물 `../../pre-edit-ast/tossctl--testtheordersseamresolvestheaccountonceandbuildsnosecondclient.ast.json`(720~758행, 분기 수 5개: B1 for 737·B2 if 738·B3 if 742·B4 if 749·B5 if 754).
- AST evidence: `ast.json`(편집 뒤, 분기 23, return 3 — 전부 클로저 안: 726 팩토리 대체, 787·810 `ast.Inspect` 콜백)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 원인·경계: a166 `3dc0c776b` 가 `consoleBroker.resolve()` 를 `resolve(ctx)` 로 바꿈 — 부모 PASS, 원인 커밋 FAIL(`../../boundary/boundary.txt`).
- 판본: 1차(9dc611feb)는 선택자 참조 수만 셈 → gstack 리뷰 1회차(적대 보이스 둘·red team)가 안 쓰는 참조 한 줄로 뚫음을 보임 → 2차(f5e13ea7d, e6fbcb3bb)가 참조·직접 호출 대입·다시 대입·단언 대상을 같이 셈.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `verifyBrokerFactory`(시험이 대체) | 세 번 새로고침에 빌드 1회 | 이 시험 | 빌드 수 ≠ 1 이면 Errorf |
| `console.go` 소스 바이트(`readSource`) | `official.New(` 0회 | 이 시험 | 있으면 Error |
| `(*lazyOrders).Orders` 선언과 본문 | 선언 1; `<수신자>.shared.resolve` 선택자 참조 1; 그 선택자를 Fun 으로 하는 직접 호출이 단일 RHS 이고 첫 LHS 가 `_` 아닌 식별자(broker)인 대입문 1; broker 에 대한 대입(AssignStmt LHS·ValueSpec 이름) 1; `broker.(consoleOrdersReader)` 단언 1 | 이 시험(AST) | 아니면 Errorf(다섯 수를 같이 찍음) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (737) | `for i < 3` 새로고침 세 번 | `seam.Orders` 호출 | — | 자신 |
| B2 (738) | `Orders` 오류 | — | Fatalf | 자신 |
| B3 (742) | `built != 1` | — | Errorf | 자신 — 변이 o2·o3 |
| B4 (749) | `official.New(` 철자 | — | Error | 자신 — 변이 o6 |
| B5 (760) | console.go 파싱 실패 | — | Fatalf | 자신(정상 소스에선 안 탐) |
| B6 (764) | `range file.Decls` | — | — | 자신 |
| B7 (766) | `(*lazyOrders).Orders` 아님 | — | continue | 자신 — 대조군(Holdings 는 여기서 걸러짐) |
| B8 (774) | 첫 순회 type-switch | — | — | 자신 |
| B9·B10 (775·776) | 선택자 텍스트 == want | `resolves++` | — | o1(2)·o7(2) |
| B11·B12·B13 (779·780·781) | 단일 RHS 직접 호출 · LHS 식별자 `_` 아님 | `feeds++`, broker 기록 | — | o5(메서드 값 — 0)·o8(결과 버림 — 0) |
| B14·B15 (791·792) | 둘째 순회 · AssignStmt | — | — | 자신 |
| B16·B17 (793·794) | LHS 가 broker | `rebinds++` | — | o9(2) |
| B18·B19·B20 (798·799·800) | ValueSpec 이름이 broker(안쪽 `var` 가림) | `rebinds++` | — | o10(2) |
| B21·B22 (804·805) | `broker.(consoleOrdersReader)` | `asserted++` | — | o10(가린 변수에 단언 — 이름은 같아 asserted 1, rebinds 2 로 잡힘) |
| B23 (813) | 다섯 수 중 하나라도 ≠ 1 | — | Errorf | o1·o2·o3·o4·o5·o7·o8·o9·o10 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newVerifyServer`·`official.New`·`consoleOrdersSeam`·`newConsoleBroker`·`seam.Orders` | 행동 절반(편집 전과 같음) | `Orders` 오류는 Fatalf | AST |
| `readSource`·`strings.Contains` | `official.New(` 금지(편집 전과 같음) | — | AST |
| `parser.ParseFile`·`ast.Inspect`·`types.ExprString` | AST 단언 | 파싱 실패 Fatalf | AST |

## State mutations and fallbacks

- 패키지 변수 `verifyBrokerFactory` 를 시험 동안 바꾸고 `t.Cleanup` 으로 되돌림(편집 전과 같음). 그 밖의 상태 변경 없음.

## 편집

- 편집 전 B5(754) `!strings.Contains(src, "l.shared.resolve()")` 하나를 B5~B23 으로 바꿈(design D11). 철자 `l.shared.resolve(ctx)` 는 `(*lazyHoldings).Holdings` 에도 있어 철자만 고치면 Orders 를 못 봄 — 변이 `o1-second-resolve-S`·`o4-unlocked-cache-S` 가 S 판본(철자만 고친 판본)에서 PASS.
- 남는 한계: 이름 기반 대조라 broker 를 함수 인자로 다른 헬퍼에 넘겨 거기서 바꿔치기하는 모양은 이 단언 밖임. 그 경로와 동시 첫 사용은 행동 시험 `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`(console_test.go) 와 `consoleBroker.resolve` 의 잠금이 짐.

## Safety conclusion

- Safe edit boundary: 시험 함수 본문의 소스 절반. 운영 `console.go` 무편집.
- High-risk impact: 없음(시험). 콘솔 조회 seam 의 계좌 조회 1회 성질을 더 좁게 지킴.
