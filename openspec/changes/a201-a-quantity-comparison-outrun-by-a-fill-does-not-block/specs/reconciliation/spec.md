## ADDED Requirements

### Requirement: 체결이 조정을 대신한 수량 불일치 차단은 체결 credit 뒤의 재조회 일치로 풀린다

대사 주기는 로컬을 읽기 전에 전역 최대 체결 id 를 기록해야 한다(SHALL). 차단 판정이 성공한 뒤, 그 비교의 수량 불일치 종목 중 그 종목 체결이 그 id 뒤로 커밋됐고 이번 수렴이 조정을 반영하지 않았으며 자동 해제 대상 수량 불일치 차단이 서 있는 종목에 그 비교의 as-of 로 체결 credit 을 기록해야 한다(SHALL). 체결 credit 은 조정 credit 과 같은 규칙으로 쓰이고 반박 · 소멸하며(SHALL), 그것으로 풀린 해제는 `FILL_APPLIED` 원인과 근거를 남긴다(SHALL). 판독이 실패하면 기록해서는 안 되고 차단을 세우는 판정을 바꿔서는 안 된다(SHALL NOT). 체결 credit 은 불일치가 그 체결 때문이었음을 증명하지 않으며, 해제는 그 뒤의 재조회 일치에만 근거한다.

#### Scenario: 로컬 읽기 뒤 체결이 앞지른 수량 불일치
- **WHEN** 대사가 로컬을 읽은 뒤, 그 종목의 조정이 커밋되기 전에 엔진 주문의 체결이 커밋되어 수렴이 그 종목에 조정을 반영하지 못했고 그 주기에 수량 불일치 차단이 섰다
- **THEN** 그보다 나중 비교가 진입을 막지 않고 그 종목에 차이를 보지 않으면 그 차단은 `FILL_APPLIED` 원인으로 해제된다

#### Scenario: 이미 선 차단 위의 같은 경합
- **WHEN** 조정 credit 이 있던 종목의 수량 불일치 차단이 활성인 채 같은 경합이 일어난다
- **THEN** 그 주기 비교가 옛 credit 을 반박하고 체결 credit 이 서며, 다음 깨끗한 비교에서 `FILL_APPLIED` 원인으로 해제된다

#### Scenario: 나중 비교가 그 종목을 분쟁으로 본다
- **WHEN** 체결 credit 보다 나중 비교가 그 종목에 차이를 본다
- **THEN** 그 credit 은 소멸하고, 그 뒤 비교가 일치해도 이 근거로는 해제되지 않는다

#### Scenario: 체결이 없었거나 조정이 반영된 종목
- **WHEN** 그 종목 체결이 로컬 읽기 전에 기록한 id 뒤로 커밋되지 않았거나, 이번 수렴이 그 종목에 조정을 반영했다
- **THEN** 체결 credit 은 기록되지 않는다

#### Scenario: 판독이 실패한다
- **WHEN** 로컬 읽기 전 체결 id 나 그 종목의 체결 워터마크를 읽지 못한다
- **THEN** 체결 credit 은 기록되지 않고 그 주기의 차단 판정은 이 요구가 없을 때와 같다

#### Scenario: 재기동
- **WHEN** 체결 credit 이 있는 채 엔진이 재기동한다
- **THEN** credit 은 복원되지 않고 그 차단은 이 근거로 해제되지 않는다

#### Scenario: 다른 종목의 분쟁이 남아 있다
- **WHEN** 체결 credit 이 있는 종목은 일치하지만 같은 비교가 다른 종목 때문에 진입을 막는다
- **THEN** 해제는 비교 전체가 진입을 막지 않는 첫 관측까지 기다린다

#### Scenario: 깨끗한 비교의 외부 포지션 분쟁과 권위 확인 실패
- **WHEN** 진입을 막지 않는 비교가 credit 이 있는 종목을 외부 포지션으로 보고, 대기 중인 영구 승격의 권위 확인이 실패한다
- **THEN** 그 종목의 조정 credit 과 체결 credit 은 오류 반환 전에 소멸하고, 그 뒤의 일치로는 이 근거로 해제되지 않는다

#### Scenario: 수량 흔적 없는 종결 증명 결속 행
- **WHEN** 이 프로세스가 missing-order 로 삽입 · 결속한 종목 행에 수량 불일치가 관측된 적이 없다
- **THEN** 그 종목에는 체결 credit 이 기록되지 않고 그 행은 종결 증명 경로로만 자동 해제된다

## MODIFIED Requirements

### Requirement: 불일치 시 진입 차단
missing local order가 permanent evidence를 얻으려면 non-empty canonical `Diff.AccountRef`, order account, tracker account가 모두 같아야 한다(SHALL). blank/foreign-account comparison은 complete order identity를 포함해도 ordinary block만 유지하고 permanent evidence를 얻어서는 안 된다(SHALL NOT).
허용 오차를 넘는 불일치가 확인되면 신규 진입은 차단되고(SHALL) 청산 경로는 유지된다(SHALL — 확정 하한 규칙). 재대사는 최소 간격(기본 30초)을 두고 수행한다. 영구 승격의 연속 실패는 계좌의 모든 blocking comparison이 공유하는 횟수가 아니라 **동일 canonical blocking dispute**가 즉시 앞선 blocking comparison에도 존재한 횟수여야 한다(SHALL). 수량 불일치의 identity는 tracker와 같은 non-empty canonical `Diff.AccountRef`·non-empty 정규화 심볼·snapshot/comparer에서 float64를 거치지 않고 보존된 exact finite-decimal local quantity·broker quantity이고, missing local order의 identity는 tracker와 같은 canonical 계좌·시장·시장-local 거래일·심볼·side·opaque order identifier의 여섯 canonical component가 모두 non-empty인 tuple이다(SHALL). opaque order identifier는 whitespace-only invalidity만 검사하고 byte-preserving해야 하므로 `"id"`와 `" id "`는 다른 identity다(SHALL). 같은 identity가 연속 3회 관측될 때만 영구 불일치로 표기하고 운영자 확인 절차를 요구한다(SHALL). 다른 dispute로 바뀌거나 exact quantity tuple 또는 canonical order identity가 바뀌거나 한 comparison에서 사라지면 그 dispute의 streak는 다시 1부터 시작한다(SHALL). 비교 하나에 같은 identity가 중복되어도 한 번만 센다(SHALL). 대사 성공 시 모든 transient streak는 리셋된다. blank/foreign diff account, blank symbol, blank·malformed·non-finite quantity, tracker와 다른 account 또는 required component가 빈 missing-order처럼 identity의 exact canonicalization을 증명할 수 없으면 그 관측은 permanent streak evidence로 쓰지 않되(SHALL NOT), ordinary block은 그대로 즉시 유지한다(SHALL). 차단 범위(계좌/시장/심볼)와 자동·수동 해제 조건은 reason-code와 함께 상태표로 정의한다(SHALL).

영구 승격의 durable write가 실패한 경우 반환 직후 pending account-wide block과 account gate는 fail-closed로 유지되어야 하고(SHALL), 해당 승격은 그 승격을 얻은 canonical dispute가 바로 다음 blocking comparison에도 존재할 때만 재시도한다(SHALL). 다음 관측이 clean이거나 해당 dispute가 사라졌으면 journal authority read를 먼저 수행해야 한다(SHALL). 같은 계좌의 durable account-wide permanent row가 있으면 durable projection과 gate를 유지하고(SHALL), authority read가 실패하면 fail-closed 상태로 error를 반환해야 한다(SHALL). authority가 durable row 부재를 확인한 경우에만 pending account-wide 승격과 retry identity를 메모리·gate에서 철회한다(SHALL). 이 규칙은 ordinary symbol block의 fail-closed durable retry 또는 이미 durable한 permanent block을 철회해서는 안 된다(SHALL NOT).
continuity를 끊은 현재 blocking comparison에 새 ordinary item이 있으면 authority read 전에 그 memory block과 gate를 latch해야 한다(SHALL). authority error 뒤 성공한 Refresh가 stale account proposal을 철회해도 그 ordinary pending block을 잃어서는 안 된다(SHALL NOT).
그 현재 blocking comparison이 adjustment credit보다 later이고 credited symbol을 계속 disputed로 보고하면 authority error 반환 전에 credit을 refuted로 만료해야 한다(SHALL). 진입을 막지 않는 comparison이 credited symbol을 외부 포지션으로 disputed로 보고 authority read가 실패해도 같다 — adjustment credit과 체결 credit 모두 error 반환 전에 refuted로 만료해야 한다(SHALL). 이후 Refresh와 clean read가 stale credit으로 ordinary block을 release해서는 안 된다(SHALL NOT).

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

해제 규칙의 정밀화(SHALL): 비영구 차단의 자동 해제는 **조정 이벤트가 반영된 뒤의 재조회 일치**, 또는 「missing-order 종목 차단은 그것을 만든 주문들의 종결 증명과 그 뒤의 일치로 풀린다」 Requirement 가 정한 **결속된 주문 전부의 종결 기록 뒤의 재조회 일치**, 또는 「체결이 조정을 대신한 수량 불일치 차단은 체결 credit 뒤의 재조회 일치로 풀린다」 Requirement 가 정한 **체결 credit 뒤의 재조회 일치**에만 근거하며, 각각 구분되는 release cause(ADJUSTMENT_APPLIED 계열, ORDER_TERMINAL_CONFIRMED, FILL_APPLIED)와 원인 기록을 남긴다. 조정도 종결 증명도 체결 credit 도 없이 우연히 일치한 단발 관측은 차단을 해제하지 못하고(SHALL NOT), 영구 불일치의 해제는 운영자 확인뿐이다(SHALL).

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
조건을 전부 갖춘 경우, 또는 「체결이 조정을 대신한 수량 불일치 차단은 체결 credit 뒤의 재조회 일치로 풀린다」 Requirement 의 체결 credit 뒤의 재조회 일치뿐이다(SHALL). 영구 불일치의 해제는 운영자 확인뿐이라는 기존 규칙도 그대로다.

#### Scenario: 수렴과 같은 주기의 관측
- **WHEN** 한 대사 주기가 수량 불일치를 계좌값으로 수렴시키고, 같은 주기에서 그 수렴을 만든 것과 **같은** 비교를 관측한다
- **THEN** 차단은 유지되고, 해제는 일어나지 않으며, 그 심볼의 credit은 다음 재조회를 위해 보존된다

#### Scenario: 조정 이후의 재조회가 일치한다
- **WHEN** 조정이 기록된 뒤 더 나중에 수집된 비교가 그 심볼에 대해 일치한다
- **THEN** 비영구 차단이 ADJUSTMENT_APPLIED 원인 기록과 함께 자동 해제되고, 그 심볼의 진입 차단이 풀린다

#### Scenario: 조정 이후의 재조회가 여전히 불일치한다
- **WHEN** 조정이 기록된 뒤 더 나중에 수집된 비교가 그 심볼에 대해 여전히 불일치한다
- **THEN** credit은 소멸하고 차단은 유지되며, 해제는 새 조정 또는 새 체결 credit 과 그 뒤의 일치를 다시 요구한다

#### Scenario: 무관한 심볼이 불일치한다
- **WHEN** 조정이 기록된 심볼은 이후 재조회에서 일치하는데 같은 비교의 다른 심볼이 불일치한다
- **THEN** 일치한 심볼의 credit은 소멸하지 않고 보존되며, 비교 전체가 일치하는 첫 관측에서 그 심볼의 차단이 해제된다

#### Scenario: 비교의 선후를 판정할 수 없다
- **WHEN** 관측하는 비교에 as-of가 없거나 시각으로 읽을 수 없다
- **THEN** 일치하더라도 credit을 사용하지 않고 차단을 유지한다

#### Scenario: 조정 없는 일치의 반복
- **WHEN** 조정이 한 번도 기록되지 않았고 missing-order 종결 증명 해제의 조건도 체결 credit 도 갖추지 않은 심볼의 재조회가 여러 주기 연속으로 일치한다
- **THEN** 어떤 주기에서도 자동 해제되지 않고 운영자 해제만 남는다

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
- **THEN** 이 경로로는 해제되지 않고 조정 이벤트 또는 체결 credit 뒤의 재조회 일치만 그 차단을 풀 수 있다

#### Scenario: 행을 삽입한 뒤 그 종목에 체결이 커밋됐다
- **WHEN** 이 프로세스가 행을 삽입하기 전에 읽은 체결 워터마크보다 나중에 그 종목의 체결 이벤트(계좌 무관) 또는 종목을 알 수 없는 체결 이벤트가 하나라도 커밋되었다
- **THEN** 이 경로로는 해제되지 않는다

#### Scenario: 영구 차단은 이 경로로 풀리지 않는다
- **WHEN** missing-order 분쟁이 같은 identity 로 연속 3회 관측되어 계좌 전체 영구 차단이 생긴 뒤 그 주문의 종결 증명과 일치하는 비교들이 관측된다
- **THEN** 종목 차단은 이 경로로 풀릴 수 있지만 계좌 전체 영구 차단은 자동 해제되지 않고 운영자 확인을 요구한다
