# Branch Test Map: `StrategyEntrySupervisor.Snapshot`

- Source SHA-256: `c713aa34dee53aa7d276e488efbd6928bc1be598e94128f66fc7f0055de392e0`; AST branch locations are authoritative.
- Revision: **modified (태스크 8.8.4, 2026-09-05)** — 한 줄 편집. 분기는 둘이고 편집
  전후 같다(AST 로 확인). 새 분기는 새 leaf 함수
  `StrategyEntrySupervisor.recordSwallowedCycleError` 안에 있으며 그 함수는 frozen
  base 에 없으므로 `Function Logic Map: not-applicable` 이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 728:2 — nil 감독자이거나 열거 밖 시장 | 기존 감독자 시험이 `Snapshot` 을 알 수 없는 시장으로 부른다 | no (base) | yes |
| B2 | if at 734:2 — 등록되지 않은 시장 | 같음 | no (base) | yes |
| 정상 경로 (`737:2`) | 등록된 시장의 상태를 읽어 낸다. 8.8.4 가 여기에 `SwallowedCycleErrors`·`FirstSwallowedFailure` 를 실었다 | `TestARefreshOnlyWorkerCountsTheCycleErrorsItSwallows` | **yes (8.8.4)** — 새 시험이 `SwallowedCycleErrors` 미정의로 컴파일 실패 | yes. 반증 셋 CAUGHT: 기록 호출 삭제 · 첫 원인 덮어쓰기 · 모든 시장을 함께 세기 |
