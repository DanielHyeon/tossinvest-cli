# a157 Tasks

Story: `STORY-TOS-a157` · **High-risk**(Guardian · execgw 첫 다리 판정 순서 · 일일 손실 분류 경로). Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 ·
변이 원장 · 독립 적대 리뷰가 의무다.

> 병행: 다른 세션(Codex)이 같은 트리에서 활동 중이다(미커밋 `docs/architecture*` · a128/a140 shadow 원장 append, a148~a156 정의).
> 이 change 는 자기 경로만 커밋한다. `_registry.yaml` · `FEAT-TOS-003` · `docs/pm/generated` 는 a157 줄만 index 에 올린다.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`STORY-TOS-a157.yaml` · `_registry.yaml` · `FEAT-TOS-003` 역링크) + proposal · design · spec 델타 · tasks —
      AST 근거 8건(`analysis/ast-evidence/`, 편집 전 HEAD `7b1748481`)
- [x] 0.2 CodeGraph hard evidence(`analysis/code-context/codegraph-baseline.md`, raw 는 이미 있음) · CodeGraphContext(또는 `not-applicable` 사유) ·
      `evidence-reconciliation.md`(인터페이스 선언 :65 대 실제 호출 :105 포함)
- [x] 0.3 proposal-freeze 적대 리뷰(분리 컨텍스트 Opus + codex) → review.md → 수리 — Opus APPROVE-WITH-FIXES(P0 0 · P1 3), codex REJECT as written(P1 1).
      사람 결정 「1-1, 2-1」 반영, 지적 전부 문서에 반영(review.md 표)
- [x] 0.4 특성 시험 — 편집 전 HEAD 에서 통과해야 한다(현재 순서를 고정함). 영수증 `analysis/red/characterization-head.txt`.
  - 뒤에 RED 로 뒤집을 행:
    - 상한 0 + Bn 일곱(B5 는 q_candidate 갈래와 소유자 LaneID/CampaignID 갈래 둘). 오늘은 `EXISTING_GUARDIAN_CAP`.
    - 입력 하나가 둘을 거는 (f) 실측 두 행(`QCandidate=0` · 비정규 Symbol).
  - 뒤에도 같아야 할 행:
    - 결속 결함 단독 B2~B8.
    - 진입 거절 단독 (a)~(e).
    - **일일 손실 한도 + B2 · B8, 측정 불가 + B2 · B8**(execgw 거절 + 엔진 `admit` 분류 · 강화 호출. 엔진 행은 B2 · B8 만 — design D3).
    - 통과 봉인 값.
    - nil Guardian.
- [ ] 0.5 편집할 기존 함수의 FLM/BTM 번들(`analysis/function-logic/`) + Pre-Edit 선언(review.md). 번들 대상:
  - `RiskGuardian.PrecheckQFinalCampaignFirstLeg`
  - `strategyFirstLegAdmissionBridge.admit`
  - `Context.NewPairedStrategyEntryProductionAssembly`
  - 판별 함수가 기존 파일에 들어가면 그 파일의 기존 함수 영향(기억 「Logic Map 은 계획보다 항상 많다」 — 새 코드는 새 파일에)

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED → 실패 확인. 영수증 `analysis/red/red.txt`.
  - 특성 시험의 「뒤집을 행」을 새 기대로 고친다.
  - 구조 시험(AST — design D6: 호출 하나 · field 문자와 `QFinalPolicyVersion` 각 한 번 · 호출보다 앞)을 더한다.
  - 판별 함수 표 시험을 더한다.
  - 조립 감독자 시험 셋(design D4 — 알림 칸 · 기록자 → `c.Log` · `EntryGate` nil/실제, 생산 감독자 대조군)을 더한다.
- [ ] 1.2 GREEN.
  - 판별 함수(새 파일).
  - `admit` 의 두 조건을 판별 함수로.
  - `PrecheckQFinalCampaignFirstLeg` 새 모양(design D1).
  - 조립 감독자 옵션(design D4).
- [ ] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `riskguardian_qfinal.go` 파일 sha 불변 · `IssuePrecheckedQFinalCampaignFirstLeg` 함수 본문 sha 불변(design D3) ·
      `check_analysis.py --change a157-first-leg-binding-before-sizing`

## 2. 검증

- [ ] 2.1 `make lint` · `go vet ./...` · `rtk proxy go test -count=1`(execgw · app/engine · risk · riskbucket · journal) · `-tags tossos_testseams` · `-race` 대상 ·
      `openspec validate --strict` → `analysis/verify/receipts.md`
- [ ] 2.2 변이 원장(`analysis/mutation/`, design D6 — 저장소 밖 사본 · pid 접미 · 무변이 대조군 GREEN · 시작 sha 단언). 비동등 전부 CAUGHT
- [ ] 2.3 독립 적대 리뷰(분리 컨텍스트 — Manager) → review.md → 수리

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a157-first-leg-binding-before-sizing`(격리 워크트리) + 독립 검증 · PM tracker 재생성 · archive
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서. 배포 시점은 사용자 지시 「P3 까지 끝나면 바로」를 따른다.

## 후속 (체크박스 아님)

- 조립 감독자를 **Run** 하게 되는 change 는 알림 dedup 상태(`dailyLossNoticeDay` · `entrySkipNoticeDay` · `entrySkipNoticed`)를 감독자에서 `Context` 로 올려야 한다.
  조립 감독자는 권한 갱신 파도마다 새로 만들어지고(design F9), 생산 감독자와 함께 돌면 같은 줄이 두 번 나온다.
