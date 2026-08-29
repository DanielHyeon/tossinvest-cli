# Function Logic Map: `holdingsCache.peek`

- Source: `internal/console/holdings.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

console-operator-overview가 신설한, **갱신하지 않는** 읽기. 개요 화면의 계약이 브로커 0콜이고 (D4: 콘솔에서 가장 오래 열려 있는 탭이 rate budget을 쓰는 탭이어서는 안 된다), `get`에는 그렇게 하는 모드가 없었다 — unheld get은 TTL이 지나는 순간 갱신한다. 0콜에 도달하는 유일한 우회로가 `get(ctx, now, true, …)`였고 화면은 그것을 '검증 중 — 갱신 보류'로 렌더한다. **검증이 돌고 있지 않은 기계에서 그 문장은 지어낸 것**이고, 지어낸 대상이 '왜 당신의 계좌 숫자가 오래됐는가'다. 리뷰 P1-5가 이 우회로를 지적했고 `peek`이 그 답이다.

한 번도 채워지지 않은 캐시는 `Wired=true, Present=false`로 돌아온다. 그것이 다섯째 미측정 사유 `never_fetched`('아직 읽지 않음')이고 실패한 fetch·보류된 fetch와 **의도적으로 구분된다** — 운영자의 다음 행동이 각각 다르다(포지션 화면을 연다 / 자격증명을 고친다 / 기다린다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` / `c.reader` | nil 허용 | `Options.Holdings` | nil이면 zero 스냅샷 (`Wired=false`) — `seam_unwired` |
| `now` | 주입 시계 | `c.now()` | 나이 계산에만 쓰인다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c == nil || c.reader == nil` | 없음 | zero `holdingsSnapshot` | `TestAnUnwiredGateLimitsSeamOnlyDarkensItsOwnPanel`의 자매 경로 + `brokerReadable`의 `!snap.Wired` 분기 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.mu.Lock` / `defer Unlock` | `snapshotLocked`의 계약(호출자가 뮤텍스를 쥔다)을 만족 | 브로커 호출 없음 | holdings.go:207 |
| `c.snapshotLocked` | 현재 내용 렌더 | 부작용 없음 | holdings.go:214 |
| (금지 바인딩) | 브로커 호출이 **하나도 없다** — 이 함수의 존재 이유 | `TestTheOverviewMakesNoBrokerCall`(호출 수) + `TestTheOverviewReadsTheBrokerCacheWithPeekAndNothingElse`(소스) | overview_test.go:236,319 |

## State mutations and fallbacks

- 캐시를 **읽기만** 한다. `tried`/`attempted`도 건드리지 않는다.
- `Held`/`HeldReason`을 **채우지 않는다**. 그것이 이 함수의 요점 — 관측하지 않은 것을 말하지 않는다.
- 부작용으로 `holdingsSnapshot.Stale()`의 문서 주석이 거짓이 됐다(`Held`와만 함께 참이라던 문장). 같은 change에서 주석을 정정했고 삭제하지 않았다 — 옛 문장이 'unheld-and-old는 불가능'이라는 오독의 근원이었기 때문이다.

## Safety conclusion

- Safe edit boundary: 신설. `get`의 갱신 판정을 건드리지 않고 같은 조립부(`snapshotLocked`)를 공유한다.
- High-risk impact: no (계좌·원장·주문 무접촉, 브로커 호출 0). 다만 이 함수가 없을 때의 대안이 **없는 검증을 근거로 든 사유 문장**이었다는 점에서 운영자 판단 표면이다 — 값이 아니라 사유의 정직성이 이 함수의 산출물이다.
