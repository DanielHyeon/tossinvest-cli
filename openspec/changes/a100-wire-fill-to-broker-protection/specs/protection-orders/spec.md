## ADDED Requirements

### Requirement: M-A 발동 측정은 durable local causal receipt를 사용해야 한다
시스템의 measurement-only verify 도구는 M-A response를 durable causal receipt로 기록해야 한다 (SHALL).
parent와 child의 official GET response는 한 프로세스의 증가 sequence와 monotonic elapsed time으로
기록해야 한다 (SHALL). 이 도구는 runtime,
engine, protection worker, attestation, trading journal에 연결되어서는 안 되며 (MUST NOT), 새 broker
mutation method나 자동 승인 권위를 추가해서는 안 된다 (MUST NOT).

M0 실행은 exact `--include-trigger --confirm-each --resume --redo conditional-trigger`와 explicit receipt
경로가 아니면 broker factory·confirmer·어떤 broker mutation보다 먼저 거부되어야 한다 (SHALL).
`--include-ttl-edge`, 다른 redo, trigger 외 모든 step의 prior PASS가 없는 상태와 함께 실행해서는 안 된다 (MUST NOT).
prior verify record에 outstanding artifact가 하나라도 있으면 cleanup prologue를 실행하지 않고 broker
factory보다 먼저 HOLD해야 한다 (SHALL).
receipt 경로는 current uid가 소유한 non-symlink 0700 parent 아래의 새 non-symlink 0600/O_EXCL 파일이어야
하며 (SHALL), versioned header/run ID의 file fsync와 parent-directory fsync가 mutation 전에 성공해야 한다
(SHALL). 실패한 receipt를 무시하고 주문을 계속해서는 안 된다 (MUST NOT).
모든 parent path component는 no-follow descriptor walk로 열어야 하고 (SHALL), 한 receipt에는 active M0 run
lease가 정확히 하나만 존재해야 한다 (SHALL). arbitrary append는 없어야 하며 lease-bound typed writer만
사용해야 한다 (SHALL). write/fsync 실패는 receipt를 영구 unusable로 만들어 이후 write/acquire를 모두
거부해야 한다 (SHALL).

M0 mutation, parent raw read와 child raw read는 동일한 concrete official client/account에서 나와야 한다
(SHALL). 별도 Broker/client가 반환한 raw result와 unrelated official attempt를 결합해 evidence를 만들면
안 된다 (MUST NOT).

parent/child official response는 lossy domain mapping 전에 읽어야 한다 (SHALL). transport는 각 HTTP
attempt의 request-start와 body-read-complete를 같은 process monotonic anchor에서 포착하고 numeric status
또는 no-response class, 401/429 attempt와 exact raw `result` bytes를 decode 전에 기록해야 한다 (SHALL).
helper 반환 뒤 시간을 response-received로 대신해서는 안 된다 (MUST NOT). 성공 payload digest는 수신한
result bytes 그대로의 `SHA-256(raw-result-bytes-v1)`이고, non-2xx/invalid envelope는 exact response body의
`SHA-256(raw-response-body-bytes-v1)`이며 extracted-fields schema는 v1이어야 한다 (SHALL). body-read가
완료된 empty response는 empty bytes의 SHA-256을 남겨야 하고 (SHALL), 실제 no-response 또는
body-read-incomplete일 때만 digest를 생략할 수 있다 (MAY).

parent receipt는 request/response conditional ID tag, client ID tag, symbol, market, type, order type,
quantity, first side/trigger, expiry, root/leg status와 child ID tag를, child receipt는 request/response order
ID tag, requested market scope와 raw symbol/side/status/quantity/filled quantity/execution price·quantity를
기록해야 한다 (SHALL). account/token/opaque ID 원문은 causal receipt에 저장해서는 안 되며 (MUST NOT),
pending/create client tag와 parent raw client tag 및 pending approved parent fields와 parent raw fields가
일치해야 한다 (SHALL). parent triggered-child tag, child checkpoint tag, child request tag와 child response
order-ID tag도 서로 일치하고 child SELL·requested market scope·symbol·quantity가 approved parent leg와
일치해야 한다 (SHALL).

trigger create의 human gate 뒤 broker call 전 unique client ID와 approved request scope/run ID는 owner-only
verify record의 pending intent로 append+fsync되어야 한다 (SHALL). create response와 exact parent checkpoint
사이 crash는 다음 M0 resume에서 official all-page raw reconciliation으로 client ID+approved fields의 unique
match를 찾아 checkpoint해야 하며 (SHALL), 그 recovery run은 mutation 없이 HOLD로 끝나야 한다 (SHALL).
zero/multiple/mismatch는 HOLD이고 자동 cancel/recreate해서는 안 된다 (MUST NOT). parent raw의 child ID는
exact child checkpoint에 먼저 append+fsync하고 그 뒤 sanitized parent causal receipt를 fsync한 뒤에만 child
GET을 시작해야 한다 (SHALL). crash/restart에서 human surface는 exact/pending 객체를 보여야 하며 (SHALL),
triggered-but-child-unobserved 객체를 자동 취소해서는 안 된다 (MUST NOT).
`pending-create`만 exact resume에서 read-only recovery에 진입할 수 있고 (SHALL), durable
`parent-created`/`parent-recovered`/`child-observed`는 broker factory 전 manual HOLD여야 한다 (SHALL).
parent POST 성공 뒤 exact-parent checkpoint write/sync가 실패한 현재 run도 typed terminal HOLD로 즉시
중단하고 이후 step을 실행해서는 안 된다 (MUST NOT).

non-empty triggered child id가 포함된 parent receipt가 durable fsync된 뒤에만 official child GET을
시작해야 한다 (SHALL). PASS는 parent child-id receipt seq가 child first-observed-fill receipt seq보다
작고 parent fsync-complete가 child request-start보다 앞설 때만 허용해야 한다 (SHALL). broker server
timestamp나 외부 shell timestamp로 이 local order를 대체해서는 안 된다 (MUST NOT).

durable parent child-ID receipt부터 durable child first-observed-fill receipt까지 모든 HTTP attempt를
기록해야 하고 (SHALL), 그 critical window의 첫 parent/child read error, 401, 429, decode/identity 불완전,
receipt write/sync 실패는 irreversible `INCONCLUSIVE/HOLD`여야 한다 (SHALL). 이후 retry 성공으로 이를
지우거나 다른 process/restart sequence를 이어 붙여 PASS로 만들면 안 된다 (MUST NOT). child fill receipt가
durable해진 뒤 parent terminal GET은 요구하지 않으며, 그 전 parent 404는 read gap으로 처리해야 한다 (SHALL).

#### Scenario: parent receipt가 child 조회보다 먼저 durable해진다
- **WHEN** parent official GET이 non-empty triggered child id를 처음 반환한다
- **THEN** 그 raw evidence receipt의 append·flush·fsync가 성공한 뒤에만 child official GET을 시작하고 두 monotonic 경계를 검증한다

#### Scenario: causal evidence 사이에 429가 발생한다
- **WHEN** triggering 관측 이후 child terminal 관측 전 parent 또는 child GET이 429나 read error를 반환한다
- **THEN** receipt에 실패를 남기고 M-A를 INCONCLUSIVE/HOLD로 종료하며 PASS를 기록하지 않는다

#### Scenario: receipt 경로가 안전하지 않다
- **WHEN** receipt 경로가 기존 파일이거나 권한·write·sync 검증에 실패한다
- **THEN** confirmer 호출과 broker mutation은 0건이며 실행을 거부한다

#### Scenario: 다른 client의 evidence를 끼워 맞춘다
- **WHEN** conditional mutation Broker와 parent/child raw-attempt source가 같은 concrete official client가 아니다
- **THEN** direct constructor와 CLI는 M0를 거부하고 mutation·PASS를 만들지 않는다

#### Scenario: receipt persistence가 한 번 실패한다
- **WHEN** typed attempt 또는 causal append의 write/fsync가 실패한다
- **THEN** receipt를 영구 poison하고 이후 append·lease·PASS를 모두 거부한다

#### Scenario: create 직후 프로세스가 사라진다
- **WHEN** broker가 parent conditional ID를 반환한 뒤 다음 read 전에 process가 종료된다
- **THEN** pre-create pending client intent로 official all-page reconciliation만 수행하고 unique exact parent를 checkpoint한 뒤 mutation 없이 HOLD한다

#### Scenario: 과거 verify artifact가 남아 있다
- **WHEN** M0 entry에서 prior verify record의 outstanding artifact가 하나 이상이다
- **THEN** cleanup prologue와 broker factory를 실행하지 않고 human cleanup/reconciliation을 요구한다

#### Scenario: child identity를 얻은 직후 프로세스가 사라진다
- **WHEN** parent raw response가 child ID를 반환한 뒤 causal receipt 완료 전 process가 종료된다
- **THEN** 먼저 fsync된 exact child checkpoint를 human reconciliation surface에 보이고 과거 sequence를 재개하지 않는다

#### Scenario: retry가 causal gap을 숨기려 한다
- **WHEN** durable parent child-ID receipt 뒤 한 HTTP attempt가 429이고 다음 retry가 성공한다
- **THEN** 두 attempt를 모두 기록하고 verdict를 INCONCLUSIVE/HOLD로 유지한다

#### Scenario: child fill 뒤 parent가 사라진다
- **WHEN** child first-observed-fill receipt가 durable해진 뒤 parent GET이 404가 된다
- **THEN** 이미 닫힌 critical window의 PASS를 뒤집기 위해 추가 parent terminal GET을 요구하지 않는다

#### Scenario: verify process가 재시작됐다
- **WHEN** parent와 child evidence가 서로 다른 process-local monotonic run에 속한다
- **THEN** 두 sequence를 결합하지 않고 M-A를 다시 시작하도록 요구한다

### Requirement: 보호는 journal에 커밋된 상태로 수렴하며 촉발은 이벤트가 아니라 상태다
시스템은 브로커 상주 보호주문을 **journal에 커밋된 포지션 상태로부터만** 계획해야 한다 (SHALL).
계획은 그 상태의 durable 커밋이 반환된 뒤에 시작해야 하며 (SHALL), 커밋 이전이나 커밋과 동시에
브로커 보호 mutation을 발행해서는 안 된다 (MUST NOT).

촉발 조건은 개별 체결 이벤트여서는 안 된다 (MUST NOT). 시스템은 **exit 관리 상태가 열려 있고
유효한 stop을 가진 모든 포지션**을 대상으로 desired 상태와 브로커의 observed 상태를 비교해
수렴해야 하며 (SHALL), 그 대상 집합은 **엔진이 체결시킨 포지션과 이미 계좌에 있던 보유를
구별해서는 안 된다** (MUST NOT).

수렴 완료는 **브로커 응답이 실제로 노출하는 값으로만** 판정해야 한다 (SHALL): M-A가 armed로
증명한 raw status이고, 종료·발동 상태가 아니며, 보호 수량이 보유 수량과 정확히 같고, trigger가
journal 값에서 유도한 값과 같을 때다. raw status와 opaque triggered child id를 접거나 버려서는
안 되며 (MUST NOT), `PAUSED`·unknown·triggering·child 미귀속을 ACTIVE로 판정해서는 안 된다
(MUST NOT).
보호 수량·trigger·만료는 journal에 기록된 값에서 정확히 유도해야 하며 (SHALL), 심볼·시각
근사나 브로커 응답 추정으로 대체해서는 안 된다 (MUST NOT).

수렴한 포지션도 주기적으로 재확인해야 한다 (SHALL). 한 번 수렴했다는 이유로 브로커 조회를
영구히 중단해서는 안 된다 (MUST NOT) — 상주 주문은 이후 취소·만료·정지될 수 있고 그때
보호 미설치 시간이 다시 시작되어야 한다.

한 포지션에 대해 동시에 둘 이상의 등록 시도가 진행되어서는 안 된다 (MUST NOT).
브로커 mutation을 전송하기 직전에 포지션 수량을 다시 읽어야 하며 (SHALL), 계획 시점과 다르면
전송해서는 안 된다 (MUST NOT). 누적 수량이 변하지 않은 체결 정정은 보호 교체를 촉발해서는 안
된다 (MUST NOT). 이미 관측한 상태와 모순되는 스냅샷은 보호를 변경해서는 안 된다 (MUST NOT).

상주 보호주문이 발동해 체결된 사실은 원장에 귀속되어야 한다 (SHALL). 그 체결을 알지 못한 채
같은 포지션에 보호를 재등록해서는 안 된다 (MUST NOT).

triggered child의 canonical owner는 fill snapshot/event보다 먼저 durable해야 한다 (SHALL).
ordinary confirmed attempt와 protection child ownership을 합친 결과가 정확히 하나가 아니면
체결을 적용해서는 안 된다 (MUST NOT). child fill이 ownership보다 먼저 durable해진 경우에는
소급 귀속하거나 delta/hook을 재생해서는 안 되며 (MUST NOT), durable `ATTRIBUTION_FAILED`
reconcile과 alert를 기록하고 계좌 대사가 복구할 때까지 새 보호 mutation을 금지해야 한다 (SHALL).

#### Scenario: 기존 보유가 첫 주기에 보호된다
- **WHEN** 보호 컬럼이 NULL인 기존 포지션이 exit 관리 상태와 유효한 stop을 갖고 있다
- **THEN** 그 포지션은 체결 없이도 수렴 대상이 되고 브로커에 상주 보호주문이 등록된다

#### Scenario: 커밋 이후에만 계획한다
- **WHEN** 포지션 상태가 durable 커밋을 반환하기 전에 프로세스가 사라진다
- **THEN** 브로커 보호 mutation은 0건이고 재기동은 커밋된 상태만을 수렴의 입력으로 삼는다

#### Scenario: 수렴한 포지션은 다시 등록되지 않는다
- **WHEN** ACTIVE 보호 수량이 보유 수량과 같고 trigger가 유도값과 같다
- **THEN** 그 주기는 브로커 mutation을 0건 발행하고 상태를 바꾸지 않는다

#### Scenario: 수량이 변하지 않은 정정
- **WHEN** 브로커가 이미 보고한 체결을 누적 수량 변화 없이 재진술한다
- **THEN** 보호 교체와 브로커 mutation은 0건이고 기존 보호주문은 그대로 유지된다

#### Scenario: 모순되는 스냅샷
- **WHEN** 직전 관측보다 작은 누적 체결 수량이 도착한다
- **THEN** 보호주문을 취소·교체하지 않고 typed reconcile reason을 기록한다

#### Scenario: PAUSED 또는 알 수 없는 raw status
- **WHEN** 수량·trigger는 같지만 raw status가 `PAUSED`이거나 판정표에 없는 값이다
- **THEN** 수렴 완료로 표시하지 않고 mutation 없이 operator alert를 유지한다

#### Scenario: child 체결이 ownership보다 먼저 관측됐다
- **WHEN** triggered child의 scoped fill snapshot이 protection child owner 등록보다 먼저 durable했다
- **THEN** 그 fill을 소급 적용하지 않고 `ATTRIBUTION_FAILED` reconcile/alert를 남기며 새 보호 mutation을 금지한다

### Requirement: 보호 미설치 시간은 관측 가능하고 상한을 갖는다
시스템은 포지션별 **보호 미설치 시간**을 측정해야 한다 (SHALL) — desired 보호가 생긴 시각부터
브로커 ACTIVE 확인까지의 경과 시간이다.
그 시간이 구성된 상한을 넘으면 관측 가능한 알림을 내야 하며 (SHALL),
수렴 실패가 조용한 무보호로 남아서는 안 된다 (MUST NOT).

기본 미설치 상한은 90초, dirty/pending 재시도는 5초에서 시작해 지수 백오프로 최대 60초,
수렴한 ACTIVE의 재확인은 60초여야 한다 (SHALL). 구성 키는
`engine.protection_convergence.unprotected_alert_seconds`, `dirty_retry_initial_seconds`,
`retry_max_seconds`, `active_recheck_seconds`여야 한다 (SHALL). block이 없으면 각각 90/5/60/60초를
사용해야 하며 (SHALL), 명시된 0·음수 또는 initial이 max보다 큰 block은 거부하고 worker를
기동해서는 안 된다 (MUST NOT).

알림 identity는 `(account, market, position, generation, cause)`여야 하며 (SHALL), 같은 미해결
episode에서 중복 발행해서는 안 된다 (MUST NOT). 해결 뒤 같은 cause가 재발하면 새 episode로
다시 발행해야 한다 (SHALL). `/position-management`는 raw status, desired/observed 수량·trigger,
현재 baseline delta, 마지막 성공/오류 시각, 미설치 경과와 같은 cause를 shared projection으로
표시해야 하며 (SHALL), ownership/상태 증거가 불완전할 때 `보호됨`으로 표시해서는 안 된다 (MUST NOT).

수렴 실패는 체결 감지·대사·reduce-only 청산을 실패시키거나 지연시켜서는 안 되며 (MUST NOT),
typed reconcile reason 기록과 알림으로 처리해야 한다 (SHALL).
수렴 worker는 all-or-nothing supervised-loop 집합에 포함되어서는 안 되며 (MUST NOT), recovery 완료
뒤 runtime의 auxiliary executor로 기동·취소·drain되어야 한다 (SHALL). ordinary cycle failure는
worker를 종료해서는 안 된다 (MUST NOT). panic 또는 예상 밖 return은
`protection.convergence.worker_stopped` stable event와 durable reconcile/alert를 남겨야 하지만
(SHALL), 다른 safety loop를 취소하거나 entry gate를 변경해서는 안 된다 (MUST NOT).

#### Scenario: 브로커가 계속 거부한다
- **WHEN** 한 포지션의 보호 등록이 상한 시간을 넘도록 계속 실패한다
- **THEN** 그 포지션을 지목하는 알림이 나고 체결 감지·대사·청산 경로는 영향을 받지 않는다

#### Scenario: 수렴 worker가 예상 밖에 종료된다
- **WHEN** 수렴 worker가 context 취소가 아닌 panic 또는 return으로 종료된다
- **THEN** 전용 stable event와 reconcile/alert가 기록되고 reconcile·exit·filldetect는 계속 실행되며 entry gate는 바뀌지 않는다

#### Scenario: 수렴 중과 실패 중의 구별
- **WHEN** 등록이 아직 확인되지 않았고 상한 이내다
- **THEN** 알림 없이 다음 주기가 재시도하며 미설치 시간이 계속 관측된다

### Requirement: 보호 상태는 additive-nullable journal 컬럼으로만 영속된다
시스템은 보호 상태를 기존 trading journal에 영속해야 한다 (SHALL): lifecycle 상태,
parent/client order id, raw status, triggered child order id, position generation, desired 시작 시각과
lifecycle applied watermark.
별도의 보호 전용 데이터베이스를 기동 의존성으로 도입해서는 안 된다 (MUST NOT).
스키마 변경은 additive이고 새 컬럼은 nullable이어야 하며 (SHALL), 기존 컬럼의 의미를 바꾸어서는
안 된다 (MUST NOT). 값이 없는 행은 「보호 미설치」로 읽혀야 한다 (SHALL).

브로커에 등록을 전송하기 **전에** 그 operation의 identity와 pending 상태를 durable하게 커밋해야
한다 (SHALL). 전송 후에만 기록하는 순서를 써서는 안 된다 (MUST NOT) — 재기동이 조회할 identity가
없으면 복구가 시장을 latch한다.

브로커 보호주문 등록이 확인되면 그 order id를 journal에 커밋해야 한다 (SHALL).

triggered child ownership은 같은 journal의 additive `protection_child_orders` registry에 기록해야 하며
(SHALL), parent/client/account/market/day/symbol/side/generation과 등록 시각을 exact 보존해야 한다
(SHALL). registry의 유일 writer는 수렴 worker가 호출하는 좁은 Journal registrar여야 하고 (SHALL),
직접 SQL writer나 fill detector의 추론 writer를 추가해서는 안 된다 (MUST NOT).

이 journal은 더 새로운 스키마를 이전 바이너리가 여는 것을 거부한다. 따라서 스키마를 올린 뒤의
롤백은 백업 복원이며, 복원된 journal은 상주 보호주문의 기록을 잃는다. 롤백 절차는 **복원 전에
브로커의 상주 보호주문 목록을 기록하고 복원 후 사람이 각 주문의 처분을 정하는 절차로 문서화되어야
한다** (SHALL). 자동 롤백이 상주 보호주문을 취소해서는 안 되며 (MUST NOT), 자동 유지가 안전을
보장한다고 기술해서도 안 된다 (MUST NOT).

#### Scenario: 스키마를 올린 뒤 롤백
- **WHEN** 보호 컬럼이 채워진 journal을 그 컬럼을 모르는 이전 바이너리가 연다
- **THEN** 그 바이너리는 journal을 거부하고, 문서화된 절차는 백업 복원 전에 브로커 상주 주문 목록을 기록하도록 요구한다

#### Scenario: 등록 전송과 응답 사이의 프로세스 손실
- **WHEN** 브로커가 보호주문을 수락한 뒤 order id를 journal에 커밋하기 전에 엔진이 종료된다
- **THEN** 재기동은 전송 전에 커밋된 operation identity로 exact broker 조회를 수행해 기존 주문을 귀속하고 attested idempotency 증명 없이 재제출하지 않는다

#### Scenario: 보호 미설치 행
- **WHEN** 보호 컬럼이 NULL인 기존 포지션 행을 읽는다
- **THEN** 「보호 미설치」로 해석되어 수렴 대상이 되고 기존 청산 경로는 영향을 받지 않는다

### Requirement: 브로커측 매도 청구권이 둘이 되는 창을 최소화하고 남는 창을 밝힌다
시스템은 브로커측 매도 청구권이 둘이 되는 창을 알려진 경로마다 최소화해야 한다 (SHALL).
브로커가 조건주문에 대해 매도가능수량을 예약하지 않기 때문이다 (측정 M13).
남는 창은 문서화되어야 하며 (SHALL), 이 계약이 청구권이 항상 하나임을 보장한다고 기술해서는
안 된다 (MUST NOT).

포지션을 **완전히 닫는 모든 청산**은 — 보호성인지 여부와 무관하게 — 발행 전에 그 포지션의
상주 보호주문 취소를 시도해야 한다 (SHALL).
취소 확인 실패가 인프로세스 매도를 막거나 지연시켜서는 안 되며 (MUST NOT), typed reason을
기록하고 즉시 매도를 진행해야 한다 (SHALL). 포지션이 flat이 되면 남아 있는 상주 보호주문을
취소해야 한다 (SHALL).

상주 보호주문이 먼저 체결되어 포지션이 이미 flat이면 시스템은 **보호 청산을 시도해서는 안
된다** (MUST NOT). 그 포지션은 보호에 실패한 것이 아니라 보호된 것이므로 그 사실을 구별되는
결과로 기록해야 한다 (SHALL). 이미 flat인 포지션에 대해 수량 0의 보호 청산을 제출 경로로
흘려보내서는 안 된다 (MUST NOT).

#### Scenario: 인프로세스 청산이 상주 주문을 앞선다
- **WHEN** 엔진이 살아 있는 상태에서 보호성 시장가 매도를 발행한다
- **THEN** 상주 보호주문 취소를 먼저 시도하고, 취소가 확인되지 않아도 매도는 지연 없이 발행된다

#### Scenario: 상주 주문이 먼저 채워졌다
- **WHEN** 브로커 상주 보호주문이 먼저 체결되어 포지션이 이미 flat인 상태에서 보호 청산 판단에 도달한다
- **THEN** 보호 청산을 시도하지 않고 「상주 주문이 보호를 완료함」으로 기록하며 수량 0의 제출은 발생하지 않는다

#### Scenario: 포지션이 flat이 되었다
- **WHEN** 어떤 경로로든 보유 수량이 0이 되었고 상주 보호주문이 남아 있다
- **THEN** 다음 수렴 주기가 그 주문을 취소한다

### Requirement: 상주 보호의 trigger는 영속된 보호 baseline에서만 유도된다
시스템은 상주 보호주문의 trigger를 영속된 보호 baseline에서만 유도해야 한다 (SHALL).
그 baseline은 exit 관리 상태에 **스칼라로 영속된 값**이어야 한다 (SHALL).
관측값으로부터의 재계산이나 근사로 trigger를 정해서는 안 되며
(MUST NOT), t0에 얼린 초기 stop을 그 baseline 대신 사용해서도 안 된다 (MUST NOT) — 초기 stop은
그 baseline의 t0 값일 뿐이고 이후의 상향분을 담지 않는다.

영속된 baseline이 상주 trigger보다 높아지면 다음 수렴 주기가 trigger를 그 값으로 교체해야 한다
(SHALL). 이미 등록된 trigger를 더 낮은 값으로 교체해서는 안 되며 (MUST NOT), baseline이 후퇴한
것으로 관측되면 보호를 변경하지 말고 typed reason으로 기록해야 한다 (SHALL) — 단조 비감소는
exit 경로가 이미 보장하는 불변식이므로 그 위반은 보호 경로가 흡수할 사건이 아니다.

시스템은 **마지막 교체 이후의 경과와 현재 baseline과의 차**를 운영자가 알 수 있게 표시해야 하며
(SHALL), 상주 trigger가 항상 현재 baseline과 같다고 보고해서는 안 된다 (MUST NOT).

#### Scenario: baseline이 올라갔다
- **WHEN** 영속된 보호 baseline이 상주 주문의 현재 trigger보다 높아졌다
- **THEN** 다음 수렴 주기가 trigger를 그 값으로 교체하고, 교체되기 전까지의 차이가 표시된다

#### Scenario: 후퇴는 거부된다
- **WHEN** 영속된 baseline이 이미 등록된 trigger보다 낮게 관측된다
- **THEN** trigger를 낮추지 않고 typed reason을 기록한다

#### Scenario: 재계산으로는 올리지 않는다
- **WHEN** 현재가와 고점으로부터 영속된 baseline보다 높은 stop을 계산할 수 있다
- **THEN** 그 계산값은 trigger가 되지 않고 영속된 baseline만 쓰인다
