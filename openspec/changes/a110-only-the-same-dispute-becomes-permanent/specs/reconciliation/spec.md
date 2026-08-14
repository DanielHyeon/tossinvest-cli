## MODIFIED Requirements

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
