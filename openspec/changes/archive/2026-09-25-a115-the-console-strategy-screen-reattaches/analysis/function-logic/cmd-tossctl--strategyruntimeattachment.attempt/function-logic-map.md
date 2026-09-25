# Function Logic Map: `strategyRuntimeAttachment.attempt`

- Source: `cmd/tossctl/httpapi_strategy_attach.go`
- AST evidence: `ast.json` (base `8688f74f` = HEAD — 이 파일은 a115 가 편집하지 않는다, :180–215, 분기 3)
- Risk scan: `risk-pattern-report.md`

> **인용 전용 번들**(a115 freeze 리뷰 P1-4) — a115 는 이 함수를 편집하지 않는다. design D1 의 재사용·펌프 논거가 이 분기들에 기댄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.resolve(a.ctx)` | (reader, live) | 콘솔판: `resolveConsoleStrategyRuntime`(재시도 경고 io.Discard) | live 아니면 침묵 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !live {` (:184) | 실패(!live) — 침묵, **붙어 있는 자리 불변**(live→sentinel 격하 금지) | — | `TestAFailedAttemptDoesNotClobberTheCurrentScreen` · `TestTheConsoleStrategyScreenReattachesAfterTheEngineRestarts` |
| B2 | `if a.reader == nil && reader != nil {` (:197) | 빈 자리(nil)에 sentinel 승격(a109 G1), seat++ — 영구 NOT_CONFIGURED 탈출로 | — | `TestAnEmptySeatTakesTheUnavailableSentinel` |
| B3 | `if announce {` (:212) | 부착 전이 1회 로그(announce). 밀려난 값은 잠금 밖에서 Close(G5) | — | `TestTheAttachmentReportsOnlyTransitions` · `TestTheReplacedReaderIsClosed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a.resolve` | 재해석(여기서만 dial) | 200ms probe, 요청 경로 밖 | AST |
| `closeEvictedStrategyReader` | 밀려난 client 의 유휴 연결 해제 | io.Closer 만 | AST |
| `a.reportAttached` | 부착 전이 1줄 | 전이 시 1회 | AST |

## State mutations and fallbacks

- trying=false · (B2) reader=sentinel · (성공) reader=client·failed=false·seat++·attached=true. a115 함의: G2(:102–112 주석)가 「시도 대상만 깨우기」 게이트를 기각했다 — 콘솔 펌프의 무조건 wake(freeze P1-2)의 근거이고, 변이 K8 이 그 게이트를 재도입하면 렌더 없는 재시작 시험이 빨개진다.

## Safety conclusion

- Safe edit boundary: 편집 없음(인용 전용).
- High-risk impact: no — 조회 전용 projection 재부착, 주문 경로 아님.
