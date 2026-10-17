# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json` — **편집 전**, :618–705, 분기 6 · 반환 7 · 호출 13, source_sha256 `cfe57e7c2822…`, 추출 HEAD `8c390aa6`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(22판 C1): exit observer 옵션에서 `Announcer: ectx.Notifier`(:639)를 뺀다. exit 쪽 announcer는 `Context.ExitObserver`의 기본값(기록 전용)이 진다 — exit 결속을 한 자리에 모은다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 조립된 엔진 Context | `engine.go` 조립 | 각 생성 실패는 오류로 반환 |
| `clk` · `logger` · `ready` | — | 호출자 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `engineFillDetector` 오류 (:621) | — | 오류 (:622) | (미실행) |
| B2 | `ReconcileDriver` 오류 (:630) | — | 오류 (:631) | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B3 | `ExitObserver` 오류 (:641) | — | 오류 (:642) | 같음 |
| B4 | `Recovery` 오류 (:646) | — | 오류 (:647) | 같음 |
| B5 | 전략 진입 감독자 오류 (:650) | — | 오류 (:651) | (미실행) |
| B6 | `AlertDeliverer` 오류 (:662) | — | 오류 (:663) | (미실행) |
| 종단 | — | `engine.NewRuntime(…)` (:666) — 루프 넷 + 보조 실행자 | 그 반환 | 같음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.ExitObserver(ExitObserverOptions{…, Escalate: ectx.Journal, Announcer: ectx.Notifier})` :634-640 | exit 관측기 | 오류는 B3 | AST · 21라운드 C1(관측 두절 강화가 이 announcer로 동기 통지) |
| `engine.NewRuntime(RuntimeOptions{…, Alerts: ectx.Notifier, Announcer: ectx.Notifier})` :666 | 런타임 — 루프 비정상 반환 알림과 감독자 승격 통지는 **동기**로 남는다(Q1 문자 해석) | — | AST |

## State mutations and fallbacks

- 상태를 바꾸지 않는 조립 함수다. 편집은 exit observer 옵션의 한 필드를 지우는 것이다. `RuntimeOptions`의 `Alerts` · `Announcer`(런타임 자신의 알림)는 그대로다.

## Safety conclusion

- Safe edit boundary: 분기 B1~B6과 호출 순서는 그대로다. `ExitObserverOptions`의 `Announcer` 필드 하나만 빠진다.
- High-risk impact: yes — 엔진 조립(알림 입구). 손절 경로 무변화.
