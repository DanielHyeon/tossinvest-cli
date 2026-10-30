# Branch Test Map: newverifyreconcilereader

| 분기 | 필요한 확인 | 상태 |
|---|---|---|
| required 오류·취소 | wire 0, 오류 반환 | 선택 fixture 또는 정적 binding 검증; 전체 분기 미검증 |
| local TLS 정상 | 최초 조회와 seq-bound client의 동일 budget | 선택 fixture 또는 정적 binding 검증; 전체 분기 미검증 |
| 잘못된 행·수·seq | 기존 refusal 보존 | 선택 fixture 또는 정적 binding 검증; 전체 분기 미검증 |
| OFF | 기존 선택과 좁은 wrapper 보존 | 선택 fixture 또는 정적 binding 검증; 전체 분기 미검증 |
