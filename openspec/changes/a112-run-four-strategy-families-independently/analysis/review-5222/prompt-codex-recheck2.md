# codex 재확인 #2 — a112 5.2.2.2 2차 수리(계좌 경계)

같은 세션의 이어 하기다. 앞 재확인(BLOCK: #2 계좌 쪽 P1 — `accountScopeCause` 가 ctx 밖의 모든 오류를 범위 국소로 분류)에 대한 2차 수리가 착지했다. 작업 디렉터리는 2차 수리
커밋의 `git archive` 트리(git 이력 없음). 읽기 전용 — 실행이 안 되면 코드 인용으로 판정하고 그렇다고 적어라.
**`~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라 — 어겼다면 출력 맨 위에 적어라.** 운영 원장 · 자격 증명도 열지 마라.

읽을 것: review.md 끝 소절 「codex 재확인 #1(face8d0d) — BLOCK(계좌 경계) · 정정 사슬 · 2차 수리」, `internal/app/engine/strategy_account_first_leg_authority.go`
`collectStrategyFirstLegAuthority`(계좌 범위 권한 부재 갈래) · 계좌 `collectMarket`, `strategy_owner_scope_authority.go` `accountScopeCause`, `internal/strategyaccount/production.go`
`LoadProductionAuthority` · `FileName`, 시험 `a112_owner_scope_trading_test.go`(`TestAnAccountLoadFailureOnOneScopeIsAFaultThatStopsTheCycle` 등) · `a112_scope_refusal_census_test.go`,
증거 `analysis/measurements/lot-5.2.2.2-fix2/`.

질문(좁게): 계좌 무결성 · I/O · ctx 실패가 범위 거절로 분류되어 같은 주기의 다른 범위 발급으로 이어지는 경로가 남는가? 「계좌 매니페스트는 시장 단위 파일이라 범위 국소 원인이
없다」는 전제가 코드상 참인가? 앞 재확인의 나머지(#1 · #3 · 위험 쪽 #2)가 이 편집으로 되살아나지 않았는가?
새 발견은 P0/P1 만 BLOCK 사유로. 맨 위 판정 한 줄 APPROVE / BLOCK, 발견 표, 항목별 한 줄, 마지막 줄 `Recommendation: …`.
