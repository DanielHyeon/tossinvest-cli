# 실제 read lifetime P1 delta

3dc0c776b에서 resolver의 child context가 반환 전 끝나 실제 read가 signal lifetime을 잃는 P1을 독립 검토가 확인했다. Calendar, Holdings, Orders, InstrumentNames 진입점에서 원래 request와 signal lifetime을 결합해 read가 끝날 때까지 유지한다. 직접 accountRef의 resolve context는 계속 보호한다.

cached client 뒤 calendar read의 signal 취소 RED→GREEN, holdings/orders/names 각각 signal 취소 GREEN 및 lifetime 제거 변이 RED, 원래 request 취소 tagged race GREEN. 기존 선택 정상/tagged race 재실행은 별도 network namespace의 lo-only/no-route에서 통과했다. 사전 AST·분기 분석은 preedit.md 및 옆 JSON. 수동 gstack 범위는 context 취소·cache·오류/부분결과와 직접 resolve 경로를 검토했다. CodeGraph 인덱스 없음, 혼합 모델 gstack·독립 delta·전체 tests/gate·실계좌/운영 미실행이다. 3dc 단독은 P1이 남은 후보로 간주한다.

최신 main 34ec59fe를 병합한 e7c71378d에서 console root owner.Close/relaunch 결속과 verify·HTTP 변경의 조합을 소스 대조하고 Go 1.25 정확 선택 normal/tagged race로 재확인했다. console worker·handler 수명 및 root Close의 대표 선택 시험을 함께 실행했다. 원문은 combined-receipts.md. 현재 변경 전체의 독립 재검증은 별도다.
