# a169 당일 abort·엔진 신원 조사

기준 source 54f1dc59b87b6221d8aadd2b737a236facab133f. 손읽기만 수행, AST/시험/실환경 접근 없음. 제안은 승인된 설계가 아니다.

## Source 사실

- a140 `openLedger`는 directory 생성 후 JSONL을 O_APPEND로 열고 `ledger.write`는 JSON 한 행 Write 및 파일 Sync 오류를 반환한다 (`tools/a140-breakout-signal-shadow/ledger.go:104-130`). 파일 최초 생성에 대한 directory fsync, process 간 lock, transaction/abort tombstone은 이 경로에 없다. `readLedger`는 ENOENT를 빈 원장으로, 잘못된 JSON/schema를 오류로 취급한다 (`:133-158`). append/Sync 실패가 반드시 온전한 미기록을 뜻하지 않으며 crash 후 마지막 행이 부분적일 수 있다.
- `probe.startSession`은 KRX 거래일 session ID를 만들고 ledger 행의 lot·policy digest·session identity를 대조해 기존 universe를 복원한다 (`shadow.go:162-211`). abort를 판단하는 분기는 해당 읽기 경로에 없고 다른 출력 경로 사용도 현재 허용한다. 같은 ledger를 두 tool process가 함께 사용하는 전역 배타 lock은 a140 package source 검색에서 확인되지 않았다. `internal/runlock`은 이름과 달리 두 process 모두 획득 가능한 advisory marker이며 오류도 무시할 수 있게 설계됐다 (`runlock.go:17-33,78-88`); a140 전용 run lock이라는 증거가 아니다.
- `enginelock.Acquire`는 journal 경로별 kernel flock (`enginelock.go:1-8,144-160`), marker는 별도 advisory다. marker는 PID, started/refreshed, build, ReadyAt, ProcInstance를 담지만 budget scope/instance/manifest는 담지 않는다 (`:165-209`). marker write 실패를 engine 시작 거절로 삼지 않는다 (`cmd/tossctl/engine.go:250-265`); Ready/Identify write 오류도 버린다 (`enginelock.go:290-303,318-328`). Fresh 및 Status.Ready만으로 활성 engine 또는 공통 budget의 증명이 되지 않는다.
- `readEngineSignal`은 process 목록·PID 및 boot/starttime 기반 proc-instance와 ready marker를 대조해 재사용 PID/오래된 marker를 구분한다 (`cmd/tossctl/engineready.go:191-246`). 그 결과는 관측 시점의 신호이며 이후 종료/restart, PID namespace·권한·열거 실패, 공통 manifest 소유는 별도다. 이 판단을 a169에 직접 재사용하려면 현재 cmd 내부 경계를 조정하거나 좁게 추출해야 한다.
- a166 ObservationStore는 dirty marker를 확인해 실패하고 (`internal/accountbudget/observation_linux.go:238-245`), marker 생성·file sync·directory sync→임시파일 전체 write 및 sync→rename 및 directory sync→marker unlink 및 directory sync 순으로 저장한다 (`:292-359`). lock과 검증된 root identity 전제에서 crash-uncertain 상태를 감지하는 선례이지 a169 상태 API는 아니다.

## 최소 설계 후보 및 미결

1. 별도 신뢰된 단일 state root/identity 아래 거래일별 session의 잠금·상태를 두고 동일 session 시작 판정과 abort 상태를 원자적으로 직렬화한다. 다른 `--out`, 다른 프로세스 또는 재시작도 동일 root를 사용해야 한다. root/lock 경로 변경, 복제, 삭제, 독립 namespace, 비협조 프로세스를 막는 신뢰·운영 전제가 없으면 당일 금지의 증명 범위는 제한된다. advisory runlock/O_APPEND만으로는 불충분하다.
2. 최초 wire 전에 durable in-progress/uncertain 상태를 만들고 crash 또는 abort 기록 실패 시 재시작이 fail closed되게 한다. 단순히 사건 후 tombstone Write+Sync만 하면 해당 쓰기 전 crash나 디스크 실패 뒤 다음 실행이 정상으로 오인할 수 있다. 성공 완료와 abort를 원자적으로 구별하고 temp/dirty/rename/fsync 및 file·dir 최초 생성 내구성을 계약해야 한다. 기존 a140 ledger는 진단/보고에 계속 쓰되 tombstone authority와 원자 결합이 없으면 ledger 한 행만 권위로 쓰지 않는다. 파일 IO 실패 시 현재 실행 즉시 중단해도 이후 재부팅 차단은 사전 durable 마커 및 root 보존이 필요하다.
3. 강제 중단도 같은 날 전체 수집을 막는지, 정상 종료 뒤 동일 session 재개·다중 universe 단일 세션 수명, 거래일 결정/장외·날짜 변경, operator recovery 권한과 보존 범위는 아직 결정할 정책이다. 하나의 owner 생존을 임시 flock으로 판단하더라도 lock만으로 durable abort 판정은 불가능하다. 검사→다음 wire 사이 다른 process abort를 어떻게 선형화할지도 별도 명세가 필요하며 fence mutex는 process-local이다.
4. active engine 신원은 marker+live PID/proc-instance+readiness를 최소 관찰 재료로 쓸 수 있으나, 엔진 자신이 실제 열린 required observed Budget의 검증된 scope/instance/manifest identity를 해당 process instance와 결속해 신뢰 가능한 수명 신호로 제공하는 새 경계가 필요하다. 자기 주장 파일이나 marker에 문자열만 추가하는 것은 binding·race 증명이 아니다. 결속 주체·권한·namespace, engine 교체/종료 감시 및 감시 실패 시 tool 동작은 설계와 fixture 필요. a166 shared event fence의 group identity와 active engine identity는 구분한다.

예상 검증(미실행): 두 process 동시 시작, abort 쓰기/rename/fsync 각 오류와 kill/restart, partial JSON, 상태 root 복제·다른 output, 정상 종료 뒤 재진입, 거래일 경계, PID reuse/marker stale·불일치/engine restart·manifest mismatch, abort와 prewire 경쟁. 라이브 설정·로그·원장·시크릿은 읽지 않았다.
