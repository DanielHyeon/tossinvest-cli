# U2 score 사전 검토

base 2d67c226e, 승인 계획 650003f72. domain.Candle의 Time·Volume과 sizing.go의 stdlib math/big 재사용. 신규 비공개 함수/타입만 추가하므로 기존 함수 FLM/BTM 대상 없음. 입력 최대25, 오류 우선순위·정수 범위·KST 자정·중복·최신20·2단계 floor·overflow를 리터럴과 부정 사례로 확인한다. 입력 불변, caller 0. ai-memory a169/U2 검색 결과 없음. 신규 의존성 및 UI 없음. 네트워크·운영·전체 tests/gate 제외.
