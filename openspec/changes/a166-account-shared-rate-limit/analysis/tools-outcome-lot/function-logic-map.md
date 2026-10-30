# a130 / d2 함수 분기 지도

기준 34ec59fe7643b3a20b680b663199f354c228f992. 사전 AST 원문 세 JSON, 변경 후 AST 세 JSON 참조. 사전 분석은 `pre-edit.md`로 기록했으나 완전한 분기별 FLM/BTM은 코드 변경 전에 작성하지 못했고 이 문서는 사후 보완이다.

- a130 기존 openCandles B1–B3: 경로·자격정보 오류 및 부재 거부. 변경 후 B4–B6: 설정·factory·client 오류 거부. 함수가 반환하는 method closure가 실제 조회 마지막까지 같은 factory를 참조한다. run이 모든 반환에서 factory Close를 호출하고 실패를 기존 오류와 결합한다.
- a130 기존 run B1–B12: 인자·원장·schema·목록·조회·preview·쓰기 오류와 출력. 기존 흐름 유지. 신규 run 첫 분기들은 close callback 존재·실패·성공 코드 변경을 처리한다. lots 0이면 fetchPrices가 candles 호출을 생략하며 factory는 열리지 않는다. preview도 lots가 있으면 기존대로 조회한다.
- d2 기존 run B1–B15: 날짜·계획·파일·종목·dry-run·설정·자격·types·수집·쓰기. runWithOptions로 옮긴 본문에 factory 설정·open·client 세 거부를 삽입한다. dry-run은 그 앞에서 반환. 기존 5분 context 및 rate-limit 오류는 유지한다.
- AST는 분기 열거를 증명한다. closure 수명·에러 결합과 defer 순서는 소스 손읽기와 선택 시험 근거다. 전체 기존 분기 재실행 주장은 아니다.
