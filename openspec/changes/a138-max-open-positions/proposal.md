# a138 — 동시에 들고 있는 종목 수에 상한을 둔다 (max-open-positions)

> **상태(2026-10-15):** proposal 단계. 사용자 결정 「최대 6」 수신(Manager 전달). **세는 범위(design D1)는 사람 결정 대기** — 저장소가
> 이미 세 종류의 포지션을 구별하고(엔진 진입 · 편입 · 외부), 계좌 단위 한도의 두 선례(a136 수동 매수는 셈, a137 일일 손실은 관리
> 포지션만)가 서로 다른 답을 주므로 저자가 고르지 않는다. 결정 전에는 production 파일을 편집하지 않는다.

## Why

사용자 지시(2026-10-15): **동시 보유 포지션 최대 6개**. 계좌가 6개보다 많은 포지션을 들게 될 신규 진입은 거부한다. 매도 · 손절 · 비상 청산 ·
flatten · 대사는 막히거나 늦어지지 않는다(§0.3). 다른 automation_gate 한도처럼 설정 가능 · 감사 · 기동 인터록 대조, 셀 수 없으면 진입 거부.

지금 HEAD 대조(a1d66f13):

- `risk-management` 「Guardian 판정 체인」이 `MAX_POSITIONS` 를 P3 **제외** 로 열거한다(spec.md:13). 보유 종목 수를 세는 한도는 어디에도 없다.
- 계좌 총계 한도의 최종 권위는 예약 트랜잭션이고, 모든 예약 tx 는 `journal.reserveRows` 한 함수를 지난다(CodeGraph callers 6 — 생산 5 + 시험 1,
  `analysis/code-context/codegraph-baseline.md`). a136 의 하루 매수 총액 판정이 이미 그 자리에 있다(AST `journal--reserverows` B4,
  `journal--checkdailybuynotional` B2 = 노출을 올리지 않는 결정은 통과, B3 = 상한 없는 EXPOSURE_RAISING 예약 거부).
- 위험 축소 발급(`RiskGuardian.IssueReduction`)은 `journal.RecordDecision` 만 부르고 예약 tx 를 지나지 않는다(AST `execgw--riskguardian.issuereduction`
  호출 목록 — `RecordDecisionAndReserve` 없음). 그래서 예약 tx 에 두는 판정은 구조적으로 매도에 닿지 않는다.
- 포지션 투영(`positions`)은 세 종류를 칼럼으로 구별한다: 엔진 진입(`entry_decision_id`), 편입(`adoption_id`), 외부(둘 다 NULL —
  대사가 계좌 보유를 조정 이벤트로 접어 넣은 것, `reconcile/external.go`). 관리 여부의 단일 술어는 `position.ExitEligible`.

## What Changes

1. **설정** — `engine.automation_gate.max_open_positions`(선택 숫자). 없음 · 0 은 기존 관례(0 = 미설정)대로 **기본값 6**. 개수는 통화와 무관하므로
   기본값은 KRW · USD 게이트 모두에 적용(a136 D5 같은 통화 문제 없음). 음수 · 정수가 아님 · NaN · Inf 는 게이트 ON 기동 거부.
2. **판정 위치** — 체인 단계가 아니라 **결정 · 예약 원자 트랜잭션 안**(`reserveRows` 5c, a136 5b 다음). 같은 트랜잭션에서 읽으므로 동시 진입이 함께 넘을 수 없다.
3. **세는 것** — 서로 다른 (시장, 종목) 의 집합: ① 살아 있는 포지션(CLOSED 가 아닌 행, **범위는 D1 사람 결정**) ∪ ② 발급일과 무관하게 아직 HELD 인
   EXPOSURE_RAISING 결정의 OPEN_EXPOSURE 예약의 종목(= 아직 포지션이 안 된 진행 중 진입) ∪ ③ 이번 결정의 종목. 집합 크기가 상한을 **넘으면**(> 6)
   LIMIT_REACHED. 이미 든 종목에 더 사는 진입은 집합을 키우지 않으므로 6 개 보유 중에도 허용된다(「6 개보다 많이 들게 될」 진입만 거부 — design D2).
   종목은 결정의 저장된 preimage(market · symbol)에서 트랜잭션 안에서 읽는다(호출자 값 불신).
4. **fail-closed** — 원장 읽기 실패 · preimage 해석 불가 · EXPOSURE_RAISING 예약에 상한 없음 → 그 진입의 결정 · 예약 롤백.
5. **적용 문** — EXPOSURE_RAISING 진입 발급 문 셋(`IssueEntry` · `IssuePrecheckedQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg`)이 같은 상한 값을
   한 곳(`g.openPositionLimit()`)에서 만들어 싣는다. `IssueReduction` · 보호 · 비상 청산 · 대사 · 체결 감지는 판정을 지나지 않는다.
6. **단일 출처 · 감사** — Guardian 이 상한을 진술(`MaxOpenPositionsLimit()`)하고 인터록이 config 해석 값과 대조, `LimitsDigest` 결속, 기동 audit 한 줄,
   인터록 acceptance 문구. 변경은 콘솔 한도 편집기 seam(키 단위 쓰기 + 필드별 audit), 확인 문구 마찰 없음. 적용은 엔진 재시작 때.
   통화 전환 저장은 이 값을 지우지 않는다(개수는 통화로 재해석되지 않음).
7. **콘솔 칸** — 고급 폼의 정수 입력 + 현재 값 표 행(미설정이면 「기본 6」). 티어 프리셋은 이 값을 갖지 않는다(nil = 키 무접촉).
8. **spec** — `risk-management`: 「Guardian 판정 체인」 제외 목록에서 MAX_POSITIONS 를 빼고 신규로(MODIFIED), 「Automation gate limits …」 상한 진술 ·
   인터록 대조(MODIFIED), 「정책 수치의 provenance」 기본값 6 의 출처(MODIFIED), 「최대 보유 포지션 수」 요구(ADDED). `engine-safety`:
   「자동화 게이트 기동 인터록」 조항 1 · 4(MODIFIED).

## What Does NOT Change

- 매도 · 손절 · 익절 · 비상 청산 · flatten · 보호 주문 · 대사 · 체결 감지의 경로와 즉시성. `IssueReduction` 무편집.
- 체인 12단계 순서 · 이름 · reason-code, 재수집 루프 · 결정 TTL, 결정 `limits_json` 5필드(`execgw.Limits`).
- 원장 스키마(v35 유지), 예약 종류 집합, 예약 해제 규칙. 새 브로커 호출 0.
- 수동 주문(토스 앱 · `tossctl order place`)을 막지 않는다(OFF = upstream).
- 다른 여섯 한도의 값 · 의미. 게이트 OFF 동작(Guardian 미구성).

## Impact

- Affected specs: `risk-management`(MODIFIED 3 · ADDED 1), `engine-safety`(MODIFIED 1).
- Affected code(예정): `internal/config/engine.go` · 새 `internal/config/open_positions.go` · `limits.go` · `limits_io.go`,
  `internal/risk/input.go`(Policy · DefaultPolicy · Validate), 새 `internal/journal/open_positions.go` + `reservations.go`(요청 칸 · 훅 한 줄),
  새 `internal/execgw/open_positions.go` + `riskguardian.go`(digest · IssueEntry) · `riskguardian_qfinal.go` · `riskguardian_first_leg.go`,
  `internal/app/engine/guardian_wiring.go` · `interlock.go` · 새 `open_positions_interlock.go`, `cmd/tossctl/limitsettings.go`,
  `internal/console/settings_limits.go`, `internal/measure/counterfactual.go`(격자 정책).
- **공유 파일**: `guardian_wiring.go riskPolicyFromAutomationGate` · `risk.DefaultPolicy` · risk-management 「Automation gate limits …」 요구는 a139 도 고친다
  → a138 을 먼저 커밋하고 a139 델타는 a138 문장 위에 쓴다(Manager 지시: 순차).
- 위험 등급: **High-risk**(Guardian · 예약 트랜잭션 · 사이징 인접). Pre-Edit · FLM/BTM · RED 선행 · 변이 원장 · 독립 적대 리뷰 의무.
- 배포: 엔진 이미지 교체. `LimitsDigest` 값이 한 번 바뀐다 — 생산 전략 바인딩은 Guardian 자신에게서 유도하므로(a136 근거 동일) 재서명 불요.

## 사람 결정 (대기 — design D1)

| ID | 질문 | 추천 |
|---|---|---|
| D1 | 세는 포지션의 범위: A 관리(엔진 진입 + 편입) / B 계좌 전체(외부 포함) / C = B + 미체결 수동 매수 주문 종목 | **B** |
| D2 | 6 개 보유 중 이미 든 종목에 더 사는 진입(scale-in)을 허용하나 | 허용(문구 그대로 — 개수가 늘지 않음) |
