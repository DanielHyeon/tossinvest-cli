# 사전 delta FLM·BTM — token nil context

기준 21b9bc85bdf1fbd7947b072a99f53554923b157c. tokenManager.exchange의 기준 AST는 pre-token.json. 기존 OFF는 nil context를 http.NewRequestWithContext에 전달해 ErrTransport 오류를 반환한다. 21b9의 enforcing 분기에서 WithTokenObservation(nil)이 ctx.Value를 호출해 선행 panic이 생긴다. helper 적용 조건에 ctx != nil을 추가하고 nil을 기존 NewRequest에 넘긴다. form·HTTP·응답·OFF 본문은 바꾸지 않는다.

BTM: nil context를 OFF와 enforcing에서 호출하여 동일 오류 identity/문자열 및 wire 0·panic 0을 확인. 기존 정상 token·poll collector fixture 회귀와 nil guard 제거 변이 RED. HTTP 예제 주소는 요청 생성 자체가 거절되므로 실제 dial 없음; 나머지 TLS는 기존 격리 namespace 사용.
