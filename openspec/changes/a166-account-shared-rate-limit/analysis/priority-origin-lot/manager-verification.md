# Manager 검증 — 보호 우선순위·replay 출처

대상 `f6874e6b0d0b87b9496f31271376614506d5018f` 및 test-only 구조 보강 `5c1d1b907a26bc4636f223215d273c5b0f7b39ae`.

Manager는 engine observe/record/submit, flatten 시작 context, borrowed Budget 배선, required replay 분기와 exact-intent journal 조회 및 action 구조 검사 원문을 손읽기했다. 보호 판단의 Query 진입 전 critical, 보호 취소·발주 critical, 익절 normal, flatten 최초 bootstrap critical의 범위가 승인 계약과 맞는다. token/ACK의 priority와 deadline 계승은 해당 fixture로 확인한다. 미입증 required replay는 count 증가·전송 전에 QueryFallback으로 넘기고 DB 오류를 숨기지 않는다. OFF HTTPReplay와 a087 admission 및 replay attestation nil을 유지한다.

독립 Terra의 Go 1.25 normal·tagged race 지정 네 패키지 원문 `/tmp/a166-priority-independent-normal.log`, `/tmp/a166-priority-independent-tagged-race.log`를 확인했다. normal은 journal/execgw/engine/cmd 각각 3.030/4.837/1.568/0.995초, tagged race는 34.012/23.761/16.824/2.017초 PASS다. 원문 압축본 21개의 SHA 대조와 변이 검출 기록도 확인했다.

Manager와 독립 검토가 식별자 등장만 세던 action 구조시험의 false-green 공백을 발견했다. 실제 typed Action 값·Orderable 결과와 해당 switch case/반환을 비교하는 delta 및 미사용 참조 반례로 보완했다. `/tmp/a166-priority-delta-independent-normal.log`, `/tmp/a166-priority-delta-independent-tagged-race.log`의 구조시험 2개와 ACK·실제 critical flatten replay·required engine 조립 3개가 통과했다. execgw/engine normal 1.616/1.124초, tagged race 6.381/5.834초다. 생산 delta는 0이며, 이는 일반적인 Go 제어흐름 증명기가 아니라 현재 분류 구조의 검사다.

모든 최종 네트워크 독립 실행은 별도 user/network namespace의 loopback만 활성화하고 외부 route가 없는 상태였다. scope 수동 gstack review 및 독립 검토이며 전체 자동 review/gate 실행 주장이 아니다. source P0/P1과 알려진 P2는 해소됐다.

잔여 한계: exit event 조회와 MarkReplayStarted는 별도 DB 연산이다. 정상 writer의 pending 재무장 거절, 미결 verdict release 제한, recovery 선행 및 후속 lifecycle append는 원문으로 대조했다. 외부 임의 DB 변경이나 비정상 직접 journal API 동시 쓰기까지 원자적으로 막는다고 주장하지 않는다. 출처 미입증의 보수 fallback은 보호 replay를 지연할 수 있다. CodeGraph는 이 worktree에 미초기화여서 hard evidence로 주장하지 않고 AST/FLM과 손읽기 범위를 구별한다.

판정: 이 로트의 비강제 병합·푸시 승인. 최신 main의 승인 문서·다른 로트를 보존하고 scoped strict OpenSpec·PM·diff를 확인한다. 전체 tests/gate, 실제 운영·실계좌·활성화·archive는 미승인/미실행이다. a166 전체 완료 판정은 후속 429·나머지 caller·도구·운영 경계 증거 뒤에 한다.
