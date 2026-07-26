# 리뷰 기록: adopt-external-positions

## 라운드 1 (2026-07-27, 적대적 리뷰 — 판정 REVISE, P1 8·P2 7·P3 4)

Manager 재검증: 하중이 실리는 5건을 코드로 직접 확인 — 전부 사실. (a) `decisions` CHECK enum 동결(execution_contract.go:144-147), (b) exitloop.go:505 RiskIntent 단언, (c) IngestExternalPositions/Converge 프로덕션 호출자 0, (d) journal.Open의 무조건 mkdir+migrate, (e) trade_outcomes에 진입/청산가 컬럼 부재.

### Manager 처분 (2판에 반영)

| 발견 | 처분 |
|---|---|
| P1 첫 관측 틱 대량 매도 | **채택 — 설계 변경**: manage-forward t0(EntryPrice·워터마크 = 편입 시점 관측가, 원가는 분석용). 리뷰어의 min(A,P_obs) 클램프만으로는 승자 즉시 부분익절이 남아 불충분 — 관측가 단일 기준으로 즉시 매도 0을 SHALL NOT로 동결(design D2, 회귀 테스트 task 2.3) |
| P1 ADOPTION class 스키마 거부 | **채택 — (b)안**: 별도 `position_adoptions` + `positions.adoption_id`(set-once). decisions·class 축 무접촉 → engine-safety·risk-management delta 불필요가 논증됨(design D1) |
| P1 메인 스펙 3개 충돌 | 부분 채택: reconciliation delta 신설. engine-safety·risk-management는 class 무접촉 설계로 충돌 자체가 소멸 |
| P1 관측 소스 부재 | 채택: reconcile 구동 루프 신설을 task 2.1로 명시(주기 60s·§0.4 계상) |
| P1 entry_decision_id 첫 변이자 | 채택: 그 컬럼은 읽기만, adoption_id set-once + 정적 스캔(design D1, task 1.2), 불변 SHALL NOT을 position-ledger delta에 성문화 |
| P1 delta 자기모순(RiskIntent.stop) | 채택: t0 요구사항 문장 재작성 + 단언 확장 task 1.4 |
| P1 편입 토글 부재 | 채택: `adoption.enabled` 기본 false(§0.2), flip은 §0.5+§0.7(design D3) |
| P1 kill switch 오도 문구 | 채택: proposal에 "긴급 중지의 정직한 서술" 절 신설(D4) — exit 일시중지는 §0.3 위반이므로 만들지 않음을 명시 |
| P2 float 손실·수수료 미측정 | 채택: 원문 decimal 보존 SHALL(position-ledger delta·task 2.2), averagePurchasePrice 비용 포함 여부는 `[미측정 — 2b 실측 대상]`으로 preimage 문서에 태그 |
| P2 pct 범위 가드 | 채택: 0<pct<1 설정 거부 SHALL + 시나리오 |
| P2 신선도 3개념 | 채택: Stabiliser + staleness ≤10s 조합 지목(D5) |
| P2 인터록 런타임 술어 | 채택: `AutomationStatus.Verified`로 정확화(D5) |
| P2 편입 후 추가 매수/부분 매도 | 채택: 동결+알림 / 수량0→completed+ADJUSTMENT_CLOSED(D6). 비율 회계 완전 정합은 후속 change로 기록 |
| P2 알림 모순 | 채택: D8로 통일(제외·실패만) |
| P2 합성 R 혼합 | 채택: trade-analytics delta 신설(adoption_id 조인·표본 수 병기) |
| P2 design.md 부재 | 채택: design.md 작성(D1~D9) |
| P3 ladder 하드코딩·guarded 스캔·lineage·재편입 루프 | 채택: 문구 정정·task 1.2 주의·lineage 형태 성문화·재편입 의도 명시 |

대시보드와의 `internal/journal` 소유권 충돌(대시보드 리뷰 P1-1)은 D9로 순차화: 대시보드 journal 조각 선행.


## 라운드 2 (2026-07-27 — 판정 REVISE, 신규 P1 7·P2 4·P3 4)

Manager 재검증(4건 코드 확인 — 전부 사실): trade_outcomes.go:175-179 조기 반환, reconcile Holding에 가격 필드 부재+전체 스냅샷 수집, breakEven이 exit_state.EntryPrice 사용, 1.8 핸드오프 landed(b6a0879).

### Manager 처분 (3판에 반영)

- **P1 알림 회귀(§0.2)**: 채택 — 무관리 보유 알림을 enabled 무관 존치(A4), 무알림은 전이 상태만. exit-policy·reconciliation delta 양쪽 복원
- **P1 편입 성과 행 부재**: 채택 — A7 신설: 엔진 청산 편입 포지션은 동결 경로 확장(매수 leg 합성)으로 성과 행 생성, **외부 매도 종결은 성과 행 없이 completed+ADJUSTMENT_CLOSED 이벤트+알림**(매도 leg 부재의 정직한 처리 — 이 결정이 P1 provenance 컬럼 충돌도 소멸시킴)
- **P1 절대 무발의 주장**: 채택 — "편입 트랜잭션은 발의를 생성하지 않음(SHALL NOT)" + "이후 가격 이동 발의는 정상"으로 정직화, **pct 하한 0.02**(하한 근거 provenance), 테스트 2종 분리(A2)
- **P1 observed_price float**: 채택 — 원문 보존 SHALL은 cost_basis 한정(원문 접근 additive 추가 태스크), observed_price는 exit 관측과 동일 경로 + `[기존 제약 float64]` 태그
- **P1 §0.4 계상 불일치**: 채택 — A6 재작성: 전체 스냅샷×2회/주기 60초, Tracker.Observe 포함, MaxPages 상한, reconciliation delta에 수치 고정
- **P1 순환 의존**: 채택 — A9: 대시보드는 entry 자격만으로 선착지, 편입 change task 2.7이 확장(콘솔 예외 명시)
- **P2 fold 가드 재정의**: 채택 — 가드는 entry_decision_id 명시 비교로 좁혀 유지(자격 술어와 분리), ExitEligible 하드코딩 교체(task 1.3)
- **P2 FK 순환·중복 컬럼**: 채택 — 단방향 참조(position_adoptions.position_id 제거), 스냅샷·권위 명시(A1)
- **P2 원장 확장 규칙 v6 고정**: 채택 — position-ledger delta에 MODIFIED 추가(v7 포함·design 참조 change-id 한정, A1~A9로 개칭)
- **P2 manage-forward 귀결**: 채택 — "편입일 가격+비용이 보호 바닥" 명시(A2·proposal·사용자 보고)
- **P3 골든 목록·task 1.4 문구·HighWater seed·Stabiliser 미수렴**: 전건 채택

## 라운드 3 (대기)

## 라운드 3 (2026-07-27 — 판정 REVISE, 잔여 P1 2·P2 2·P3 4)

라운드 2 P1은 6건 중 5건 종결 확인(알림 복원·provenance 충돌 소멸·계상 일치·순환 의존 해소를 리뷰어가 코드·문서 대조로 검증). Manager 재검증: realized_r 산식(trade_outcomes.go:200-211)·핸드오프 TTL 상수 — 인용 그대로.

### Manager 처분 (4판에 반영)

- **P1 observed_price 신선도 공백**: 채택 — "편입 tx 직전 시세 관측·staleness ≤ 15s 초과 시 연기" SHALL(exit-policy delta·tasks 2.2), §0.4 계상에 편입 후보 심볼당 시세 1콜 추가(reconciliation delta·A6·tasks 2.1)
- **P1 매수 leg 기준가 이원화**: 채택 — **observed_price 단일화**(분자·분모 t0 epoch 일치), cost_basis는 기록·표시용으로 격하·cost_basis_src 산식 분기 삭제(A7·tasks 1.6)
- **P2 매도 leg 조인**: 채택 — exit_events→mutation_attempts→fill_events 명시 조인 SHALL(A7·tasks 1.6, 시간창 매칭 금지)
- **P2 대시보드 시점 상대 괄호**: 채택 — (a)안: 괄호 삭제·자격 정의를 exit-policy/position-ledger에 위임(대시보드 스펙), task 2.7은 델타 없는 구현이 됨
- **P3 4건**(A6 예시 수치·proposal 태그·Impact 콘솔·무관리 확정 래치 구현 노트): 전건 채택

## 라운드 4 (대기 — 확인만)
