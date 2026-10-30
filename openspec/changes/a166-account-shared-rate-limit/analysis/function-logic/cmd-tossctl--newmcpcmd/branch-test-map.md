# Branch Test Map: `newMCPCmd`

| 분기 | 시험 목표 | 상태 |
|---|---|---|
| eligible + required 설정 오류 | 서버나 공식 요청 전에 오류 | RED → GREEN; bypass 변이 RED |
| WTS-only | factory 열지 않고 기존 동작 | 선택 WTS fixture GREEN |
| OFF eligible | 기존 공식 client·주문 게이트 | 선택 WTS fixture GREEN |
| eligible + required 정상 | root factory 동일 budget 사용 | 후속 local fixture 필요 |
