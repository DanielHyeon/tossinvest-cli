# a112 3.8 · 4.5 처분 표 (2026-10-04, Manager 판정 — 감사 후 최소 보강, 시험 전용)

입력: 감사 `audit-3.8-4.5.md`(절 → 기존 시험 사상). 이 표는 감사의 PARTIAL · GAP 행마다 이 로트가 더한 시험과 남는 이연을 적는다.
「새」 = 이 로트가 더한 시험(변이 원장 `../lot-3.8-4.5/mutation-3.8-4.5.tsv`).

## 3.8 — 후보 출처(Toss rank · volume · flow)는 읽기 전용 발견 증거

| 절 | 감사 판정 | 이 로트 | 처분 |
|---|---|---|---|
| a 읽기 전용 발견 증거 | COVERED(구조, 기존) | 새 `strategyflow/a112_candidate_origin_test.go` `TestTossRankEvidenceAloneCannotStandInForOfficialClosedBars`(KR · US, 무태그): 승인 후보(순위 백분위 지표) + 라우팅만으로는 권한 계수 0 · 수량 0 · 후보 증거 digest 는 따로 보존 · 레인 증거 자리는 비어 있음 | 닫음 |
| b 공식 마감 봉 대체 불가 | PARTIAL | 같은 시험: 봉 증거 없음 → LANE_REFUSAL(타입 있는 거절) · 대조(같은 후보 · 같은 라우팅 + 공식 봉 증거) → 제안 성립, 레인 증거 = 봉 스냅숏 digest(후보 digest 아님) | 순수 코어 · 흐름 층 닫음. **생산 절반 이연**: 공식 어댑터가 전이 · 최종 거부 시점에 재검증하는 경로는 3.6(생산자, 생산 importer 0) + 결정 49 벽 뒤 — 감사 「Clauses to defer」 |
| c 세션 대체 불가 | PARTIAL | 새 `breakoutlane/a112_session_authority_test.go`(무태그): 정규장 아님 · 미마감 봉 거절(NewClosedBar), 다른 세션 봉이 섞인 스냅숏 거절(KR · US) — 감사가 짚은 미시험 갈래 둘(types.go NewClosedBar · machine.go validStructuralEvidenceInput) | 순수 코어 닫음. 생산 절반 이연(b 와 같은 사유) |
| d 거래 가능 여부 대체 불가 | GAP | — | **이연(not-applicable in a112)**: breakout 입력에 거래 가능 여부 자리가 없고(`breakoutlane.EvidenceInput` · `strategyflow.BreakoutRequest` 에 해당 필드 0), 생산 경로가 없다(3.6 미착지 · 벽) — 감사 「Clauses to defer」 |
| e Toss 수동 조건 주문 변경 경로 없음 | COVERED(레인 · 증거 · 후보) / PARTIAL(엔진 전략 dispatch) | 새 `app/engine/a112_no_conditional_order_from_strategy_test.go` `TestNoStrategyFileReachesAConditionalOrderMutator`(무태그): 엔진 생산 파일에서 조건부 주문 변경자 · 의도 타입 이름은 보호 어댑터 broker.go 에만 · 전략 파일 0 | 닫음 |

## 4.5 — 짝 KR/US 생산 통합: 8 서술자 · 정확한 계보 · 매트릭스 이행 거절 · 암묵 활성화 없음

| 절 | 감사 판정 | 이 로트 | 처분 |
|---|---|---|---|
| a 8 서술자(생산 경로) | PARTIAL(경로 8/8, 제안 6/8 중 생산 적재 continuation 만) | 새 `strategyproposal/a112_paired_lane_batch_test.go`(태그): reversal · weekly 를 **실 LoadProductionAuthorityBatch** 로 KR · US(서명 매니페스트 · 실 증거 저장소 · 실 RouteSet · weekly 는 실 원장 예약) → 제안 측 6/6(벽 아래). breakout 두 시장 정상 부재(고장 아님). 새 `strategyflow/a112_production_breakout_integration_test.go`(태그): 생산 Evaluate · Propose 에 breakout KR · US → 흐름 층 8/8 | 벽 아래 닫음. **벽 너머 이연**: breakout 제안 · dispatch 의 생산 적재는 결정 49(ErrBreakoutEvidenceUnavailable) — 해제는 B · C 선행(tasks 6.4 행 · review.md 6.4) |
| b 정확한 route/proposal 계보 | PARTIAL | 위 두 시험: 제안 계보의 레인 · 버전 · 수평선 · 라우터 증거 = RouteSet 결정, 레인 증거 = 봉인 스냅숏, 캠페인 = 매니페스트, weekly 는 원장 예약 결속(ID · 주 · record digest); breakout 은 라우터 · 레인 증거 = 순수 코어 스냅숏 digest, 설정 = 코어 설정 digest | 닫음(벽 아래) |
| c 매니페스트 이행 거절 | PARTIAL | 새 `strategyrouter/a112_matrix_migration_refusal_test.go`(무태그): 현재 스키마 · 새 서명 · 새 핀의 세 가족 경로 매니페스트를 **적재기**(LoadProductionRouteAuthority)가 거절(KR · US), 세 가족 활성화 매니페스트도 두 시장 모두 Unavailable | 닫음. `MigrateLegacy`(생산 호출자 0 — ROADMAP a070 이월)는 근거로 세지 않음 |
| d 암묵 desired/effective 없음 | COVERED(성분) · PARTIAL(스펙 문구) | 새 `app/engine/a112_no_implicit_activation_test.go`(태그): 실 원장 + 서명된 네 가족 경로 매니페스트(후보 전부 Desired/Effective ON — 전제 확인)를 실 적재기로 적재, 활성화 핀 없는 생산 관문(실 loadFamilyActivation) = 미선언, 레인 여덟 DORMANT · 봉투 0 · 투영 OFF/OFF/UNOBSERVED | 닫음 |
| e 암묵 LIVE 없음 | GAP(행동) | 같은 시험: 같은 설정 디렉터리(생산 c.Paths.ConfigDir — 운영 설정 config.json 의 trading.allow_live_order_actions 가 사는 곳)와 원장의 바이트가 전후 동일 | 닫음 |

## BTM L 행(4.5-a 제안 측과 같은 갈래) — `btm-remeasure-prod.tsv`

reversal · weekly 의 생산 lane-input 경로 15 행 중 8 이 이 로트의 실 적재기 시험으로 진입(L→ENTERED), 남은 7 은 실패-닫힘 오류 갈래(L→R).
