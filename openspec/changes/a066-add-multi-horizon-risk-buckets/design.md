## Context

기존 Guardian은 개별 주문, 총 노출과 일손실을 결정+예약 transaction으로 강제한다. 후속 lane는 KR/US와 short/medium horizon에서 여러 leg를 만들므로, 개별 주문이 기존 한도 안이어도 합산 집중도가 커질 수 있다. 같은 symbol을 두 lane가 동시에 획득하면 exit 책임과 scale-in 예약도 갈라진다.

이 change는 기존 Guardian 체인과 예약 권위를 약화하지 않고 그 뒤에 더 보수적인 다차원 cap을 추가한다. a065 campaign identity를 소비하지만 campaign core나 lane별 비율을 소유하지 않는다.

## Goals / Non-Goals

**Goals:**

- horizon, market, sector와 symbol별 exposure를 합산해 수량을 보수적으로 cap한다.
- 최종 수량 `q_final`이 lane가 제안한 `q_candidate`를 절대 초과하지 않게 한다.
- symbol/market의 owning lane와 campaign을 원자 예약한다.
- scale-in, partial fill, retry와 restart에도 예약·실노출을 중복 없이 재구성한다.
- entry loss lock과 bucket 장애가 위험 감소 경로를 막지 못하게 한다.

**Non-Goals:**

- 전략별 목표 수량·leg 비율 산출
- 운영 한도 수치의 자동 완화 또는 market 간 자본 자동 이전
- short, leverage 또는 비공식 FX source 도입
- live lane/automation 활성화

## Decisions

### D1. Bucket은 금액 단위의 직교 dimension 교집합이다

모든 exposure-raising 요청은 horizon(`SHORT|MEDIUM`), market(`KR|US`), strategy risk bucket ID/version, sector와 symbol bucket key를 가져야 한다. 각 bucket은 account base currency의 monetary limit, filled monetary exposure, HELD monetary reservation, valuation provenance, snapshot version과 freshness를 제공한다. unknown horizon/market/strategy/sector 또는 환산 불능은 별도 무제한 bucket으로 보내지 않고 진입을 거부한다. strategy bucket은 개별 lane ID와 같을 수도 있지만 여러 lane을 한 정책으로 묶을 수 있는 server-owned versioned risk identity이며 caller 문자열을 그대로 신뢰하지 않는다.

각 dimension을 가중 합산하는 대안은 한 dimension 초과가 다른 여유로 상쇄되므로 기각한다.

### D2. Monetary reservation은 worst price, fees와 FX haircut을 사용한다

buy quantity `q`의 예약 금액은 account base currency minor unit에서 다음 순수 decimal 함수로 계산한다.

`reserve(q) = ceil_minor(q × worst_executable_price_quote × fx_rate_quote_to_base × fx_haircut_multiplier + worst_case_fees_base(q))`

`worst_executable_price_quote`는 공식 주문 계약상 해당 intent가 체결될 수 있는 보수적 최고 가격이며 source/version/observed-at을 가진다. `fx_haircut_multiplier`는 1 이상이고 같은 통화이면 rate와 multiplier가 1이다. fee 함수와 FX policy도 version/digest를 가진다. 입력이 missing/stale/non-positive이거나 haircut이 1 미만, arithmetic overflow 또는 minor-unit ceil이 불가능하면 진입을 거부한다. caller price, mid-price 또는 fee 0 fallback은 사용하지 않는다.

각 bucket cap은 `reserve(q) <= monetary_remaining`을 만족하는 최대 non-negative 정수 q다. fee가 비선형일 수 있으므로 단순 per-share 나눗셈을 권위로 쓰지 않고 bounded monotone search 또는 동등한 exact algorithm을 사용한다.

### D3. q_final은 q_candidate와 모든 금액 cap의 최솟값이다

canonical field는 lane 제안 `q_candidate`와 최종 `q_final`이다. `q_final = min(q_candidate, q_existing_guardian, q_horizon, q_market, q_strategy, q_sector, q_symbol)`이며 각 cap은 위 monetary reservation 함수로 도출한 non-negative integer다. overflow, stale snapshot, price/fee/FX/currency 부재는 `q_final=0` typed refusal이다. 이 계층은 수량을 늘리는 multiplier를 갖지 않는다.

수량 0은 주문 intent가 아니라 typed refusal이다. `q_candidate`, `q_final`, reservation input/digest, 각 binding monetary cap과 산출 수량을 decision evidence로 저장한다.

### D4. q_final 확정 뒤 GuardianDecision과 예약을 원자 발급한다

기존 Guardian chain은 mutation 없는 precheck와 `q_existing_guardian` cap만 산출한다. bucket calculator가 `q_final`과 각 monetary reservation을 확정한 뒤에만 `(account, market, symbol, prospective-or-actual position generation)` unique owner row, 모든 HELD monetary reservation과 **q_final을 봉인한 GuardianDecision**을 expected snapshot version 아래 하나의 transaction으로 commit한다. q_candidate나 precheck quantity로 GuardianDecision을 먼저 기록해서는 안 된다. 같은 lane/campaign의 후속 leg만 기존 owner를 재사용할 수 있고 경쟁 lane는 전부 rollback된다.

사전 in-memory lock은 다중 process/restart를 막지 못하므로 권위로 사용하지 않는다. a065 prospective generation token은 first fill 전 owner key가 되고 실제 successor generation에 set-once 결합된다.

### D5. filled와 HELD monetary exposure를 함께 계상한다

bucket usage는 authoritative Position/fill projection에 귀속된 filled monetary exposure와 미해소 HELD monetary reservation의 합이다. 각 deduplicated positive fill delta의 Position apply transaction은 horizon, market, strategy, sector와 symbol 모든 적용 bucket에 대해 원 reservation policy/rounding으로 계산한 `transfer_delta = proportional_reserved_allocation(new_cumulative_fill) - previously_transferred`와 실제 fill price, allocated fee 및 fill-time persisted FX에 따른 `actual_delta`를 함께 기록한다. 해당 fill의 `filled_delta = max(transfer_delta, actual_delta)`이고 `transfer_delta`만 HELD에서 차감한다. actual이 더 크면 usage는 그 차이만큼 증가하며, 실제 가격이 낮다는 이유로 transfer보다 낮춰 여유를 만들 수 없다.

실제 fill price, fee 또는 FX provenance가 unknown이면 fill, watermark와 Position apply를 거부하거나 rollback하지 않는다. 계산 가능한 proportional HELD amount는 provisional filled floor로 이동하되 actual amount를 0으로 간주하지 않고 durable `UNKNOWN_ACTUAL_RISK`를 모든 적용 bucket/owner에 latch해 evidence가 authoritative하게 보완될 때까지 신규 exposure를 차단한다. actual이 나중에 확정되면 같은 fill identity에서 filled amount를 `max(transfer, actual)`로 단조 보완하고 이미 반영된 fill을 다시 적용하지 않는다.

각 transaction은 `filled + remaining HELD - monetary limit`의 positive delta를 horizon, market, strategy, sector와 symbol 각각에 overage로 저장한다. 어느 cap을 초과해도 fill/Position과 risk-reducing 경로는 보존하고 durable `RISK_OVERAGE` latch로 신규 exposure만 차단한다. partial, replacement 및 predecessor late fill도 같은 rule을 사용한다. retry는 fill identity/cumulative watermark로 delta 0이고 crash는 Position, HELD transfer, filled amount, all-bucket overage/latch를 모두 commit하거나 모두 rollback한다. cancel/expiry는 미체결 held 잔량에 대응하는 금액만 release한다. event replay는 owner, monetary reservation, actual provenance, overage와 usage snapshot을 결정적으로 재구성하며 불일치는 entry를 차단한다.

### D6. Owner release는 이전 generation의 protection/sell claim clean을 요구한다

owner release는 Position generation CLOSED/수량 0, pending exposure-raising mutation 및 HELD reservation 부재, reconciliation의 broker 수량 0 확인뿐 아니라 이전 generation에 귀속된 active/pending broker protection order, protection replace/recovery saga, pending sell/reduce-only claim, sell mutation attempt와 unresolved fill observation이 모두 없다는 journal/broker attestation을 요구한다. 하나라도 unknown, stale 또는 남아 있으면 owner를 유지하고 새 generation entry를 차단한다. release가 protection이나 sell claim을 자동 취소·삭제해서는 안 된다.

### D7. 손실 lock과 bucket failure는 entry admission에만 적용한다

daily/horizon loss lock과 bucket snapshot 장애는 EXPOSURE_RAISING decision/leg만 차단한다. RISK_REDUCING, stop, emergency exit, reconciliation과 fill observation은 bucket 계산, owner 획득 또는 FX 수집을 기다리지 않고 기존 경로로 진행한다.

별도 entry-only port를 둬 위험 감소 호출자가 실수로 bucket admission을 거칠 수 없게 한다.

### D8. 완화(해제)는 사람 승인·audit 된 별도 기록이고, 자동은 조이기만 한다 (5.5, 사용자 결정 2026-09-28)

사용자가 승인한 원칙:
- 자동 경로는 조이기만 한다.
- 완화와 해제는 actor `OPERATOR`만 할 수 있다. 사람 승인 참조 문자열이 있어야 하고, Auditor audit 줄이 commit **앞**에 쓰여야 한다.
- 동시에 들어온 조이기는 보수 쪽이 이긴다.
- 진입점은 journal API와 tossctl `mutating: true` 명령뿐이다(명령은 엔진 제어 endpoint 를 거친다 — 아래 "경로"). 대화형 에이전트는 이 명령을 자동 실행하지 않는다. 콘솔 버튼은 두지 않는다.

이 change 는 메커니즘까지만 만든다. 실제 해제는 별도의 사람 행위이고, 운영 원장에서는 실행하지 않는다.
형태는 `TransitionOperatingMode`(operating_mode.go:346–470)를 따른다.
- AUTO 는 완화를 요청할 수 없다.
- 방향은 트랜잭션 안에서 현재 상태에 대해 판정한다.
- approval 과 Auditor 가 둘 다 있어야 한다.
- audit 는 commit 앞에 쓴다.

**진입 손실 잠금 해제.**
- v35 는 additive 다. v33 파일은 커밋된 기록이므로 바꾸지 않는다.
  - `risk_bucket_entry_loss_lock_events`(append-only)를 둔다. 잠금이 열려 있을 때 들어온 활성화는 `REAFFIRM` 행(원인, 시각)으로 남긴다. 예전에는 아무것도 쓰지 않았다.
  - `risk_bucket_entry_loss_lock_releases`(append-only)를 둔다. `lock_seq UNIQUE`, actor `OPERATOR`, approval, reason, released_at 를 갖는다.
  - v33 의 first-cause-wins 트리거는 v35 migration 안에서 DROP/CREATE 로 "열린 잠금 1건" 트리거로 바꾼다. 해제 기록이 없는 잠금이 있으면 같은 범위의 새 잠금 INSERT 를 거절한다.
- 효력 있는 잠금은 해제 기록이 없는 잠금이다. 범위(account×market×horizon)마다 열린 잠금은 최대 한 건이다.
- 해제 요청은 두 값을 결속한다.
  - `LockSeq`: 해제할 잠금.
  - `ExpectedLastEvent`: 운영자가 본 그 잠금의 마지막 REAFFIRM seq. REAFFIRM 이 없으면 0.
- 트랜잭션 안에서 다음 중 하나라도 맞지 않으면 stale 로 거절한다.
  - 그 잠금이 열려 있다.
  - 그 범위의 열린 잠금이 바로 그 잠금이다.
  - 마지막 REAFFIRM seq 가 기대값과 같다.
- 그래서 운영자가 본 뒤에 새로 들어온 조이기가 해제를 이긴다(동시 조이기는 보수 쪽 승리).
- 해제 뒤의 활성화는 새 잠금을 연다.

**RISK_OVERAGE latch 해제.**
- 해제 대상은 owner generation 하나다. 그 owner 의 `risk_bucket_owners.risk_overage_latched` 와, 같은 generation 의 `risk_bucket_reservations.risk_overage_latched` 를 0 으로 되돌린다.
- 해제하지 않는 것:
  - `overage_minor` 수치. 이력으로 남는다.
  - `UNKNOWN_ACTUAL_RISK`. 실제 증거가 완성되면 스스로 풀린다(`clearResolvedUnknownLatches`).
  - scope latch 행.
- 요청은 `ExpectedStateDigest` 를 결속한다. 운영자가 읽기 전용 `risk-latch-show` 로 본 마지막 상태 봉인 digest 다.
  - 사이에 체결이나 latch 가 끼면 봉인이 다시 찍혀 digest 가 바뀌고, 해제는 stale 로 거절된다.
  - 이것이 ABA 방지다. 결속하는 값이 판정한 바이트다.
- 같은 트랜잭션 안에서 세 가지를 한다.
  - append-only 해제 기록(`risk_bucket_latch_releases`)을 쓴다.
  - 상태를 다시 봉인한다(`recordRiskBucketStateTx`).
  - commit 앞에 audit 를 쓴다.
- 해제 직후 다음 체결이 여전히 한도를 넘으면 `recomputeOverageLatches` 가 다시 latch 한다(보수).
- **순서.** latch 해제(운영자 승인)가 먼저이고, 기존 owner 해제(`releaseRiskBucketOwner`: broker zero 등의 검사)가 그 뒤다. owner 해제의 `owner_latch` 검사(owner.go:930)는 latch 해제 뒤에야 열린다.

**engine lock 을 잡지 않는 이유.**
- 해제는 journal 의 BEGIN IMMEDIATE 로 직렬화된다. 판정도 트랜잭션 안에서 현재 상태에 대해 다시 한다. 직렬성은 이것으로 충분하다.
- engine lock 을 잡으려면 엔진을 멈춰야 한다. 엔진이 멈추면 보호가 UNWIRED 인 동안 손절이 없다.
- 해제가 손절의 연속성을 깨는 조건이 되면 안 된다. 그래서 reconcile-resolve 선례와 달리 lock 을 잡지 않는다.

**경로: 엔진 제어 endpoint (Manager 판정 2026-09-29, a092 완화 명령 가족 계약).**
- 원장은 단일 writer(엔진)다(journal.go `DefaultBusyTimeout` 주석, engine.go "두 프로세스가 한 DB 를 마이그레이션"). 첫 구현
  (bce793a7)의 CLI 는 `journal.Open` 으로 원장을 직접 썼다. `journal.Open` 은 마이그레이션을 하므로, 새 CLI 바이너리가 도는
  엔진 밑에서 스키마를 올릴 수 있었다. 계약 대조에서 드러나 고쳤다.
- tossctl 해제 명령은 엔진이 발행한 position-policy 제어 endpoint(같은 listener · bearer 토큰 · private descriptor)의 두 route
  (`/v1/risk-relaxation/entry-lock-release`, `/v1/risk-relaxation/risk-latch-release`)에 요청한다. a079 격리 해제처럼
  `PositionPolicyCommandService` 의 선택 capability 로 발견되므로, capability 없는 빌드는 route 집합이 그대로다.
- 엔진 프로세스가 자기 journal 핸들과 자기 audit 로그(`Context.Audit`)로 journal API 를 부른다. audit 로그가 nil 이면 거절한다.
  `(*audit.Log)(nil).RecordAction` 이 nil 을 돌려주기 때문이다.
- 엔진이 돌지 않으면 CLI 는 거절한다(원장을 대신 열지 않는다). 사유는 "엔진이 없으면 진입도 없다"이고, 메시지는 엔진 기동 후
  재시도를 말한다. 엔진은 latch 가 걸린 채로 안전하게 기동한다.
- **통지.** 커밋 **뒤** 엔진이 원장 alert(`engine.risk_relaxation`, outbox = critical 전용)를 enqueue 한다. 담는 것은 대상 ·
  운영자 · 승인 참조 · 해제 번호다. enqueue 가 실패해도 해제는 유효하다. 결과가 `Notified=false` 를 싣고, CLI 는
  「완화됨·통지 실패」로 0 이 아닌 코드로 끝난다.
- 읽기 전용 `risk-latch-show` 는 `journal.OpenReadOnly`(mode=ro, query_only)로 직접 읽는다. writer 가 아니다.
- 응답을 못 읽은 경우는 "결과 불명"으로 말한다. 재시도는 안전하다. 이미 풀린 대상의 해제는 결속이 stale 로 거절한다.
- 아무것도 바꾸지 않은 journal 거절은 이름 있는 거절로 건넌다(리뷰 R1 P2 · R2 P3). 봉인 불일치(`ErrRiskBucketReplayMismatch`,
  봉인 없이 원장이 움직인 owner)는 `state_mismatch`, audit 쓰기 실패(`ErrRiskRelaxationAuditFailed`)는 `audit_unavailable` 이다.
  CLI 는 둘 다 "거절, 아무것도 안 풀림"으로 말한다. `internal`(결과 불명)은 이름 없는 실패에만 남는다.
- 외부 전송으로 나가는 통지 제목·본문에는 계좌 식별자를 싣지 않는다(안전 불변식 8, 리뷰 R2 P2). 대상은 시장·범위로 말하고
  (`entry_loss_lock:KR/SHORT`, `risk_owner:US/AAPL/<generation>`), 계좌를 포함한 전체 대상은 원장 payload 와 CLI 결과에만 둔다.
  같은 §8 문제를 a090 r2 가 다른 자리에서 찾았다. 사용자 큐의 해당 항목과 같은 가족이다(Manager 2026-09-29).
- **통지 적재는 a092 기록 입구 밖의 직접 기록자다 — 명명된 잔여(Manager 판정 2026-09-29 Q5 (a)).** a092 델타(engine-safety)는
  "입구를 거치지 않고 원장에 직접 쓰는 기록자는 … 행을 넣기 **전에** 자기 진입 차단 사유를 세워야 한다(SHALL)"라고 쓰고,
  정보성 critical 예외를 두지 않는다. 완화 통지가 자기 진입 차단 사유를 세우는 것은 의미가 맞지 않는다. 그래서 이행으로 규범을
  충족한다: a092 가 알림기의 기록 전용 입구(`n.mu` 아래 `RecordAlert`, 기록자별 `remindAfter`, 0 허용)를 착지시키면
  `notifyRelaxation` 의 호출 한 자리를 그리로 옮긴다(task 5.5.5). 그때까지는 `journal.EnqueueAlert` 직접 적재다. a092 쪽
  census 가 이 자리를 세고 있어 양방향 교차가 성립한다.
  a092 r23 판정(Manager 승인)으로 이것이 **유일 경로**가 됐다: 완화 통지에는 세울 자기 차단 사유가 없어 「먼저 잠금」 자체가
  불가능하다. 상호 조건: a092 는 "a066 이행 커밋 인용"을 자기 archive 선행 조건으로 갖고, a066 의 task 5.5.5 는 a092
  `RecordAlert` 착지를 선행 조건으로 갖는다.

**audit 줄은 시도의 기록이다 (Manager 판정 2026-09-29 Q6 (i)).**
- commit 앞 audit 줄의 value 는 `release_attempt` 다. 완료가 아니라 의도의 기록이라 롤백돼도 거짓이 아니다. "audit 가 commit 앞"
  원칙은 이 시도 줄이 충족하고, 결과는 원장 상태가 말한다.
- 시도 줄 뒤에 commit 이 실패하면(요청 문맥 끊김 등) 같은 action·setting 으로 `not_committed` 보상 줄을 쓴다. audit 쓰기는 요청
  문맥을 받지 않으므로 클라이언트가 끊겨도 기록된다(리뷰 R2 probe 2 의 "released 인데 롤백" 거짓 audit 를 닫음).
- 비용(CX-2 · R2 P2 · R1/R3 P3, 명명된 잔여): 해제는 audit fsync 동안 엔진의 단일 journal 연결을 쥔다. 실측(사본, ext4, 50회,
  2026-09-29): audit `RecordAction` 중앙값 4.6ms · 해제 전체 9.1ms(p95 10.7ms) · 활성화 커밋 4.8ms. fsync 1회 규모로 유계이고,
  audit 파일은 journal 과 같은 디스크라 같은 실패 영역이다. 선례 `TransitionOperatingMode` · reservation release 도 같은 형태다.
  트랜잭션은 요청 문맥으로 돌아 클라이언트 절단(5s)이 연결 점유의 상한 노릇을 한다.

**audit.** action 두 종을 둔다(동결 census 에 등록한다).
- `risk_bucket.entry_lock_release`: setting 은 `entry_loss_lock:<account>/<market>/<horizon>`.
- `risk_bucket.overage_latch_release`: setting 은 `risk_owner:<account>/<market>/<symbol>/<generation>`.
- Auditor 가 실패하면 아무것도 바뀌지 않는다.

## Risks / Trade-offs

- [다차원 예약 deadlock] → 정규화된 bucket key 정렬 후 하나의 journal transaction에서 획득한다.
- [price/fee/FX/strategy/sector evidence stale] → 진입만 fail closed하고 risk-reducing 경로는 독립시킨다.
- [partial/replacement/late fill 이중 계상 또는 cap 초과] → fill identity watermark와 tx-scoped proportional transfer/actual max를 사용하고 overage·unknown은 fill을 버리지 않고 신규 entry latch로 보존한다.
- [owner가 영구 잔존] → CLOSED generation과 protection/sell claim clean을 증명하는 idempotent release를 제공하며 unknown이면 보수적으로 유지한다.
- [보수 cap으로 기회 감소] → 안전한 의도이며 완화는 사람 승인·audit 된 해제 기록으로만 수행한다(D8).

## Migration Plan

1. monetary bucket policy, owner, reservation과 event 구조를 additive migration으로 추가한다.
2. reservation/cap 계산기를 pure function으로 구현해 price, nonlinear fee, FX haircut, minor-unit ceil, boundary, overflow, stale/missing과 `q_final <= q_candidate` property를 검증한다.
3. journal 원자 예약·owner race·partial/replacement/predecessor-late fill, actual price/fee/FX unknown, all-bucket overage와 restart/crash replay를 fixture로 검증한다.
4. Guardian에 shadow evaluation을 연결해 기존 승인 수량과 cap 차이를 기록하되 주문 수량은 변경하지 않는다.
5. 후속 runtime change에서 보호 readiness와 사람 승인 조건을 충족한 뒤에만 authoritative entry admission으로 연결한다.

Rollback은 authoritative binding을 OFF로 유지/제거하고 기존 Guardian 경로를 보존한다. 신규 스키마를 구버전이 열면 ErrSchemaTooNew로 fail closed하며 owner/예약 행을 임의 삭제하지 않는다.

## Open Questions

- KRW account에서 US exposure를 평가할 공식 FX source, haircut과 freshness budget은 authoritative binding 전 별도 운영 설정으로 확정해야 한다. 확정 전 US exposure-raising `q_final`은 0으로 fail closed한다.
