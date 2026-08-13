# Function Logic Map: `Collector.Collect`

- Source: `internal/reconcile/snapshot.go` (245-295)
- AST evidence: `ast.json` — AST 기준 branches **8** / returns 6 / calls 17 / defers 0 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `9cb72654892fda5456759435653fc5bf01f9f616e6a4dd7d2dbf87f916497c90`

**a102 §1(D1b)이 편집한 함수다.** 편집은 원인 wrap 동사 네 개(`%v` → `%w`)와 주석뿐이다.

## 두 판 — 구조는 안 바뀌었다

| | 1판 (편집 전) | **2판 (편집 후, 이 문서)** |
|---|---|---|
| 위치 | `snapshot.go:232-282` | **`:245-295`** (주석 13줄만큼 밀렸다) |
| 분기 | 8 | **8** (동일) |
| 이탈 | 6 | **6** (동일) |
| 호출 | 17 | **17** (동일) |
| source SHA-256 | `9d93f80e…` | `9cb72654…` |

**분기·이탈·호출이 하나도 안 변한 것이 D1b의 성질이다** — 바뀐 것은 오류가 실어 나르는
정체뿐이고, 제어 흐름은 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Orders` | non-nil | `validate()`(`:340`) | B1 `:246` — 오류를 **wrap 없이** 그대로 돌려준다 |
| `c.Positions` | non-nil | 같음 | 같음 |
| `c.Balance` | non-nil | 같음 | 같음 |
| `c.Currencies` | 최소 1개 | 같음 | 같음 — 현금 없는 스냅샷은 스냅샷이 아니다 |
| `c.MaxPages` | >0 (0이면 50) | B2 `:251` | 상한을 넘으면 `ScanOrders`가 오류 |
| `clk` | non-nil | `c.clock()`(`:353`) | 없음 |

> **불변식 하나가 이 함수 전부를 지배한다**: "or none". 오류 경로 여섯이 모두
> `Snapshot{}`을 돌려주고, **부분 스냅샷은 존재할 수 없는 값**이다. a102는 이 불변식을
> 건드리지 않는다 — 바꾸는 것은 **오류가 무엇을 실어 나르는가**뿐이다.

## Branches and early returns

`ast.json`의 열거를 그대로 옮긴다.

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:246` | `c.validate() != nil` | 없음 | `:247` `Snapshot{}, err` (**sentinel wrap 없음**) |
| B2 | `:251` | `maxPages <= 0` | `maxPages = 50` | — |
| B3 | `:260` | `ScanOrders` 오류 | 없음 | `:261` `%w: walking the open-order list: %w` |
| B4 | `:263` | `range raws` | 반복 | — |
| B5 | `:265` | `parseBrokerOrder` 오류 | 없음 | `:266` `%w: %w` |
| B6 | `:275` | `c.holdings` 오류 | 없음 | `:276` `%w: sweeping the holdings: %w` |
| B7 | `:281` | `range c.Currencies` | 반복 | — |
| B8 | `:283` | `BuyingPower` 오류 | 없음 | `:284` `%w: reading the %s buying power: %w` |

Returns: `:247` · `:261` · `:266` · `:276` · `:284` · `:294` (AST 6개).

> 편집 전에는 `%w`가 `ErrPartialSnapshot` 자리에만 있고 원인 자리에는 없었다. 그래서
> `errors.Is(err, ErrPartialSnapshot)`는 성립하고 `errors.Is(err, official.ErrRateLimited)`는
> **B3·B5·B6·B8 어디서도 성립하지 않았다.** 이제 둘 다 성립한다 — 다중 `%w`는 Go 1.20+이고
> go.mod는 1.25다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.validate` | 배선 검사 | B1 — 재시도 없음 | `:246` |
| `clk.Now` | as-of / completed 스탬프 | 실패 없음 | `:255` · `:293` |
| `execgw.ScanOrders` | 미체결 목록 전 페이지 | **재시도 없음.** 429는 `official.ErrRateLimited`로 올라온다 | `:259` |
| `parseBrokerOrder` | 페이로드 해독 | 못 읽으면 스냅샷 폐기 | `:264` |
| `c.holdings` | 보유 1회 읽기 | 재시도 없음 — 같은 429 표면 | `:274` |
| `c.Balance.BuyingPower` | 통화별 현금 | 재시도 없음 — 같은 429 표면 | `:282` |
| `fmt.Errorf` ×4 | `ErrPartialSnapshot` + **원인** wrap | 이제 원인의 정체가 사슬에 남는다 | `:261` · `:266` · `:276` · `:284` |

**프로덕션 호출자는 셋이다** (실측 `rg -n '\.Collect\(' internal/ cmd/`, non-test):

| 호출자 | 오류 처리 | D1b의 영향 |
|---|---|---|
| `internal/reconcile/recovery.go:372` | 429면 백오프, 그 밖은 즉시 실패 | **a102 §1이 여기를 고쳤다** |
| `internal/app/engine/reconcileloop.go:491`·`:505` | `cycle.Err`에 담고 다음 주기가 받는다 | 무변 — 아무도 `errors.Is`로 분류하지 않는다 |
| `internal/flatten/liquidate.go:596` | `%w`로 감싸 올린다 | 무변 |

`official.Err*`를 `Collect` 오류에 대고 검사하는 소비자는 a102 이전에 **한 곳도 없었다**
(실측 `rg -n 'official\.Err' internal/app/engine/ internal/flatten/` → 0건). 따라서
`%w` 추가는 **판정을 넓히기만 하고 좁히지 않는다** — 보수 방향이다.

## State mutations and fallbacks

- `Collector`의 필드를 쓰지 않는다. 상태는 지역 `snap` 하나이고 오류 경로는 그것을 버린다.
- fallback은 하나뿐이다: `c.holdings`가 `RawPositionsReader`를 우선한다(`:303`).
  같은 엔드포인트 1회 요청이므로 §0.4 예산은 동일하다.
- **a102가 바꾼 상태는 없다.** 바뀐 것은 반환 오류의 `errors.Is` 사슬뿐이다.

## Safety conclusion

- Safe edit boundary: `:261` · `:266` · `:276` · `:284`의 **원인 자리 동사 하나씩.**
  포맷 문자열의 나머지·인자 순서·반환값은 바뀌지 않았다.
- High-risk impact: **yes** — 대사 스냅샷 수집. 다만 편집은 오류의 **정체 사슬**만 넓힌다.
- 회귀 방어: `ErrPartialSnapshot` 판정 무회귀를 기존 4개 지점
  (`snapshot_test.go:186`·`:201`·`:214`·`:231`)과 `internal/app/engine/reconcileloop_test.go:717`이
  이미 지고 있고, a102가 `TestCollectStillReportsAPartialSnapshot`으로 한 겹을 더했다.
- **남은 구멍 하나**: B5(`:266`)의 wrap은 뮤테이션이 살아남는다. 이유는
  `branch-test-map.md`에 적었다 — 그 원인을 만드는 `parseBrokerOrder`(`compare.go:640`)가
  자기 원인을 `%v`로 지우므로, 이 자리에서 되살릴 정체가 애초에 없다.
