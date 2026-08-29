# order-execution Specification

## Purpose
엔진 주문 실행의 상태 모델·journal 내구성·멱등성(멱등 재생 포함)·재시도·위험 예약·RECONCILE·fail-closed 요구사항.
## Requirements
### Requirement: Intent-Mutation-주문 분리 모델
주문 실행 모델은 세 노드를 분리해야 한다(SHALL): 불변 OrderIntent(의도), MutationAttempt(PLACE/CANCEL/AMEND 시도 — 각각 독립 수명주기), 브로커 주문 노드(주문번호 단위). 토스 공식 API의 cancel/modify는 새 주문번호를 발급하므로, 주문 노드 간 관계는 lineage edge(`replaces`, 원주문 체결수량, 요청 잔여수량, 새 주문번호)로 기록해야 한다(SHALL). 엔진 경로의 lineage 기록은 journal DB 안에서 트랜잭션으로 수행한다(SHALL).

#### Scenario: 부분체결 중 정정
- **WHEN** 부분체결된 주문에 AMEND가 성공해 새 주문번호가 발급되면
- **THEN** 원주문 노드는 체결수량과 함께 종결되고, 새 주문 노드가 `replaces` edge·잔여수량과 함께 생성되며, 두 기록은 동일 트랜잭션에서 커밋된다

#### Scenario: 다단계 정정 체인
- **WHEN** 정정이 2회 연속 수행되면
- **THEN** lineage 체인으로 최초 주문번호에서 현재 주문번호를 결정적으로 해소할 수 있다

### Requirement: MutationAttempt 수명주기
각 MutationAttempt는 RECORDED → DISPATCH_STARTED → (ACKED | IN_DOUBT) → 종결(CONFIRMED | NOT_DISPATCHED | FAILED_CONFIRMED | UNRESOLVED_IN_DOUBT) 단계를 가져야 한다(SHALL). RECORDED는 fsync 완료 후에만 DISPATCH_STARTED로 진행하며(SHALL), RECORDED 단계에서 결정 참조(decision_id·safety_class·generation)·멱등키·canonical wire body·serializer version이 함께 불변 영속된다(SHALL — 기록 API의 공개 계약 확장을 수반한다). 재생은 저장된 wire body만 사용한다(SHALL NOT 재구성). 멱등키는 결정에서 유도된 값이며 확인 토큰의 canonical 입력에 포함되지 않는다(SHALL NOT — CLI confirm token 무변경). 재시작 시 RECORDED에서 멈춘 attempt는 NOT_DISPATCHED로 안전 종결하고, DISPATCH_STARTED 이후는 해소 완료 전까지 같은 심볼의 신규 mutation을 차단한다(SHALL — 전 클래스).

#### Scenario: 전송 시작 전 크래시
- **WHEN** RECORDED까지만 기록된 attempt가 재시작 시 발견되면
- **THEN** NOT_DISPATCHED로 종결되고 어떤 차단도 발생하지 않는다

#### Scenario: 전송 중 크래시
- **WHEN** DISPATCH_STARTED로 기록된 attempt가 재시작 시 발견되면
- **THEN** 영속된 멱등키·wire body로 해소 절차가 시작된다

#### Scenario: 직렬화 규칙 변경 후 재생
- **WHEN** 바이너리 업데이트로 직렬화 규칙이 바뀐 뒤 이전 attempt의 재생이 필요하면
- **THEN** 저장된 wire body가 그대로 사용되어 본문 불일치가 발생하지 않는다

### Requirement: IN_DOUBT 해소
IN_DOUBT 해소의 목적은 **정체 회수**다 — 주문이 접수됐는지, 접수됐다면 어떤 식별자인지 확정하는 것이며, 정체를 모르는 채 다시 주문을 내는 것은 금지된다(SHALL NOT).

공식 API의 주문 생성은 클라이언트 제공 멱등키를 지원한다(openapi `clientOrderId`: "동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다", 유효 10분). 해소 순서(SHALL):

1. **멱등 재생**: RECORDED 단계에 영속된 동일 키·동일 wire body의 재요청 응답에서 식별자를 회수한다. 다음은 **진입점 자신의 의무**다(SHALL — 호출자 전제 금지): 대상 attempt가 IN_DOUBT 상태인지 확인, 능력 attestation 플래그 확인 `[미측정 — 2b 전 비활성]`, **재생 1회마다** `elapsed(전송 시작) < TTL − margin`(margin 기본 60초, 설정 주입 — 경과 근거가 로컬 시계이므로 마진 없는 경계 사용 금지(SHALL NOT)) 재검사, 회수 상한(기본 2회)·최소 간격 준수, 저장된 wire body 외의 본문을 구성·전송할 수 없는 형태(SHALL — 진입점 입력은 attempt 식별자뿐). 재생은 새 attempt를 만들지 않고 nonce를 소비하지 않으며 해소 기록(횟수·시각)으로 남는다(SHALL).

   재생 응답 분류에 dispatch 분류기를 사용해서는 안 된다(SHALL NOT — 재생의 422는 의미가 다르다): 2xx+식별자 → 회수·CONFIRMED; `409 request-in-progress`(openapi) → 원 요청 처리 중, 대기 후 재시도(상한 미소비); `422 idempotency-key-conflict`(openapi) → 원 주문에 대해 아무것도 증명하지 않으므로 FAILED_CONFIRMED 금지(SHALL NOT), 프로그램 오류로 UNRESOLVED + critical 알림; 응답 유실 → 기록 후 상한 도달 시 조회 폴백.
2. **조회 대조 (폴백)**: 창 초과, 미검증, 키 충돌, 멱등키를 받지 않는 mutation(취소·정정)은 journal fingerprint로 미체결·종결 양쪽 목록을 pagination 완주하며 대조한다. `status` 파라미터는 그룹 라벨이고 `PARTIAL_FILLED`는 양쪽 그룹에 모두 속하므로(openapi), 유일성 판정 전에 **orderId 기준 dedup**을 수행해야 한다(SHALL — dedup 없이는 부분 체결 주문이 이중 매칭되어 영구 주차된다). 조회 응답은 멱등키를 싣지 않으므로 키로 매칭할 수 없다(SHALL NOT).
3. **부재 판정**: 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 + 매수가능금액·보유수량 delta 교차 확인 후에만. 관측 창 동안 같은 심볼에 다른 mutation이 전송되었다면 delta 교차 확인은 무효이며 자동 FAILED_CONFIRMED는 금지된다(SHALL NOT — 부재를 증명할 수 없으므로 UNRESOLVED로).
4. **해소 불능**: UNRESOLVED_IN_DOUBT로 해당 심볼 신규 진입 영구 차단, 운영자 해소만 허용.

재시작 복구가 호출 순서를 소유한다(SHALL): 미종결 attempt 순회 → IN_DOUBT이고 재생 적격이면 Gateway의 해소 전용 재생 진입점 → 부적격·실패 시 조회 폴백. 해소 절차 자체는 mutator를 갖지 않는다(P1 불변식 유지 — 재생 진입점은 Gateway 소속). 조회 대조의 유일 매칭을 위해 심볼당 in-flight mutation 1개 제한은 **모든 safety class에** 유지된다(SHALL — 동시 반대 방향 제출은 브로커도 `422 opposite-pending-order-exists`로 거부한다(openapi)). §0.3은 동시성이 아니라 해소의 우선·유계성(재생이 이를 단축)과 수동 flatten의 순서화된 saga로 달성한다.

#### Scenario: 멱등 재생으로 정체 회수
- **WHEN** 능력 검증 완료 상태에서 제출 응답이 유실되고 TTL−margin 이내에 해소가 시작되면
- **THEN** 저장된 wire body의 재요청 응답에서 orderId를 회수해 CONFIRMED로 종결하며 두 번째 주문은 생성되지 않는다

#### Scenario: 두 번째 재생의 시간 재검사
- **WHEN** 첫 재생 후 재시도 시점에 경과 시간이 TTL−margin을 넘었으면
- **THEN** 두 번째 재생은 수행되지 않고 조회 폴백으로 전환된다

#### Scenario: 재생 중 409
- **WHEN** 재생 요청이 409 request-in-progress를 받으면
- **THEN** 상한을 소비하지 않고 대기 후 재시도하거나 조회 폴백으로 전환한다

#### Scenario: 부분 체결 주문의 양쪽 목록 출현
- **WHEN** 조회 폴백이 같은 orderId를 미체결·종결 목록 양쪽에서 발견하면
- **THEN** dedup 후 단일 매칭으로 CONFIRMED 종결되며 이중 매칭 주차가 발생하지 않는다

#### Scenario: 관측 창 오염
- **WHEN** 부재 판정 창 동안 같은 심볼의 다른 mutation이 전송되었으면
- **THEN** delta 교차 확인이 무효화되어 자동 FAILED_CONFIRMED 없이 UNRESOLVED로 처리된다

#### Scenario: 해소 불능
- **WHEN** 관찰 기간 내 존재도 부재도 증명되지 않으면
- **THEN** UNRESOLVED_IN_DOUBT로 표기되어 해당 심볼 신규 진입이 영구 차단되고 운영자 알림이 발송된다

### Requirement: 브로커 상태 파생
주문 종결 상태는 공식 API의 원시 status와 `(canceledAt, execution.filledQuantity, quantity, lineage)`에 대한 우선순위 파생 함수로 결정해야 한다(SHALL). 파생은 문서화된 OrderStatus 10개 전체를 다룬다(SHALL — openapi): PENDING, PENDING_CANCEL, PENDING_REPLACE, PARTIAL_FILLED, FILLED, CANCELED, REJECTED, CANCEL_REJECTED, REPLACE_REJECTED, REPLACED. status와 다른 필드의 모순 조합·미지 값은 UNKNOWN_BROKER_STATE로 fail-closed 처리한다(SHALL: 해당 심볼 신규 진입 차단 + 알림 + 운영자 해제 경로).

`CANCEL_REJECTED`·`REPLACE_REJECTED`는 "별도 주문 레코드로 생성됨"(openapi — 원주문은 이전 상태로 복귀). 취소·정정의 해소 절차는 이 별도 레코드를 인지해야 하며(SHALL), 레코드의 구체 형태는 `[미측정 — 2b]`이므로 귀속 실패는 외부 분류가 아니라 RECONCILE로 처리한다(SHALL — fail-closed).

#### Scenario: 10값 status의 정상 파생
- **WHEN** status=CANCELED이고 canceledAt·filledQuantity가 정합인 주문을 파생하면
- **THEN** UNKNOWN이 아니라 CANCELLED(부분 체결 시 부분체결 후 취소)로 판정된다

#### Scenario: 취소 거부 레코드 관측
- **WHEN** 취소 요청 후 CANCEL_REJECTED 상태의 별도 주문 레코드가 관측되면
- **THEN** 원주문은 이전 상태로 복귀한 것으로 파생되고, 별도 레코드는 외부 주문으로 분류되지 않는다

#### Scenario: 미지의 status 값
- **WHEN** 파생 함수가 알 수 없는 status 문자열을 받으면
- **THEN** UNKNOWN_BROKER_STATE로 fail-closed 처리되고 알림이 발송된다

### Requirement: Journal 내구성
Journal은 SQLite 단일 파일로 하되(SHALL): 저장 경로는 `$TOSSOS_DATA_DIR` > `$XDG_DATA_HOME/tossos` > `~/.local/share/tossos` 순으로 해석하고, 로컬 저널링 파일시스템 allowlist(ext4/xfs/btrfs) 밖이면 기동을 거부한다(SHALL). intent 기록은 `BEGIN IMMEDIATE` + `synchronous=FULL`로 커밋 성공 후에만 제출을 진행하고(SHALL), 스키마는 버전 필드와 additive migration 규칙을 가지며, 손상 감지 시 기동을 거부한다(SHALL). 마이그레이션 직전 자동 백업을 수행하고 복원 절차를 문서화·테스트한다(SHALL) — 구버전 바이너리 실행은 ErrSchemaTooNew 기동 거부이므로 롤백 수단이 아니다.

#### Scenario: 커밋 실패 시 제출 차단
- **WHEN** journal 트랜잭션 커밋이 실패하면
- **THEN** 브로커 제출은 수행되지 않는다

#### Scenario: DB 손상
- **WHEN** 기동 시 journal 무결성 검사가 실패하면
- **THEN** 기동이 거부되고 복구 안내가 출력된다

### Requirement: Retry Matrix 산출물
재시도 정책은 endpoint×method×오류 클래스 표로 스펙 산출물화해야 한다(SHALL). 최소 규정: **주문 mutation은 어떤 오류에도 자동 재시도 금지 — 단, 멱등 재생은 재시도가 아니라 정체 회수다**(같은 키·같은 본문의 재요청은 유효 창 안에서 새 주문을 만들 수 없다 — openapi; IN_DOUBT 해소 요구의 조건·상한·비활성 규칙을 따른다). 이 구분은 재시도 정책 구현부의 문서화된 근거에도 반영한다(SHALL). 조회는 재시도 예산(횟수·총 시간)과 bounded jitter, 429는 Retry-After 상한 준수, 401/403은 즉시 신규 진입 차단 + 알림, 필수 조회의 staleness 임계 초과 시 신규 진입 차단. 표의 수치는 구현 시 확정하고 표 없이 구현하지 않는다(SHALL NOT).

#### Scenario: 재생과 재시도의 구분
- **WHEN** 멱등키 없이 전송된 mutation이 모호한 오류로 끝나면
- **THEN** 어떤 자동 재전송도 발생하지 않는다 (재생은 영속된 키·본문이 있는 attempt의 해소 절차에서만)

#### Scenario: 필수 조회 장기 실패
- **WHEN** 잔고 조회가 재시도 예산을 소진하고 staleness 임계를 초과하면
- **THEN** 신규 진입이 차단되고 조회 복구 후 자동 해제된다

### Requirement: 시간 규율
모든 시간 판정(제출 시각 창, staleness, 안정화 간격, 거래일 경계)은 주입 가능한 clock을 사용해야 하며(SHALL), 시장별 timezone(KST/ET)과 DST 전환을 명시적으로 처리해야 한다(SHALL). 거래일 경계는 시장별 규칙으로 정의한다.

#### Scenario: DST 전환일의 US 세션 판정
- **WHEN** 미국 DST 전환일에 세션 판정을 수행하면
- **THEN** ET 기준 정확한 장 시간이 사용된다

### Requirement: Fail-closed 분기
공식 주문 경로에서 interactive auth challenge 요구, USD 주문의 통화 잔고 부족, 미지원 주문 유형은 제출 없이 거부하고 사유 코드와 함께 기록·통지해야 한다(SHALL). 자동 환전·자동 승인은 금지된다(SHALL NOT). 차단·거부 사유는 안정적 reason-code enum으로 정의한다(SHALL).

#### Scenario: USD 잔고 부족 매수
- **WHEN** USD 매수 intent의 사전 잔고 확인이 부족을 반환하면
- **THEN** 제출 없이 reason code와 함께 거부·통지된다

### Requirement: 원자적 위험 예약
계좌 전체에 걸친 한도(총 개방 노출, 일일 손실, 현금)의 판정과 그 결과의 예약은 하나의 journal 트랜잭션 안에서 수행되어야 한다(SHALL). 서로 다른 심볼에 대한 동시 결정이 같은 스냅샷을 각각 통과해 합산 한도를 초과하는 것은 허용되지 않는다(SHALL NOT).

브로커 조회를 이 트랜잭션 안에서 수행해서는 안 된다(SHALL NOT — journal은 단일 커넥션이므로 네트워크 왕복 동안 모든 mutation 기록이 막힌다). 스냅샷은 트랜잭션 밖에서 수집하고, 안에서 as-of·staleness를 검증한 뒤 예약을 삽입하며, 불충족이면 롤백·재수집한다(SHALL). 재수집은 횟수 상한(기본 3회)과 총 데드라인을 가지며 초과 시 fail-closed로 거부한다(SHALL). 예약 산술은 decimal 문자열 연산이며 float 누적을 사용하지 않는다(SHALL NOT). one-shot nonce의 소비 기록은 전송 시작 기록과 같은 트랜잭션에서 수행한다(SHALL).

예약은 attempt에 결속된다(SHALL — Prepare 시 attempt 참조 backfill). 해제 트리거:

- **파생된 브로커 종결 상태 도달** — 상태 파생 함수가 FILLED·CANCELLED 계열·REJECTED로 판정한 경우와 NOT_DISPATCHED·FAILED_CONFIRMED 종결. 만료 추정으로 해제해서는 안 된다(SHALL NOT — 미체결 만료가 어떤 status로 나타나는지는 `[미측정 — 2b]`이며, 파생 함수는 설명 불가 조합을 UNKNOWN으로 보존한다)
- nonce **미소비** 상태의 결정 만료 (소비 후 만료는 해제하지 않는다 — 주문이 접수됐을 수 있다)
- **운영자 해제** — UNKNOWN_BROKER_STATE·UNRESOLVED_IN_DOUBT가 잡은 예약의 유일한 출구(SHALL 제공). 근거 기록·audit 필수
- 일일 손실 예약의 시장별 거래일 경계 소멸(SHALL)

해제는 종결 상태의 journal 기록과 같은 트랜잭션에서 원자적으로 수행한다(SHALL — 관측 생산자가 무엇이든 기록 시점에 해제가 따라온다).

#### Scenario: 동시 다심볼 결정
- **WHEN** 총 개방 노출 한도의 잔여분이 1건분만 남은 상태에서 서로 다른 두 심볼의 결정이 동시에 요청되면
- **THEN** 하나만 예약에 성공하고 다른 하나는 한도 초과로 거부된다

#### Scenario: UNKNOWN 상태가 잡은 예약
- **WHEN** attempt의 브로커 상태가 UNKNOWN_BROKER_STATE로 fail-closed 되어 예약이 유지되면
- **THEN** 운영자 해제 경로가 근거 기록과 함께 제공되고 알림이 발송된다

#### Scenario: nonce 소비 후 만료
- **WHEN** nonce가 소비된 뒤 응답이 유실되고 결정 유효 시간이 지나면
- **THEN** 예약은 만료를 이유로 해제되지 않고 해소 완료까지 유지된다

#### Scenario: 거래일 경계
- **WHEN** 일일 손실 예약을 보유한 채 거래일이 바뀌면
- **THEN** 그 예약은 소멸하고 새 거래일의 한도가 온전히 사용 가능하다

### Requirement: RECONCILE 상태
권위 값의 불일치는 산식으로 보정하지 않고 RECONCILE 상태로 전이해야 한다(SHALL). 진입 조건: 소비자가 필요로 하는 시점의 브로커 보유·매도가능 조회 불가 또는 staleness 한계 초과, 로컬 파생 수량과 브로커 스냅샷의 불일치, 같은 브로커 식별자가 상충하는 계좌·심볼 컨텍스트에 출현.

RECONCILE 상태는 journal에 영속된다(SHALL — scope·원인·증거·진입/해제 시각). journal 기록이 권위이고, 기존 in-memory 차단 기계(EntryGate 래치·reconcile Tracker)는 기동 시 journal에서 재구성되는 투영이다(SHALL — 재시작이 차단을 잃어서는 안 된다). 해제는 재조회 일치와 원인 기록을 요구한다(SHALL).

RECONCILE 중 신규 진입과 수량 확대는 금지되고(SHALL NOT), 위험 축소는 **확정 하한 수량**으로만 허용된다(SHALL):

```text
확정 하한 = max(0, min(신선한 브로커 보유수량, 신선한 매도가능수량) − 로컬 미체결 SELL 수량)
신선 = 스냅샷 나이 ≤ staleness 한계 · 스냅샷 부재 시 0 (자동 경로의 로컬 매도 없음)
```

로컬 파생 수량을 하한을 올리는 근거로 사용해서는 안 된다(SHALL NOT — 외부 주문 제외로 실제보다 클 수 있다). 매도가능수량의 의미는 `[미측정 — 2b]`이나 min()에서 하한을 낮추는 방향으로만 사용한다. 수동 flatten은 자체 신선 조회를 수행하므로 이 규칙의 대상이 아니다(§0.3 — P1 동작 보존).

#### Scenario: 수량 불일치 시 청산 요청
- **WHEN** RECONCILE 상태에서 청산이 요청되면
- **THEN** 확정 하한 공식의 수량까지만 제출되고 초과분은 해소 후로 보류된다

#### Scenario: 재시작 후 차단 유지
- **WHEN** RECONCILE 상태에서 프로세스가 재시작되면
- **THEN** journal에서 상태가 재구성되어 진입 차단이 유지된다

### Requirement: 총계 한도의 계산 계약
총 개방 노출·일일 손실·현금은 계산 계약이 정의되어야 하며(SHALL), 정의되지 않은 양에 예약을 걸어서는 안 된다(SHALL NOT). 확정된 구조 결정(수치는 후속 change가 **보수 방향으로만** 대체):

- 자동 진입은 LIMIT 전용이다(SHALL — 시장가 진입의 노출 평가가는 정의되지 않는다). 진입 노출 평가 = 지정가 × 수량 + 과대 추정 비용
- 노출은 long-only gross: 보유 평가액 + 미체결 매수 평가액 + 유효 진입 예약의 합(SHALL)
- 일일 손실은 실현 손익(비용 차감 후) 기준(SHALL — 미실현 포함 여부는 후속이 보수 방향으로만 추가)
- 통화 정규화: 최신 환율, staleness 한계 초과 시 fail-closed(SHALL)
- 외부 수동 거래는 브로커 스냅샷을 통해서만 반영된다(SHALL)
- staleness 보수 기본값: 계좌·보유 스냅샷 10초, 환율 60초. 후속 change는 보수 방향(축소) 또는 사람 승인·audit로만 변경(§0.9)

입력이 stale하거나 미지이면 fail-closed로 진입을 거부한다(SHALL).

#### Scenario: 환율 stale
- **WHEN** 외화 자산의 원화 환산에 필요한 환율이 staleness 한계를 넘으면
- **THEN** 총 개방 노출 판정이 fail-closed로 진입을 거부한다

#### Scenario: 시장가 진입 시도
- **WHEN** 자동 경로에서 시장가 진입 의도가 평가되면
- **THEN** LIMIT 전용 규칙으로 거부된다

### Requirement: 브로커 식별자의 opaque 취급
브로커가 발급하는 주문 식별자는 opaque token으로 취급해야 한다(SHALL — openapi는 `orderId`에 형태·패턴을 계약하지 않는다). 공백 제거 후 빈 값은 거부하고(SHALL), 저장은 수신 원문 그대로 하며(SHALL — trim·대소문자·숫자 변환 금지), 비교는 바이트 동일로 한다(SHALL). 형태·prefix·길이 패턴의 검증·해석을 추가해서는 안 된다(SHALL NOT). 식별자는 계좌 스코프와 함께 저장하고(SHALL), 생성 응답의 식별자는 상세조회 round-trip으로 실재를 확인하며(SHALL), round-trip 실패·부재 시 CONFIRMED로 종결하지 않고 IN_DOUBT로 남겨 해소 절차가 확정한다(SHALL). 같은 식별자가 상충하는 계좌·심볼 컨텍스트에 나타나면 RECONCILE로 전이한다(SHALL).

#### Scenario: 빈 식별자
- **WHEN** 생성 응답의 orderId가 비어 있으면
- **THEN** ACK로 처리되지 않고 IN_DOUBT 해소가 시작된다

#### Scenario: round-trip 실패
- **WHEN** 생성 응답의 식별자가 상세조회에서 확인되지 않으면
- **THEN** CONFIRMED가 아니라 IN_DOUBT로 남고 해소 절차가 정체를 확정한다

### Requirement: 체결 정정 이벤트
누적 체결 수량이 동일한데 평균 체결가 또는 체결 금액이 변경된 관측은 수량을 재반영하지 않고 EXECUTION_CORRECTION 이벤트로 기록해야 한다(SHALL). 감지를 위해 체결 스냅샷은 체결 금액을 함께 저장하고 관측 payload는 `filledAmount`를 읽는다(SHALL). 정정 이벤트 삽입과 스냅샷 갱신은 체결 반영과 같은 트랜잭션에서 수행한다(SHALL — prev가 존재하는 유일한 지점). 평균 체결가는 부분 체결마다 바뀌는 값이므로(openapi) 중복 판정 키에 포함해서는 안 된다(SHALL NOT). 누적 수량 감소는 fail-closed를 유지한다.

#### Scenario: 수량 동일·평균가 변경
- **WHEN** 같은 주문의 누적 수량이 동일하고 평균 체결가만 달라진 스냅샷이 관측되면
- **THEN** 수량 delta 없이 정정 이벤트가 기록되고 스냅샷이 같은 트랜잭션에서 갱신된다

#### Scenario: 동일 관측 반복
- **WHEN** 이미 반영된 것과 동일한 스냅샷이 다시 관측되면
- **THEN** 정정 이벤트가 중복 삽입되지 않는다

### Requirement: 조건주문 mutation은 durable execution contract를 따른다
조건주문 create/replace/cancel은 canonical body, serializer version, client order identity, mutation attempt와 broker identifier를 submit 전에 영속해야 한다 (SHALL).

#### Scenario: create 응답 유실
- **WHEN** broker가 create를 처리했으나 응답이 유실된다
- **THEN** attempt는 IN_DOUBT로 남고 attested idempotency/reconciliation 절차 외 재제출을 금지한다

#### Scenario: replace가 새 ID를 반환
- **WHEN** 보호 정정이 새 conditional ID를 만든다
- **THEN** old/new identifier와 유효성 전환을 한 saga generation에 기록한다

### Requirement: strategy provenance는 주문까지 끊기지 않는다
entry decision, RiskIntent, mutation attempt, broker order와 fill은 candidate-life ID, threshold set/evidence digest와 lane ID/version을 결정적으로 연결해야 한다 (SHALL). 전략 provenance가 없는 legacy RiskIntent의 canonical bytes는 바뀌어서는 안 된다 (MUST NOT).

#### Scenario: 정상 체결
- **WHEN** strategy order가 체결된다
- **THEN** fill과 열린 position에서 원 candidate와 lane version을 역추적할 수 있다

#### Scenario: 재시작 중 중복 decision
- **WHEN** 같은 canonical decision이 재시작 뒤 다시 계산된다
- **THEN** deterministic identity와 duplicate guard가 두 번째 LIVE order를 차단한다

### Requirement: 승인 계획은 단계가 지목할 객체를 이름한다

검증 실행의 승인 계획은 각 mutating 단계가 실제로 지목할 **대상 객체**를 이름해야 한다(SHALL).
이미 등록된 조건주문 위에서 동작하는 단계의 계획 줄은 그 조건주문의 종목을 실어야 하며,
실행의 probe 종목을 실어서는 안 된다(SHALL NOT — 2026-07-29 실측: `conditional-modify`가
`005930`으로 계획되고 `333430`으로 요청되어 두 번 연속 `ErrOutsidePlan`으로 멈췄고,
그 결과 이 도구가 남긴 조건주문을 이 도구가 제거할 수 없었다).

대상을 이름할 수 없는 mutating 단계는 계획에 오르지 않아야 한다(SHALL — 모르는 것을 승인받지
않는다는 수량 규칙과 같은 방향이다). 종목이 비어 있는 계획 줄은 종목이 있는 요청을 인가해서는
안 된다(SHALL NOT — 승인 목록의 한 줄은 무엇에 대한 요청인지 말해야 하고, 말하지 않는 줄은
무엇이든 인가하는 줄이 되어서는 안 된다).

계획 밖 요청이 아무것도 전송하지 않고 실행 전체를 멈추는 레일은 무변경이다(SHALL — 이 요구는
목록을 정확하게 만들 뿐 인가를 느슨하게 하지 않는다).

#### Scenario: 이전 실행이 남긴 조건주문의 정정

- **WHEN** probe 종목과 다른 종목에 이 검증이 등록한 조건주문이 살아 있는 기록으로
  조건주문 정정·취소 단계를 계획하면
- **THEN** 두 계획 줄은 그 조건주문의 종목을 싣고, 실행이 같은 종목으로 보내는 요청이 인가된다

#### Scenario: 아직 등록되지 않은 조건주문

- **WHEN** 같은 실행에서 조건주문 등록부터 도는 계획을 만들면
- **THEN** 정정·취소 줄은 등록이 쓸 보유 종목을 싣고, 등록 후 실행이 보내는 요청과 일치한다

#### Scenario: 대상을 이름할 수 없는 단계

- **WHEN** mutating 단계의 대상 종목을 계좌에서 정할 수 없으면
- **THEN** 그 단계는 계획에서 제외되고 사유가 표시되며, 어떤 요청도 인가되지 않는다

#### Scenario: 종목 없는 줄의 인가 범위

- **WHEN** 종목이 비어 있는 계획 줄만 있는 상태에서 종목이 있는 요청이 만들어지면
- **THEN** 인가되지 않고 아무것도 전송되지 않으며 실행이 멈춘다

### Requirement: 정리는 진행 중인 측정이 기다리는 객체를 건드리지 않는다

검증 실행의 정리 prologue는 아직 끝나지 않은 측정이 기다리는 객체를 취소 대상으로 삼아서는 안 된다(SHALL NOT).
기록은 각 객체에 대해 **어느 단계의 판정이 그 객체를 놓아주는지**를 말할 수 있어야 하며(SHALL),
정리는 그 단계가 **붙잡음이 선언된 뒤에** terminal 판정을 받은 경우에만 그 객체를 대상으로
삼는다(SHALL).

이 규칙은 조건주문에 한정되지 않는다(SHALL — 2026-07-29 판정: 조건주문 발동이 만드는 child
주문은 **체결되어야 하는** 주문이므로 취소되면 측정 자체가 성립하지 않는데, 현재 규칙은 모든
주문을 무조건 대상으로 삼는다).

붙잡음의 해제는 **기록의 위치**로만 판정한다(SHALL — 기록은 append-only이므로 index만 단조이고
취소 줄은 zero time을 싣는다). 경과 시간을 근거로 붙잡힌 객체를 취소 대상으로 되돌려서는
안 된다(SHALL NOT — 대기가 길다는 것은 측정 실패가 아니라 시장 조건이 아직 오지 않았다는
뜻이다).

붙잡을 단계를 기록이 말하지 않는 객체는 기존 판정을 그대로 따른다(SHALL — 조건주문은
조건주문 취소 단계의 판정이 gate이고, 주문은 gate가 없다. 이 change 이전에 기록된 모든 줄의
판정이 바뀌지 않아야 한다).

한 측정이 만든 객체들은 **같은 사슬 식별자**를 실을 수 있어야 한다(SHALL — 정정이 새 식별자를
발급하고 옛 식별자가 즉시 404가 되는 실측(M40) 뒤에도 둘이 한 측정의 것임을 기록이 말해야
한다).

#### Scenario: 발동을 기다리는 측정이 남긴 주문

- **WHEN** 어떤 단계가 자기 판정이 나올 때까지 붙잡겠다고 선언한 주문이 기록에 남아 있고
  그 단계가 아직 terminal 판정을 받지 않았으면
- **THEN** 그 주문은 정리 대상이 아니고 승인 목록에 오르지 않는다

#### Scenario: 측정이 끝나면 놓아준다

- **WHEN** 붙잡은 단계가 그 선언 뒤에 terminal 판정을 기록하면
- **THEN** 그 객체는 다시 정리 대상이 되어 승인 목록에 오른다

#### Scenario: 선언보다 앞선 판정은 놓아주지 않는다

- **WHEN** 붙잡은 단계의 마지막 판정이 그 붙잡음이 선언되기 **전에** 기록된 것이면
- **THEN** 그 객체는 정리 대상이 아니다

#### Scenario: 붙잡을 단계를 말하지 않는 기존 기록

- **WHEN** 사슬·붙잡음 필드가 없는 기존 기록으로 정리 대상을 계산하면
- **THEN** 조건주문은 조건주문 취소 단계를 gate로, 주문은 무조건 대상으로 — 이 change 이전과
  같은 목록이 나온다

#### Scenario: 정정을 건너서도 한 사슬로 남는다

- **WHEN** 붙잡힌 조건주문이 정정으로 새 식별자를 받으면
- **THEN** 새 객체가 같은 사슬 식별자를 싣고, 붙잡음도 새 객체로 이어진다
