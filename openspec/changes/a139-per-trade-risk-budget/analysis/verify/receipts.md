# a139 검증 영수증

워킹트리 = base `69fc3896` + a139 편집(GREEN 커밋 직전), 2026-10-15 KST. `rtk proxy` 로 잘리지 않은 출력.
`./...` 은 병행 세션의 미추적 `tools/a140-breakout-signal-shadow/`(main 미선언, 작업 중)가 빌드를 깨므로 `go list ./... | grep -v a140-breakout-signal-shadow` 로 돌림 — 그 디렉터리는 손대지 않음.

| 명령 | 대상 | 결과 |
|---|---|---|
| `gofmt -l internal cmd` | 전체 | 출력 0 |
| `go vet <PKGS>` · `go vet -tags tossos_testseams <PKGS>` | a140 도구 제외 전체 | exit 0 · exit 0 |
| `go test -count=1 -timeout 40m` | risk · execgw · measure · config · strategydispatch · strategyflow · riskbucket · cmd/tossctl · console | 전부 ok |
| `go test -count=1 -timeout 40m ./internal/app/engine` | engine 전체 | ok 303.0s |
| `go test -tags tossos_testseams` | execgw · app/engine · cmd/tossctl · strategyflow · weeklyvaluelane · continuationlane · reversallane | 전부 ok |
| `go test -race -run A139` | risk · execgw · app/engine | ok · DATA RACE 0 |
| `openspec validate a139-per-trade-risk-budget --strict --no-interactive` | — | valid |
| `python3 tools/logic-map/check_analysis.py --change a139-per-trade-risk-budget` | 번들 12 | 커밋 전: 미추적 시험 이름 경고만 |
| journal | — | 돌리지 않음 — a139 는 journal 을 편집하지 않고 journal 시험은 `RiskBudget` · `DefaultPolicy` 를 읽지 않음(grep 0, freeze 리뷰 실측과 같음) |

기존 시험 5 건(freeze 리뷰 P2-2 실측)의 처리: `TestOrderAboveThePerOrderCapIsRefused` 손절 폭을 200 으로(기대 사유 유지) · `TestQuantityAboveTheRiskBudgetIsRefused` 새 예산 수량(17/16) ·
`TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit` 16.91 · testseams `TestQFinalAccountBaseGuardianPairedKRUS` · `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`
는 rig 예산을 옛 값으로 명시(짝 계약을 재는 시험 — 1/3 은 US 위험 상한만 묶어 짝을 깸; 1/3 사이징은 a139 시험이 잼).
RED: `analysis/red/red-risk.txt`(컴파일 RED) · `red-app-engine.txt`(컴파일 RED) · `red-app-engine-behaviour.txt`(행동 RED — 예산 100000 · acceptance 칸 없음).
