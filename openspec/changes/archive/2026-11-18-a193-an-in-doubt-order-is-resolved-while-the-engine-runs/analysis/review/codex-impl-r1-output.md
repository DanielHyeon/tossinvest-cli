**BLOCK — P1 2건, P2 1건. P0는 발견하지 못했습니다.**

기준 `6be702985 → e391565e523c0475eca710b6cafc7ad4d3293f38`. 지정 코드·시험과 계약을 정적으로 검토했습니다. 파일 변경·시험·네트워크·운영 접근 0. 작업 트리도 전후 깨끗합니다.

**[P1] attempt가 종결되면 실패한 critical 알림을 재시도하지 못한다**

- **근거:** [in_doubt_settle.go:176](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:176), [같은 파일:212](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:212), [같은 파일:347](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:347), [recovery.go:30](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/journal/recovery.go:30).
- **왜:** `record()` 실패는 래치를 세우지 않지만, 실패한 이벤트도 보관하지 않는다. 첫 알림 실패 뒤 해소가 CONFIRMED·park로 끝나거나, **park 성공 뒤 park 알림만 실패하면**, 다음 `PendingAttempts`에서 대상이 사라진다. 따라서 “the next sweep retries”는 성립하지 않는다. 특히 비-exit park에는 대신 재시도할 `recordParkCause`도 없다. 기록 실패에 따른 진입 차단은 해당 attempt의 통지를 대신하지 못한다.
- **시험 공백:** [시험:669](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/a193_in_doubt_settle_test.go:669), [시험:877](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/a193_in_doubt_settle_test.go:877)은 Resolver를 계속 실패시켜 대상이 IN_DOUBT에 남는 경우만 잰다.
- **최소 제안:** 실패한 알림을 attempt 조회 결과와 별도로 유지하고 다음 스윕에서 재시도한다. 알림 실패 때문에 해소를 멈추지는 않는다. 실제 Resolver로 종결시킨 뒤 기록자를 복구하는 시험을 추가하고, 실패 이벤트 보존·재시도 삭제 변이를 포함한다.

**[P1] 해소 오류 로그에 브로커 주문번호가 그대로 실린다**

- **근거:** [in_doubt_settle.go:197](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:197), [durability.go:600](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/journal/durability.go:600), [durability.go:605](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/journal/durability.go:605), [log.go:194](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/obs/log.go:194).
- **왜:** 목록에서 주문을 찾았어도 CONFIRMED 전이의 기존 체결 증거 조회나 시각 정렬이 실패하면, 원장 오류에 `orderID`가 들어간다. Resolver는 이를 그대로 반환하고 새 실행자는 **계좌만** 마스킹해 `FieldError`로 기록한다. Logger에도 추가 마스킹이 없다. 성공 결과의 `Detail`을 제외한 것만으로 오류 경로까지 보호되지는 않는다.
- **최소 제안:** 새 실행자의 오류 로그는 attempt id와 고정 오류 분류로 제한한다. 주문번호를 포함한 해소 오류를 주입해 실제 로그 출력을 검사하고, 원문 오류를 다시 출력하는 변이를 추가한다. 현재 민감값 시험은 이벤트 내용만 검사하며 로그를 검사하지 않는다.

**[P2] 손절 비대기 시험은 실제 손절 경로를 측정하지 않고 실시간 스케줄링에도 의존한다**

- **근거:** [a193_in_doubt_settle_test.go:534](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/a193_in_doubt_settle_test.go:534).
- **왜:** 이름이 `exit`인 루프는 실제 `ExitObserver`가 아니라 카운터 증가와 `time.Sleep`뿐이다. 이 시험이 증명하는 것은 보조 실행자와 별도 goroutine의 동시 진행이다. 공유 원장·알림기 등을 거치는 다른 포지션의 손절 제출까지 증명하지 않는다. 또한 50ms 동안 최소 5회 진행 조건은 부하 높은 환경에서 정상 구현도 실패할 수 있고, 실패 분기에서는 `cancel()`에 도달하지 않는다.
- **최소 제안:** 실제 ExitObserver와 가짜 제출기를 사용해 해소가 막힌 동안 다른 포지션의 제출 도달을 채널로 확인한다. 시간은 실패 시한으로만 사용하고 취소·drain을 cleanup에 등록한다. 변이 원장의 m13 대체 근거도 이 검증 범위에 맞춘다.

작성자 증거는 재실행하지 않았습니다. 구현 로그는 선택적으로 읽었고, 변이 README의 `26/26 CAUGHT`는 **작성자 기록**입니다. 패키지 검증 실패 기록과 해당 선택 시험의 후속 성공 기록도 구분했습니다. 현재 HEAD 전체 PASS나 배포 승인을 뜻하지 않습니다.

외부 스킬·ai-memory·CodeGraph 재조회는 경로·네트워크 제한으로 수행하지 않았습니다. 기존 저장소 증거와 현재 소스를 대조했습니다. `impeccable: N/A — 백엔드 실행자·시험 리뷰이며 UI 변경 없음`.

확인했고 문제없던 항목:

- D0 대상 거름: `IN_DOUBT ∧ PLACE`만 선택하며 원장의 `recorded_at, rowid` 순서를 유지한다.
- 스윕 순서: 모든 대상의 처음 봄·stale 평가 후 직렬 해소하며, 종료 후 5초 쉰다.
- 알림 key·창: 지정 key와 창 0, 성공 시 래치, 원장 읽기 성공 시 연속 초기화가 구현돼 있다.
- Q-E 정상 경로: 무장 intent 일치 시 실행자 park 알림을 생략한다. 목록 읽기 실패 시 중복 가능성은 승인된 잔여다.
- 취소 경로: 같은 ctx가 Resolver·원장 전이까지 전달되고, park 쓰기 실패 뒤에는 Gate.Block에 도달하지 않는다. 기록 도중 취소 경합은 D3의 기존 잔여다.
- 실행 구조: 복구 뒤 Auxiliary로 시작하며, 실행자 맵은 단일 실행 경로에서 접근한다. 새 goroutine 생성·재생·주문 mutation 호출은 없다.
- 변경 경계: a166 census와 지정 배선 고정 네 함수는 변경되지 않았다. 해소의 강제 시간 상한 부재·required 모드 임차 지연은 이미 공시된 잔여로 구분했다.