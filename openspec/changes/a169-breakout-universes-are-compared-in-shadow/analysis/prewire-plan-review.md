# a169 prewire fence 독립 설계 검토

대상 `/home/daniel/tossos-work/a169-prewire-fence` HEAD b0f7f1d9381fd88a49ab0d5052693514f7a019c0 및 Manager 미커밋 문서. source 손읽기; AST/시험·운영 미실행. 현재 stable cursor부터 과거 gap 제외는 새 사용자 결정이며 재질문 대상이 아니다. 한정 구현 READY 전 P1 하나의 동시성 순서가 필요하다.

## P1 — sticky 최초 검사와 대기 중 실패의 경합
plan:15,23은 Acquire 전 sticky를 확인하고 Acquire 뒤 shared state를 확인한다고 쓰지만 **Acquire 뒤 동일 fence sticky 재확인**과 cleanup 결과 발표의 순서가 없다. 요청 B가 sticky=nil을 읽고 Acquire를 기다리는 동안 요청 A가 같은 fence의 Finish/cleanup 오류를 내면 B의 사전 검사는 이미 끝났다. event state가 이미 정상 commit된 postcommit fsync 오류 등은 이벤트 변화만으로 A의 local failure를 항상 재구성할 수 없다. B가 A 다음 lease를 획득한 뒤 전송될 수 있어 '같은 fence의 후속 요청 거절'을 입증하지 못한다.

최소 계약: fence owner/유효성 및 sticky를 **lease 획득 후, Begin 이전 shared 검사와 동일한 전송 선형화 영역에서 재검사**하고, A에서 관측한 429·Begin/Finish 및 확정된 cleanup failure를 B가 전송 허가를 받을 수 있기 전에 sticky에 게시한다. `responseOwner`의 EOF/Close/cancel은 RoundTrip 반환 이후 별도 종료 경로이며 `transport_linux.go:366-375`에서 gate.stop/lease.Release가 실행된다. Release 자체 실패는 결과가 lock 해제 후에야 알려질 수 있으므로, 단순 'release 뒤 sticky 저장'으로는 경쟁 wire0을 증명하지 못한다. 동일 fence 내 admission/cleanup의 좁은 동기화나 그에 준하는 선형화 순서를 명세하고 deadlock/무기한 대기 주장을 피할 것. 단, 일반 Aux busy와 단순 요청 취소의 sticky 대상 여부는 plan:23,29와 일치시키며 예기치 않은 세션 중단을 만들지 않아야 한다.

필수 RED: 동일 fence·다른 collector의 B가 A의 postcommit Finish 오류 또는 body Close/Release cleanup 오류 **이전부터 대기**하도록 동기화한 뒤, A 오류가 확정된 이후 B token/GET Write0을 단언한다. 단순 순차 호출만으로는 위 경합을 잡지 못한다. 외부 동일 group writer의 429→B 검사도 별도 유지한다. 이미 Write 허가를 통과한 요청을 소급 취소한다고 주장하지 않는다.

## 확인된 범위와 제약
- `observation_linux.go:616-644` zero cursor Snapshot은 event lock 아래 읽은 유효 상태의 현재 Event/Gap을 반환한다. no pending이면 그 읽기가 기준점 선형화로 가능하며 historical ring overrun을 baseline에서 무시하는 것은 사용자 선택과 일치한다. Snapshot 이후 새 사건은 fence의 고정 cursor와 prewire 검사에서 걸러야 한다. current pending/dirty/읽기 오류는 과거 gap으로 축소하지 않는다. 현재 Snapshot의 generation 변경 경합은 pending 분기에서 별도 실패한다(:645-671).
- `withLease`(:361-370)의 owner pointer·released·wire 검증과 event lock read를 Acquire 이후 Begin 전 사용하면 동일 정상 참여자끼리 wire→event 순서를 지키며 타 writer Finish가 체크와 Write 사이 끼지 못한다는 설계가 가능하다. 그 상태에서 Snapshot을 다시 호출하면 자신의 wire를 Busy로 볼 수 있어 plan의 내부 read 분리가 타당하다. nil/zero/wrong Budget의 명시 fence를 absence로 내리지 않는다. 다른 프로세스가 같은 file ID라도 fence는 작성 Budget pointer와 결속한다는 지원 제한을 명확히 유지한다.
- 기존 Begin은 orphan을 gap으로 바꾸므로(:440-470) 검사 전에 Begin하지 않는 순서가 필요하다. 검사 실패의 body close/lease release 및 Begin 성공 여부에 따른 settlement, HTTP 응답을 받은 뒤 Finish 실패 시 기존 수용 의미 유지가 필수다. `WithTokenObservation`은 context 값을 보존(`observed.go:54-60`)하지만 실제 token/401/adopt/mint 전파는 local TLS fixture로 증명해야 한다. fence 없는 OFF/required 기존 경로 및 현재 collector 동작 회귀 필요.
- baseline 과거 overrun 허용은 현재 state 검증 이후 cursor만 고정하는 경우다. 현재 pending/dirty/poison 지우기·새 instance 전환과 다르다. 재시작시 매번 새 fence를 만들어 과거 사건을 지우지 못하게 하는 당일 abort 영속성, 활성 엔진/profile identity, 실제 참여자 결속은 후속 미완이다. 로컬 first Write 경계는 broker 도착 시각이나 첫 429 예방을 보증하지 않는다.

정본 delta의 no-fence 동작 보존, 사용자 결정, 설계 미완/측정 보류 표기는 일치한다. 위 경합의 명세 및 시험 반례를 반영한 뒤 재검토해야 한다. 전체 tests/gate·실제 broker/운영·코드 변경 없음.

## §5 재검토

Manager 추가 plan:43–53 및 `transport_linux.go:167–185,250–272,319–375`, `budget_linux.go:436–458`을 손읽기했다. 이전 P1은 설계상 해소: B의 Acquire 후 sticky 재검사와 A의 fence mutex 아래 Release 및 반환 오류 게시가 wire unlock→오류 저장 사이를 직렬화한다. body wrapper는 기존 responseOwner의 body.Close→finish 순서를 이용해 Close 오류를 release 전에 알릴 수 있다. gate.stop/Begin/Finish에서 lease mutex를 보유한 채 fence mutex를 얻지 않고, Acquire·네트워크/body drain 동안 fence mutex를 잡지 않는 조건도 명시됐다. postcommit·bodyClose·unlock후Release오류에 대한 대기 B RED와 정상 대조가 추가됐다.

구현시 prewire 검사에서 fence mutex를 놓은 다음 실패 cleanup helper가 다시 fence mutex를 획득해야 한다. 같은 비재진입 mutex를 쥔 채 helper를 호출하면 self-deadlock이다. 검사→Begin 이후 알려진 failure의 게시와 해제 helper까지 전 경로의 lock order를 AST/경합 fixture로 증명해야 하며, 이는 §5의 지정 순서를 구체적으로 이행하는 검증 조건이다. 실제 cleanup fault 주입이 wire unlock 후 반환인지 확인해 false-green을 피한다. 단순 context 취소의 정상 settlement 및 Aux busy 비sticky 처리도 상호 배타적으로 검증한다.

재검토 결론: 추가 P0/P1 미발견, 위 로트에 한정된 **bounded READY**. 설계 정합 판정이며 구현·동시성 시험 통과, 전체 a169 freeze·측정/운영 승인이 아니다. 시험·코드 변경 없음.
