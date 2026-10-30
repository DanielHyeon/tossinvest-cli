# Branch Test Map: buildReconcileResolveRuntime

| 분기 | 증거 | 상태 |
|---|---|---|
| required 오류/취소 | token·Accounts wire 0, 원인 유지 | 설정 오류 GREEN; 취소 미실행 |
| local TLS 성공 | 공식 client로 token/Accounts 및 계좌 판정 | 선택 GREEN |
| OFF | 기존 오류·계좌 선택 회귀 | 단일 fake 계좌 GREEN; 전체 분기 미실행 |
| 옵션 누락 변이 | 실행 전 static guard; 별도 net namespace | static RED→GREEN; 옵션 누락 변이 미실행; factory 우회 변이 RED |
