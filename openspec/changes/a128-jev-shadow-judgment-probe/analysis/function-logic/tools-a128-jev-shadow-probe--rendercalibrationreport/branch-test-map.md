# Branch Test Map: `renderCalibrationReport`

B1~B18 은 base(bbe4ceaa6)와 같은 분기다 — 기존 report 시험(`report_test.go`)이 그대로 통과하고, base 코드로 만든 골든이 새 출력의
접두로 바이트 그대로 남음을 아래 첫 시험이 고정한다. s4 가 더한 분기는 B19~B22.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1~B18 | 옛 원장(top10·s3 선별·라벨 결손·미라벨·gap) 보고가 base 와 같음 | `TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay` | no (골든은 base 코드 산출물) | yes |
| B19 | 적중 규칙 있는 질문마다 시간대 표 | `TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay` | yes (state 해시 검사 삭제 변이 → FAIL) | yes |
| B20 | 적중 규칙 없는 질문은 시간대 표 없음 | `TestUnknownQuestionGetsNoTimeOfDayTable` | no | yes |
| B21 | 로트마다(top10·s3 로트 각자 표) | `TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay` | no | yes |
| B22 | 시장마다(KR·US·ALL) | `TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay` · `TestTimeOfDayTableJudgesAtThirtyLabels` | yes (J3 `>=`→`>` 변이 → FAIL) | yes |
