# a169 prewire secondary source review — draft only

기준 HEAD b80997c1cb21019c033cca0419f490bcc3057212 위 미커밋 source. 작성자 변경 중이므로 고정 SHA 판정 아님. 원문 손읽기, AST·시험 미실행.

- `prewire_fence_linux.go:24-41`은 observed Snapshot의 현재 cursor와 generation을 저장하고 pending 거부, 과거 overrun 플래그는 생성 거절에 쓰지 않는다. `:56-96`은 Budget 포인터, lease 소유·Begin 전, scope/instance·generation 역행·event/gap 차이·pending을 확인하며 generation 증가는 허용한다. Snapshot을 lease 아래 호출하지 않는다.
- `transport_linux.go:198-214,264-293`은 선택 fence가 있을 때 사전 검사와 Acquire 후 재검사, 실패 시 body Close와 lease release를 사용한다. `:304-348,400-470`에서 Begin/Finish·429·cleanup 실패 sticky와 미전송 ErrNotSent 경계를 손읽었다. nil fence 경로는 분리되어 있으나 행동동등은 아직 독립 시험하지 않았다. `prewire_fence_linux.go:98-117`은 mutex 보유 중 Release 반환 오류까지 보존하며 body.Close 오류는 Release 전에 latch한다. fence mutex 아래 network drain은 보이지 않는다.
- `internal/accountbudget/observed.go:54-60`의 token observation은 parent context를 감싸므로 context value 보존, `internal/official/token.go:132-149`가 token 호출에 적용한다. 실제 401/adopt/mint와 교차 프로세스 fixture는 `internal/accountbudgetruntime/a169_fence_integration_test.go`에 존재; 실행·결과 검증 안 함.
- 새 API의 production caller는 확인 범위에서 아직 확인하지 못했다. 고정 snapshot에서 전체 non-test 호출 검색 재확인 필요. source 손읽기상 즉시 재현되는 P0/P1은 찾지 못했으나 증명이나 PASS 아님. 동시 release/Finish/body 오류 fixture·정상 cleanup 대조, nil/owner/역행 및 race, no-fence 회귀 원문 결과가 최종 평가에 필요.

## 고정 source 대조 결과

고정 HEAD `2764aa7016182eb6d62ced48871618900d09abdc`. 당시 b80997c HEAD 위 미커밋으로 읽은 파일이 이번 commit에 들어간 것이므로 b80997c..2764 비교에는 새 source 전체가 나온다. 고정 파일을 다시 읽어 위 소견과 같은 분기를 확인했다. `rg 'NewObservationFence\(' --glob '*.go' --glob '!**/*_test.go'` 결과는 Linux/unsupported 선언 두 곳뿐이며 production 생성 호출 0이다. private context key 생성은 `ObservationFence.Context`뿐이므로 production 결속 호출도 확인되지 않았다.

무 fence source 한정: 기존 context에서 key 부재면 새 fence 분기 모두 건너뛰며 요청 body Close 오류 결합, 기존 collector, Begin/Finish, gate.stop 후 lease Release, responseOwner의 Close·ErrNotSent 판정 순서는 기존과 동일하게 읽힌다 (`transport_linux.go:191-196,216-277,297-348,400-480`). HTTP endpoint/JSON parser·구성은 diff 대상이 아니다. 이는 손읽기 근거이며 동적 HTTP/JSON byte 동일성 또는 전체 회귀 PASS 주장이 아니다. source 범위 P0/P1 발견 없음; 독립 시험·race·Manager 최종 판정은 별도다. RoundTrip FLM/BTM의 사후 작성 절차 미준수는 수정된 source가 없어지는 것이 아니며 기존 기록을 유지해야 한다.
