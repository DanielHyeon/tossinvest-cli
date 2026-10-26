# Branch Test Map: `candidateFilterMatrixFromMarkets`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 시장 순서대로 map·세션 구성 | `TestCandidateFilterMatrixPreservesMarketOnlyKeys` | `p2-red-output.txt`: `Registered` field 없음 | `p2-green-output.txt` 통과 |
| B2 | descriptor를 해당 시장 key에 저장 | 같은 시험 | 같은 RED | 같은 GREEN |
| B3 | 처음 본 key를 합집합 순서에 추가 | 같은 시험 | US 전용 key를 표현할 상태 없음 | KR→US first-seen 두 행 확인 |
| B4 | 합집합 key마다 KR·US cell 생성 | 같은 시험 | 같은 RED | 두 행 생성 확인 |
| B5 | KR 미등록 | 같은 시험 | 같은 RED | KR 미등록/US 등록 확인 |
| B6 | US 미등록 | 같은 시험 | 같은 RED | KR 등록/US 미등록·실제 template unavailable 문구 확인 |
| B7 | KR 미등록이면 US label 사용 | 같은 시험 | 같은 RED | US 전용 행 label 확인 |
| B8 | 두 시장 등록·모든 field 동일일 때만 공통 | `TestCandidateFilterMatrixMergesEqualMarketDescriptors`, `TestCandidateFilterMatrixKeepsCategoryDifferenceSeparate` | 기존 `Category` 비교는 이미 GREEN | equal은 공통, Category 차이는 분리 확인 |
