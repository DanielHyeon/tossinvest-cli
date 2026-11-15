# Function Logic Map: `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` (cmd/tossctl)

- Source: `cmd/tossctl/console_test.go`(편집 뒤 720~782행). 편집 전 logic-map 산출물 `../../pre-edit-ast/tossctl--testtheordersseamresolvestheaccountonceandbuildsnosecondclient.ast.json`(720~758행, 분기 5: B1 for 737·B2 if 738·B3 if 742·B4 if 749·B5 if 754).
- AST evidence: `ast.json`(편집 뒤, 분기 9, return 2 — 둘 다 클로저 안: 726 팩토리 대체, 774 `ast.Inspect` 콜백)
- Risk scan: `risk-pattern-report.md`(일치 0)
- 원인·경계: a166 `3dc0c776b` 가 `consoleBroker.resolve()` 를 `resolve(ctx)` 로 바꿈 — 부모 PASS, 원인 커밋 FAIL(`../../boundary/boundary.txt`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `verifyBrokerFactory`(시험이 대체) | 세 번 새로고침에 빌드 1회 | 이 시험 | 빌드 수 ≠ 1 이면 Errorf |
| `console.go` 소스 바이트(`readSource`) | `official.New(` 0회 | 이 시험 | 있으면 Error |
| `(*lazyOrders).Orders` 선언과 본문 | 선언 1, `<수신자>.shared.resolve` 선택자 참조 1 | 이 시험(AST) | 아니면 Errorf(수 둘을 같이 찍음) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (737) | `for i < 3` 새로고침 세 번 | `seam.Orders` 호출 | — | 자신 |
| B2 (738) | `Orders` 오류 | — | Fatalf | 자신 |
| B3 (742) | `built != 1` | — | Errorf | 자신 — 변이 o2·o3 |
| B4 (749) | `official.New(` 철자 | — | Error | 자신 — 변이 o6 |
| B5 (758) | console.go 파싱 실패 | — | Fatalf | 자신(정상 소스에선 안 탐) |
| B6 (762) | `range file.Decls` | — | — | 자신 |
| B7 (764) | FuncDecl 아님·이름 ≠ Orders·수신자 ≠ `*lazyOrders`·수신자 이름 없음 | — | continue | 자신 — 대조군(Holdings 는 여기서 걸러짐) |
| B8 (771) | 선택자 텍스트 == `<수신자>.shared.resolve` | `resolves++` | — | 자신 — 변이 o1(2)·o2·o3·o4(0)·o5(1) |
| B9 (777) | `found != 1 \|\| resolves != 1` | — | Errorf | 자신 — 변이 o1·o2·o3·o4 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newVerifyServer`·`official.New`·`consoleOrdersSeam`·`newConsoleBroker`·`seam.Orders` | 행동 절반(편집 전과 같음) | `Orders` 오류는 Fatalf | AST |
| `readSource`·`strings.Contains` | `official.New(` 금지(편집 전과 같음) | — | AST |
| `parser.ParseFile`·`ast.Inspect`·`types.ExprString` | 새 AST 단언 | 파싱 실패 Fatalf | AST |

## State mutations and fallbacks

- 패키지 변수 `verifyBrokerFactory` 를 시험 동안 바꾸고 `t.Cleanup` 으로 되돌림(편집 전과 같음). 그 밖의 상태 변경 없음.

## 편집

- 편집 전 B5(754) `!strings.Contains(src, "l.shared.resolve()")` 하나를 B5~B9 로 바꿈. 이유(design D11): 철자 `l.shared.resolve(ctx)` 는 `(*lazyHoldings).Holdings` 에도 있어(파일 안 2회), 철자만 고치면 Orders 가 공유 resolver 를 버려도 통과함 — 변이 `o4-unlocked-cache-S`·`o1-second-resolve-S` 가 S 판본(철자만 고친 판본)에서 PASS 로 그것을 잼.
- 다른 분기(B1~B4)는 편집 전 B1~B4 와 같은 항목·같은 행.

## Safety conclusion

- Safe edit boundary: 시험 함수 본문의 소스 절반 하나. 운영 `console.go` 무편집.
- High-risk impact: 없음(시험). 콘솔 조회 seam 의 계좌 조회 1회 성질을 더 좁게 지킴.
