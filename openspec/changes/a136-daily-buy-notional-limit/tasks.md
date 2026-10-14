# a136 Tasks

Story: `STORY-TOS-a136` · **High-risk**(Guardian 발급 · 예약 트랜잭션 · 사이징 인접) — Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 ·
독립 적대 리뷰 의무(docs/WORKFLOW.md 비례 원칙의 의무 범위).

> 사용자 결정 2026-10-14 수신(design 「사람 결정」 원문). D9=b 의 수동 매수 집계는 **a136b 분리 제안**(Manager 승인 대기) — a136 은 원장 집계만 코딩한다.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a136.yaml`, `_registry.yaml`, `FEAT-TOS-003` 역링크) + tracker 재생성 + proposal
- [x] 0.2 design(D0 근거표 · D1 경계 4안 · D2 집계 3안 · D3 위치 · D4 경계 · D5 설정 · D6 스냅샷 · D7 seam · D8 · D9) · spec 델타 · tasks,
      `openspec validate a136-daily-buy-notional-limit --strict --no-interactive` rc=0 — proposal 단계 AST 근거 5건(`analysis/ast-evidence/`, 게이트 번들 아님)
- [x] 0.3 **사용자 결정**: D1=D · D2=c · D4=≥ · D5=기동 거부 · D9=b(수동 매수 합산 — a136b 분리 제안) — design 「사람 결정」 원문 · review.md
- [ ] 0.4 결정 반영(spec · design 수정 시 requirement 변경 리뷰) + proposal-freeze 리뷰(autoplan 4관점 + 적대 Eng) → review.md
- [ ] 0.5 `capture_change_base.py --change a136-daily-buy-notional-limit` 로 구현 직전 base 고정(병행 a137 착지 뒤가 되도록 Manager 와 순서 조정)
- [ ] 0.6 CodeGraph hard evidence · CodeGraphContext · evidence reconciliation(`analysis/code-context/`) — 발급 문 셋 · `reserveRows` · `recordLimitSave` · 인터록,
      `codegraph affected <files> --filter '*_test.go'`
- [x] 0.7 D2-c 실측: 결정 → attempt → 주문 → 체결 수량 결합으로 「체결 수량 0 종결」을 판정할 수 있는지, NOT_DISPATCHED · EXPIRED_UNCONSUMED 의
      해제 사유 행 모양 — **증명 가능**(design 「D2-c 실측」 Z1~Z3, 그 밖은 전부 센다)
- [ ] 0.8 편집할 기존 함수의 FLM/BTM 번들(`analysis/function-logic/`): `RiskGuardian.IssueEntry` · `IssuePrecheckedQFinalEntry` ·
      `IssuePrecheckedQFinalCampaignFirstLeg` · `reserveRows`(또는 그 호출 자리) · `ExposureLimitsFor` · `riskPolicyFromAutomationGate` ·
      `risk.Policy.Validate` · `DefaultPolicy` · 인터록 대조 함수 · `SaveEngineGateLimits` · `recordLimitSave` · `AutomationGate.Limits` + Pre-Edit 선언(review.md)

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED — spec 시나리오 8개 각각의 시험이 base 에서 실패(`analysis/red/`): 동시 발급 경합 · 체결 뒤 잔존 · 무체결 제외 · 경계(07:59/08:00 KST,
      서머타임 양쪽) · 계산 불가 fail-closed · 매도 무관(reduce-only · 손절) · US frozen FX 금액 · 기본값/0/음수/USD 무값 · 인터록 여섯째 칸 불일치 거부
- [ ] 1.2 GREEN — 설정 필드 · Policy · Limits · 원자 트랜잭션 안 합산(새 파일) · 발급 문 셋 · 인터록 · 한도 seam audit
- [ ] 1.3 콘솔 한도 편집기 숫자 칸(규칙 12 · 13 — ui-skills 1~3개, 렌더 확인, 버튼 미클릭, 확인 문구 없음). 커지면 후속 change 로 분리하고 여기 기록
- [ ] 1.4 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change`

## 2. 검증

- [ ] 2.1 `gofmt -l` · `go vet ./...` · `rtk proxy go test -count=1`(`internal/journal` · `internal/execgw` · `internal/risk` · `internal/config` ·
      `internal/app/engine` · `cmd/tossctl` · `internal/console`) · `-tags tossos_testseams` · `make test-race` 대상 · `openspec validate --strict`
- [ ] 2.2 변이 원장(`analysis/mutation/`): ≥→> · 체결분 제외 · 경계 오프셋 0 · 매도 경로에 판정 삽입 · 합계 오류를 0 으로 대체 · 동시성(트랜잭션 밖 합산)
- [ ] 2.3 독립 적대 리뷰(분리된 컨텍스트, 적대 Eng 포함) → review.md

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a136-daily-buy-notional-limit` + 독립 검증
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서. 실계좌 한도 값 변경 · 토글 flip 은 사람이 직접.
