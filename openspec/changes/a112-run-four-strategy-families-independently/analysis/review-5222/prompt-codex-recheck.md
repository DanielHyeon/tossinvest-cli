# codex 재확인 — a112 5.2.2.2 리뷰 수리 로트

같은 세션의 이어 하기다. 앞 판정(BLOCK: #1 활성화 없는 두 항목 쌍이 마지막 권한 통과, #2 적재 단계 오류가 범위 거절로 분류, #3 CAS 「항상 STALE」 과장)에 대한
수리가 착지했다. 작업 디렉터리는 수리 커밋의 `git archive` 트리(git 이력 없음). 읽기 전용 — 실행이 안 되면 코드 인용으로 판정하고 그렇다고 적어라.
**`~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라 — 어겼다면 출력 맨 위에 적어라.** 운영 원장 · 자격 증명도 열지 마라.

읽을 것: `openspec/changes/a112-run-four-strategy-families-independently/review.md` 끝 절 「5.2.2.2 리뷰 라운드(80ae96a5) — 합본 · 처분 · 수리 로트」,
`internal/riskbucket/production_snapshot_authority.go`(`ErrProductionRiskScopeRefused` · 감싸기 두 자리 · latch 분리), `internal/app/engine/strategy_owner_scope_authority.go`
(`riskScopeCause` · `accountScopeCause` · `strategyScopeRefusal.Unwrap`), `strategy_account_first_leg_authority.go` `collectStrategyFirstLegAuthority`(활성화 없는 개수 관문 · 범위 원인 분류)
· 계좌 `collectMarket`, `strategy_risk_authority.go` 위험 `collectMarket`, `strategy_first_leg_admission.go` `admit`, `strategy_dispatch_cycle.go` `dispatch`,
`strategy_entry_supervisor.go` `buildProductionStrategyMarketWorker`, 시험 `a112_owner_scope_trading_test.go` · `a112_scope_refusal_census_test.go` · `a112_first_leg_owner_scope_seal_test.go`,
증거 `analysis/measurements/lot-5.2.2.2-fix/`.

질문(좁게): 세 발견이 **종결**됐는가?
1. 활성화 없는 시장에서 마지막 권한의 수용 집합이 7ab8cd12(5.2.2.2 이전)와 같은가 — 새로 통과하는 입력이 남는가.
2. 원장 · 무결성 · ctx 결함이 범위 거절 타입으로 분류되어 같은 주기의 다른 범위 발급으로 이어지는 경로가 남는가(적재기 → 범위 칸 → 1차 레그 → admit → dispatch → 전달 몸통 사슬 전체).
   반대로 범위 국소 거절(정책 밖 종목 · scope latch · 계좌 매니페스트 부재)이 결함으로 오분류되는 곳은? sentinel 이 결함 원인에 감싸이는 자리가 있는가.
   riskbucket 편집이 판정(거절 집합)을 바꾸는가.
3. CAS 문구 정정이 코드와 맞는가.
새 발견은 P0/P1 만 BLOCK 사유로. 맨 위 판정 한 줄 APPROVE / BLOCK, 발견 표, 항목별 한 줄, 마지막 줄 `Recommendation: …`.
