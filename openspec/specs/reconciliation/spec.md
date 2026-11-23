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
missing local order가 permanent evidence를 얻으려면 non-empty canonical `Diff.AccountRef`, order account, tracker account가 모두 같아야 한다(SHALL). blank/foreign-account comparison은 complete order identity를 포함해도 ordinary block만 유지하고 permanent evidence를 얻어서는 안 된다(SHALL NOT).
허용 오차를 넘는 불일치가 확인되면 신규 진입은 차단되고(SHALL) 청산 경로는 유지된다(SHALL — 확정 하한 규칙). 재대사는 최소 간격(기본 30초)을 두고 수행한다. 영구 승격의 연속 실패는 계좌의 모든 blocking comparison이 공유하는 횟수가 아니라 **동일 canonical blocking dispute**가 즉시 앞선 blocking comparison에도 존재한 횟수여야 한다(SHALL). 수량 불일치의 identity는 tracker와 같은 non-empty canonical `Diff.AccountRef`·non-empty 정규화 심볼·snapshot/comparer에서 float64를 거치지 않고 보존된 exact finite-decimal local quantity·broker quantity이고, missing local order의 identity는 tracker와 같은 canonical 계좌·시장·시장-local 거래일·심볼·side·opaque order identifier의 여섯 canonical component가 모두 non-empty인 tuple이다(SHALL). opaque order identifier는 whitespace-only invalidity만 검사하고 byte-preserving해야 하므로 `"id"`와 `" id "`는 다른 identity다(SHALL). 같은 identity가 연속 3회 관측될 때만 영구 불일치로 표기하고 운영자 확인 절차를 요구한다(SHALL). 다른 dispute로 바뀌거나 exact quantity tuple 또는 canonical order identity가 바뀌거나 한 comparison에서 사라지면 그 dispute의 streak는 다시 1부터 시작한다(SHALL). 비교 하나에 같은 identity가 중복되어도 한 번만 센다(SHALL). 대사 성공 시 모든 transient streak는 리셋된다. blank/foreign diff account, blank symbol, blank·malformed·non-finite quantity, tracker와 다른 account 또는 required component가 빈 missing-order처럼 identity의 exact canonicalization을 증명할 수 없으면 그 관측은 permanent streak evidence로 쓰지 않되(SHALL NOT), ordinary block은 그대로 즉시 유지한다(SHALL). 차단 범위(계좌/시장/심볼)와 자동·수동 해제 조건은 reason-code와 함께 상태표로 정의한다(SHALL).

영구 승격의 durable write가 실패한 경우 반환 직후 pending account-wide block과 account gate는 fail-closed로 유지되어야 하고(SHALL), 해당 승격은 그 승격을 얻은 canonical dispute가 바로 다음 blocking comparison에도 존재할 때만 재시도한다(SHALL). 다음 관측이 clean이거나 해당 dispute가 사라졌으면 journal authority read를 먼저 수행해야 한다(SHALL). 같은 계좌의 durable account-wide permanent row가 있으면 durable projection과 gate를 유지하고(SHALL), authority read가 실패하면 fail-closed 상태로 error를 반환해야 한다(SHALL). authority가 durable row 부재를 확인한 경우에만 pending account-wide 승격과 retry identity를 메모리·gate에서 철회한다(SHALL). 이 규칙은 ordinary symbol block의 fail-closed durable retry 또는 이미 durable한 permanent block을 철회해서는 안 된다(SHALL NOT).
continuity를 끊은 현재 blocking comparison에 새 ordinary item이 있으면 authority read 전에 그 memory block과 gate를 latch해야 한다(SHALL). authority error 뒤 성공한 Refresh가 stale account proposal을 철회해도 그 ordinary pending block을 잃어서는 안 된다(SHALL NOT).
그 현재 blocking comparison이 adjustment credit보다 later이고 credited symbol을 계속 disputed로 보고하면 authority error 반환 전에 credit을 refuted로 만료해야 한다(SHALL). 이후 Refresh와 clean read가 stale credit으로 ordinary block을 release해서는 안 된다(SHALL NOT).

ordinary block과 account-wide permanent가 동시에 pending이면 permanent retry를 deterministic하게 먼저 journal에 시도해야 한다(SHALL). ordinary write의 반복 실패나 map iteration order가 continuity가 증명된 permanent retry를 선점해서는 안 된다(SHALL NOT).

representable ordinary sibling은 unrepresentable blank-symbol pending block보다 먼저 deterministic하게 영속해야 한다(SHALL). blank row 거절이 valid sibling의 durable journal authority를 굶겨서는 안 된다(SHALL NOT).

서로 다른 canonical quantity 문자열이 같은 float64로 충돌하면 수량 일치로 판정해서는 안 된다(SHALL NOT). `0.3`과 `0.30000000000000004`처럼 한쪽의 짧은 canonical decimal spelling과 그 spelling의 binary-expanded round-trip representation임을 증명할 수 있고, 서로 다른 float64로 표현되며 larger magnitude에서 한 float64 ULP 이내인 artifact만 예외로 일치할 수 있다(MAY). 단지 한 ULP 이내라는 이유로 서로 다른 exact integer/decimal을 일치 처리해서는 안 된다(SHALL NOT). 상대값 `1e-9 * scale`처럼 한 ULP보다 넓은 실제 decimal 차이를 tolerance로 숨겨서는 안 된다(SHALL NOT).
largest finite float에서 next representable value가 infinity라서 ULP 계산 결과가 finite하지 않으면 tolerance를 적용해서는 안 되며(SHALL NOT), exact canonical equality만 일치로 인정한다(SHALL).

blank·malformed·non-finite quantity는 양쪽 문자열이 동일해도 valid equality로 판정해서는 안 되며(SHALL NOT), comparer는 ordinary fail-closed mismatch를 생성하되 permanent streak evidence는 부여하지 않아야 한다(SHALL NOT).

raw broker/local quantity의 blank·malformed·non-finite 여부는 zero canonicalization 또는 external-position 분류 전에 검사해야 한다(SHALL). unreadable 또는 negative broker-only exposure는 long-only projection에서 ordinary fail-closed mismatch이며(SHALL), valid positive broker-only holding의 external provenance 비차단 계약은 유지한다(SHALL).

production `RawPositionsReader`에서 수집한 holding quantity는 blank/invalid 원문을 `Comparer`까지 보존해야 하며(SHALL), Collector가 이를 zero로 재작성해 validation을 우회해서는 안 된다(SHALL NOT).
holdings quantity의 snapshot digest는 같은 dedicated evidence vocabulary를 사용해 blank와 exact zero를 서로 다른 읽기로 취급해야 한다(SHALL). 다른 optional decimal field의 blank vocabulary는 변경하지 않는다(SHALL NOT).

normalized symbol이 빈 ordinary quantity mismatch는 `EntryGate`와 `Tracker.EntryAllowed` 모두에서 모든 candidate를 account-safe하게 차단하되(SHALL), real account-permanent block과 in-memory identity를 충돌시키거나 empty-symbol `QUANTITY_MISMATCH` journal row로 영속해서는 안 된다(SHALL NOT). 그 row는 재시작 시 기존 account-wide permanent 모양과 구분할 수 없으므로, write는 error를 반환하고 pending ordinary gate를 유지해야 하며(SHALL), Restore가 이 invalid ordinary observation을 operator-only permanent로 재구성해서는 안 된다(SHALL NOT).

known-nondurable blank-symbol pending block은 audited operator release에서 존재하지 않는 durable row의 해제를 요구해서는 안 되며(SHALL NOT), operator identity/note와 실제 durable rows의 원자적 release가 확인된 뒤 memory/gate에서 제거할 수 있다(SHALL).

해제 규칙의 정밀화(SHALL): 비영구 차단의 자동 해제는 **조정 이벤트가 반영된 뒤의 재조회 일치**, 또는 「missing-order 종목 차단은 그것을 만든 주문들의 종결 증명과 그 뒤의 일치로 풀린다」 Requirement 가 정한 **결속된 주문 전부의 종결 기록 뒤의 재조회 일치**에만 근거하며, 각각 구분되는 release cause(ADJUSTMENT_APPLIED 계열, ORDER_TERMINAL_CONFIRMED)와 원인 기록을 남긴다. 조정도 종결 증명도 없이 우연히 일치한 단발 관측은 차단을 해제하지 못하고(SHALL NOT), 영구 불일치의 해제는 운영자 확인뿐이다(SHALL).

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

#### Scenario: 같은 수량 분쟁의 영구 승격
- **WHEN** 같은 계좌·심볼·canonical local/broker quantity tuple이 세 번의 연속 blocking comparison에 존재하면
- **THEN** 기존 account-wide `reconciliation_mismatch_permanent`가 durable하게 기록되고 운영자 확인 전까지 자동 해제되지 않는다

#### Scenario: 서로 다른 수량 분쟁은 streak를 공유하지 않는다
- **WHEN** 세 번의 연속 blocking comparison이 서로 다른 심볼 또는 서로 다른 exact local/broker quantity tuple을 담으면
- **THEN** ordinary symbol block은 각각 즉시 유지되지만 세 관측을 합친 account-wide permanent block은 만들어지지 않는다

#### Scenario: canonical missing-order identity가 다르다
- **WHEN** opaque order identifier가 같아도 시장·거래일·심볼·side 중 하나가 다른 missing local order가 연속 관측되면
- **THEN** 각 ordinary block은 유지되지만 서로의 permanent streak를 이어받지 않는다

#### Scenario: opaque order identifier는 byte-preserving이다
- **WHEN** 다른 scope field는 같지만 opaque order identifier의 원본 bytes가 `"id"`와 `" id "`처럼 다르면
- **THEN** ordinary block은 유지되지만 두 order는 서로의 permanent streak를 이어받지 않는다

#### Scenario: 다른 계좌의 missing-order는 tracker 승격 증거가 아니다
- **WHEN** missing local order의 canonical account가 tracker account와 다르면
- **THEN** fail-closed ordinary block은 유지되지만 tracker account의 permanent streak를 얻지 않는다

#### Scenario: blank 또는 다른 계좌 comparison은 missing-order 승격 증거가 아니다
- **WHEN** complete missing local order의 account는 tracker와 같지만 `Diff.AccountRef`가 blank이거나 다른 계좌이면
- **THEN** fail-closed ordinary block은 유지되지만 permanent streak와 durable account-wide row는 생성하지 않는다

#### Scenario: incomplete missing-order identity는 승격 증거가 아니다
- **WHEN** missing local order의 account·market·market-local trading day·symbol·side·opaque order identifier 중 하나라도 canonicalization 뒤 비어 있으면
- **THEN** ordinary block은 유지되지만 해당 item은 permanent streak 횟수를 얻지 않으며 같은 불완전 tuple이 반복되어도 permanent로 승격되지 않는다

#### Scenario: 증명할 수 없는 streak identity
- **WHEN** blocking item의 quantity가 blank·malformed·non-finite이거나 canonical order identity를 permanent-streak key로 정규화할 수 없으면
- **THEN** 신규 진입 차단은 유지되고 해당 item은 그 관측에서 영구 승격 횟수를 얻지 않는다

#### Scenario: 비어 있는 수량 분쟁 scope는 승격 증거가 아니다
- **WHEN** quantity mismatch의 `Diff.AccountRef`가 비었거나 tracker account와 다르거나 normalized symbol이 비어 있으면
- **THEN** ordinary fail-closed block은 유지되지만 해당 item은 permanent streak 횟수를 얻지 않는다

#### Scenario: 서로 다른 큰 exact decimal은 충돌하지 않는다
- **WHEN** snapshot/comparer를 통과한 float64 충돌 가능 큰 decimal quantity tuple이 연속 관측되면
- **THEN** ordinary 차단은 유지되지만 서로의 permanent streak를 이어받지 않는다

#### Scenario: float64가 같은 서로 다른 exact quantity는 불일치다
- **WHEN** broker와 local의 서로 다른 canonical decimal quantity가 같은 float64 값으로 변환되면
- **THEN** comparer는 이를 match로 숨기지 않고 ordinary quantity mismatch를 생성한다

#### Scenario: relative epsilon은 실제 decimal 차이를 숨기지 않는다
- **WHEN** 두 exact quantity의 차이가 `1e-9 * scale` 이내지만 larger magnitude의 한 float64 ULP보다 크면
- **THEN** comparer는 ordinary quantity mismatch를 생성하고 tolerance-zero 계약을 유지한다

#### Scenario: largest finite float의 ULP는 무한 tolerance가 아니다
- **WHEN** 한 quantity가 largest finite float이고 다른 exact quantity가 `1`이면
- **THEN** comparer는 infinite ULP tolerance로 match하지 않고 ordinary quantity mismatch를 생성한다

#### Scenario: exact integer 차이는 한 ULP라도 round-trip artifact가 아니다
- **WHEN** quantity가 `9007199254740992`와 `9007199254740994`처럼 서로 다른 exact integer이지만 parsed float 간격은 한 ULP이면
- **THEN** comparer는 짧은 decimal의 binary-expanded spelling 증거가 없으므로 ordinary quantity mismatch를 생성한다

#### Scenario: 동일한 invalid quantity도 ordinary 불일치다
- **WHEN** broker와 local quantity가 같은 `NaN`, infinity 또는 malformed spelling이면
- **THEN** comparer는 이를 유효한 match로 숨기지 않고 ordinary fail-closed mismatch를 생성하며 permanent streak evidence는 부여하지 않는다

#### Scenario: unreadable broker-only quantity는 external exposure가 아니다
- **WHEN** 로컬 포지션이 없고 broker holding quantity가 blank·malformed·non-finite이면
- **THEN** comparer는 ordinary fail-closed mismatch를 생성하고 permanent streak evidence는 부여하지 않는다

#### Scenario: negative broker-only quantity는 external exposure가 아니다
- **WHEN** long-only 로컬 포지션이 없고 broker holding quantity가 finite negative이면
- **THEN** comparer는 ordinary fail-closed mismatch를 생성하고 permanent streak evidence는 부여하지 않는다
- **AND** valid positive broker-only holding만 nonblocking external provenance로 남는다

#### Scenario: Collector가 unreadable raw holding을 보존한다
- **WHEN** production raw holdings read가 blank·malformed·non-finite quantity를 반환하면
- **THEN** Collector는 그 evidence를 zero로 바꾸지 않고 Comparer까지 전달하며 ordinary fail-closed/no-promotion 계약이 동일하게 적용된다

#### Scenario: blank holding과 exact zero는 서로 안정화 증거가 아니다
- **WHEN** 연속 raw holdings read 중 하나의 quantity가 blank이고 다른 하나가 exact zero이면
- **THEN** snapshot digest는 두 읽기를 다르게 유지하고 stabiliser는 순서와 무관하게 stable로 판정하지 않는다
- **AND** 두 blank 읽기 또는 두 exact-zero 읽기는 각각 서로 corroborate할 수 있다

#### Scenario: 빈 심볼 ordinary row는 permanent 모양으로 영속하지 않는다
- **WHEN** normalized symbol이 빈 quantity mismatch를 journal-backed tracker가 관측하면
- **THEN** `EntryGate`와 `Tracker.EntryAllowed`가 모든 candidate를 ordinary reason으로 차단하고 pending block은 유지되지만 empty-symbol quantity row는 쓰지 않고 error를 반환하며, 재시작 시 operator-only permanent로 복원되지 않는다

#### Scenario: blank-symbol pending은 valid sibling durability를 굶기지 않는다
- **WHEN** 한 comparison에 blank-symbol ordinary block과 representable symbol block이 함께 있으면
- **THEN** representable sibling은 먼저 durable하게 기록되고 blank block은 memory/gate에 pending으로 남아 error를 반환한다

#### Scenario: blank-symbol pending은 earned sibling release를 굶기지 않는다
- **WHEN** blank-symbol pending과 durable valid-symbol ordinary block이 함께 남아 있고, valid symbol이 adjustment 후 later clean comparison으로 release를 얻으면
- **THEN** valid symbol release는 blank pending error보다 먼저 durable하게 기록되어 memory/gate와 restart projection에서 사라진다
- **AND** blank-symbol block만 memory/account-safe gate에 pending으로 남고 error를 반환한다

#### Scenario: 실패한 permanent write 뒤 continuity가 끊긴다
- **WHEN** 같은 dispute의 threshold 관측에서 account-wide permanent write가 실패하고 다음 관측이 clean이거나 그 dispute가 사라지면
- **THEN** 아직 durable하지 않은 pending account-wide 승격은 철회되며 이후 stale write로 재시도되지 않는다

#### Scenario: permanent write 실패 직후 gate는 닫혀 있다
- **WHEN** threshold의 account-wide permanent journal write가 실패하여 `Observe`가 error를 반환하면
- **THEN** 다음 authoritative comparison이 continuity를 판정할 때까지 pending permanent account gate는 닫힌 채 유지된다

#### Scenario: permanent write 응답만 유실됐다
- **WHEN** account permanent journal write가 durable하게 커밋됐지만 timeout/error를 반환하고 다음 관측에서 earning dispute continuity가 끊기면
- **THEN** authority read가 durable row를 복원하며 account gate는 operator release 전까지 닫힌다

#### Scenario: pending 철회 authority read가 실패한다
- **WHEN** failed permanent write 뒤 continuity가 끊겼지만 active journal authority를 읽을 수 없으면
- **THEN** pending account gate를 철회하지 않고 error를 반환한다
- **AND** 이미 관측된 continuity break는 기록되어 이후 earning key가 돌아와도 stale streak나 pending permanent write를 재사용하지 않는다

#### Scenario: authority failure 전에 현재 ordinary mismatch를 latch한다
- **WHEN** pending permanent continuity를 끊은 blocking comparison에 새 symbol mismatch가 있고 authority read가 실패하면
- **THEN** 새 ordinary memory block과 symbol gate는 error 반환 전에 fail-closed로 latch된다
- **AND** 이후 성공한 `Refresh`가 stale account proposal을 철회해도 그 ordinary pending block과 symbol gate는 유지된다

#### Scenario: authority failure도 refuted adjustment credit을 보존하지 않는다
- **WHEN** adjustment credit보다 later인 blocking comparison이 credited symbol을 계속 disputed로 보고 pending-permanent authority read가 실패하면
- **THEN** error 반환 전에 그 credit은 refuted로 만료되고 ordinary block/gate는 유지된다
- **AND** 이후 성공한 `Refresh`와 clean read만으로 stale credit을 재사용해 block을 release하지 않는다

#### Scenario: Refresh는 continuity-broken pending proposal을 permanent로 만들지 않는다
- **WHEN** continuity break의 authority read가 실패해 non-durable pending account gate가 남은 뒤, 이후 `Refresh`가 durable permanent row 부재를 확인하면
- **THEN** `Refresh`는 pending proposal을 철회하고 `Failures`를 threshold로 올리거나 permanent state를 생성하지 않는다
- **AND** ordinary pending block과 이미 기록된 continuity break는 보존한다

#### Scenario: known-nondurable blank block을 운영자가 해제한다
- **WHEN** blank-symbol ordinary block이 ambiguous durable write를 거절해 memory에만 pending이고 운영자가 identity와 note로 Resolve하면
- **THEN** 존재하지 않는 journal row release를 요구하지 않고 실제 durable rows를 원자적으로 해제한 뒤 memory/gate를 함께 정리한다

#### Scenario: 실패한 permanent write 뒤 같은 dispute가 계속된다
- **WHEN** 같은 dispute의 threshold 관측에서 account-wide permanent write가 실패하고 바로 다음 blocking comparison에도 같은 dispute가 존재하면
- **THEN** ordinary gate를 계속 닫은 채 그 account-wide permanent write를 재시도한다

#### Scenario: ordinary와 permanent가 함께 pending이다
- **WHEN** 같은 earning dispute가 계속되지만 ordinary enter와 account permanent enter가 모두 아직 durable하지 않으면
- **THEN** 다음 persistence pass는 account permanent retry를 먼저 시도하고 ordinary retry 실패가 이를 굶기지 않는다

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

### Requirement: 조정 credit은 그것을 만든 비교가 아니라 그 다음 재조회가 쓴다

조정 credit은 그 조정이 계산된 **비교의 as-of를 함께 기록해야 한다**(SHALL).

같은 비교를 관측하는 호출은 credit을 사용해서도 소멸시켜서도 안 된다(SHALL NOT).
그 관측이 보고 있는 것은 조정 **이전**에 수집된 세계이므로 "조정 이벤트가 반영된 뒤의
재조회"가 아니다. 조정을 발행한 비교와 그 조정을 관측하는 비교가 같은 것일 때 credit이
소멸하면, 승인된 자동 해제는 어떤 입력에서도 도달할 수 없다.

credit의 소멸과 사용은 **심볼 단위로** 판정해야 한다(SHALL). 어떤 심볼의 credit은 그것이
기록된 비교보다 엄격히 나중의 as-of를 가진 관측이 **그 심볼에 대해 여전히 불일치**할 때만
소멸한다(SHALL). 같은 관측에서 다른 심볼이 불일치한다는 이유로 소멸시켜서는 안 된다
(SHALL NOT) — 그 심볼의 조정은 그 심볼의 재조회가 답하며, 무관한 심볼의 불일치는 그
질문에 대한 답이 아니다.

두 비교의 선후를 판정할 수 없으면 credit을 사용해서는 안 된다(SHALL NOT) — as-of가
없거나, 시각으로 읽을 수 없거나, 관측의 비교가 credit의 비교보다 앞선 경우가 모두
여기에 해당하며, 차단 유지가 보수 방향이다.

이 요구사항은 해제의 증거 요건을 약화해서는 안 된다(SHALL NOT). 조정이 기록되지 않은
심볼은 재조회가 아무리 여러 번 일치해도 조정 경로로는 자동 해제되지 않으며, 그 밖의 자동 해제는
「missing-order 종목 차단은 그것을 만든 주문들의 종결 증명과 그 뒤의 일치로 풀린다」 Requirement 의
조건을 전부 갖춘 경우뿐이다(SHALL). 영구 불일치의 해제는 운영자 확인뿐이라는 기존 규칙도 그대로다.

#### Scenario: 수렴과 같은 주기의 관측
- **WHEN** 한 대사 주기가 수량 불일치를 계좌값으로 수렴시키고, 같은 주기에서 그 수렴을 만든 것과 **같은** 비교를 관측한다
- **THEN** 차단은 유지되고, 해제는 일어나지 않으며, 그 심볼의 credit은 다음 재조회를 위해 보존된다

#### Scenario: 조정 이후의 재조회가 일치한다
- **WHEN** 조정이 기록된 뒤 더 나중에 수집된 비교가 그 심볼에 대해 일치한다
- **THEN** 비영구 차단이 ADJUSTMENT_APPLIED 원인 기록과 함께 자동 해제되고, 그 심볼의 진입 차단이 풀린다

#### Scenario: 조정 이후의 재조회가 여전히 불일치한다
- **WHEN** 조정이 기록된 뒤 더 나중에 수집된 비교가 그 심볼에 대해 여전히 불일치한다
- **THEN** credit은 소멸하고 차단은 유지되며, 해제는 새 조정과 그 뒤의 일치를 다시 요구한다

#### Scenario: 무관한 심볼이 불일치한다
- **WHEN** 조정이 기록된 심볼은 이후 재조회에서 일치하는데 같은 비교의 다른 심볼이 불일치한다
- **THEN** 일치한 심볼의 credit은 소멸하지 않고 보존되며, 비교 전체가 일치하는 첫 관측에서 그 심볼의 차단이 해제된다

#### Scenario: 비교의 선후를 판정할 수 없다
- **WHEN** 관측하는 비교에 as-of가 없거나 시각으로 읽을 수 없다
- **THEN** 일치하더라도 credit을 사용하지 않고 차단을 유지한다

#### Scenario: 조정 없는 일치의 반복
- **WHEN** 조정이 한 번도 기록되지 않았고 missing-order 종결 증명 해제의 조건도 갖추지 않은 심볼의 재조회가 여러 주기 연속으로 일치한다
- **THEN** 어떤 주기에서도 자동 해제되지 않고 운영자 해제만 남는다

### Requirement: 재시작 복구는 rate limit을 영구 실패와 구분한다

재시작 복구의 계좌 스냅샷 수집이 브로커의 rate limit으로 거부되면 엔진은 그 거부를 안정화 판정의 입력으로 삼지 않고 유한한 백오프 후 재시도해야 한다(SHALL). 읽기가 일어나지 않았으므로 그 거부는 계좌가 불안정하다는 증거가 아니다. 백오프는 같은 계좌 예산을 쓰는 read-only 서베이의 재시도 간격보다 짧아서는 안 된다(SHALL NOT).

재시도의 총 대기는 유한해야 하며(SHALL), 상한 소진 시 복구는 오늘과 같이 미완료로 실패하고 그 사유에 rate limit과 기다린 시간을 명시해야 한다(SHALL). 대기 동안과 소진 후 모두 신규 주문 게이트는 닫혀 있어야 한다(SHALL) — 이 요구는 복구를 더 끈질기게 만들 뿐, 부분 데이터로 거래를 시작하는 것을 허용하지 않는다.

rate limit이 아닌 수집 오류의 처리와 안정화 판정 자체는 이 요구의 대상이 아니며 오늘과 같아야 한다(SHALL).

#### Scenario: 429 두 번 뒤 안정 스냅샷

- **WHEN** 복구의 스냅샷 수집이 rate limit으로 두 번 거부된 뒤 안정된 읽기가 이어지면
- **THEN** 복구는 완주하고, 두 거부는 안정화 시도 횟수에 계상되지 않으며, 각 대기는 서베이의 백오프 간격 이상이다

#### Scenario: rate limit 대기 상한 소진

- **WHEN** rate limit 재시도의 누적 대기가 상한을 넘으면
- **THEN** 복구는 미완료로 실패하고 사유에 rate limit과 총 대기 시간이 있으며 진입 게이트는 닫힌 채 남는다

#### Scenario: rate limit이 아닌 오류는 오늘과 같다

- **WHEN** 스냅샷 수집이 rate limit이 아닌 오류로 실패하면
- **THEN** 복구는 재시도 없이 즉시 미완료로 실패한다 — 기존 동작 그대로다

#### Scenario: 백오프 중 종료 신호

- **WHEN** rate limit 백오프 대기 중에 프로세스 종료가 요청되면
- **THEN** 대기는 즉시 중단된다 — 손절·비상 청산의 즉시성을 약화하지 않는다

### Requirement: missing-order 종목 차단은 그것을 만든 주문들의 종결 증명과 그 뒤의 일치로 풀린다

엔진은 **실행 중인 이 프로세스의 missing-order 관측이 직접 삽입한** 비영구 종목 차단 행만, 그 행이 살아 있는 동안 그 종목에서 관측된 missing-order 주문
**전부**가 장부에 깨끗한 종결(`terminal` 이고 `fail_closed` 아님)로 기록되고, 그 뒤 진입을 막지 않는 비교가 그 종목에 차이를 보지 않고 다시 그보다
나중의 비교도 같은 증명으로 통과할 때 스스로 해제해야 한다(SHALL). 그 행이 살아 있는 동안 같은 종목의 수량 불일치·외부 포지션이 관측됐거나, 행을
삽입하기 **전에** 읽은 체결 워터마크 뒤에 그 종목의 체결(계좌 무관) 또는 종목을 알 수 없는 체결이 커밋됐으면 이 경로로 해제해서는 안 된다(SHALL NOT).

#### Scenario: 종결 기록 뒤의 두 비교가 missing-order 차단을 푼다
- **WHEN** 이 프로세스가 한 주문의 missing-order 로 종목 행을 삽입했고, 그 주문이 장부에 깨끗한 종결로 기록된 뒤, 그 기록보다 나중 as-of 의 비교가 진입을 막지 않고 그 종목에 차이를 보지 않으며, 그보다 더 나중의 비교도 같은 증명으로 통과한다
- **THEN** 다른 유지 조건(수량·외부 포지션 흔적, 워터마크 뒤 체결, 신원·계좌·lineage, 같은 번호의 외부 주문)이 없을 때 첫 비교는 차단을 유지하고, 둘째 비교에서 차단이 해제되어 그 종목의 진입 게이트가 열린다

#### Scenario: 종결 기록보다 이른 비교는 풀지 않는다
- **WHEN** 결속된 주문의 종결이 기록됐지만 관측하는 비교의 as-of 가 그 기록 시각과 같거나 이르다
- **THEN** 차단은 유지된다

#### Scenario: 결속된 주문 하나라도 깨끗한 종결이 아니면 풀지 않는다
- **WHEN** 결속된 주문 중 하나가 비종결이거나, 종결이지만 `fail_closed` 이거나, 그 신원의 스냅숏이 없다
- **THEN** 차단은 유지된다

#### Scenario: 수량 분쟁이 겹친 차단은 조정 규칙만 푼다
- **WHEN** missing-order 차단이 활성인 동안 같은 종목의 수량 불일치가 한 번이라도 관측되었다
- **THEN** 이 경로로는 해제되지 않고 조정 이벤트 뒤의 재조회 일치만 그 차단을 풀 수 있다

#### Scenario: 행을 삽입한 뒤 그 종목에 체결이 커밋됐다
- **WHEN** 이 프로세스가 행을 삽입하기 전에 읽은 체결 워터마크보다 나중에 그 종목의 체결 이벤트(계좌 무관) 또는 종목을 알 수 없는 체결 이벤트가 하나라도 커밋되었다
- **THEN** 이 경로로는 해제되지 않는다

#### Scenario: 영구 차단은 이 경로로 풀리지 않는다
- **WHEN** missing-order 분쟁이 같은 identity 로 연속 3회 관측되어 계좌 전체 영구 차단이 생긴 뒤 그 주문의 종결 증명과 일치하는 비교들이 관측된다
- **THEN** 종목 차단은 이 경로로 풀릴 수 있지만 계좌 전체 영구 차단은 자동 해제되지 않고 운영자 확인을 요구한다

### Requirement: 종결 증명 해제의 결속은 이 프로세스의 삽입 증명에서만 생기고 모호하면 차단을 유지한다

결속은 프로세스 기억이며 해제를 늦추거나 포기하게만 해야 한다(SHALL). 결속은 그 missing-order 관측의 차단 행 쓰기가 행을 **실제로 삽입했다**는 장부의
반환(삽입 여부와 행 id)이 있을 때만 생겨야 하며(SHALL), 이 프로세스가 직접 삽입하지 않은 행 — 재기동 전 행, 복원·갱신으로 들어온 행, 다른 생산자가 만든
행, 이미 있던 행에 겹친 missing-order — 은 이 경로로 해제해서는 안 된다(SHALL NOT). 신원이 비었거나, 결속·행·추적기·비교의 계좌가 모두 같지 않거나,
lineage 가 옮겨졌거나, 판정 읽기가 실패하거나, as-of 를 읽을 수 없거나, 결속된 번호와 바이트가 같은 주문이 open 목록에 있으면 유지해야 한다(SHALL).

#### Scenario: 재기동 뒤 남은 missing-order 행은 운영자 해제로 돌아간다
- **WHEN** missing-order 차단 행이 남은 채 엔진이 재기동하고, 이후 그 주문들의 종결 증명과 일치하는 비교들이 관측된다
- **THEN** 이 경로로는 해제되지 않고 운영자 해제만 남는다

#### Scenario: 이미 있던 행에 겹친 missing-order 는 결속을 만들지 않는다
- **WHEN** 다른 생산자가 만든 활성 종목 행이 있는 종목에서 missing-order 가 관측되어 차단 행 쓰기가 아무것도 삽입하지 않는다
- **THEN** 결속은 생기지 않고 그 행은 이 경로로 해제되지 않는다

#### Scenario: 삽입 증명이 없는 블록은 풀지 않는다
- **WHEN** 차단 행 쓰기가 오류를 돌려주어 삽입 여부를 알 수 없거나, 첫 쓰기가 실패한 블록이 나중 재시도에서 삽입된다
- **THEN** 결속은 생기지 않고 이 경로로 해제되지 않는다

#### Scenario: 같은 추적기의 복원은 결속을 버린다
- **WHEN** 결속이 있는 추적기가 장부에서 블록을 다시 복원한다
- **THEN** 결속과 무장은 사라지고 그 행은 이 경로로 해제되지 않는다

#### Scenario: 브로커가 종결된 주문을 열린 주문으로 보여 준다
- **WHEN** 결속된 주문번호와 바이트가 같은 주문이 그 비교의 외부 주문으로 계좌 open 목록에 있다
- **THEN** 차단은 유지된다

#### Scenario: 정정 lineage 로 옮겨진 주문은 풀지 않는다
- **WHEN** 결속된 missing-order 의 현재 주문번호가 원래 주문번호와 다르거나, 그 번호를 겨눈 정정 시도나 대체 기록이 장부에 하나라도 있다(확정 여부 무관)
- **THEN** 차단은 유지된다

#### Scenario: 판정 읽기가 실패한다
- **WHEN** 종결 스냅숏이나 lineage 를 읽는 장부 조회가 실패한다
- **THEN** 차단은 유지되고 그 주기는 실패하지 않는다

### Requirement: 종결 증명 해제는 같은 트랜잭션의 재판정 뒤 구분되는 원인으로 기록되고 다른 규칙을 바꾸지 않는다

종결 증명 해제는 행 id 를 고정하고 `released_at` 이 비어 있을 때만, 판정 근거(행·결속 주문의 종결 스냅숏·lineage·삽입 뒤 체결)를 다시 읽는 **같은 장부
트랜잭션 안에서** 같은 판정을 다시 통과해야 커밋되어야 한다(SHALL). release cause `ORDER_TERMINAL_CONFIRMED` 와 근거를 장부에 기록한 **뒤에만** 진입
게이트에서 사라져야 하며(SHALL), 계좌번호·주문번호·수량·가격 없는 로그를 남겨야 한다(SHALL). 해제하지 않은 결속의 유지 사유는 바뀔 때마다
종목과 사유만 담은(계좌번호·주문번호·수량·가격 없는) 로그로 남겨야 한다(SHALL). 이 경로는 수량 불일치 해제 규칙, 차단 조건과 시점,
영구 승격, 청산·손절 경로를 바꿔서는 안 된다(SHALL NOT).

#### Scenario: 판정과 해제 사이에 증명이 바뀐다
- **WHEN** 판정 읽기 뒤 해제 커밋 전에 결속 주문의 스냅숏이 `fail_closed` 로 바뀌거나 그 종목의 체결이 커밋된다
- **THEN** 같은 트랜잭션의 재판정이 실패해 해제되지 않는다

#### Scenario: 해제 기록이 실패한다
- **WHEN** 종결 증명 해제의 장부 기록이 실패한다
- **THEN** 차단과 진입 게이트는 그대로 남고 그 주기는 편입 판정으로 가지 않는다

#### Scenario: 해제 근거가 구분된다
- **WHEN** 종결 증명으로 차단이 해제된다
- **THEN** 장부 행의 release cause 는 `ORDER_TERMINAL_CONFIRMED` 이고 로그 이벤트에는 계좌번호·주문번호·수량·가격이 없다

### Requirement: 귀속 실패 종목 차단은 재조회 비교를 근거로 운영자가 해제할 수 있다

운영자 해제 명령은 종목 범위의 귀속 실패(`ATTRIBUTION_FAILED`) 차단을 수량 불일치 차단과 같은 조건으로 해제할 수 있어야 한다(SHALL). 그 조건은 엔진 잠금 아래 명령 자신의 안정 재조회 비교가 진입을 막는 차이(수량 불일치, 원장에만 있는 주문)를 보지 않는 것과 운영자 신원·확인 내용이다. 그 차이가 남아 있으면 해제해서는 안 된다(SHALL NOT). 해제는 운영자 해제 원인과 근거로 장부에 남아야 한다(SHALL). 이 요구는 귀속 실패 차단을 자동으로 풀지 않고, 차단을 거는 조건과 수량 불일치 차단의 해제 규칙을 바꾸지 않는다(SHALL NOT).

#### Scenario: 귀속 실패 종목 차단을 운영자가 해제한다
- **WHEN** 한 종목에 귀속 실패 차단만 활성이고, 엔진이 정지된 상태에서 운영자가 신원과 확인 내용을 주어 해제 명령을 실행하며 명령의 재조회 비교가 진입을 막는 차이를 보지 않으면
- **THEN** 그 차단은 운영자 해제 원인과 그 근거로 장부에서 해제되고, 다음 기동의 대사는 그 종목을 차단으로 복원하지 않는다

#### Scenario: 진입을 막는 차이가 남아 있다
- **WHEN** 귀속 실패 차단이 활성인데 명령의 재조회 비교가 진입을 막는 차이를 보면
- **THEN** 명령은 거절하고 차단은 바뀌지 않는다

#### Scenario: 해제 대상 종목에 외부 보유가 있다
- **WHEN** 귀속 실패 차단이 활성인 종목에 장부가 설명하지 못하는 보유가 있고 재조회 비교가 진입을 막는 차이를 보지 않으면
- **THEN** 차단은 해제되고 그 보유는 이후 대사의 통상 경로로 다뤄진다

#### Scenario: 수량 불일치와 귀속 실패가 함께 있다
- **WHEN** 수량 불일치 차단과 귀속 실패 종목 차단이 함께 활성이고 재조회 비교가 진입을 막는 차이를 보지 않으면
- **THEN** 두 차단이 한 번의 명령으로 함께 해제된다

#### Scenario: 귀속 실패는 스스로 풀리지 않는다
- **WHEN** 귀속 실패 차단이 활성인 채 엔진의 대사가 그 종목에 차이 없는 비교를 여러 번 관측하면
- **THEN** 그 차단은 해제되지 않는다

### Requirement: 해제 근거가 정의되지 않은 원인이 있으면 운영자 해제는 아무것도 풀지 않는다

해제 근거가 정의되지 않은 원인의 활성 차단이 하나라도 있으면 운영자 해제 명령은 아무것도 해제하지 않고 그 원인을 이름으로 말하며 거절해야 한다(SHALL). 해제 가능한 원인의 목록은 한 곳에 있어야 한다(SHALL). 계좌 전체 범위의 귀속 실패와 식별자 충돌(`IDENTIFIER_CONFLICT`)은 해제 가능한 원인이 아니다(SHALL NOT).

#### Scenario: 근거가 정의되지 않은 원인이 함께 있다
- **WHEN** 귀속 실패 종목 차단과 식별자 충돌 차단이 함께 활성이면
- **THEN** 명령은 식별자 충돌을 이름으로 말하며 거절하고 두 차단 모두 바뀌지 않는다

#### Scenario: 계좌 전체 범위의 귀속 실패
- **WHEN** 계좌 전체 범위의 귀속 실패 차단이 활성이면
- **THEN** 명령은 거절하고 그 차단은 바뀌지 않는다
