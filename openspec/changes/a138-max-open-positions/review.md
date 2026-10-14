# a138 review

## 사람 결정 기록 (2026-10-15)

Manager 가 전달한 사용자 결정 — 원문은 design.md 「사람 결정」.

- 1차: 「max 6」 — 계좌가 6 개보다 많은 포지션을 들게 될 신규 진입 거부, 매도 · 손절 · 비상 청산 · flatten · 대사는 막지도 늦추지도 않음.
- 2차(사용자 답 「권장대로」): **D1 = B**(계좌 전체 — 관리 + 대사가 접어 넣은 외부 보유 + HELD 진입 예약 + 이번 종목, 서로 다른 (시장, 종목)) ·
  **D2 = scale-in 허용**(이미 든 종목의 진입은 개수를 늘리지 않음). spec 델타는 B 로 쓴 문장 그대로, scale-in 문장 · 시나리오도 이미 있음.

## Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a138-max-open-positions / 1.1~1.4
- 대상 심볼(편집하는 기존 함수 — 번들 analysis/function-logic/, 편집 전 ast-base.json = base 소스):
    internal/journal/reservations.go          reserveRows (5c 호출 한 줄) · ReserveRequest(칸 하나)
    internal/execgw/riskguardian.go           RiskGuardian.IssueEntry (요청 칸 하나) · RiskGuardian.LimitsDigest (입력 한 줄)
    internal/execgw/riskguardian_qfinal.go    RiskGuardian.IssuePrecheckedQFinalEntry (요청 칸 하나)
    internal/execgw/riskguardian_first_leg.go RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg (요청 칸 하나)
    internal/risk/input.go                    DefaultPolicy · Policy.Validate (분기 하나)
    internal/app/engine/guardian_wiring.go    riskPolicyFromAutomationGate (대입 한 줄) · newProductionRiskGuardian (정수 판정 분기 하나 — freeze P2-3)
    internal/journal/reservations.go          ErrReservationLimitExceeded 문서(변수 — 함수 아님, freeze P3-2)
    internal/app/engine/interlock.go          newAutomationStatus · acceptanceDetail · verifyGate(7c 한 단) · recordGateSettings(끝에 한 줄)
    internal/config/engine.go                 mergeEngine (복사 한 줄)
    internal/config/limits_io.go              Service.LoadRawEngineGate · AutomationGate.Limits · gateMembersOf (멤버 분기 하나)
    internal/config/limits.go                 GuardianLimits.Validate (분기 하나)
    (internal/measure/counterfactual.go gridPolicy 는 편집하지 않음 — 체인은 개수 상한을 읽지 않고 DefaultPolicy 의 6 이 검증을 통과함)
    시험 fixture: internal/app/engine/interlock_test.go stubGuardian(칸 · 메서드) — 그 밖의 fixture 는 검사기가 요구하는 대로 번들 추가
  판정이 새로 읽는 표: positions · risk_reservations(+decisions) · reconcile_states(④, freeze P1-1) — 읽기만, 스키마 무변경
  새 파일(새 leaf): internal/journal/open_positions.go · internal/execgw/open_positions.go · internal/config/open_positions.go ·
                   internal/app/engine/open_positions_interlock.go
  콘솔(1.3, 번들은 그때 추가): cmd/tossctl/limitsettings.go · internal/console/settings_limits.go
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (reserveRows 생산 호출자 5 = 모든 예약 tx)
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/ (불일치 1 — CodeGraph 가 DefaultPolicy 호출자 하나 누락, HEAD 로 해소)
- 기존 동작 파악 근거: HEAD 읽기 + analysis/ast-evidence/ + 기존 시험(reservations_test · a136_daily_buy_test · riskguardian_test · interlock_test · guardian_wiring_internal_test · limits_io_test)
- Function Logic Map / Branch Test Map: analysis/function-logic/<pkg>--<func>/ (하네스 analysis/harness/a138_flm.py — base 로 만들고 편집 뒤 after 로 재추출)
- upstream 상속 테스트 영향: no — 게이트 OFF 는 Guardian 을 만들지 않음, 새 설정 키는 float 로 읽어 OFF 설정 읽기를 깨지 않음, 매도 경로 무편집
- 실패 테스트 선행 작성: yes (analysis/red/ — base + a138 시험 = 기호 미정의, 그다음 판정 함수를 빈 몸으로 둔 상태의 행동 RED)
- 설정·DB·journal 변경과 rollback: 설정 키 1(선택, 없음 = 기본 6) · 원장 스키마 변경 0 · rollback = 이전 이미지(키는 옛 빌드가 무시)
- 안전 불변식 §0 위반 여부 검토: 통과 — 매도 · 손절 · 청산 · 대사 · 체결 감지는 reserveRows 를 지나지 않음(IssueReduction 은 RecordDecision 만, AST), 판정은 거부 방향만 추가
```

## proposal-freeze 리뷰 (2026-10-15)

- 보이스: 분리된 컨텍스트의 적대 리뷰어 1(code-reviewer — 적대 Eng + CEO 범위 + 운영 DX), 읽기 전용, HEAD a1d66f13 → 0e6c42d6(D1/D2 결정 커밋이 리뷰 중 들어옴 — D1-B 기준으로 봄).
  `openspec validate --strict` rc=0, 인용 소스 6 파일 sha256 = ast.json 일치.
- 판정: **APPROVE-WITH-FIXES** — P0 0 · P1 2 · P2 3 · P3 7.
- 수용 · 반영(이 커밋, Go 편집 전):
  - P1-1 투영 거절 갭 — 종결 체결이 예약을 해제한 뒤(RecordFill B28 512) 투영이 거절되면(ProjectPosition B5 101) 포지션 행 없이 종목 범위 RECONCILE 만 남아
    ①·② 어디에도 없음 → **④ 종목 범위 활성 RECONCILE 중 S 에 없는 종목을 한 칸씩** 셈(design D2, spec ①, 시나리오 추가). AST 근거 `journal--journal.recordfill` ·
    `journal--projectposition` 추가. RED 추가.
  - P1-2 ② 가 OPEN_EXPOSURE 만 셈 → (a) 채택: **종류 무관**(닫힌 집합 셋을 `IN` 으로 — 색인 유지), Gateway `checkReservation` 인가 술어와 같음. 변이 원장에 kind 필터.
  - P2-1 대사 주기 60초 · 안정 사이클 정정(design D0 · D1 · D7, spec 명시 한계).
  - P2-2 D7 열거 보강(운영자 해제 대기 HELD · 엔진 밖 매도 · dust · KR+US 합산 · permissive 쪽 운영자 해제 · 엔진 밖 정정) — spec 「명시 한계」 문장에도.
  - P2-3 정수 판정 헬퍼 하나 `config.WholeOpenPositions`(IsInf 먼저), 생산 factory 가 Validate 앞에서 운영자 값을 문구에 담아 거부, RED 에 +Inf · 1e300 · 2^63 · 생산 factory 경로.
  - P3-1 D8 색인 서술 정정(CLOSED 이력까지 훑음) — EXPLAIN 시험 대신 지연 실측을 영수증에. P3-2 sentinel 문서 갱신(Go 편집 때). P3-3 이름 통일.
    P3-4 spec audit 문장(콘솔 audit best-effort + 기동 audit 둘째 기록, 0 = 「미설정 — 기본 6」). P3-5 D4 한정. P3-7 proposal Impact 정정.
  - P3-6 절차: 리뷰가 본 워킹트리의 a138 시험 · RED 영수증 · Pre-Edit · base-commit.txt 는 **이 저자(같은 change 의 teammate)**가 D1/D2 결정 뒤 리뷰와 병행해 만든 것이다
    (리뷰어가 「peer 세션」으로 본 것은 오인). 사실: proposal-freeze 리뷰 완료 **전**에 RED 시험을 썼다 — production 파일은 리뷰 전 0 편집. 리뷰 반영으로 RED · Pre-Edit 를
    다시 맞춘다(아래 갱신). base-commit.txt 는 첫 Go 커밋 직전 HEAD 로 다시 고정한다.
