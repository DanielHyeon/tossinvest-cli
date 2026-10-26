# CodeGraph hard evidence — a174

- 기준 HEAD: `01e61a756484cbdcf415bad669f532361e0cffc8`
- CodeGraph: `1.6.0`, worktree에서 초기화·색인 완료
- `Console.handleDashboard` 정의: `internal/console/pages.go:111`
- 호출자: `ANY /verify-console` 라우트 하나(`internal/console/console.go:752`)
- 직접 callee: `chromeOnRequest`, request query, `snapshot`, `currentRun`, `run.snapshot`, `render`, `dashboardPage`
- 영향: `pages.go:handleDashboard`, `/verify-console` 라우트 두 심볼
- `settingsPage.GatePreflight` 정의: `internal/console/settings_operating.go:84`; 호출자는 `GateBlockers` 하나이며 callee는 한도 `Validate`, 거래 정책 `Complete`·`Missing`, `limitSummary`임.
- `settingsPage.LimitTierCards` 정의: `internal/console/settings_limits.go:232`; 카드의 `Preview.Rows`가 현재·프리셋 다섯 한도 값을 이미 보유함.

처분: `GatePreflight`·`LimitTierCards`·설정 저장 handler는 편집하지 않는다. 기존 함수 본문 변경은 `handleDashboard` 리터럴 한 줄로 제한하고, 나머지 로직은 새 leaf 파일에 둔다.
