# a138 Tasks

Story: `STORY-TOS-a138` · **High-risk**(Guardian 발급 · 예약 트랜잭션 · 사이징 인접) — Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 ·
독립 적대 리뷰 의무(docs/WORKFLOW.md 비례 원칙의 의무 범위).

> 순서: a138 의 Go · spec 을 전부 커밋한 뒤 a139 production 파일(공유: `riskPolicyFromAutomationGate` · `risk.DefaultPolicy`)을 만진다(Manager 지시).

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a138.yaml`, `_registry.yaml`, `FEAT-TOS-003` 역링크) + tracker 재생성 + proposal
- [x] 0.2 design(D0 근거표 · D1 세는 범위 3안 · D2 경계 · D3 위치 · D4 설정 · D5 단일 출처 · D6 · D7 거절되는 정상 입력 · D8 · D9) · spec 델타 · tasks,
      `openspec validate a138-max-open-positions --strict --no-interactive` rc=0 — proposal 단계 AST 근거 6건(`analysis/ast-evidence/`, 게이트 번들 아님)
- [x] 0.3 CodeGraph hard evidence · CodeGraphContext(not-applicable 사유) · evidence reconciliation(`analysis/code-context/`)
- [ ] 0.4 **사용자 결정**: D1(세는 범위 A/B/C) · D2(6 개 보유 중 scale-in 허용 여부) — design 「사람 결정」 원문으로 기록
- [ ] 0.5 결정 반영(spec · design 수정 시 requirement 변경 리뷰) + proposal-freeze 리뷰(적대 Eng + CEO 범위 + DX, 분리 컨텍스트) → review.md
- [ ] 0.6 편집할 기존 함수의 FLM/BTM 번들(`analysis/function-logic/`, 편집 전 `ast-base.json` + 편집 뒤 `ast.json`) + Pre-Edit 선언(review.md)
- [ ] 0.7 base 고정 — `base-commit.txt` = 이 change 첫 Go 커밋의 부모

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED — spec 시나리오 각각의 시험이 base 에서 실패(`analysis/red/`): 일곱 번째 종목 거부 · 든 종목 scale-in · 동시 진입 · 편입/외부 셈(D1 답) ·
      셀 수 없음 fail-closed · 매도 무관 · 기본값/0/음수/비정수 · 인터록 진술 불일치 · 상한 없는 예약 · digest 결속
- [ ] 1.2 GREEN — 설정 필드 · 해석 · Policy · 예약 tx 판정(`journal/open_positions.go`) · 발급 문 셋 · 인터록 · 기동 audit · digest · 격자 정책
- [ ] 1.3 콘솔 한도 편집기 정수 칸(규칙 12 · 13 — ui-skills 1~3개, 렌더 확인, 버튼 미클릭, 확인 문구 없음)
- [ ] 1.4 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change a138-max-open-positions`

## 2. 검증

- [ ] 2.1 `gofmt -l` · `go vet ./...` · `rtk proxy go test -count=1`(journal · execgw · risk · config · app/engine · cmd/tossctl · console · measure) ·
      `-tags tossos_testseams` · `make test-race` 대상 · `openspec validate --strict` → `analysis/verify/receipts.md`
- [ ] 2.2 변이 원장(`analysis/mutation/`): `>`→`>=` · 종목 중복 제거 제거 · HELD 예약 갈래 제거 · 외부/편입 갈래 제거 · 노출 미증가 조기 통과 반전 ·
      상한 nil 통과 · preimage 오류를 0 으로 · 발급 문 하나 상한 누락 · 인터록 대조 제거 · digest 입력 제거
- [ ] 2.3 독립 적대 리뷰(분리된 컨텍스트, 적대 Eng 포함) → review.md

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a138-max-open-positions` + 독립 검증
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서. 실계좌 한도 값 변경 · 토글 flip 은 사람이 직접.
