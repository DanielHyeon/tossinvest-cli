# KR/US Multi-Market Strategy Lanes Implementation Plan

Date: 2026-08-03
Status: Proposal freeze approved — implementation may start with RED tasks
Feature: FEAT-TOS-013
Scope: KR and US strategy evidence, campaign/leg state, risk buckets, six market-specific lanes, horizon routing, protection readiness, runtime wiring, and operations

## Objective

`strategy-runtime`을 실제 전략 실행 경로에 연결하되 한국 시장을 먼저 운영한 뒤 미국 시장을 시작하는 직렬 계획으로 만들지 않는다. 공통 계약을 먼저 고정하고, 각 전략군에서 KR·US 전용 레인을 같은 개발 웨이브로 구현·검증한다. 운영 활성화는 구현 순서와 분리하며 시장별로 독립적인 사람 승인을 요구한다.

## Non-Negotiable Invariants

- KR 안정화는 US 설계·구현의 선행 조건이 아니다.
- KR과 US는 캘린더, 증거, 레인 상태, 위험 예산, 실행 소유권, 활성화 상태를 독립적으로 가진다.
- LIVE 주문, 자동매매 토글, 운영 레인 활성화는 이 변경 작업에서 자동 실행하지 않는다.
- 토글 OFF는 신규 진입을 upstream에서 차단하되 기존 포지션의 손절·비상 청산 경로는 보존한다.
- 주문은 공식 Open API 게이트웨이만 사용하며 `Guardian`과 보호 주문 준비 상태를 우회하지 않는다.
- 모든 외부 증거는 point-in-time과 출처를 보존하고, 필수 증거가 없거나 오래되면 신규 진입을 fail closed 한다.
- 한 포지션/캠페인의 주문 소유자는 언제나 하나이며 EXIT가 ENTRY보다 우선한다.

## Delivery Graph

```text
Wave 0: SDD 문서 동결
  a064..a073 proposal/spec/design/tasks + PM tracker + independent review
                         |
Wave 1A: 공통 계약·RED·순수 계산기 (병렬)
  a064 evidence  ||  a065 campaign/leg  ||  a066 risk buckets
           \              |                /
            +-------------+---------------+
                          |
Wave 1B: authoritative integration (의존 순서)
  a064 evidence.db + consumed-snapshot contract
       || a065 prospective generation / campaign journal migration
                          |
                 a066 final risk transaction
                          |
Wave 2: KR/US 전략군과 라우터 (동시)
  a067 KR+US continuation || a068 KR+US reversal
  a069 KR+US weekly value || a070 multi-market horizon router
                          |
Wave 3: 실행 안전성과 런타임 (계약 병렬, 통합 순차)
  a071 KR+US protection readiness || a072 KR+US runtime wiring
                          |
Wave 4: 운영 표면
  a073 console/API/Compose dormant deployment
```

Wave 1의 병렬 범위는 계약, RED fixture, adapter/pure calculator다. a064는 별도
`evidence.db`를 사용하고 trading journal에는 소비된 snapshot ID/digest만 기록한다. a065와
a066의 authoritative journal migration version은 구현 전에 연속 번호로 예약하며, a066의
통합 transaction은 a064 snapshot contract와 a065 campaign identity가 들어온 뒤 연결한다.

Wave 2의 각 변경은 하나의 시장을 완료한 뒤 다른 시장으로 넘어가지 않는다. 예를 들어 a067에서는 KR continuation과 US continuation을 나란히 RED/GREEN/REFACTOR/VERIFY하며, 두 시장 계약과 테스트가 모두 충족되어야 변경이 완료된다.

## Change Sequence

### Wave 0 — Documentation and Proposal Freeze

1. `a064-add-multi-market-strategy-evidence`
2. `a065-add-position-campaign-leg-core`
3. `a066-add-multi-horizon-risk-buckets`
4. `a067-add-kr-us-continuation-lanes`
5. `a068-add-kr-us-reversal-lanes`
6. `a069-add-kr-us-weekly-value-lanes`
7. `a070-add-multi-market-horizon-router`
8. `a071-wire-kr-us-protection-readiness`
9. `a072-wire-multi-market-strategy-runtime`
10. `a073-operate-multi-market-strategy-lanes`

각 change에 proposal, delta spec, design, tasks, PM story를 작성한다. a065, a066, a071, a072는 high-risk 변경으로 adversarial engineering review를 포함한다. 모든 proposal이 동결되고 `openspec validate --all --strict --no-interactive`가 통과하기 전에는 production 코드를 변경하지 않는다.

### Wave 1 — Common Contracts in Parallel

#### a064 Multi-Market Strategy Evidence

- KR/US 가격·거래량·시장 참여·재무 증거의 공통 envelope를 정의한다.
- KR은 KRX/OpenDART, US는 공식 market data/SEC EDGAR 경계를 사용한다.
- 출처, 관측 시각, 유효 시각, 수집 시각, freshness, revision을 저장한다.
- historical 조회는 `source_available_at <= evaluation_at`과
  `ingested_at <= ingestion_cutoff`을 동시에 만족해야 한다.
- 대량 evidence는 append-only `evidence.db`에 저장하고 single-writer trading journal에는
  실제 decision이 소비한 immutable snapshot ID/digest만 기록한다.
- 필수 증거 누락과 lane-scoring 결측을 구분한다.

#### a065 Position Campaign and Leg Core

- 시장·전략과 독립적인 campaign/leg 상태 기계를 만든다.
- 8:4:2, 2:4:8, 주 1회 최대 7회 같은 레인별 배분 정책을 주입 가능하게 한다.
- 재시작 후 ledger로부터 상태를 결정론적으로 복원한다.
- 중복 intent와 중복 fill을 멱등 처리하고 stop은 후퇴시키지 않는다.
- prospective position generation을 journal CAS로 예약하고 leg/order-attempt별 cumulative
  watermark와 replacement/late-fill lineage를 합산한다.
- 교체 전 predecessor의 늦은 fill도 immutable broker-order watermark와 Position에 정확히 한
  번 반영한다. 수량 cap 초과나 lineage ambiguity는 fill을 버리지 않고 campaign을
  `RECONCILE`로 보내 신규 exposure만 차단한다.

#### a066 Multi-Horizon Risk Buckets

- market, horizon, strategy, sector, symbol 수준의 예산을 결합한다.
- 최종 주문 수량은 언제나 upstream 후보 수량 이하이다.
- 신규 진입 제한은 손절·비상 청산을 차단하지 않는다.
- KR과 US 예산 고갈 및 오류가 서로 전파되지 않게 한다.
- `q_candidate`에서 모든 bucket cap을 적용해 `q_final`을 확정한 뒤에만 final
  RiskIntent/GuardianDecision과 reservations를 한 transaction으로 발급한다.
- 실제 fill의 가격·수수료·FX가 conservative HELD transfer보다 크면 차액을 모든 적용
  bucket의 durable overage로 기록한다. 실제 위험을 확정할 수 없어도 fill/Position은 보존하고
  `RISK_OVERAGE`/`UNKNOWN_ACTUAL_RISK` latch로 신규 exposure만 차단한다.

### Wave 2 — KR and US Lanes in the Same Wave

#### a067 Continuation Lanes

- KR flow continuation 레인과 US participation continuation 레인을 동시에 구현한다.
- 두 레인은 공통 campaign/leg 계약을 사용하되 신호 증거와 lineage를 공유하지 않는다.
- 양 시장 모두 8:4:2 배분과 재시작/중복 이벤트 테스트를 가진다.

#### a068 Reversal Lanes

- KR absorption reversal 레인과 US dislocation reversal 레인을 동시에 구현한다.
- 양 시장 모두 2:4:8 배분을 사용한다.
- 마지막 leg는 sweep, market-structure break, reclaim을 모두 확인해야 한다.
- lane은 typed invalidation/refusal만 방출하고, 공통 exit engine만 청산 decision 권한을
  가진다. 공통 exit가 신규 진입/추가 매수보다 먼저 평가되는지 검증한다.

#### a069 Weekly Value Lanes

- KR OpenDART weekly value 레인과 US EDGAR weekly value 레인을 동시에 구현한다.
- 두 레인은 최대 7개 leg, 주 1회 한도, 매 leg 재검증, stop 한도를 독립 적용한다.
- 정정 공시와 point-in-time 재무 데이터가 미래 정보를 유출하지 않게 한다.

#### a070 Multi-Market Horizon Router

- KR/US 캘린더와 DST를 독립 처리한다.
- 한 market-symbol-position에 하나의 실행 owner만 할당한다.
- owner key는 `(account, market, canonical symbol, position_generation)`이며 horizon은
  eligibility/attribution일 뿐 ownership key가 아니다.
- 시장별 rate limit, 오류, 재시도, 활성화 상태를 격리한다.
- market/horizon capability는 admission/anti-replay subscope이고 실제 호출량은 provider의
  physical endpoint/reset-generation quota 하나에서 차감해 quota를 복제하지 않는다.
- 한 시장의 휴장·장애가 다른 시장의 레인 평가를 막지 않는지 검증한다.

### Wave 3 — Protection and Runtime Integration

#### a071 KR/US Protection Readiness

- 신뢰 root, key rotation/revocation, monotonic serial과 exact broker identity/query 의미까지
  서명된 attestation이 증명할 때만 market readiness를 `WIRED`로 만든다.
- 엔진 중단 후에도 보호 주문이 유효한지 확인하고, 교체·복구를 멱등화한다.
- 보호 준비가 `UNWIRED`이거나 typed refusal이면 신규 진입을 차단하고 청산은 허용한다.
- 어떤 테스트나 배포 단계도 LIVE 주문 또는 운영 토글을 자동 실행하지 않는다.

#### a072 Multi-Market Strategy Runtime

- evidence → scheduler → router → lane → risk → Guardian → dispatch → official gateway 경로를 supervised loop로 연결한다.
- KR/US loop를 같은 런타임 릴리스에 포함하되 lease, cursor, backoff, health를 시장별로 분리한다.
- OFF 상태에서 신규 진입이 생성되지 않고 기존 청산/보호 경로가 유지되는지 검증한다.
- 복수 replica에서 lease ownership evidence로 중복 실행을 차단한다.
- dispatch lease는 claim/검증 시 성공 또는 거부 terminal state로 원자 전이하며 권한 digest가
  A→B→A로 복구돼도 재활성화되지 않는다. durable owner epoch/fencing token과
  `ISSUED→CLAIMED→SUBMITTING→SUBMITTED|AMBIGUOUS|REFUSED` 상태를 사용한다.

### Wave 4 — Operations and Dormant Deployment

#### a073 Console, API, and Compose

- console/API가 KR/US 레인 상태, readiness, freshness, lineage, 성과를 독립 노출한다.
- Compose build/deploy 시 레인은 OFF/default 상태를 유지한다.
- 배포 전 image digest/schema/config/activation preimage를 고정하고 부분 replace 첫 실패 시
  이미 교체된 subset까지 pinned image로 되돌린다.
- 배포 후 health/readiness와 dormant no-entry 동작을 확인한다.
- 운영 활성화는 별도의 사람 승인 절차로 남긴다.

## Test Strategy

각 change는 기존 함수 내부 로직을 수정하기 전에 Function Logic Map과 Branch Test Map을 작성한다. 구현은 RED → GREEN → REFACTOR → VERIFY 순서로 진행한다.

- unit: evidence freshness, signal purity, leg allocation, risk intersection, routing ownership
- property/fuzz: idempotence, non-retreating stops, `q_final <= q_candidate`, no duplicate owner
- integration: restart reconstruction, independent KR/US failure domains, OFF-state exit preservation
- contract: official data adapters, API schemas, runtime descriptors, protection attestations
- UI/API: market별 상태와 lineage가 혼합되지 않는지 확인
- deploy: `docker compose config`, build, dormant up, health/readiness, no live mutation

## SDD and Release Gates

각 change 또는 합의된 batch마다 다음을 수행한다.

1. `make sdd-sync`
2. `make sdd-check`
3. `make gate CHANGE=<change-id>`
4. relevant Go/unit/integration tests
5. independent code/security/architecture review
6. PM tracker regeneration and OpenSpec status update

모든 change가 검증된 뒤 feature branch를 커밋·push하고 main에 non-destructive merge한다. main에서 전체 테스트와 Compose build를 다시 실행한 뒤 push한다. 배포는 OFF/default 설정으로 수행하고 post-deploy dormant checks를 통과해야 한다.

## External Configuration and Known Boundaries

- OpenDART 수집은 사용자 제공 API key가 필요하다. key가 없으면 해당 KR weekly value 신규 진입은 fail closed 하며 secret을 코드·로그·커밋에 기록하지 않는다.
- SEC EDGAR 호출은 공식 access policy와 식별 가능한 User-Agent를 준수한다.
- 외부 데이터 장애는 해당 증거/시장/레인으로 격리한다.
- 코드 배포 완료와 자동매매 운영 승인 완료는 서로 다른 상태다.

## Expected Outcome

동일 릴리스에서 KR과 US의 continuation, reversal, weekly-value 레인이 각각 독립된 lineage와 위험 예산으로 연결된다. 구현과 dormant 배포는 병렬 시장 지원을 완성하지만, LIVE 주문과 운영 토글은 명시적인 사람 승인 전까지 비활성 상태로 남는다.

## GSTACK REVIEW REPORT

| Review | Trigger | Why | Runs | Status | Findings |
|--------|---------|-----|------|--------|----------|
| CEO Review | `/plan-ceo-review` | Scope & strategy | 1 | CLEAR | Modular ten-change plan selected; KR/US same-wave premise confirmed |
| Codex Review | `/codex review` | Independent 2nd opinion | 0 | NOT RUN | Three independent repository-agent voices used instead |
| Eng Review | `/plan-eng-review` | Architecture & tests (required) | 3 | CLEAR | Initial blockers plus two semantic rounds folded; no remaining critical gap |
| Design Review | `/plan-design-review` | UI/UX gaps | 1 | CLEAR | Existing read-only runtime pattern reused; all partial/unavailable/dormant states specified |
| DX Review | `/plan-devex-review` | Developer/operator experience gaps | 1 | CLEAR | Typed first refusal, shared schema and bounded dormant deploy verification specified |

**CROSS-MODEL:** Manager scope review plus three disjoint author/reviewer voices agreed on the final contracts after two adversarial correction rounds.

**VERDICT:** CEO + ENG + DESIGN + DX CLEARED — ready to implement RED tasks; deployment remains dormant/OFF.

NO UNRESOLVED DECISIONS
