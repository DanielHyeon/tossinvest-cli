# Branch Test Map: `Runtime.escalate`

- Source: `internal/app/engine/runtime.go` (:442-482); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-engine.json`(연결 워크트리 `b910173a`, `internal/app/engine` 시험 37개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-app-engine--runtime.escalate.json(AST)`에 보존.
- 재번호: 새 B2(`ErrModeAnnouncementFailed` 갈래), 편집 전 B2(승격 오류) → B3.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 459:2 | 승격기 · 계좌 없음 → 반환 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | if at 463:2 | 커밋됨 · 통지 기록 실패 → 「승격됨」 경고 | `TestA092AnUnannouncedSustainedTighteningIsReportedAsTightened` | Z18 CAUGHT(`TestA092AnUnannouncedSustainedTighteningIsReportedAsTightened`) | 블록 463.69-471.3을 시험 1개가 실행, PASS |
| B3 | if at 472:2 | 승격 오류 → 「재시작이 푼다」 경고 | `TestTheDegradationAlertGoesOutEvenWhenTheModeTransitionFails` | 해당 없음(분기 불변) | 블록 472.16-478.3을 시험 1개가 실행, PASS |
