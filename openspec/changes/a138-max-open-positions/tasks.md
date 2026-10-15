# a138 Tasks

Story: `STORY-TOS-a138` · **High-risk**(Guardian 발급 · 예약 트랜잭션 · 사이징 인접) — Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 ·
독립 적대 리뷰 의무(docs/WORKFLOW.md 비례 원칙의 의무 범위).

> 순서: a138 의 Go · spec 을 전부 커밋한 뒤 a139 production 파일(공유: `riskPolicyFromAutomationGate` · `risk.DefaultPolicy`)을 만진다(Manager 지시).

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a138.yaml`, `_registry.yaml`, `FEAT-TOS-003` 역링크) + tracker 재생성 + proposal
- [x] 0.2 design(D0 근거표 · D1 세는 범위 3안 · D2 경계 · D3 위치 · D4 설정 · D5 단일 출처 · D6 · D7 거절되는 정상 입력 · D8 · D9) · spec 델타 · tasks,
      `openspec validate a138-max-open-positions --strict --no-interactive` rc=0 — proposal 단계 AST 근거 6건(`analysis/ast-evidence/`, 게이트 번들 아님)
- [x] 0.3 CodeGraph hard evidence · CodeGraphContext(not-applicable 사유) · evidence reconciliation(`analysis/code-context/`)
- [x] 0.4 **사용자 결정**(2026-10-15 「권장대로」): D1 = B(계좌 전체) · D2 = scale-in 허용 — design 「사람 결정」 원문 · review.md
- [x] 0.5 결정 반영 + proposal-freeze 리뷰(적대 Eng + CEO 범위 + DX, 분리 컨텍스트) → review.md — APPROVE-WITH-FIXES(P1 2 · P2 3 · P3 7), 전부 문서에 반영(④ RECONCILE · ② 종류 무관 · 정수 헬퍼 등)
- [x] 0.6 편집할 기존 함수의 FLM/BTM 번들(`analysis/function-logic/`, 편집 전 `ast-base.json` + 편집 뒤 `ast.json`) + Pre-Edit 선언(review.md) — 생산 19 번들은 편집 전에 base 로 만듦(하네스 `analysis/harness/a138_flm.py base`), 콘솔 · 시험 fixture 15 는 검사기가 요구한 뒤 base worktree 에서 `ast-base.json` 추출 — 34 번들, 둘은 `revision: base`(본문 무편집 · 이웃 편집)
- [x] 0.7 base 고정 — `base-commit.txt` = `ab5edc48`(이 change 첫 Go 커밋의 부모)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — spec 시나리오 각각의 시험이 base 에서 실패(`analysis/red/`): 일곱 번째 종목 거부 · 든 종목 scale-in · 동시 진입(최종 거부가 MAX_OPEN_POSITIONS) · 편입/외부 셈(D1 = B) ·
      투영 거절 갭(④) · DAILY_LOSS 전용 HELD 셈(②) · RELEASED 미셈 · 같은 종목 다른 시장 두 칸 · 정수 헬퍼(+Inf · 1e300 · 2^63) · 생산 factory 거부 문구 ·
      셀 수 없음 fail-closed · 매도 무관 · 기본값/0/음수/비정수 · 인터록 진술 불일치 · 상한 없는 예약 · digest 결속
- [x] 1.2 GREEN — 설정 필드 · 해석 · 정수 헬퍼 · Policy · 예약 tx 판정(`journal/open_positions.go` ①~④) · 발급 문 셋 · 생산 factory 사전 거부 · 인터록 7c · 기동 audit · digest. 시험 fixture 10 곳에 넉넉한 상한, `seedExistingRiskReservation` 의 `'{}'` preimage 를 정규형으로(생산이 만들 수 없는 모양 — 판정이 셀 수 없음으로 거부)
- [x] 1.3 콘솔 한도 편집기 정수 칸(규칙 12 `fixing-accessibility` 1개 · 규칙 13 렌더 확인 — `analysis/verify/receipts.md`, 버튼 미클릭, 확인 문구 없음) · 저장 audit(0 = 「미설정 — 기본 6」) · census 시험 둘 등록. RED 는 콘솔 GREEN 뒤 세 파일을 HEAD 로 되돌려 관찰(`red-console.txt` — 순서 역전 기록)
- [x] 1.4 REFACTOR · 편집 뒤 AST 재추출(`a138_flm.py after`) · `check_analysis.py --change a138-max-open-positions` — 커밋 전 경고는 미추적 시험 이름뿐

## 2. 검증

- [x] 2.1 `gofmt -l` · `go vet ./...` · `rtk proxy go test -count=1`(journal · execgw · risk · config · app/engine · cmd/tossctl · console · measure) ·
      `-tags tossos_testseams` · `make test-race` 대상 · `openspec validate --strict` → `analysis/verify/receipts.md`
- [x] 2.2 변이 원장(`analysis/mutation/`): `>`→`>=` · 종목 중복 제거 제거 · HELD 예약 갈래 제거 · 외부/편입 갈래 제거 · 노출 미증가 조기 통과 반전 ·
      상한 nil 통과 · preimage 오류를 0 으로 · 발급 문 하나 상한 누락 · 인터록 대조 제거 · digest 입력 제거 · ③ 제거 · CLOSED 필터 제거 · 시장 차원 무시 ·
      state=HELD 필터 제거 · kind 필터 OPEN_EXPOSURE 만 · ② safety_class 필터 제거(구조 시험) · ④ RECONCILE 갈래 제거 · 정수 판정 IsInf 제거
      — 하네스 `analysis/mutation/a138_mutate.py`, round1(86442ad7) 28/30 · **round2(7e292659) 37/39 CAUGHT, 생존 2 = 동등 변이(IsInf · IsNaN)** — `round1.md` · `round2.md`
- [x] 2.3 독립 적대 리뷰(분리된 컨텍스트, 적대 Eng 포함) → review.md — APPROVE-WITH-FIXES(P0 0 · P1 0 · P2 1 · P3 6), P2 · P3-1~4 수리(7e292659), 잔재 P3-5 · P3-6(배포 전 사람 확인) 기록

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a138-max-open-positions` + 독립 검증
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서. 실계좌 한도 값 변경 · 토글 flip 은 사람이 직접.
