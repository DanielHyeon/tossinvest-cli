# 전략 정확성·손절 정책 단계별 실행 계획

작성: Astra. 2026-10-18. 현 단계는 설계이며 구현·운영 완료가 아니다.

| 순서 | Change | 완료 산출물 | 선행/승인 |
|---|---|---|---|
|1|a148 clarify-parker-risk-reward-contract|공간/목표/실현 RR 계약과 호환 진단|설계 freeze|
|2|a149 compute-session-vwap-and-ema9|세션 지표 순수 계산·검증|a148|
|3|a150 bind-strategy-stop-sizing-and-exit|전략별 손절→수량→원장→exit 연결 검증|a148,a149|
|4|a151 compare-stop-policies-offline|고정/구조/ATR 및 별도 EMA 비교|a150|
|5|a152 qualify-volume-profile-evidence|데이터 지원/불지원 판정|a149; 4와 독립 조사 가능|
|6|a153 evaluate-cost-aware-strategy-variants|비용·holdout·의존표본 반영 비교|a151,a152|
|7|a154 preview-validated-stop-policy-options|검증/승인 상태를 구분한 선택 UI|a150,a153|
|8|a155 observe-versioned-strategy-decisions|주문 권한 없는 replay/관측 adapter|a149,a150,a153,a154; 실제 수집 별도 승인|
|9|a156 gate-live-strategy-policy-rollout|후속 leg·마감·복구와 live 승인 증거|a155 실제 관측 증거 및 사람 승인; 현재 보류|

1~3 정확성 계약부터 순차 수행한다. a152 자료조사는 계산 이후 병행 가능하나 다른 범위의 변경을 한 커밋에 묶지 않는다. a153 자료 부족도 정직한 inconclusive 결과로 완료할 수 있지만, a154에서는 검증되지 않은 선택을 활성화하지 않는다. a155 도구 완료와 실제 관측은 다르며 a156은 실증/승인 미완료 상태로 archive하지 않는다.

Manager 직접 설계 → Terra 단일 writer → 별도 독립 reviewer 및 gstack → Manager 결과물 확인 → 전체 gate → Terra 범위 커밋·푸시·정식 archive/PM sync. 구현 전에 change별 최신 base와 영향 분석을 고정한다. 기존 a128·a147 관측, r1, 운영 설정, live 계좌는 이번 설계에 포함하지 않는다.
