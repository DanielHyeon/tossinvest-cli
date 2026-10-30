# 구조 검사 구현·선택 검증

기준 source는 engine poll 게시 `bc8ae32ee53bffeb89a14508da3dbd8be562d079`를 통합한 `50d9247a6e4620ccf6304571fc14397c07c44363`. 앞의 plan은 구현 전 계획이다. 이번 결과는 그 계획의 실제 구현 범위와 한계를 우선한다. 생산 소스 수정 없음. UI 없음: impeccable N/A — backend 정적 시험만.

`cmd`, `internal`, `tools`의 비test Go 전체를 parser로 탐색하며 build tag도 제외하지 않는다. 공식 constructor, budget open/factory, HTTP Do/RoundTrip 및 일부 stdlib direct 함수의 selector 호출을 116행의 파일·함수·receiver·분류·근거 목록과 대조한다. 모르는 새 selector, 해당 함수값 escape, 대상 dot import, 알려진 세 shadow 이외의 import alias shadow는 오류다. 기존 세 shadow는 해당 함수의 본문 형태를 별도 고정했다. 다만 `Budget`, `NewClient`, `RoundTrip`의 비호출 selector는 기존 필드/타입 용례로 목록에 기록한다. 식별자만의 동적 호출, reflection, unsafe, 별도 wrapper, endpoint 값의 전역 변경까지 일반적으로 해석하지 않는다. `reason`은 source 손읽기 분류이지 Go 타입 검증 결과가 아니다. 따라서 이것은 전체 가능한 송신의 증명이나 운영 activation 승인이 아니다.

핵심 경계 함수 83개는 go/format으로 정규화한 선언 전체를 `shapes.json`과 비교한다. 분기·대입·return·receiver를 모두 고정하므로 의미 있는 수정을 포함한 어떤 본문 변경에도 재검토가 필요하다. Manager 승인된 한정 shape 접근이며 범용 call graph가 아니다. Required root→observed runtime→same budget official/token, engine/flatten→replay, a112 clone, probe와 도구 owner, OFF 분기 및 offline provision/check를 포함한다. 새 engine poll `buildGateway`, `NewEnginePollRetrier`, reader binding, `pollContext`, collector 오류, `Retrier.Query`, Budget key/remaining을 추가했다. source 손읽기에서 poll binding은 같은 budget/client와 account reference를 확인하고 collector 결과를 query retry에 연결하며 기존 fallback과 전송 자체를 바꾸지 않는다. shape의 텍스트 동일성은 그 동작의 독립 실행 증거를 대체하지 않는다.

사전 조사: 기존 parser 검사와 `go/parser`, `go/ast`, `go/format` 재사용. 추가 의존성 없음. 기존 source-sites 검색은 이전 기준의 탐색 기록이며 최종 목록은 `sites.json`. 새 테스트만 작성해 기존 함수 본문 변경 및 FLM/BTM 적용 대상 없음. source-copy 변이 여덟 개는 실제 source 파일과 네트워크를 변경하지 않는다. root 오류의 OFF fallback, factory raw New, token DefaultClient, replay raw transport, probe raw send, 죽은 guard + raw return, 새 미분류 Do, Do 함수값 escape 각각을 검사기가 거절한다. 원문에서 거절 사유를 확인할 수 있다.

Go 1.25.0 정확 선택 normal/tagged race 각 1회: `TestA166WiringCensusBaseline`, `TestA166WiringCensusRejectsBypasses`와 8개 subtest, `TestA166CancelRetryAdmissionRefusalPreservesFirstAmbiguous`, `TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome`, `TestA166ReplayAdmissionRefusalPreservesInDoubtJournal` PASS. 앞 두 회귀는 fake broker, replay는 local TLS와 `a166FixtureDial`의 정확 loopback 주소 결속이다. 각 실행은 별도 user/net namespace, lo만 있고 route 없음. 대상 패키지에서 setns/unshare 경로 없음 확인. gzip -n 원문 및 압축해제 SHA를 `receipts/`에 보존했다. 원문 마지막 두 명령의 각각 PASS를 확인했다.

수동 gstack checklist 범위: test-only diff와 두 JSON baseline 검토. SQL/데이터 변경, LLM, shell 입력, UI, 새 enum, 배포 경로 없음. 파일 traversal은 오류와 symlink를 거절하고 변이는 temp 복사본만 사용한다. 시험은 병렬 실행하지 않고 source 변경 없는 namespace에서 실행했다. 범위 내 새 P0/P1 없음. 전문 자동 specialist dispatch, 설치·업그레이드·telemetry는 실행하지 않았다. 설계 특성상 shape는 작은 무관한 변경에도 실패하고 수동 재분류가 필요하다.

미실행: 전체 test/gate, 전체 tag/OS, 실제 broker·계좌·운영·activation, 실행 중 참여자의 identity/namespace 검증, 임의 동적 호출 전체 해석. engine poll 다른 회귀 전체와 후속 consumer 변경의 source 합성은 별도 확인 대상이다. 독립 검토 및 Manager 최종 승인 전 후보이며 2.4 전체 완료나 a166 archive를 뜻하지 않는다.
