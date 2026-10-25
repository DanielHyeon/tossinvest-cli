# 증거 조정 — a174

- CodeGraph hard evidence와 현재 `pages.go`·`settings_operating.go`·`settings_limits.go`·관련 시험은 일치함.
- CodeGraphContext는 새 worktree 미등록으로 미가용이며, 이 불일치는 보조 도구의 범위 문제이지 현재 소스 충돌이 아님.
- `GatePreflight()`는 한도·거래 정책을 절 단위로만 판정하므로 다섯 한도 행으로 재분해하지 않음.
- `LimitTierCards()`의 `Preview.Rows`를 사용해 값·방향·통화 설명을 옮기며 별도 계산을 만들지 않음.
- 미배선·읽기 실패의 영값 표시 위험은 새 `startupReadiness`에서 행별 `unknown`과 원인으로 처리하고 RED 시험으로 고정함.
