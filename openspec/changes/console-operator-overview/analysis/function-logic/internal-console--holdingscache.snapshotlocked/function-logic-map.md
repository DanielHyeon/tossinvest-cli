# Function Logic Map: `holdingsCache.snapshotLocked`

- Source: `internal/console/holdings.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`get`과 `peek`이 공유하는 스냅샷 조립부. console-operator-overview가 `get`의 본문에서 뽑아냈다. 호출자가 뮤텍스를 쥐고 있고, **갱신이 있었는지 없었는지를 결정하는 것도 호출자**라는 것이 계약이다. `Held`/`HeldReason`을 여기서 채우지 않는 것이 추출의 요점 — 그래야 `peek`이 그 어휘 없이 같은 값을 만들 수 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.mu` | 호출자가 이미 잠갔다 | `get`/`peek` | 잠그지 않고 부르면 데이터 경쟁 — 패키지 내부 함수이고 호출자는 둘뿐이다 |
| `c.present` / `c.rows` / `c.at` / `c.lastErr` | `refreshLocked`가 쓴 값 | 브로커 응답 | 해당 없음 |
| `now` | 주입 시계 | 호출자 | `at`보다 이르면 나이를 0으로 고정한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c.present` | `snap.Age = now.Sub(c.at)` | 없음(계속) | `TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing`(Age=90s), `TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt`(Present=false) |
| B2 | `snap.Age < 0` | `snap.Age = 0` | 없음(계속) | 역행 시계 방어 — 주입 시계를 되감는 직접 테스트는 없다. 음수 나이가 화면에 나가지 않는다는 것만 보장한다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append([]domain.Position(nil), c.rows...)` | 행 복사 | 호출자가 캐시 내부 슬라이스를 잡지 못하게 한다 | holdings.go:218 |
| `now.Sub(c.at)` | 나이 | 순수 계산 | ast.json calls |
| (금지 바인딩) | 브로커·원장·파일 무접촉 | 메모리 읽기만 한다 | ast.json calls |

## State mutations and fallbacks

- 캐시를 읽기만 하고 아무 필드도 쓰지 않는다.
- `Wired: true`를 무조건 세운다 — nil 판정은 호출자(`get` B1, `peek` B1)가 이미 했다.
- 행 슬라이스를 복사하므로 화면이 캐시의 내부 배열을 공유하지 않는다.

## Safety conclusion

- Safe edit boundary: 신설(추출). `get`이 만들던 값과 동일하되 `Held`/`HeldReason` 두 필드를 뺐다.
- High-risk impact: no (계좌·원장·주문 무접촉의 순수 조립). `Wired=true`를 무조건 세우므로 nil 판정을 호출자에서 빼면 미배선이 '읽지 않음'으로 보이게 되는 것이 이 함수의 유일한 오용 방식이다.
