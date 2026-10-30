# Function Logic Map: buildReconcileResolveRuntime

기준 c412c0e55의 AST는 ast.json. CodeGraph 인덱스 부재; 원본 손읽기.

| 분기 | 기존 결과 | 보존 |
|---|---|---|
| config/경로/자격 실패 또는 자격 없음 | 오류 반환 | 기존 분류 |
| Accounts 오류 | 공식 계좌 해석 오류 | ctx 및 원인 유지, required 실패는 wire 0 |
| 계좌 참조 거절 | 오류 반환 | 계좌 선택 기준 유지 |
| journal 열기/복원 실패 | 오류 반환, 복원 실패 시 journal 닫음 | 순서·정리 유지 |
| 성공 | collector와 tracker 구성 | 공식 read/order 경로 유지 |

client 생성 한 지점만 root factory로 변경. 시험 전용 officialOptions는 authority를 부여하지 않고 local TLS로 봉인한다.
