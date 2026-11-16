# Function Logic Map: `engineFillDetector` (cmd/tossctl)

- Source: `cmd/tossctl/engine.go` 762~787 (편집 뒤). 편집 전: `../../pre-edit-ast/tossctl--enginefilldetector.ast.json`(762~786, sha ec77f4d643ec).
- AST evidence: `ast.json`(편집 뒤 sha 2365d757c29d). 분기 B1(if@764)·B2(if@765) — 편집 전과 같음. return@766·return@771. call 4→5(`engineFillAlerts`@785).
- Risk scan: `risk-pattern-report.md`.
- 호출자(CodeGraph 1.6.0): `engineRuntime`(engine.go:655) + 시험 2, 이 change 의 시험 1.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `hints` | nil(운영) 또는 검증된 소비자 | 호출자 | 검증 실패 → 오류(B2) |
| `ectx.Notifier` | nil 또는 `*obs.Notifier` | 엔진 컨텍스트 | nil → `Alerts` nil(경보 안 함) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@764) | hints 있음 | — | — | `TestTheHintConsumerIsValidatedAtAssembly` |
| B2 (if@765) | hints 검증 실패 | — | return@766 | 같음 |
| return@771 | 조립 — **편집: `Alerts: engineFillAlerts(ectx.Notifier)` 필드 1줄** | — | 감지기 | `TestA191TheDetectorIsWiredToTheEngineNotifier` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFillAlerts`(@785, 새 파일 `a191_fill_alerts.go`) | nil `*obs.Notifier` 를 nil 인터페이스로 | 실패 없음 | `TestA191FillAlertsFilterTheNilNotifier` |

## State mutations and fallbacks

- 없음. nil 알림기 → 경보 없음(감지·종결 확인은 그대로 동작).

## Safety conclusion

- 배선 1줄. 분기 변화 0. 알림기 nil 거름은 새 함수(패닉 방지).
