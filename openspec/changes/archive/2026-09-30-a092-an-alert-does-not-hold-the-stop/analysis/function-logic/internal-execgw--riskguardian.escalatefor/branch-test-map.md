# Branch Test Map: `RiskGuardian.escalateFor`

- Source: `internal/execgw/riskguardian.go` (:640-656); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-execgw.json`(연결 워크트리 `b910173a`, `internal/execgw` 시험 14개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-execgw--riskguardian.escalatefor.json(AST)`에 보존.
- 재번호: 새 B3(`ErrModeAnnouncementFailed` 갈래, B2 안).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 641:2 | 다른 거절 → 승격 없음 | `TestAChainRefusalIssuesNothing`, `TestOtherChainRefusalsAreNotTriggers` | 해당 없음(분기 불변) | 블록 641.56-643.3을 시험 2개가 실행, PASS |
| B2 | if at 644:2 | 승격 오류 | (미실행) | 해당 없음(분기 불변) | 블록 좌표 없음(조건이 여러 줄이거나 본문이 비어 하네스가 같은 줄 블록을 못 잡음) — 행동 증거는 RED 칸 |
| B3 | if at 646:3 | 커밋됨 · 통지 기록 실패 → 「승격됨」 오류 | `TestA092AnUnannouncedDailyLossTighteningIsReportedAsTightened` | Z19 CAUGHT(`TestA092AnUnannouncedDailyLossTighteningIsReportedAsTightened`) | 블록 646.56-651.4을 시험 1개가 실행, PASS |
