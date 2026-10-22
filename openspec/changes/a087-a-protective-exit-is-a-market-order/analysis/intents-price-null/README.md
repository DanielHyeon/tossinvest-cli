# a087 tasks 3.7 — `intents.price` NULL 소비부 전수 열거

- 측정 순간: 2026-10-22, HEAD `0686c0f0a`(브랜치 `feat/a112-four-family-runtime`), codegraph **1.6.0**(`codegraph sync .` 직후)
- 모집단: 저장소의 **비시험 Go 파일 969 개**(`openspec/`·`.git`·`vendor`·`.sdd`·`.codegraph` 제외). Python 도구·SQL 파일·
  외부 소비자(사람의 sqlite3 조회 등)는 **안 봤음** — 이 열거는 그 채널에 대해 아무것도 말하지 않음.
- 생산자 쪽(참고): `Prepare` 가 `nullableString(req.Intent.Price)` 로 `""` → NULL 을 씀(`durability.go:403`·`:802`).
  취소 intent 도 같은 길 — `Gateway.Cancel` 의 `price` 0 → `priceString(0)` = `""` → NULL(`execgw/gateway.go:1045`·`:1099`).

## 방법 (3차 재리뷰 A-P3-3 그대로)

| 단계 | 도구 | 산출 |
| --- | --- | --- |
| ① SQL 문자열 | `census.go`(`//go:build ignore`, go/parser) — 최상위 선언 하나의 문자열 리터럴을 이어 붙여 `intents` 를 읽고(FROM/JOIN/쉼표 조인) `price` 낱말 **또는** `SELECT *`·`별칭.*` 를 담는 선언을 셈 | `census-output.txt` |
| ② 접근자 호출자 | CodeGraph `codegraph callers`(심볼 단위) ∪ `flow.go`(go/ast — 접근자 결과 변수와 range 원소를 따라가 모든 등장을 「필드 읽기 / 통째 전달」로 분류, 통째 전달 받는 함수의 매개변수를 고정점까지 다시 따라감) | `codegraph-callers.txt` · `flow-output.txt` |
| 함수 내부 주장 | `tools/logic-map` AST 번들(`ast/*.ast.json`, 생성 시 `source_sha256` 일치 확인) | `ast/` |

재생성(저장소 루트): `go run ./openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/intents-price-null/census.go`,
같은 디렉터리 `flow.go`. AST 는 `go build -o /tmp/logicmap ./tools/logic-map && /tmp/logicmap --file <f> --func <F>`.

## ① 결과 — SQL 자리 3 (그중 price 를 실제로 읽는 것 2)

| # | 선언 | price 읽기 | 판정 |
| --- | --- | --- | --- |
| S1 | `const intentSelect`(`internal/journal/durability.go:698`) | 열 이름 `price` | **읽음** — `scanIntent` 가 `sql.NullString` 으로 받아 `.String`(NULL → `""`) (`ast/journal--scanintent.ast.json`: 분기 B1 ErrNoRows·B2 오류, 그 뒤 무조건 `in.Price = price.String`) |
| S2 | `Journal.LiveOrdersForSymbol`(`internal/journal/fills.go:1854`) | `coalesce(i.price, '') AS price`(`:1864`) → `c.price`(`:1880`) | **읽음** — SQL 에서 `""` 로 바뀐 뒤 `string` 으로 Scan(`ast/journal--journal.liveordersforsymbol.ast.json` B4) |
| — | `confirmedFillOwners`(`fills.go:565`) `[select-star]` | `c.*`·`SELECT *` 가 도는 CTE(`all_candidates`)의 열에 price 없음 | **안 읽음** — not-applicable(별표가 덮는 열 집합에 price 가 없음, 같은 선언의 SELECT 목록 실측) |

## ② 결과 — 접근자 호출 자리 12 (비시험)

CodeGraph 와 텍스트(`flow.go`)의 차이 — 숨기지 않음:

- `LiveOrdersForSymbol` 의 `reconcileFloor.localOpenSells`(`internal/app/engine/exitwiring.go:266`)를 **CodeGraph 는 못 찾음**(비시험 호출자 2: `noteHeldProposal`·`clearTheSymbol`), `flow.go` 는 찾음 → 합집합에 넣음.
- CodeGraph 는 `LookupIntent` 의 호출자로 인터페이스 선언 `attempt_thaw_command.go:39` 을 적음 — 실제 호출식은 `releaseThawedProposal`(`:106`, `repo.LookupIntent`). 합집합은 호출식 쪽.
- `scanIntent` 호출자 2(`Prepare`·`LookupIntent`)는 둘이 일치.

| 자리 | 접근자 | `.Price` 를 읽는가 (flow-output) | NULL 처분 |
| --- | --- | --- | --- |
| `Journal.Prepare` (`durability.go:389`) | scanIntent | 통째 비교 `existing != req.Intent`(`:392`, `ast/journal--journal.prepare.ast.json` B6) | **시험** `TestA087ReprepareOfAMarketIntentIsNotAMutation` — NULL 이 `""` 로 대칭 왕복하지 않으면 시장가 intent 의 둘째 attempt(취소)가 `ErrIntentMutated` 로 막힘 |
| `Journal.LookupIntent` (`durability.go:694`) | scanIntent | 반환만 | **시험(기존)** `TestPrepareStoresEmptyPriceAsNull`(`durability_test.go:101`) — 저장 NULL · 읽기 `""` |
| `ExitObserver.clearTheSymbol` (`exitloop.go:1630`) | LiveOrdersForSymbol | **예** — `workingOrderPrice(order.Price)`(`:1681`) | **시험** `TestA087AnEmptyWorkingOrderPriceIsZeroNotAFailure` + `TestA087LiveOrdersForSymbolReadsANullPriceAsEmpty`. `workingOrderPrice` B1(`exit_held_proposal.go:112`) 빈 값 → `0, nil`; 이것이 실패면 `clearTheSymbol` B10(`exitloop.go:1682` `qerr != nil \|\| perr != nil`)이 청소를 실패시켜 보호 청산을 보류함. 0 은 `OrderRef.Price` → `orderTypeFor(0)`="MARKET"(`gateway.go:1088`) 로 취소 intent 에 실림 |
| `ExitObserver.noteHeldProposal` (`exit_held_proposal.go:161`) | LiveOrdersForSymbol | 아니오(Side·OrderID·범위 필드만) | not-applicable — price 미사용. 행 Scan 자체는 S2 시험이 덮음 |
| `reconcileFloor.localOpenSells` (`exitwiring.go:266`) | LiveOrdersForSymbol | 아니오(Side·Quantity) | not-applicable — price 미사용. Scan 은 S2 시험 |
| `ExitObserver.followExitFills` (`exit_order_log.go:295`) | LookupIntent | 아니오(범위 필드) | not-applicable — price 미사용 |
| `PositionPolicyCommandService.releaseThawedProposal` (`attempt_thaw_command.go:106`) | LookupIntent | 아니오(AccountRef) | not-applicable — price 미사용 |
| `Gateway.attemptTargets` (`execgw/gateway.go:834`) | LookupIntent | 아니오(Symbol·Market) | not-applicable — price 미사용 |
| `Resolver.Resolve` (`execgw/indoubt.go:243`) | LookupIntent | 통째 전달 4 → 아래 행 | 아래 |
| ↳ `newMatcher` (`indoubt.go:652`, Resolve·resolveAmend 경유) | — | **예** — `parseOptionalDecimal(intent.Price)`(`:657`) | **시험** `TestA087AnEmptyIntentPriceMatchesAnyBrokerPrice`. `parseOptionalDecimal` B1(`:627`) 빈 값 → 0; `matcher.matches` B4(`:697` `m.price > 0 && …`) 가 꺼져 가격 와일드카드 — 종목·방향·수량(B1~B3)·제출 창(B5·B6)은 그대로 거름(시험이 수량 불일치 거절과 가격 있는 대조군을 같이 봄) |
| ↳ `absenceCorroborated` (`indoubt.go:445`) → `intentNotional` (`:545`) | — | **예** — `strings.TrimSpace(intent.Price)`(`:550`) | **시험** `TestA087AnEmptyIntentPriceHasNoNotional`. `intentNotional` B2(`:551`) 파싱 실패 → 0; `absenceCorroborated` B8(`:492` `notional > 0 && …`) 매수가능금액 교차 확인이 꺼짐 — 부재 확증은 그 앞 B7(보유수량 불변)에 기댐. 매도에서 B8 이 무엇을 거르는지는 [미측정] → `issues.md` I-R10 |
| ↳ `windowContaminated`·`resolveCancel`·`resolveAmend`·`parkAlert` | — | 아니오(`resolveCancel` 은 매개변수 미사용, `resolveAmend` 는 `newMatcher` 로만 전달) | not-applicable — price 미사용(`newMatcher` 는 위 행) |
| `Gateway.ReplayInDoubt` (`execgw/replay.go:237`) | LookupIntent | 아니오(Symbol·Market) + `parkAlert` | not-applicable — price 미사용. 재생 전송은 `mutation_attempts.wire_body` 를 씀(intents.price 아님 — 재생 경로는 tasks 4.6) |
| `Journal.blockedFor` (`journal/recovery.go:130`) | LookupIntent | 아니오(Market·Symbol) | not-applicable |
| `Recovery.confirmAcked` (`reconcile/acked_boot.go:46`) | LookupIntent | 아니오(Symbol) | not-applicable |
| `Recovery.blockedSymbol` (`reconcile/recovery.go:424`) | LookupIntent | 아니오(Market·Symbol) | not-applicable |

**결론**: ①∪② 에서 NULL 가격이 닿는 자리는 SQL 2 + Go 소비 4(`Prepare` 비교·`workingOrderPrice`·`parseOptionalDecimal`·`intentNotional`)이고,
넷 다 빈 값을 실패가 아닌 「가격 없음」(0·와일드카드·대칭 비교)으로 처리함 — 시험 6 개(신규 5 · 기존 1)로 고정.
변이 7/7 CAUGHT(대조군 7/7 PASS) — `../pin-mutations/mutants-2026-10-22.log` M12~M16(3.7 분), M10·M11(3.5·3.6 분).
production 편집 0.
