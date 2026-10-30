# a169 shared 429 prewire 연구 (read-only)

기준 54f1dc59b87b6221d8aadd2b737a236facab133f. 아래는 원문 손읽기이며 AST/시험 증명이 아니다. 초안 analysis/a166-bridge-draft.md는 freeze/구현 승인이 아니다. ai-memory 해당 marker scope 검색 a169 OR a166 결과 없음.

## 최소 선형화 후보

1. required observed v2 및 동일 scope/instance의 Budget을 검증한 뒤 관측 baseline을 만든다. `Budget.ObservationSnapshot`은 closed/unobserved/nil을 거절하고 store로 위임한다 (`internal/accountbudget/consumer_snapshot_linux.go:8`). `Snapshot`은 cursor의 scope/instance와 미래 cursor를 검증하고 ring overrun을 계산한다 (`observation_linux.go:616`, `:675`). baseline은 pending 없는 안정 view의 현재 cursor를 채택하고 시작 이전 event는 새 event로 판정하지 않는다. 과거 gap/overrun을 무시할지, 보수적 거부할지와 Busy 재시도/기한은 초안에서 미결이다. pending Busy를 orphan으로 추정하면 안 된다.
2. 각 a169 logical request 및 OAuth/401 재전송에 동일 불변 baseline과 취소 상태를 context로 결속한다. `WithTokenObservation`은 기존 context와 collector를 보존하며 kind/key만 바꾼다 (`observed.go:49`, `:54`); `tokenManager.exchange`는 이를 적용해 동일 client로 전송한다 (`internal/official/token.go:132`). 401 경로 전체에 같은 context가 유지되는지 구현 시 실제 official fixture로 증명해야 한다. 기존 collector는 logical call 자기 오류/429만 latch하며 그룹 공유 fence는 아니다 (`transport_linux.go:197`).
3. `Transport.RoundTrip`의 `Acquire` 성공 직후, 자기 `Begin` 이전에 그 lease 소유를 검증하는 좁은 store 내부 읽기를 추가하는 것이 최소 후보다 (`transport_linux.go:227-245`). queue 선택과 wire 취득은 `budget_linux.go:549-570`, auxiliary는 `:612-639`; wire를 가진 상태에서는 다른 정상 참여자의 Finish/Begin이 진행될 수 없으므로 event lock으로 직렬 읽고 baseline 이후 event/gap, overrun, scope/instance 불일치, 오류, 기존 pending을 확인한다. 기존 pending은 다른 정상 owner가 wire를 소유할 수 없으므로 orphan/미확정으로 fail closed; Begin을 먼저 호출해 orphan을 gap으로 변환할 필요는 없다. 외부 비협조 조작은 이 전제 밖이다.
4. `Snapshot`을 그대로 wire lease 아래 호출하면 pending 검사에서 `wire.lock` 재획득을 시도하고 자기 소유를 Busy로 오인한다 (`observation_linux.go:643-671`); 따라서 lease 소유의 내부 read/view 재사용이 필요하며 범용 callback은 불필요하다. wire→event lock 순서만 사용하고 event lock을 놓은 뒤 Begin; wire lease는 유지한다. `Begin`은 이전 pending을 gap으로 바꾸고 자기 pending을 생성한다 (`:440-475`). 기존 pending을 우선 거부하면 자신의 pending에 대한 오탐과 Begin 전 orphan 변환 누락을 피한다. event 자체는 Finish에서 헤더 상태에 따라 저장된다 (`:484`, `:545-559`).
5. 검사 성공 후 Begin, cancellation 확인, gated first Write 순서를 지킨다. `gatedConn.Write`는 stop/context를 잠금 아래 점검하고 first-write 표식을 둔다 (`transport_linux.go:143-164`). 이 순서의 선형화는 wire lease 보유 중 fence read이며 실제 Write가 그 뒤 실행된다. 취소 경쟁에서 이미 first Write를 통과한 요청을 없던 것으로 주장할 수 없다. 실패 시 pretransport body Close, 성공한 sequence만 NoResponse Finish, lease Release 및 collector 오류 결합; wire 0이면 ErrNotSent. 기존 처리 `transport_linux.go:254-272`를 재사용하되 Begin 실패의 불명확한 durable pending은 추측해 Finish하지 않는다. 이미 보낸 요청의 HTTP 응답은 Finish 오류로 바꾸지 않고 collector/cleanup으로 후속 시도 차단한다 (`:319-357`).

## 미결 및 별도 blocker

- baseline 전에 기록된 gap/overrun을 영구적 위험으로 볼지, 안정 cursor로 시작 전 이력 제외할지 명시 필요. 읽기와 baseline 설정 사이, monitor 실패와 다음 first Write 사이의 동기화 및 여러 동시 logical request의 abort latch 소유권/수명·취소 계약 필요. baseline view 사용 후 무감시 시간의 이벤트는 다음 lease 검사에서 포착되어야 한다.
- 현재 Snapshot은 pending 확인 중 generation 변경 시 오류이므로 Busy를 bounded retry할지 명시해야 한다. Finish 실패 후 postcommit cleanup처럼 state 내구 성공과 pending/dirty가 남는 경우를 구분하고, 모든 오류를 경계에서 fail closed하되 이미 응답한 요청 결과는 보존한다.
- 동일 budget identity는 실제 활성 엔진 identity/프로필 결속의 증명이 아니다. same-day abort의 durable 기록, 다중 프로세스 및 재시작 방지 역시 별도 미해결이다. 설계 승인이나 운영 준비 주장 불가.
- 검증 계획: 공유 두 process local TLS에서 상대 429 완료→release→다음 token/GET Write 0, pending orphan, event/gap overrun 및 instance mismatch, 자체 Begin 정상, Begin/Finish fault, cancel vs first Write, 401/adopted/minted token context 및 원 응답 유지. 네트워크/시험은 본 조사에서 실행하지 않았다.
