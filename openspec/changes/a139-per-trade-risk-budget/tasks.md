# a139 Tasks

Story: `STORY-TOS-a139` · **High-risk**(사이징) — Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 · 독립 적대 리뷰 의무.

> 순서: a138 의 Go · spec 이 전부 커밋된 뒤에 production 파일을 만진다(Manager 지시). archive 는 a138 뒤(델타가 a138 문장 위에 쓰였음).

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a139.yaml`, `_registry.yaml`, `FEAT-TOS-003` 역링크) + tracker 재생성 + proposal
- [x] 0.2 design(D0 소비자 전수 · D1 내림 단위 · D2 digest/인터록/감사 · D3 기본 정책 · D4 정책 버전 · D5 거절되는 정상 입력 · D6 매도) · spec 델타 · tasks,
      `openspec validate a139-per-trade-risk-budget --strict --no-interactive` rc=0 — AST 근거 12건(`analysis/ast-evidence/`, 게이트 번들 아님)
- [x] 0.3 CodeGraph hard evidence · CodeGraphContext(not-applicable 사유) · evidence reconciliation(불일치 1 — CodeGraph 가 DefaultPolicy 호출 하나 누락, HEAD 로 해소)
- [x] 0.4 proposal-freeze 리뷰(적대 Eng + CEO 범위, 분리 컨텍스트) → review.md — APPROVE-WITH-FIXES(P1 2 · P2 4 · P3 4), 전부 문서에 반영(D0 출력 소비자 · audit 유도 불가 · 버전 v2 · provenance MODIFIED)
- [x] 0.5 편집할 기존 함수의 FLM/BTM 번들(`riskPolicyFromAutomationGate` · `newProductionRiskGuardian` · `DefaultPolicy` · `newAutomationStatus` · `recordGateSettings` · `acceptanceDetail` + 고칠 시험 fixture) + Pre-Edit 선언(review.md)
- [x] 0.6 base 고정 — `base-commit.txt` = `69fc3896`(첫 Go 커밋의 부모 — 병행 a128 커밋이 5bbbdc23 뒤에 들어와 그 위로 다시 고정)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — 유도(100,000 KRW → 33,333 · 50 USD → 16.66 · 50.75 USD → 16.91 · 0.02 USD → 오류 · 표에 없는 통화 → 오류 · 음수/NaN → 오류) · 생산 정책 ·
      기본 정책 = 유도 · 체인 사이징 상한 1/3(경계: 예산이 정확히 한 단위를 덮음) · q_final Guardian 상한 1/3 · audit 한 줄(값 = 생산 유도의 답, detail 에 규칙) · 매도 무관
      + freeze 리뷰: Guardian 상한 0 첫 다리 = `EXISTING_GUARDIAN_CAP`(분류 고정 특성 시험) · 게이트 OFF · 일일 손실 0 기동 성공(audit 빈 값) · JPY / 0.02 USD 게이트 ON →
      `ErrLimitsRequired`(factory 사전 검사, audit 단계 아님) · 정책 버전 v2.
      기존 시험 5 건(실측): `TestOrderAboveThePerOrderCapIsRefused`(손절 폭을 좁혀 notional 단에 계속 닿게 — 기대 사유 유지) · `TestQuantityAboveTheRiskBudgetIsRefused`(새 예산의
      수량으로) · `TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit`(50.75 → 16.91) · testseams `TestQFinalAccountBaseGuardianPairedKRUS` ·
      `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`(Guardian 소유 수량이 의도대로 줄어든 값으로 — 시험이 재는 것이 Guardian 사이징)
- [x] 1.2 GREEN — `risk/risk_budget.go` · `riskPolicyFromAutomationGate` · `DefaultPolicy` · `recordGateSettings` · `acceptanceDetail`
- [x] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change a139-per-trade-risk-budget`

## 2. 검증

- [x] 2.1 `gofmt -l` · `go vet ./...` · `rtk proxy go test -count=1`(risk · execgw · app/engine · measure · journal) · `-tags tossos_testseams` · `openspec validate --strict`
      → `analysis/verify/receipts.md`
- [ ] 2.2 변이 원장(`analysis/mutation/`): 나눗수 3→2 · 내림→올림 · 최소 단위 USD 2→3 · KRW 0→2 · 0 허용 · 유도 우회(= dailyLoss) · 유도 실패 시 dailyLoss 로 폴백 · DefaultPolicy 값 ·
      audit 줄 제거 · audit 만 dailyLoss · factory 사전 검사 제거 · 버전 v1 되돌림
- [ ] 2.3 독립 적대 리뷰(분리된 컨텍스트, 적대 Eng 포함) → review.md

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a139-per-trade-risk-budget` + 독립 검증 (a138 archive 뒤)
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서.
