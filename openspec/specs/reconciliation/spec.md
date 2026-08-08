# reconciliation Specification

## Purpose
토스 계좌를 최종 권위로 하는 로컬 상태 대사 계약(스냅샷 순서·허용 오차·안정화), 재시작 복구 시퀀스, 불일치 시 진입 차단·청산 유지 요구사항을 정의한다.
## Requirements
### Requirement: Reconciliation 계약

로컬 상태와 토스 계좌의 대사는 명시된 계약을 따라야 한다(SHALL): 스냅샷은 (미체결 목록 pagination 완주 → 보유 → 잔고) 고정 순서로 구성하고 as-of 시각을 기록하며, 부분 실패한 스냅샷은 폐기한다(SHALL). 비교 키는 계좌·심볼·lineage 해소 후 현재 주문번호. 수량 허용 오차 0, 평균단가는 decimal 문자열 비교 + 문서화된 epsilon(진입 차단 판정에서 제외). 안정화는 최소 간격(기본 2초)을 둔 연속 2회 동일 스냅샷으로 판정한다(SHALL). 로컬 intent와 매칭되지 않는 브로커 주문·포지션은 external provenance로 분류한다. 충돌 시 토스 계좌가 항상 우선한다(SHALL).

로컬 포지션 상태의 출처는 **Position 투영**(체결 이벤트+조정 이벤트 — position-ledger)이며(SHALL), fills-only 파생과 별도의 두 번째 포지션 계산을 두지 않는다(SHALL NOT). 비교는 심볼 수준에서 수행하고 투영은 비-CLOSED 인스턴스의 합으로 축약한다(SHALL — 보유 스냅샷의 market 차원 제공 여부는 `[미측정]`이며, 제공 전까지 심볼 합산이 비교 단위다). external 분류된 브로커 보유는 조정 이벤트로 투영에 편입하되(SHALL — 청산 수량 판정이 실제 보유를 알아야 한다), exit 관리 자격 없는 포지션의 발견은 `adoption.enabled`와 무관하게 알림을 발송하고(SHALL — 전이 상태 제외; exit-policy 편입 계약이 알림·편입 규칙의 정본), enabled=true이면 편입 후보 판정으로 이어진다(SHALL).

**대사 구동 루프**(SHALL — adopt-external-positions design A6): 엔진은 주기 60초의 구동 루프로 전체 스냅샷 수집(미체결 pagination ≤ MaxPages 50 + holdings 1콜 + 통화별 잔고 N콜) → Stabiliser(수집 2회 필요) → 비교·fold → `Tracker.Observe`(차단·해제 상태기계 — 확정 하한 캡의 전제) → 편입 후보 판정을 수행한다. 이 루프가 "재대사 최소 간격 30초" 요구사항의 주기적 재대사 절차이며, §0.4 계상은 정상 상태 주기당 수집 2회 × (1 + 1 + 통화 수)콜 + 편입 후보 발생 시 시세 배치 1콜(후보 전체 한 번에 — observed_price, staleness 상한은 exit-policy 편입 계약), 상한은 MaxPages다(SHALL 문서화). fold의 "entry 결정 상속 금지" 가드는 `entry_decision_id` 명시 비교로 유지한다(SHALL — 편입 포지션(adoption_id만 설정)에 fold가 착지하는 것은 정상 재대사 경로다).

#### Scenario: 외부 수동 주문 발견
- **WHEN** 로컬 journal에 없는 미체결 주문이 계좌 조회에서 발견되면
- **THEN** external로 분류·기록되고 엔진 상태와 분리 추적되며 알림이 발송된다

#### Scenario: 외부 포지션의 투영 편입
- **WHEN** 로컬 투영이 0인 심볼에 브로커 보유가 발견되면
- **THEN** 외부 분류 조정 이벤트로 투영에 편입되고, 무관리 상태로 확정되면 알림이 발송되며, enabled=true이면 exit-policy 편입 계약의 판정으로 이어진다

#### Scenario: 편입 후 재대사
- **WHEN** adoption_id가 설정된 포지션이 다음 대사 주기에 다시 관측되면
- **THEN** fold 가드(entry 결정 상속 금지)에 걸리지 않고 수량 비교가 정상 수행된다

### Requirement: 재시작 복구
프로세스 재시작 시 엔진은 journal의 미확정 intent 해소 → 계좌·미체결·체결 조회 → 로컬 상태 재구성 순서의 복구를 완료한 후에만 신규 주문을 허용해야 한다(SHALL). 엔진은 공식 계좌 목록의 첫 레코드에 비어 있지 않은 계좌번호와 양수 account sequence가 함께 있을 때만 그 레코드를 기본 계좌로 해석해야 하며(SHALL), 둘 중 하나라도 유효하지 않으면 뒤 레코드로 건너뛰지 않고 기동을 거부해야 한다(SHALL). 클라이언트에 양수 sequence가 명시되어 있으면 엔진은 그 값이 첫 레코드 sequence와 정확히 같을 때만 기동해야 한다(SHALL). 기동이 이 레코드를 성공적으로 해석했다면 같은 공식 클라이언트의 재시작 복구는 동일 레코드의 sequence를 재사용하고 첫 account-scoped 조회 전에 계좌 목록을 다시 요청해서는 안 된다(SHALL NOT).

#### Scenario: 복구 완료 전 주문 시도
- **WHEN** 복구 절차가 완료되기 전에 신규 주문 요청이 발생하면
- **THEN** 요청은 거부되고 복구 미완료 사유가 반환된다

#### Scenario: 기동 계좌 해석의 sequence 재사용
- **WHEN** 엔진 기동 계좌 해석이 한 번 성공한 뒤 재시작 복구가 첫 account-scoped 스냅샷 조회를 수행하면
- **THEN** 같은 sequence가 `X-Tossinvest-Account`에 사용되고 `/api/v1/accounts`는 다시 호출되지 않는다

#### Scenario: 엔진의 명시된 sequence 일치
- **WHEN** 엔진의 공식 클라이언트가 명시적 account sequence 7로 구성되고 첫 계좌 레코드 sequence도 7이면
- **THEN** 기동과 이후 account-scoped 조회는 같은 계좌번호와 sequence를 유지한다

#### Scenario: 엔진의 명시된 sequence 불일치
- **WHEN** 엔진의 공식 클라이언트가 명시적 account sequence 99로 구성되고 첫 계좌 레코드 sequence가 7이면
- **THEN** 엔진은 journal을 열기 전에 계좌 해석 실패로 기동을 거부한다

#### Scenario: 첫 계좌 레코드가 불완전함
- **WHEN** 공식 계좌 목록의 첫 레코드에 계좌번호가 없거나 sequence가 양수가 아니고 뒤 레코드는 유효하면
- **THEN** 엔진은 뒤 레코드를 암묵적으로 선택하지 않고 계좌 해석 실패로 기동을 거부한다

### Requirement: 불일치 시 진입 차단
허용 오차를 넘는 불일치가 확인되면 신규 진입은 차단되고(SHALL) 청산 경로는 유지된다(SHALL — 확정 하한 규칙). 재대사는 최소 간격(기본 30초)을 두고 수행하며, 연속 3회 실패 시 영구 불일치로 표기하고 운영자 확인 절차를 요구한다(SHALL). 대사 성공 시 실패 카운터는 리셋된다. 차단 범위(계좌/시장/심볼)와 자동·수동 해제 조건은 reason-code와 함께 상태표로 정의한다(SHALL).

해제 규칙의 정밀화(SHALL): 비영구 차단의 자동 해제는 **조정 이벤트가 반영된 뒤의 재조회 일치**에만 근거하며 신규 release cause(ADJUSTMENT_APPLIED 계열)와 원인 기록을 남긴다. 조정 없이 우연히 일치한 단발 관측은 영구 차단을 해제하지 못하고(SHALL NOT), 영구 불일치의 해제는 운영자 확인뿐이다(SHALL).

조정 이벤트의 발행은 compare-and-append여야 한다(SHALL): 스냅샷 수집과 조정 커밋 사이의 체결 반영 경쟁을 막기 위해, 조정 커밋 트랜잭션 안에서 기대 이전 값(투영 수량)과 체결 watermark의 불변을 재검증하고, 어긋나면 조정을 폐기하고 재수집한다(SHALL — 뒤늦은 조정이 최신 체결을 이중 차감해서는 안 된다).

#### Scenario: 수량 불일치 감지
- **WHEN** 로컬 포지션 수량과 계좌 수량이 다르면
- **THEN** 신규 진입이 차단되고, 청산 주문은 확정 하한 기준으로 계속 가능하며, 알림이 발송된다

#### Scenario: 조정 반영 후 자동 해제
- **WHEN** 조정 이벤트가 반영되고 재조회가 일치하면
- **THEN** 비영구 차단이 ADJUSTMENT_APPLIED 원인 기록과 함께 자동 해제된다

#### Scenario: 조정과 체결의 경쟁
- **WHEN** 스냅샷 수집 후 조정 커밋 전에 같은 심볼의 체결이 반영되었으면
- **THEN** 조정은 기대 이전 값 불일치로 폐기되고 재수집이 수행되어 이중 차감이 발생하지 않는다

#### Scenario: 영구 차단의 운영자 해제
- **WHEN** 영구 불일치로 승격된 뒤 재조회가 일치하면
- **THEN** 자동 해제되지 않고 운영자 확인을 요구한다

### Requirement: 보호주문 불일치는 신규 진입을 차단하고 수렴한다
reconciliation은 broker conditional orders와 local protection saga를 비교하고 missing, duplicate, orphan, quantity mismatch를 typed discrepancy로 격리해야 한다 (SHALL).

#### Scenario: broker orphan
- **WHEN** 계좌에 local saga가 모르는 활성 조건주문이 있다
- **THEN** 자동 취소하거나 귀속을 추정하지 않고 RECONCILE로 격리하며 신규 진입을 차단한다

#### Scenario: flatten
- **WHEN** 운영자가 포지션 전량 flatten을 승인한다
- **THEN** 2초 안에 관련 보호주문의 terminal cancel과 broker sellable quantity를 확인한 경우에만 기존 reduce-only liquidation을 실행한다

#### Scenario: flatten cancel이 모호함
- **WHEN** cancel 응답이 유실되거나 trigger 경합으로 2초 안에 terminal 상태를 확인할 수 없다
- **THEN** saga를 `IN_DOUBT`로 격리하고 신규 exposure를 차단하며 최우선 reconcile과 사람 emergency action을 요구하고 oversell 가능한 blind liquidation을 제출하지 않는다

### Requirement: External order provenance remains stable across cycles

An order classified as external because no local intent, confirmed mutation attempt, or lineage owns it SHALL remain external across polling cycles. Persisting a broker observation MUST NOT by itself turn that order into local open exposure, and its later absence MUST NOT create a missing-local-order mismatch or contribute to permanent promotion.

#### Scenario: Stored external observation is absent later
- **WHEN** reconciliation runs after a previously stored external open-order observation disappears from the broker open list
- **THEN** the corrected local state contains no such open order, the diff contains no missing order for it, and the tracker failure counter does not increase because of it

### Requirement: Open-order comparison preserves canonical identity

Local reconstruction and authoritative broker comparison SHALL preserve account, market, market-local trading day, symbol, side, and opaque order identifier as one canonical identity. A broker order that shares only the opaque identifier MUST NOT satisfy a locally owned order in another scope. If broker evidence omits a field needed to distinguish multiple local candidates, the comparison MUST remain blocking instead of declaring a clean account.

#### Scenario: Same identifier appears in another canonical scope
- **WHEN** a locally owned order is absent but the broker snapshot contains the same opaque identifier with a different market, trading day, symbol, or side
- **THEN** reconciliation reports the local order as missing and the broker order as external or conflicting, and operator recovery refuses release

#### Scenario: Broker scope is insufficient for reused local identifiers
- **WHEN** more than one local canonical order shares one opaque identifier and the broker payload cannot distinguish their required scope
- **THEN** the comparison cannot match either by identifier alone and remains blocking for operator investigation

#### Scenario: Two reused canonical orders remain active across sessions
- **WHEN** the same account and market have two non-terminal engine-owned orders on different trading days that share one opaque identifier
- **THEN** local reconstruction retains both orders and reconciliation cannot discard the prior-session order merely because a newer scoped row exists

### Requirement: Reconcile release is durable before it is visible

The tracker SHALL preserve an existing block and its entry-gate projection until the corresponding journal release commits. If persistence fails, the current reconciliation cycle MUST stop before automatic adoption. Permanent release SHALL require explicit operator identity, explanatory evidence, engine exclusion, and a fresh stable authoritative comparison with no blocking diff.

#### Scenario: Durable release fails
- **WHEN** a matching post-adjustment comparison proposes a release but the journal write fails
- **THEN** the tracker and gate remain blocked and the cycle performs no adoption or price read

#### Scenario: Operator resolves a proven-clean account
- **WHEN** the engine is stopped, three official snapshots taken at least two seconds apart agree, corrected local state has no blocking diff, and the operator explicitly confirms a release with identity and note
- **THEN** supported active states are released atomically with `OPERATOR` evidence before the tracker/gate are cleared

#### Scenario: Operator release sees a blocking diff
- **WHEN** the fresh corrected comparison still contains a quantity mismatch or locally owned missing order
- **THEN** the release is refused and every active journal state remains unchanged

#### Scenario: Prior-session lineage could match the fresh broker snapshot
- **WHEN** a prior account session has a replacement child whose identifier appears in the fresh selected-account snapshot
- **THEN** the corrected local state keeps the current-session parent distinct and the recovery command refuses release if that current order is missing

### Requirement: Startup reservation recovery preserves canonical ownership

Before the engine can take a decision or prune spent-nonce evidence, startup SHALL run reservation recovery. The sweep SHALL release a held reservation from a terminal fill snapshot only when the reservation's decision-bound confirmed attempt and intent exactly match the snapshot account, market, market-local trading day, symbol, and side, and that canonical scope has exactly one intent owner. A cross-scope or ambiguous snapshot MUST leave risk headroom held; ambiguity MUST record an account-wide identifier conflict. Spent nonce evidence referenced by a held reservation MUST remain retained regardless of the ordinary age cutoff.

#### Scenario: Reused order identifier has a terminal snapshot in another scope
- **WHEN** a held reservation's order identifier has a terminal snapshot from another account, market, trading day, symbol, or side
- **THEN** startup recovery releases nothing and the reservation remains held

#### Scenario: Old spent nonce still protects a held reservation
- **WHEN** a spent nonce is older than the retention cutoff but its decision still owns a held reservation
- **THEN** retention preserves the nonce and a later startup cannot release the hold as expired-unconsumed
