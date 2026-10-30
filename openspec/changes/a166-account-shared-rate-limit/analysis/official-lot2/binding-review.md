# lot2b token 클라이언트 결속 확인 API

기준 `2c561732119f313619e04aa07da1235ee9cb1729`. 신규 leaf `UsesAccountBudget`는 기존 함수 내부 수정이 없어 기존 함수 FLM/BTM은 N/A다. 이 기록은 구현 후 작성했다. `configMu.RLock` 아래 sealed·enforced·expected identity·정확 wrapper·tokenManager와 client의 동일 hc를 검사한다. 반환은 bool뿐이며 credential·budget·HTTP pointer를 노출하지 않는다. 금전 authority와 구분해 local test client도 결속 확인이 가능하다. 정상 옵션은 봉인 뒤 no-op이며 결속 검사는 그 lock과 동기화된다. 호출 후 임의 무단 메모리 변조까지 방지하는 lease가 아니고 기존 설정의 private 봉인 불변식에 의존한다.

RED는 메서드 부재로 컴파일 실패했다. nil·foreign·OFF·미봉인·부분 필드 누락·wrapper/token drift 및 로컬 비금전 결속, 동시 옵션 재실행을 검사한다. token manager 검사와 wrapper identity 검사를 제거한 각 변이가 해당 분기에서 실패했다. 수동 gstack 범위 검토는 오류 경계·비밀 노출·권한 혼동·race를 대상으로 했다. 전체 mixed-model 리뷰·운영·전체 gate는 미실행이다.
