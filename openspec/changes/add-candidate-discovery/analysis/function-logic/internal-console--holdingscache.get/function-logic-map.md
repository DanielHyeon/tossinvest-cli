# Function Logic Map: `holdingsCache.get`

- Source: `internal/console/holdings.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

브로커 캐시의 **갱신하는** 읽기. 이 프로세스가 계좌에 요청을 보내는 유일한 경로다. 이 change가 바꾼 것은 스냅샷 조립부를 `snapshotLocked`로 뽑고 `Held`/`HeldReason`을 그 위에 덮는 형태로 바꾼 것이며, 판정 분기 둘(B1·B2)은 그대로다. 추출의 이유는 `peek`이 같은 조립을 **hold 어휘 없이** 써야 했기 때문이다(리뷰 P1-5).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` / `c.reader` | nil 허용 | `Options.Holdings` | 둘 중 하나라도 nil이면 `Wired=false` 스냅샷 — 화면은 `seam 미배선`으로 렌더한다 |
| `hold` | bool | `Console.verifyHold` — in-process run 또는 runlock 마커 신선도 | true면 나이와 무관하게 아무것도 가져오지 않는다 |
| `holdReason` | hold의 사유 문장 | 같은 곳 | hold=false면 빈 문자열 |
| `now` | 주입 시계 | `c.now()` | 해당 없음 |
| `c.ttl` | ≥15초(스펙 하한), 실제 30초 | `holdingsTTL` | 0 이하면 생성자가 기본값으로 되돌린다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c == nil || c.reader == nil` | 없음 | `{Held, HeldReason}`만 실린 zero 스냅샷 (`Wired=false`) | `TestAnUnwiredOrdersSeamSaysSoAndLeavesEveryOtherScreenWorking`의 자매 케이스 + `TestThePositionsScreenRendersWithEitherSourceMissing` |
| B2 | `!hold && (!c.attempted || now.Sub(c.tried) >= c.ttl)` | `refreshLocked` — 브로커 1콜, `tried`/`attempted` 갱신 | 없음(계속) | `TestRefreshingInsideTheTTLCostsNoBrokerCall`, `TestARefreshIsExactlyOneHoldingsCall`, `TestAFailingBrokerIsNotRetriedOnEveryPageLoad`, `TestAVerificationInProgressSuspendsTheRefresh` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.mu.Lock` / `defer Unlock` | single-flight | fetch를 뮤텍스 안에서 한다 — 두 탭이 동시에 새로고침해도 브로커 1콜. 느린 브로커가 두 페이지를 함께 지연시키는 것이 그 대가이며, 로컬 1인 콘솔에서는 rate budget이 더 비싸다 | holdings.go:176 |
| `c.refreshLocked` | 허용된 1콜 | `context.WithTimeout(ctx, holdingsTimeout=10s)`. 실패는 이전 판독을 **버리지 않고** `lastErr`만 채운다 | holdings.go:237 |
| `c.snapshotLocked` | 현재 내용 렌더 | 부작용 없음 | holdings.go:214 |
| (금지 바인딩) | `HoldingsReader`는 메서드 1개(`Holdings`)뿐이다. 브로커 클라이언트를 이 패키지가 쥐지 않는다 | `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`가 `Options.Holdings→{Holdings}`로 고정 | static_test.go:869 |

## State mutations and fallbacks

- `c.rows/at/present/lastErr/tried/attempted`는 `refreshLocked`만 쓴다.
- `tried`는 **성공이 아니라 시도**를 기록한다 — 429를 답하는 브로커가 페이지 로드마다 다시 불리는 것을 막는다.
- 실패 대체: 이전 판독을 유지하고 `Error`를 함께 실어 보낸다. 5분 된 목록 + '브로커 조회 실패'가 빈 표보다 낫다.
- hold=true는 **콜드 캐시도 콜드로 유지**한다 — '검증 중 — 데이터 없음'이 그 상태다.

## Safety conclusion

- Safe edit boundary: 스냅샷 조립부 추출과 `Held` 덮어쓰기. B1·B2의 판정식과 `refreshLocked` 호출 조건은 무변경.
- High-risk impact: yes (계좌 게이트웨이 — 이 프로세스에서 브로커 요청을 만드는 유일한 함수이고, B2의 `!hold`가 실계좌 검증의 rate budget을 보호하는 양보 규칙 그 자체다)
