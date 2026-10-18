# a157 Tasks

Story: `STORY-TOS-a157` · **High-risk**(Guardian · execgw 첫 다리 판정 순서). Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 ·
독립 적대 리뷰가 의무다.

> 병행: 다른 세션(Codex)이 같은 트리에서 활동 중이다(미커밋 `docs/architecture*` · a128/a140 shadow 원장 append, a148~a156 정의).
> 이 change 는 자기 경로만 커밋한다. `_registry.yaml` · `FEAT-TOS-003` 은 a157 줄만 index 에 올린다.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`STORY-TOS-a157.yaml` · `_registry.yaml` · `FEAT-TOS-003` 역링크) + proposal · design(F1~F8 · D1~D6) · spec 델타 · tasks —
      AST 근거 8건(`analysis/ast-evidence/`, 편집 전 HEAD `7b1748481`)
- [ ] 0.2 CodeGraph hard evidence(`analysis/code-context/codegraph-baseline.md`, raw 는 이미 있음) · CodeGraphContext(또는 `not-applicable` 사유) ·
      `evidence-reconciliation.md`(인터페이스 선언 :65 대 실제 호출 :105 포함)
- [ ] 0.3 proposal-freeze 적대 리뷰(분리 컨텍스트 + codex) → review.md → 수리. 사람 판정 둘(D2-c 수용 여부 · D5 `EntryGate`)을 리뷰 결과와 함께 사용자에게 묻는다
- [ ] 0.4 특성 시험 — 편집 전 HEAD 에서 통과해야 한다(현재 순서를 고정함): 상한 0 + 결속 결함 B2 · B5 · B8 → 오늘은 상한 0 거절(뒤에 RED 로 뒤집을 행),
      결속 결함 단독 B2~B8 · 진입 거절 단독 (a)~(e) · 통과 봉인 값 → 오늘과 같음(뒤에도 같아야 할 행). 영수증 `analysis/red/characterization-head.txt`
- [ ] 0.5 편집할 기존 함수의 FLM/BTM 번들(`analysis/function-logic/` — `RiskGuardian.PrecheckQFinalCampaignFirstLeg` ·
      `Context.NewPairedStrategyEntryProductionAssembly`) + Pre-Edit 선언(review.md)

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED — 특성 시험의 「뒤집을 행」을 새 기대로 고침 + nil Guardian 행 + 조립 감독자 알림 두 행(생산 감독자 대조군) → 실패 확인.
      영수증 `analysis/red/red.txt`
- [ ] 1.2 GREEN — `PrecheckQFinalCampaignFirstLeg`: nil 가드 + B2~B8 을 `PrecheckQFinalEntry` 앞으로(입력 출처는 design D1 표대로, `policyVersion` 은 한 번 계산).
      `(*Context).NewPairedStrategyEntryProductionAssembly`: 옵션 두 칸(D5 사람 결정이 1이면 `EntryGate` 도)
- [ ] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `PrecheckQFinalEntry` · `IssueQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg` 소스 sha 불변 확인 ·
      `check_analysis.py --change a157-first-leg-binding-before-sizing`

## 2. 검증

- [ ] 2.1 `make lint` · `go vet ./...` · `rtk proxy go test -count=1`(execgw · app/engine · risk · riskbucket · journal) · `-tags tossos_testseams` · `-race` 대상 ·
      `openspec validate --strict` → `analysis/verify/receipts.md`
- [ ] 2.2 변이 원장(`analysis/mutation/`, design D6 — 저장소 밖 사본 · pid 접미 · 무변이 대조군 GREEN · 시작 sha 단언). 비동등 전부 CAUGHT
- [ ] 2.3 독립 적대 리뷰(분리 컨텍스트 — Manager) → review.md → 수리

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a157-first-leg-binding-before-sizing`(격리 워크트리) + 독립 검증 · PM tracker 재생성 · archive
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서. 배포 시점은 사용자 지시 「P3 까지 끝나면 바로」를 따른다.
