# Branch Test Map: `TestThePresetControlsAskForNoTyping`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록 tier 카드 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B2 | 금지 입력 목록 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B3 | 금지 입력 존재 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B4 | inline 확인 handler 존재 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B5 | 가짜 form 속성 목록 순회 | 이 함수 | RED: 현 구현에 속성 존재 | GREEN: 속성 제거 |
| B6 | 가짜 form 속성 존재 | 이 함수 | RED: 현 구현에 속성 존재 | GREEN: 속성 제거 |
| B7 | Preview 행 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B8 | tier 셀 값 불일치 | 이 함수 | RED: 구조 단언 부재 | GREEN: Preview.To와 일치 |
| B9 | 실제 방향 부재 | 이 함수 | RED: 구조 단언 부재 | GREEN: Preview.Axis와 일치 |
| B10 | 최초 설정·통화 변경 | 이 함수 | RED: 표 아래 문구 부재 | GREEN: note 표시 |
| B11 | note 위치·내용 불일치 | 이 함수 | RED: 표 아래 문구 부재 | GREEN: Note와 일치 |
