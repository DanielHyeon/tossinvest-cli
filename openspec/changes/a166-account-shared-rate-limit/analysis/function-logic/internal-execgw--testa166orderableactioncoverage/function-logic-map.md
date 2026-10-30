# TestA166OrderableActionCoverage Pre-Edit

기존 AST의 함수 전체 Action 참조 수집은 미사용 참조를 실제 분류로 오인한다. go/types의 Action 상수 값과 Orderable 결과를 구하고 replayOrigin의 실제 action switch case만 비교한다. 미사용 참조 변이는 빠진 case를 검출해야 한다. 생산 함수 변경 없음; test-only 함수 변경.
