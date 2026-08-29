# Function Logic Map: `Console.positions`

- Source: `internal/console/portfolio.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

포지션 화면 조립. 이 change가 바꾼 것은 인라인이던 원장 읽기 15줄을 `c.livePositions(ctx)` 한 줄 호출로 교체한 것이며, 순서·실패 처리·조인 인자는 그대로다. 이 함수는 브로커 캐시의 **갱신하는** 읽기(`holdings.get`)를 부르는 화면이고, 개요는 그래서 이 함수를 재사용하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.now()` | 주입 시계 | `Options.Now` | 해당 없음 |
| `c.verifyHold(now)` | (보류 여부, 사유) | in-process run + runlock 마커 신선도(5분 상한) | 보류면 브로커 호출 0콜, 사유가 화면에 뜬다 |
| `c.holdings.get(...)` | 브로커 절반 | `Options.Holdings` | 미배선·실패·보류 각각 다른 문장 |
| `c.livePositions(ctx)` | 원장 절반 | `Options.JournalPath` | 부분 답 + 사유 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`, `TestThePositionsScreenRendersWithEitherSourceMissing`, `TestAnUnmanagedHoldingIsLabelledExactlyOnce`, `TestAVerificationInProgressSuspendsTheRefresh` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.verifyHold(now)` | 양보 규칙 판정 | 이 프로세스의 run은 메모리, 다른 프로세스는 runlock mtime 신선도 — 파일에 이 프로세스의 일을 묻는 것은 자기 writer와의 경쟁이다 | holdings.go:267 |
| `c.holdings.get(ctx, now, hold, why)` | TTL 내 재요청·다중 탭은 추가 호출 없음, 갱신 1회는 holdings 1콜 | 스펙 `rate budget 보호`가 고정 | holdings.go:171 |
| `c.livePositions(ctx)` | 원장 읽기 | 부분 실패는 사유와 함께 유지 | portfolio.go:405 |
| `joinPositions(...)` | (market, symbol)로 조인 | 한쪽이 비어도 다른 쪽만으로 렌더 — 조인 실패가 화면 실패가 되지 않는다 | portfolio.go:458 |
| (금지 바인딩) | 브로커는 `HoldingsReader`(메서드 1개), 원장은 `journal.ReadOnly` | 정적 가드 둘이 고정 | static_test.go:940, static_test.go:1542 |

## State mutations and fallbacks

- `positionsView` 값 하나를 만든다. 쓰기 부작용은 브로커 캐시의 내부 갱신뿐이며 그것은 `get`이 소유한다.
- `v.Journal.Readable()`를 조인에 넘긴다 — 원장을 못 읽었을 때 '관리 외'가 아니라 '불명'이 되게 하는 인자다.

## Safety conclusion

- Safe edit boundary: 원장 읽기 블록의 함수 호출로의 교체. 조인 인자 세 개와 `holdings.get` 호출 형태는 무변경.
- High-risk impact: yes (원장 + 계좌 게이트웨이 — 이 화면의 갱신이 실계좌 요청을 만들고, 그 결과로 붙는 관리/미관리 라벨이 운영자의 편입 판단 입력이다)
