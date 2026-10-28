# Branch Test Map: TestRecordCarriesNoUnmaskedAccountNumber

- Source: `internal/verifylive/record_test.go` (156-177, 편집 전). AST branches 5. 시험 함수 자신이 대상이라 「Test」 열은 이 함수와 변이 판정.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 실행 오류도 기록은 검사 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a | yes |
| B2 | 기록 읽기 실패 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a(시험 기반 경로) | yes |
| B3 | 평문 계좌 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a | yes |
| B4 | 마스크 꼬리 부분 문자열 | `TestRecordCarriesNoUnmaskedAccountNumber` | yes — 우연 일치 5/3000(결함) | 2995/3000 |
| B5 | 열린 기록에 마스크 | `TestRecordCarriesNoUnmaskedAccountNumber` | n/a | yes |
