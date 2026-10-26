# Branch Test Map: `TestGuardianPresetsRenderAsOneMatrixPerCurrency`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 통화 집합 수집 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B2 | matrix 수 불일치 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B3 | 행 수 불일치 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B4 | 통화 consequence 수 불일치 | 이 함수 | RED: 프리셋별 attribute | GREEN: current 1회 |
| B5 | consequence 속성 총수 불일치 | 이 함수 | RED: 프리셋별 attribute | GREEN: current 1회 |
| B6 | tier 카드 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B7 | tier form 수 불일치 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B8 | 통화별 가정 결과 문장 수 불일치 | 이 함수 | RED: 문맥 없는 결과 문장 | GREEN: 통화별 1회 |
| B9 | hidden tier 부재 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B10 | preview 행 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B11 | 현재값 불일치 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B12 | preset 값 불일치 | 이 함수 | 기존 보호 | GREEN에서 유지 |
