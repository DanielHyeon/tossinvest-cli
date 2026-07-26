# Tasks: adopt-external-positions

> 선행: ① 리뷰 라운드 3 FREEZE 판정 ② **add-operator-dashboard의 journal 조각(RO open·계좌 질의) landed 후 착수**(design A9). 실효는 게이트 ON + `adoption.enabled` 사람 승인 이후. `internal/console`은 task 2.7(대시보드 자격 표시 확장 — 대시보드 landed 후) 외에는 건드리지 않는다.

## 1. 원장 (v7, additive만) [T]

- [x] 1.1 `position_adoptions` 테이블 + `positions.adoption_id` 컬럼(design A1 DDL이 정본 — 단방향 참조, DEFAULT 없는 nullable ADD COLUMN) — decisions 테이블·CHECK enum 무접촉, ErrSchemaTooNew(§0.6), **스키마 골든 목록 갱신**(schema_test wantTables·v6 테이블 목록 2곳)
- [x] 1.2 `adoption_id` **set-once 전용 tx API** + 정적 스캔 가드(소유 파일 외 UPDATE 언급 거부; `entry_decision_id`에는 어떤 쓰기도 추가하지 않음). 편입 코드의 `internal/journal` 파일 배치를 기존 guarded-column 전문 스캔과 정합시킬 것
- [x] 1.3 exit 자격 **단일 술어 함수**(`entry_decision_id OR adoption_id`) — 소비 지점 4곳(exitloop 열거·position_projection·position/provenance·reconcile external) 중 fold 가드는 **`entry_decision_id` 명시 비교로 좁혀 유지**(자격 술어와 분리 — design A1), `ExternalPositionAlert.ExitEligible` 하드코딩 false를 자격 술어로 교체, drift 테스트
- [x] 1.4 exit_state open의 **자격 출처별 분기 신설**(편입 포지션은 결정 조회가 성립하지 않으므로 LookupDecision 경로를 타지 않는다): position_adoptions에서 EntryPrice=observed_price·InitialStop=synthetic_stop(HighWater는 개설 규칙이 entry로 자동 seed). lineage `ADOPTION → POSITION → EXIT_EVENT` 질의 arm 추가
- [x] 1.5 조정으로 수량 0 → exit_state completed + **ADJUSTMENT_CLOSED exit_event + 알림, trade_outcome 행 없음**(design A7 — 편입·엔진 포지션 공통, 고아 방지)
- [x] 1.6 **성과 동결의 편입 분기**(design A7): computeTradeOutcome의 `decisionID==""` 조기 반환을 adoption_id 분기로 확장 — 매수 leg은 position_adoptions에서 합성(수량 = adoption.quantity, **기준가 = observed_price 단일** — cost_basis는 산식 제외·기록용), 매도 leg은 이 인스턴스 귀속 **전 발의자 매도 fill**(exit 루프: exit_events 체인 / flatten: 발의자 결정·saga 참조 — 심볼 시간창 매칭 금지, flatten 종결 포함·빈 매도 leg 동결 금지), realized_r 분모는 합성 initial_risk. trade_outcomes 스키마 무변경

## 2. 편입 파이프라인 [T]

- [x] 2.1 **엔진 reconcile 구동 루프 신설**(design A6): 주기 60초, 전체 스냅샷(미체결 pagination ≤ MaxPages + holdings + 통화별 잔고 — 부분 실패 폐기) → Stabiliser(수집 2회) → 비교·fold → **Tracker.Observe 구동**(확정 하한 캡의 전제) → 편입 판정. §0.4 계상 문서화(주기당 수집 2회 × (2+통화 수)콜 + 편입 후보 시세 배치 1콜·MaxPages 상한). 실행 술어 `AutomationStatus.Verified`
- [x] 2.2 편입 판정·영속: enabled=true + 비RECONCILE + 신선 조건(Stabiliser·staleness ≤ 10s) + 비제외 → position_adoptions 영속(observed_price는 **편입 tx 직전 시세 관측·staleness ≤ 15s — 초과 시 연기**, exit 관측과 동일 경로 `[기존 제약 float64]`; cost_basis는 **원문 문자열 보존** — official 어댑터에 holdings 원문 접근 additive 추가) → adoption_id set-once → exit_state open. 크래시 복구·재대사 시 기편입 인식
- [x] 2.3 **manage-forward 테스트 2종**(design A2): ① 편입 트랜잭션은 매도 발의 0건 ② 편입 관측가와 다른 첫 exit 관측(P1≠P0, 원가 대비 ±50% 포함)에서 래칫 규칙이 정상 적용됨(발의 발생이 정상임을 고정)
- [x] 2.4 config: `adoption.enabled`(기본 false — zero-value 안전 테스트), `default_stop_pct`(**0.02 ≤ pct < 1** 검증·거부·하한 근거 provenance), `exclude_symbols`. enabled flip audit(§0.5)
- [x] 2.5 알림·이벤트(design A4): **무관리 보유 알림은 enabled 무관 존치**(기존 동작 — §0.2), 편입 성공 이벤트가 대체, 제외·실패 알림, 전이 상태만 무알림, 외부 수량 증가 감지 알림
- [x] 2.6 trade-analytics 구분 집계(adoption_id 조인, 표본 수 병기)
- [x] 2.7 대시보드 자격 표시를 편입 포함으로 확장(design A9 — 대시보드 landed 후, `internal/console` 소폭 수정의 유일 예외)

## 3. 완료 게이트 [M]

- [x] 3.1 §0 검토 기록: 편입 tx 무발의(A2)·기본 OFF(§0.2)·flip 승인(§0.7)·알림 존치(A4)·긴급 중지 서술(A5), flatten이 편입 포지션을 덮는지 테스트
- [x] 3.2 테스트 전수(-race)·`openspec validate adopt-external-positions --strict`
- [x] 3.3 `make gate CHANGE=adopt-external-positions`
