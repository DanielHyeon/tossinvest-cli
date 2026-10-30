# Action 구조 시험 delta

대상 f6874e6b0d0b87b9496f31271376614506d5018f. 독립 검증에서 발견된 P2: 기존 TestA166OrderableActionCoverage는 함수 전체에서 Action 이름이 존재하면 case에 없더라도 통과했다. 기존 함수 편집 전 AST는 `action-coverage-pre.json`, 편집 후 AST/FLM/BTM은 function-logic에 있다. 생산 소스 변경 없음.

go/types로 선언된 Action 상수의 실제 값을 수집하고 각 값의 Orderable을 실행한다. replayOrigin 내부의 `switch exitpolicy.Action(action)` case와 실제 Normal/Critical 반환만 수집하여 이 값 집합 및 보호 action 구분과 비교한다. 상수의 중복 값은 값으로 처리하며 실제 case 중복, 비상수 case, 예상 외 반환 및 누락은 실패한다. `TestA166OrderableUnusedReferenceIsRejected`는 정상 action을 case에서 제거하면서 `_ = exitpolicy.ActionRatchetPartial` 참조를 남긴 입력을 거절한다.

`action-unused-red.log`는 신규 helper 추가 전 컴파일 RED; `action-cases-green.log`는 선택 일반 PASS; `action-cases-race.log`, `action-cases-tag-race.log`는 Go 1.26 race/tag race PASS; `action-cases-go125.log`는 Go 1.25 선택 normal PASS. 누락 검사 제거 overlay는 `action-unused-mutation.log`의 assertion 실패로 검출했다. patch 원문은 gzip -n `action-unused-mutation.patch.gz`, 해제 SHA-256은 `action-unused-mutation.sha256`과 일치한다.

수동 gstack review는 test-only delta의 case 누락·값 동등성·오류 처리·실행 범위를 점검했다. 추가 발견 없음. 이 구조 시험은 runtime 분기와 wire 검증을 대체하지 않는다. 새 네트워크 호출, 전체 테스트/gate, 운영·배포는 수행하지 않았다. impeccable N/A — test-only. 신규 의존성 없음.
