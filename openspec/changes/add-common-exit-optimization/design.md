## Context

TossOS의 `internal/exitpolicy`에는 StockOS의 초기 BALANCED ladder와 별도의 R 단위 RATCHET이 이미 decimal 연산으로 이식돼 있다. 그러나 production `ExitObserver`는 새 포지션을 항상 RATCHET으로 열고, LADDER 상태에는 rung index만 저장할 뿐 정확한 policy ID를 저장하지 않는다. 외부 매수분은 현재가를 t0로 삼아 synthetic stop과 adoption record를 만들지만 역시 RATCHET으로 고정된다.

StockOS의 현재 공통 정책 선택 계약은 활성 포지션 스냅샷이 공통 기본보다 우선하고, `COMMON_LADDER_HYBRID_50`을 권장하되 사람 승인 전에는 기존 실행값을 바꾸지 않는다. HYBRID_50은 잔량 기준 부분익절로 원수량 약 50%를 남긴 뒤 high-water runner에 넘긴다. TossOS에는 ATR/EMA/VWAP 입력이 없으므로 runner의 공통 입력으로 이미 관측하는 `high_water`와 StockOS runtime 기본 `exit_trail_stop_pct=6.5`를 사용한다.

이 변경은 자동 매도 판단, journal schema, 외부 편입, 운영 설정과 웹 POST를 함께 건드리는 high-risk 변경이다. 기존 손절·비상 청산의 즉시성, baseline 단조성, proposal→journal→submit 순서, Guardian reduce-only 경로는 바뀌면 안 된다.

## Goals / Non-Goals

**Goals:**

- StockOS의 BALANCED/RUNNER/HYBRID_50 정책 수치를 등록하고 서버가 권위 목록을 소유한다.
- 공통 정책 미승인 상태는 현재 RATCHET 동작과 byte-equivalent하게 유지한다.
- 새로 관리되는 자체 진입·외부 매수 포지션에 정확한 policy ID를 영속 스냅샷한다.
- HYBRID_50의 T4 이후 잔량 보호선을 `max(previous, fixed rung floor, high_water × 0.935)`로 계산한다.
- 콘솔에서 정책의 rung, 부분익절 효과, trailing 의미와 현재값을 보고 사람이 승인한다.
- 정책 저장은 config의 다른 바이트, LIVE gate와 trading toggle을 변경하지 않고 audit에 남긴다.

**Non-Goals:**

- 기존 활성 포지션을 새 공통 정책으로 자동 rebind하지 않는다.
- 임의 rung 편집, per-symbol override, backtest/optimizer 실행, 정책 자동 추천 변경은 포함하지 않는다.
- ATR/EMA/VWAP를 새로 조회하거나 KIS 호출 수를 늘리지 않는다.
- 설정 저장만으로 엔진을 재시작하거나 주문을 만들거나 LIVE 권한을 켜지 않는다.
- 원격 접속/VPN/Compose는 이 변경의 gate 통과 뒤 별도 change로 수행한다.

## Decisions

### 1. 등록 정책은 코드의 immutable registry가 권위다

등록 ID와 수치는 다음으로 고정한다.

- `COMMON_LADDER_BALANCED`: targets 1.5/2.5/4.0/6.0, floors 0/1.0/2.0/3.5, remaining partials 0/0.25/0.25/1.0, final full.
- `COMMON_LADDER_RUNNER`: targets 2.5/4.5/7.0/999.0, floors 0/2.0/3.5/5.0, remaining partials 0/0.15/0/0, no fixed final full.
- `COMMON_LADDER_HYBRID_50`: targets 1.8/3.0/4.8/6.5, floors 0/1.2/2.5/3.8, remaining partials 0/0.25/1/3/0, no fixed final full.

HYBRID_50만 권장 표식을 갖는다. config에는 ID만 저장하고 rung 배열은 저장하지 않는다. 자유 형식 rung JSON은 오타·순서 역전·부분 설정이 곧 live exit geometry가 되므로 기각한다.

### 2. 기존 behavior 보존은 빈 공통값으로 표현한다

`engine.exit_policy.common_policy`가 없거나 빈 값이면 선택은 `RATCHET`이다. migration과 default config는 값을 채우지 않는다. 알 수 없는 non-empty ID는 `Rejected`로 보존하고 engine observer 구성을 거부한다. 조용한 RATCHET fallback은 운영자가 승인한 값을 무시하므로 기각한다.

### 3. 포지션별 policy ID를 additive schema v9에 저장한다

`exit_states.policy_id TEXT NULL`과 `position_adoptions.exit_policy_id TEXT NULL`을 추가한다.

- 새 자체 진입 포지션: exit state가 처음 열릴 때 process가 snapshot한 공통 선택을 `policy_kind`와 `policy_id`로 한 번 기록한다.
- 외부 매수분: adoption transaction에 선택 ID를 먼저 기록하고, crash recovery는 그 record에서 exit state를 연다.
- 기존 RATCHET row의 NULL은 RATCHET을 뜻한다.
- 기존 LADDER row의 NULL은 이 버전 이전에 존재한 유일한 ladder인 `default_v1`로 읽어 기존 의미를 보존한다. migration은 historical row를 rewrite하지 않는다.

공통값 변경으로 기존 row를 update하는 대안은 활성 rung을 다른 표로 재해석하므로 기각한다.

### 4. evaluator는 상태의 policy ID로 매번 registry를 해석한다

observer 전체에 하나의 ladder를 가정하지 않고 각 LADDER state의 `policy_id`로 registry를 조회한다. state가 알 수 없는 ID를 가지면 그 포지션은 판단을 보류하고 critical refusal을 남긴다. `checkLadderPolicyStillFits`는 선택된 policy를 인자로 받아 활성 rung 수와 저장 baseline을 교차검증한다.

테스트용 custom ladder 주입 seam은 기존 테스트 호환을 위해 policy ID가 일치할 때만 registry보다 우선한다. production config는 등록 ID만 사용한다.

### 5. runner 보호선은 기존 baseline max-composition에 한 후보로 추가한다

HYBRID_50이 마지막 rung에 도달한 뒤 각 관측에서:

`trail = high_water × (1 - 6.5 / 100)`

를 계산하고 기존 `ComputeProtectedStop` 후보에 넣는다. 결과는 언제나 이전 baseline 이상이다. 같은 tick에서 high-water가 오르고 현재가가 새 보호선 아래라면 승격된 보호선을 먼저 적용하고 breach를 판단한다. 이는 기존 evaluator의 “promotion before breach” 순서와 위험 축소 방향을 유지한다.

RUNNER는 StockOS 수치를 보존하지만 T4 sentinel 999%에 의존해 자동 final exit을 만들지 않는다. 이 change에서는 HYBRID_50 runner handoff만 high-water trailing을 활성화한다. 외부 매수분은 adoption 현재가가 entry/high-water t0이므로 과거 수익으로 즉시 rung이 발동하지 않는다.

StockOS A168 계약에 맞춰 외부 매수분이 RUNNER를 snapshot한 경우에는 같은 target/floor를 사용하되 모든 partial ratio를 0으로 해석한다. 즉 외부 RUNNER는 자동 부분익절 없이 보호선 승격과 breach 전량보호만 수행한다. 저장 ID는 파생 ID가 아니라 `COMMON_LADDER_RUNNER`를 유지하고, immutable registry policy를 position origin에 따라 안전하게 복제해 평가한다.

### 6. 최적화 화면은 별도 최소 capability seam을 쓴다

`/optimization` GET과 `/optimization/exit-policy` POST를 추가한다. POST는 기존 `session0(mutating(...))` 체인을 그대로 사용한다. console seam은 `Load() (config.ExitPolicy, error)`와 `Save(config.ExitPolicy) error`만 노출하며 broker, journal writer, gate writer를 전달하지 않는다.

설정 writer는 `engine.exit_policy` 값 span만 surgical replace/insert하고 같은 lock·atomic rename 계약을 사용한다. cmd wiring은 실제 파일 before-image를 읽은 뒤 audit에 old/new를 기록한다. 저장 성공 뒤 화면은 “다음 엔진 기동부터 반영, 기존 포지션 불변”을 표시한다.

### 7. 주문 안전 체인은 변경하지 않는다

policy evaluation의 Proposal은 기존 `record → arm → Guardian reduction intent → execution gateway` 경로만 사용한다. 최적화 GET/POST는 broker 호출을 소유하지 않는다. gate/trading 설정과 policy 설정은 타입과 seam을 분리해 한 POST가 다른 키를 운반할 수 없게 한다.

## Risks / Trade-offs

- [Risk] policy ID 없이 생성된 기존 LADDER row가 잘못 해석될 수 있음 → 이 버전 이전 production의 유일한 ladder `default_v1`로만 legacy mapping하고 테스트로 고정한다.
- [Risk] config 변경과 외부 adoption 사이 crash가 다른 정책을 적용할 수 있음 → adoption record에 policy ID를 같은 transaction으로 저장하고 recovery는 config가 아니라 record를 읽는다.
- [Risk] high-water trailing이 같은 tick에서 더 빠른 청산을 만들 수 있음 → 기존 promotion-before-breach 규칙과 단조성 property test로 위험 축소 방향만 허용한다.
- [Risk] 1주 포지션의 부분익절 rounding이 0이 될 수 있음 → 기존 proposal quantity 검증과 refusal 경로를 유지하며 state-only로 성공 처리하지 않는다.
- [Risk] 최적화 POST가 gate나 trading key를 덮어쓸 수 있음 → dedicated message type, byte-splice writer, outside-block byte preservation test를 둔다.
- [Risk] 설정 저장은 running observer에 즉시 반영되지 않음 → 화면에 다음 재기동 적용을 명시하고 observer는 시작 시 snapshot을 유지한다.
- [Risk] 외부 RUNNER가 자체 진입 RUNNER의 T2 15% 부분익절을 상속할 수 있음 → adopted origin에서는 partial ratio를 모두 0으로 만드는 StockOS A168 호환 평가와 전용 테스트를 둔다.

## Migration Plan

1. v9 migration으로 nullable 두 컬럼만 추가하고 pre-migration backup을 생성한다.
2. registry/config parser/writer를 배포하되 common policy 기본은 empty로 둔다.
3. policy-aware state opening과 external adoption snapshot을 배포한다.
4. evaluator의 per-state resolution과 HYBRID_50 runner test를 통과시킨다.
5. `/optimization` 화면과 audit wiring을 배포한다.
6. 운영자가 화면에서 수치를 검토하고 별도 클릭으로 HYBRID_50을 승인한다.
7. 엔진을 사람이 재시작한 뒤 신규 자체/외부 포지션의 policy snapshot과 audit을 관측한다.

Rollback은 pre-migration backup과 이전 binary로 한다. 이전 binary는 v9 journal을 거부하므로 backup 복원이 필요하다. config의 `engine.exit_policy`는 이전 binary가 모르는 키로 보존되며 기존 surgical writers가 블록 밖 바이트를 유지한다.

## Open Questions

없음. 공통값은 HYBRID_50 권장·사람 승인, trailing gap은 StockOS runtime 기본 6.5%, 기존 포지션 자동 rebind 없음으로 확정한다.
