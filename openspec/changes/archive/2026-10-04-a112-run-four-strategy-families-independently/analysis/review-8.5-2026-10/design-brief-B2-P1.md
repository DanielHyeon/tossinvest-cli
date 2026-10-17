# 설계 브리프 — 8.5 codex r2 P1(B2 조기 활성화 읽기가 취소된 수집을 붙잡음) 처분 (코드 변경 0, 2026-10-04)

입력: `_codex-r2/REPORT.md` P1 · P2 셋. 기준 커밋 0b441270(이동) / 그 부모 f473d815(이동 전). 좌표는 HEAD 556c3c1f.

## 1. 전-이동 노출 실측 — 「새로 생긴 위험」 vs 「원래 있던 위험」

**읽기 자체.** 활성화 적재는 `strategyrouter.LoadProductionFamilyActivation` → `readProductionRouteFile`(production_owner_unix.go:12 — `Lstat · Open · ReadAll · Lstat`,
ctx 없음, 상한 64 KiB · 0400 정규 파일). 미선언(핀 빈 값)은 **첫 줄**(:433)에서 돌아와 ctx · 파일 어느 것도 건드리지 않는다. 오늘 생산은 활성화 핀 0 건(8.7.2 측정) →
**생산의 활성화 파일 읽기는 이동 전후 모두 0 회**다. 아래 모집단은 핀을 선언한 배포(활성화 로트 이후)의 것이다.

**같은 모양의 기존 읽기.** 제안 단계의 이웃 적재기 넷이 전부 같은 모양이다 — 시장별 goroutine 안의 ctx 없는 파일 읽기 + `<-outcomes` join(ctx select 없음):
route `strategy_route_authority.go:105-131`(경로 매니페스트 `readProductionRouteFile`) · proposal `strategy_proposal_authority.go:257-283`(제안 매니페스트
`strategyproposal.readProductionFile`, `LoadProductionAuthorityBatch` 는 읽기 **앞**에서만 `ctx.Err` 확인) · risk `strategy_risk_authority.go:152-178`
(`riskbucket.readProductionRiskFile`) · account(같은 패턴). 즉 「읽기에 들어간 뒤의 취소에 응답하지 않음」은 **이동 전부터 파도 전체에 있던 결함**이다.

**주기 종류별 — 제안 수집(collectMarket) 안에서 ctx 를 안 보는 파일 읽기 수(핀 선언 시장 기준):**

| 주기 종류(닫힘) | 이동 전(f473d815) | 이동 뒤(0b441270) | 증가 |
|---|---|---|---|
| ROUTE_NOT_READY | 0 | 0 | 0 |
| FX_NOT_READY · INTERNAL_FAILURE(설정) · AUTHORITY_INVALID(열쇠) · INTERNAL_FAILURE(중복 종목) | 0 | 1(활성화) | **+1** |
| AUTHORITY_INVALID(제안 적재 실패) | 1(제안 매니페스트) — 실패 지점에 따라 0~1 | 그 + 1(활성화) | **+1** |
| PROPOSAL_PRODUCTION_FAULT · 관문 뒤 닫힘 일곱 · 성공 | 1(제안) + 1(활성화) | 같음 | 0 |

같은 파도에서 route 적재기는 제안 수집 **앞**에 이미 경로 매니페스트를 ctx 없이 읽었다(모든 종류 공통, 핀 무관). 그래서 새로 생긴 위험의 정확한 크기는:
**핀 선언 시장에서 「제안 적재 앞에서 닫히는 다섯 종류」 주기에 작은 로컬 파일 읽기 1 회가 더해짐**이고, 「읽기 중 취소 비응답」이라는 성질 자체는 원래 있던 결함이다.
codex 재현은 주입한 stalled reader 로 150 ms 대기를 보였다 — 실제 로컬 0400 64 KiB 파일이 묶이는 조건은 파일 시스템 정지(원격/고장 마운트)다. 진입 경로라 손절 경로와 무관.

## 2. carried activation 소비자 census — 실패 주기(항목 0 · Ready=false)에서 값을 실제로 읽는 곳

`strategyProposalMarketAuthority.familyActivation()` 의 생산 호출 여섯(grep):

| 자리 | 실패 주기에서 | 효과 |
|---|---|---|
| `strategy_entry_supervisor.go:545` → `lanes.evaluate(..., promotion, inputs)` | **읽음** | 레인 관측 Desired/Effective(→ 투영 `lanes[].desired/effective`) · 레인 step 이 승격으로 돎(입력 없음 → REFUSED; 영값이면 DORMANT) |
| `strategy_dispatch_handoff.go:38` `dispatchHandoffs` | 읽음 | 검증이면 범위별 handoff(항목 0 → 전달 0), 아니면 시장 handoff 하나(준비 안 됨 거절) — 어느 쪽도 dispatch 콜백 0(codex r2 실측) |
| `strategy_account_first_leg_authority.go:161` 계좌 적재 | 항목 0 이 먼저 거절(`len(entries)==0`) | 없음 |
| `strategy_account_first_leg_authority.go:311` 1차 레그 권한 | 도달 안 함(전달 0) | 없음 |
| `strategy_dispatch_cycle.go:114 · :135` 보호 하한 · lease 상한 | 도달 안 함 | 없음 |
| `strategy_proposal_authority.go:203` ResultAuthority | 결과 0 이면 먼저 not-ready | 없음 |

→ 항목 2 의 12 carry 가 지키는 실효는 **관측 정확성 하나**다: 활성화가 검증된 시장이 FX · 설정 · 적재 실패로 닫힌 주기에도 레인 관측 · 투영이 「승격됨(ON)」 을 보인다
(영값이면 그 주기만 OFF/DORMANT 로 보여, 운영자는 장애 주기마다 활성화가 꺼졌다 켜지는 것으로 읽는다). 주문 · lease · 승격 기록 · 잠금 복구 세대에는 영향 0
(잠금 복구 세대는 schedule 값 — `strategy_entry_supervisor.go:544`).

## 3. 선택지

| | 내용 | 노출 | 항목 2 | 변경 폭 | 핀 형태 |
|---|---|---|---|---|---|
| (a) | 관문 계산을 제안 적재 뒤로 되돌림 | 이동 전과 같음(+1 제거) | 재개방 — 실패 주기 관측이 OFF/DORMANT(위 2 의 관측 정확성 상실, 주문 무영향) | `collectMarket` 1 줄 + census 표 · 시험 되돌림 | 8.8.4 census 표를 옛 순서로, carriage 시험의 기대 반전 |
| (b) | ctx 장착 reader(읽기를 ctx 와 경합) | 취소 응답 | 유지 | 공유 `readProductionRouteFile`(route 적재기 공용) 서명 변경; ctx 경합은 읽기 goroutine 이 읽기 수명만큼 남음(파일 시스템 정지면 무기한) — codex 의 「goroutine 안 새는」 제약을 **만족 못함**(OS 읽기는 취소 불가) | stalled reader 주입 · 취소 후 반환 시한 단언 |
| (c) | 필요 폐포만 지연 계산 | 줄어듦 | 부분 — 열둘 모두가 관측을 위해 값이 필요하므로 「필요 폐포만」 이 곧 (a) 또는 이동 유지로 수렴 | — | — |
| (d) | 읽기 앞 `ctx.Err` 선확인 + `collect` join 의 ctx select | 이미 취소된 ctx 는 읽기 전 즉시 닫힘(현재도 Load 첫머리에서 성립); 읽기 중 취소는 join 이 즉시 돌려줌(읽기 goroutine 은 버퍼 채널에 결과를 두고 읽기 수명만큼 남음 — 기존 네 적재기와 같은 수명) | 유지 | **기존 결함 수리** — join 패턴 넷(route · proposal · risk · account)이 같은 값이라 한 곳만 고치면 이웃이 남는다(정정 단위 = 값) | 적재기마다 stalled reader seam + 취소 뒤 시한 안 반환 · 다른 시장 결과 무손상 · goroutine 수 상한(읽기 해제 뒤 0) |

## 4. 권고

**이동 유지 + (d) 를 별도 로트로(이동과 범위 분리).** 근거: 이동이 더한 것은 핀 선언 시장의 실패 주기에 작은 로컬 읽기 1 회이고, 그 읽기가 취소에 응답하지 않는다는 성질은 파도의
적재기 넷 전부가 이미 가진 결함이다 — (a) 로 되돌려도 그 결함은 FX 성공 주기 · 다른 세 적재기에 그대로 남고 항목 2 의 관측 정확성만 잃는다. (d) 는 이동과 무관하게
네 적재기의 join 을 한 규칙(취소되면 즉시 반환 · 미완 시장은 INTERNAL/취소 사유)으로 닫는다. (b) 는 OS 읽기가 취소 불가라 goroutine 잔존 제약을 못 지키고 공유 reader 를 넓힌다 —
비권고. 핀(d): 적재기 넷 각각 stalled reader seam(태그) 으로 「읽기 진입 뒤 취소 → collect 가 시한 안 반환, 막힌 시장만 취소 사유, 다른 시장 결과 그대로, 읽기 해제 뒤 goroutine 0」.
활성화 로트 선행 표시(핀 선언 전 착지)로 ROADMAP 에 걸어도 된다 — 오늘 생산 노출 0(핀 0).

**정정(보이스 2 도착 뒤):** (d) 는 보이스 2 의 P1 을 닫지 못한다 — 시장 goroutine 이 끝난 뒤 ctx 가 취소되면 join 의 select 는 `outcomes` 를 먼저 받을 수 있고(둘 다 준비면 무작위),
그 시장 결과는 이미 낡은 관문으로 READY 다. (d) 와 아래 6 절(X/Y)은 서로 대체가 아니라 직교다.

## 5. P2 처분 제안(한 줄씩)

- **P2-1 nil getenv → 사유가 FX_NOT_READY 에서 INTERNAL_FAILURE 로:** `loadFamilyActivation` 을 nil-안전하게(getenv nil → 미선언이 아니라 Unavailable 오류 반환, panic 없음 → 관문
  되돌림 · 사유 FX_NOT_READY 유지). 생산 생성자는 nil 을 `os.Getenv` 로 바꾸므로(strategy_proposal_authority.go:249) 도달은 손으로 만든 적재기뿐 — 다중 실패 조합 행동 시험 1 개로 핀.
- **P2-2 lane_id 원문 · 개행이 오류에:** 서술자 위치 `descriptors[%d]` + 필드명만 싣고 원문 값은 빼거나 `%q` + 길이 상한 — 같은 로트에서 `TestEveryActivationRefusalNamesItsField…` 표
  기대 갱신 + 개행 주입 시험.
- **P2-3 골든 ↔ 리터럴 미결속:** breakoutlane 시험이 골든 `rvol_counterfactual_ppm[0]` 을 읽어 그 값 정확히 · 1 ppm 아래의 반사실 기록을 행동으로 단언(리터럴을 상수로 바꾸지 않아도 결속) —
  시험 전용.

## 6. 보이스 2 P1 — 관문 스냅숏이 제안 적재 시간만큼 낡음 (갱신 2026-10-04, Manager 지시: X / Y 평가)

**기전(보이스 2 실증, `voice2-output.md`):** `familyGateFor` 의 입력 중 시간에 대해 변하는 것은 둘 — `ctx.Err()`(`LoadProductionFamilyActivation` 이 :440 진입 · :488 출구에서 봄)와
**디스크의 매니페스트 내용**(철회 · 교체). `observedAt` 은 주기 고정값이라 만료 판정은 시점 무관, env 는 프로세스 내 불변, 레인 목록은 같은 `*Lane` 포인터.
`strategyproposal.LoadProductionAuthorityBatch` 는 진입에서만 `ctx.Err` 를 보므로 적재 중 취소 · 적재 중 철회가 옛 순서에서는 관문에 보이고 새 순서에서는 그 주기에 안 보인다
(옛=FAMILY_GATE_CLOSED · handoff 0, 새=READY · handoff 2). 하류 `strategyaccount/production.go:123` 의 `ctx.Err()` 가 취소 쪽을 막을 개연성은 높지만 우연한 안전이고,
**철회 쪽은 하류에 막는 자리가 없다**(계좌 적재기는 활성화를 다시 읽지 않는다 — 1주기 동안 철회된 활성화로 조정 · handoff).

**정정 대상(「값 동등」 주장, 좌표):** `function-logic/…strategyproposalauthorityloader.collectmarket/branch-test-map.md:4` · `measurements/lot-8.8.4-B/pre-edit/…collectmarket/function-logic-map.md:11`
(「판정 자리 불변 · 옮긴 값은 … 조정 관문으로만」) · 코드 주석 `strategy_proposal_authority.go:319-321`(「판정 … 제자리 그대로」) · `review-8.5-2026-10/brief.md:66`(「값 동등」, 리뷰 입력 — 기록으로 두고 합본에 정정 각주).
Y 면 앞 셋은 **참이 되도록 고쳐진다**(아래), X 면 문장을 「판정 자리 불변 · 관문 스냅숏 시점은 적재 앞(철회 1주기 지연)」으로 바꿔야 한다.

| | X: 조기 계산 유지 + 조정 직전 `ctx.Err()` 재확인 | Y: 조기 계산(carry) 유지 + **판정은 옛 자리에서 재계산** |
|---|---|---|
| 구현 모양 | 조정 직전 `if ctx.Err()!=nil && <선언된 시장> { gate = rolledBack }` | 옛 줄 `gate = loader.familyGateFor(ctx, …)` 를 `coordinateMarketProposals` 직전에 **되살림**(조기 줄은 그대로) |
| 취소 경합 | 닫힘(fail-closed) | 닫힘 — `familyGateFor` 자신이 판정(편집 전과 같은 함수 · 같은 시점) |
| 적재 중 철회 · 교체 | **열림** — 그 주기는 철회 전 활성화로 조정(허용 방향 1주기) | 닫힘 — 편집 전과 같은 시점의 디스크를 읽음 |
| 토글 OFF(미선언) | **함정**: 「선언된 시장」 판별을 X 가 다시 해야 함 — 빠뜨리면 미선언 시장이 취소 주기에 FAMILY_GATE_CLOSED(upstream 과 다름, 불변식 3). 판별을 넣으면 `familyGateFor` 의 Undeclared 규칙 사본이 둘째 자리에 생김(판정 이중화) | 해당 없음 — 미선언은 :433 첫 줄 선반환, 재계산도 같은 함수 |
| 판정 동등성 | 근사 — 새 갈래(ctx 재확인) 하나가 판정에 추가, 동등성은 시험으로만 | **구성상 자명** — 조정 이후 13 줄은 편집 전과 바이트 동일한 판정 입력, 차이는 조정 앞 닫힘 여섯이 싣는 진단 값뿐 |
| carry 의미 | 조기값이 조정 · 판정 · carry 공용 | 조기값 = 조정 **앞** 닫힘(FX · 설정 · 열쇠 · 중복 · 적재 · 고장)의 진단 전용; 조정 **뒤** 닫힘 · READY 는 재계산값을 실음(편집 전과 같음) |
| 남는 비대칭 | 철회 1주기 지연(판정) | 조정 앞 닫힘에서 조기값과 「그 주기에 재계산했다면 나왔을 값」이 철회 경합에서 갈림 — **관측 전용**(그 갈래는 조정 · handoff 0) — 명시 기록 |
| 비용 | 0 | 조정 도달 주기에 활성화 읽기 2회(64 KiB 로컬 · 0400; 오늘 핀 0 → 0회) |
| codex P1 · (d) 와 결합 | 읽기 수 = HEAD 와 같음; (d) 는 별개 | 조정 도달 주기 읽기 +1(편집 전 1 → 2) — 둘 다 같은 ctx 없는 reader, (d) 의 join 수리가 같은 규칙으로 덮음(읽기 개수와 무관) |
| 보이스 2 ctx-race 시험으로 핀? | 취소 쪽만 — 철회 쪽은 통과 못 함(READY) | 둘 다 |

**권고: Y.** 근거 — 판정을 편집 전 자리 · 같은 함수로 되돌리면 동등성이 시험 이전에 구성으로 성립하고(Manager 경향과 같음), X 는 철회 경합을 남기면서 Undeclared 판별 사본을
둘째 판정 자리에 만든다(토글 OFF 불변식에 닿는 이중 규칙). 비용은 핀 선언 뒤에만 생기는 소형 읽기 1회.

**Y 의 시험 핀 형태(응답 로트):**
1. 보이스 2 ctx-race 를 저장소 시험으로 승격(태그 `tossos_testseams`): 적재 끝에 `cancel()` → 기대 FAMILY_GATE_CLOSED · 제안 0 · handoff 0, 레인 1 · 2 두 모양.
2. **철회 경합(X 와 Y 를 가르는 핀):** `loadActivation` 스텁이 호출 횟수로 답 — 1회차 verified(gen 7), 2회차 Unavailable → 기대 FAMILY_GATE_CLOSED · handoff 0.
   그리고 1회차 verified gen 7 · 2회차 verified gen 8 → 기대 READY 가 **gen 8** 을 싣고, 같은 스텁에서 FX 미준비 닫힘은 gen 7 을 실음(조기/판정 값의 역할 분리 고정).
3. 미선언 + 적재 중 취소 → 기존 경로 그대로(관문 영값, FAMILY_GATE_CLOSED 아님) — 토글 OFF 핀.
4. 구조 핀: `TestTheThirteenProposalClosures…` 를 「조기 계산 = 경로 준비 가드 직후 · 판정 계산 = `coordinateMarketProposals` 바로 앞 문장 · 조정의 gate 인자는 둘째 대입」으로 갱신(AST).
5. 변이: 둘째 대입 삭제 → 2 · 1 이 RED; 둘째 대입을 조기 줄 자리로 옮김 → 2 RED; 첫째 대입 삭제 → carriage 시험 RED.

## 7. P2 합본 표 (codex r2 + 보이스 2)

| ID | 출처 | 내용 | 처분 제안 |
|---|---|---|---|
| P2-a | codex | nil getenv → INTERNAL_FAILURE(FX_NOT_READY 이어야) | 승인됨 — `loadFamilyActivation` nil-안전(→ Unavailable), 다중 실패 행동 시험 1 |
| P2-b | codex | descriptor 오류에 `lane_id` 원문 · 개행 | 승인됨 — `descriptors[%d]` + 필드명, 원문 배제(또는 `%q` + 상한), 개행 주입 핀 |
| P2-c | codex + 보이스 2 P2-3 (동일) | 골든 `rvol_counterfactual_ppm[0]` ↔ 리터럴 `1_200_000` 미결속 | 승인됨 — 골든 값 · 1 ppm 아래 경계 행동 시험(시험 전용), 한 처분 |
| P2-d | 보이스 2 P2-1 | 읽기 결함 오류가 둘째 `%w` 로 `ErrProductionRouteUnavailable` 도 만족(1,541/13,714) — 편집 전에는 아니었음 | **안쪽 오류를 `%v` 로 접기** — 편집 전 `errors.Is` 사슬로 복원(동등성 회복, 판정 영향 0 실측); 핀: 읽기 결함 표에서 `errors.Is(err, ErrProductionRouteUnavailable)==false` |
| P2-e | 보이스 2 P2-2 | config `market` 항이 판정을 못 바꿈(읽기 실패 · 몸통 `validMarket` 이 가림) — 편집 전부터 | **기록** — 행동상 다른 항이 먼저 닫음을 FLM 에 명시, 메시지 시험이 유일 핀임을 그 행에 적음; 읽기 스텁 핀은 seam 을 새로 여는 비용이라 비권고 |

## 8. Manager 잠정 판정 기록 (2026-10-04, 합본 때 번복 사유 없으면 확정)

- 보이스 2 P1 → **Y 채택**. 근거 ① 판정 = 같은 함수 · 같은 순간이라 동등성이 구성상 복원 ② X 는 Undeclared 규칙을 둘째 판정 자리에 복제(두 판정이 서로를 가리는 금지 패턴).
- (d) 는 이 P1 을 못 닫음 — (d) 로트와 Y 는 별개로 각자 진행. (d) 는 ROADMAP 면제 불가 선행 「핀 선언 전 착지」.
- 핀: 6 절 1~5 그대로(#2 철회 경합이 X/Y 판별 시험, #4 AST 구조, 변이 3종).
- 주장 정정 3곳(BTM:4 · pre-edit FLM:11 · 코드 주석 319-321)을 Y 의 실제 동작 문구로; `brief.md:66` 은 기록 보존 + 합본 각주.
- P2: a · b 기승인, c 한 처분, d = `%v` 접기 + `errors.Is(…ErrProductionRouteUnavailable)==false` 핀,
  e = 기록 전용 — FLM 행에 가림 가드 둘(디렉터리 읽기 실패 · 몸통 `validMarket`)을 명명하고 그 블록을 닫아 두는 등식을 적는다.
- 잔여 비대칭(철회 경합 시 조정 앞 닫힘 여섯이 싣는 낡은 관측값)은 이 브리프(6 절)와 응답 로트의 FLM 에 명시 기록.
- 응답 로트는 보이스 1 · 3 합본 → 최종 판정 뒤 1회(생산 편집 일괄). 그 전까지 코드 금지.
