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

## 라운드 2 (대기)
