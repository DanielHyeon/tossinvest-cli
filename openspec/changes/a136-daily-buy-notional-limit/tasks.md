# a136 Tasks

Story: `STORY-TOS-a136` · **High-risk**(Guardian 발급 · 예약 트랜잭션 · 사이징 인접) — Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 ·
독립 적대 리뷰 의무(docs/WORKFLOW.md 비례 원칙의 의무 범위).

> 사용자 결정 2026-10-14 수신(design 「사람 결정」 원문, 1차 · 2차). 수동 매수 합산 · 콘솔 칸도 a136 안. 순서: 핵심 → 수동 매수 → 콘솔.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a136.yaml`, `_registry.yaml`, `FEAT-TOS-003` 역링크) + tracker 재생성 + proposal
- [x] 0.2 design(D0 근거표 · D1 경계 4안 · D2 집계 3안 · D3 위치 · D4 경계 · D5 설정 · D6 스냅샷 · D7 seam · D8 · D9) · spec 델타 · tasks,
      `openspec validate a136-daily-buy-notional-limit --strict --no-interactive` rc=0 — proposal 단계 AST 근거 5건(`analysis/ast-evidence/`, 게이트 번들 아님)
- [x] 0.3 **사용자 결정**: D1=D · D2=c · D4=≥ · D5=기동 거부 · D9=b(수동 매수 합산 — 2차 결정으로 a136 안) — design 「사람 결정」 원문 · review.md
- [x] 0.4 결정 반영(spec · design 수정 시 requirement 변경 리뷰) + proposal-freeze 리뷰(적대 Eng + CEO 범위 + DX, 분리 컨텍스트) → review.md — APPROVE-WITH-FIXES, P1 3 수리
- [x] 0.5 `capture_change_base.py --change a136-daily-buy-notional-limit` 로 base 고정 — `f504cc52`(첫 Go 커밋 82edde52 의 부모)
- [x] 0.6 CodeGraph hard evidence · CodeGraphContext · evidence reconciliation(`analysis/code-context/`) — 발급 문 셋 · `reserveRows` · `recordLimitSave` · 인터록,
      `codegraph affected <files> --filter '*_test.go'`
- [x] 0.7 D2-c 실측: 결정 → attempt → 주문 → 체결 수량 결합으로 「체결 수량 0 종결」을 판정할 수 있는지, NOT_DISPATCHED · EXPIRED_UNCONSUMED 의
      해제 사유 행 모양 — **증명 가능**(design 「D2-c 실측」 Z1~Z3, 그 밖은 전부 센다)
- [x] 0.8 편집할 기존 함수의 FLM/BTM 번들(`analysis/function-logic/`): `RiskGuardian.IssueEntry` · `IssuePrecheckedQFinalEntry` ·
      `IssuePrecheckedQFinalCampaignFirstLeg` · `reserveRows`(또는 그 호출 자리) · `ExposureLimitsFor` · `riskPolicyFromAutomationGate` ·
      `risk.Policy.Validate` · `DefaultPolicy` · 인터록 대조 함수 · `SaveEngineGateLimits` · `recordLimitSave` · `AutomationGate.Limits` + Pre-Edit 선언(review.md)
      — 핵심 31 번들(시험 fixture 15 포함, 편집 전 `ast-base.json` + 편집 뒤 `ast.json`). **4 함수는 FLM 전에 편집됨(P1-3, review.md)**. 콘솔 · 수동 매수 편집분은 1.3 · 1.5 에서 추가

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — spec 시나리오 8개 각각의 시험이 base 에서 실패(`analysis/red/`): 동시 발급 경합 · 체결 뒤 잔존 · 무체결 제외 · 경계(07:59/08:00 KST,
      서머타임 양쪽) · 계산 불가 fail-closed · 매도 무관(reduce-only · 손절) · US frozen FX 금액 · 기본값/0/음수/USD 무값 · 인터록 상한 진술 불일치 거부
- [x] 1.2 GREEN(핵심) — 설정 필드 · Policy · 원자 트랜잭션 안 합산(`journal/daily_buy.go`) · 발급 문 셋 · 인터록 · 기동 audit · digest · 통화 전환 지움
      — RED `analysis/red/red-{clock,journal,risk,config,execgw,app-engine}.txt`(base + 시험 = 기호 미정의), 변이 round2 18/18 CAUGHT
- [x] 1.5 수동 매수 합산(D9=b) — refresher · 스냅숏 · 중복 제거 · paired FX · fail-closed(design D9) + RED · 변이
      — `internal/manualbuy`(순수 합산) · `app/engine/manual_buy_refresh.go`(보조 실행자 30초) · `execgw/manual_buy.go`(보드 · base 환산) ·
      journal 신선도 2분 · 하루 라벨 검사 · `QFinalEntryIssuance.PairedFXAuthority`(첫 다리 조립이 짝 FX 에서 채움).
      RED `analysis/red/red-1.5-*.txt`(핵심 c8783ad9 + 시험 = 기호 미정의), 변이 round3
- [x] 1.3 콘솔 한도 편집기 숫자 칸(규칙 12 · 13 — ui-skills 1~3개, 렌더 확인, 버튼 미클릭, 확인 문구 없음). 2차 결정으로 a136 안 — 분리하지 않음
      — `GuardianLimits.DailyBuy`(nil = 그 키를 그대로 — 티어는 값이 없어 USD 프리셋을 만들지 않음), 고급 폼 칸 · 현재 값 표 행 · 저장 뒤 문장, audit 은 저장 뒤
      파일 값으로. 규칙 12: `ui-skills-root` → `fixing-accessibility` 1개(label · aria-describedby · inputmode). 규칙 13: 격리 config(USD 무값)로 로컬 콘솔을
      띄워 `/settings/daily` 를 chrome-devtools 로 렌더 — 행 문구 「USD 계좌에는 기본값이 없어 엔진이 기동을 거부한다」, 칸 · 도움말 연결, 콘솔 오류 0, 요청 1(200).
      접힌 고급 폼은 스크립트로 열어 확인했고 저장 · 토글 · 승인 버튼은 누르지 않음. 티어 레지스트리는 USD 값을 구조적으로 요구하지 않음(티어는 이 칸을 갖지 않음)
- [x] 1.4 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change` — 58 번들(편집 전 `ast-base.json` + 편집 뒤 `ast.json`), 남은 경고는 미추적 시험 이름(커밋으로 해소) · 착지 기록 전 창 안내뿐

## 2. 검증

- [x] 2.1 `gofmt -l` · `go vet ./...` · `rtk proxy go test -count=1`(`internal/journal` · `internal/execgw` · `internal/risk` · `internal/config` ·
      `internal/app/engine` · `cmd/tossctl` · `internal/console`) · `-tags tossos_testseams` · `make test-race` 대상 · `openspec validate --strict`
      — `analysis/verify/receipts.md`
- [x] 2.2 변이 원장(`analysis/mutation/`): ≥→> · 체결분 제외 · 경계 오프셋 0 · 매도 경로에 판정 삽입 · 합계 오류를 0 으로 대체 · 동시성(트랜잭션 밖 합산)
      — round1~4, 최종 32/32 CAUGHT
- [ ] 2.3 독립 적대 리뷰(분리된 컨텍스트, 적대 Eng 포함) → review.md

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a136-daily-buy-notional-limit` + 독립 검증
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서. 실계좌 한도 값 변경 · 토글 flip 은 사람이 직접.
