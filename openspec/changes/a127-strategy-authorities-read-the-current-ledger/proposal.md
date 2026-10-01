# a127 · 전략 권한 적재기는 지금 원장을 읽는다

- **Feature**: `FEAT-TOS-013` — Concurrent KR and US multi-horizon strategy lanes
- **Story**: `STORY-TOS-a127`
- **Spec**: `strategy-runtime`(ADDED 1)
- **위험 등급**: **High-risk**(위험 권한 · 진입 경로의 적재기)

> **출처**: a112 태스크 5.2.2.2 실측(2026-10-01) — ROADMAP 「a112 이월」 R1, 영수증
> `openspec/changes/a112-run-four-strategy-families-independently/analysis/measurements/lot-5.2.2.2/schema-pin-receipt.log`.
> Manager 배정(2026-10-01): 레인 활성화의 경성 선행. 착수 실측이 같은 결함의 **둘째 자리**(route 적재기)를 찾았고 Manager 가 한 change 로
> 묶었다(판정 2026-10-01 — 같은 결함 · 같은 수리 모양 · 같은 진입점, route 가 risk 보다 앞이라 하나만 고치면 활성화가 여전히 0).

## Why

엔진의 생산 전략 권한 적재기 둘이 원장 스키마를 **동결 리터럴 27** 과 정확 일치로 비교한다:

| 적재기 | 핀 | 비교 |
|---|---|---|
| 위험 snapshot 권한 `riskbucket.LoadProductionRiskSnapshotAuthority` | `internal/riskbucket/production_snapshot_authority.go:33` `productionRiskJournalSchema = 27` | `:375` `version != productionRiskJournalSchema` → `:376` 거절 |
| route 권한 `strategyrouter.LoadProductionRouteAuthorityBatch` | `internal/strategyrouter/production.go:38` `productionRouteJournalV = 27` | `:609` 같은 비교 → `:614` 거절 |

원장은 v35 다(`internal/journal/schema.go:6`, 2026-08-05 a084 부터 28 이상). supervisor 는 두 적재기에 엔진이 연 **실제 원장 경로**를 넘긴다
(`internal/app/engine/strategy_entry_supervisor.go:300-305` · `:328`). 그래서 **생산 route 권한과 위험 권한은 어느 범위에서도 ready 가 될 수
없다** — 서명 활성화를 해도 전 범위 `AuthorityUnavailable`, 1차 레그 0. 오늘은 생산 서명 활성화가 0 이라 동작 변화가 없다.

**왜 여태 안 보였나 — 가린 픽스처.** 두 적재기의 시험은 전부 `PRAGMA user_version=27` 로 만든 **축소 원장**을 읽었다
(`riskbucket/production_snapshot_authority_test.go:219`, `strategyrouter/production_test.go:216`, `app/engine/strategy_risk_authority_test.go:224`).
픽스처가 핀과 같은 값을 써서 결함이 자기 시험을 통과했다. a112 는 실제 원장 흐름을 세우다 위험 적재기 쪽만 마주쳤고, 시험 다리
`a112MirrorLedgerIntoRiskStub`(원장 행을 v27 stub 로 복사)와 트립와이어 `TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal`
로 표시해 두었다. route 쪽은 a126 · a112 문서 어디에도 기록이 없었다(grep 0) — 착수 실측이 찾았다. a126 의 snapshot digest replay 시험도
같은 이유로 v27 축소 원장에서 생성기를 돌렸다(a126 review R3 절).

**전수(census).** journal 패키지 밖에서 실제 원장을 읽기 전용(`mode=ro` · `query_only`)으로 여는 비시험 파일은 6 개이고, 엔진 원장을 여는
것은 이 둘뿐이다(나머지 넷 — candidate · performance · strategyevidence · journal 자신 — 은 자기 저장소). 원장 스키마를 리터럴로 고정한
비시험 상수도 이 둘뿐이다.

**구조가 결함이다.** 마이그레이션 규칙(`schema.go:8-21`)상 원장은 가산적으로만 자라고, 그때마다 리터럴 핀은 깨진다. 고칠 것은 값 27 이
아니라 「마이그레이션과 함께 움직이지 않는 핀」이다.

## What Changes

- 두 적재기의 스키마 조건을 **호출자가 주입한 현재 스키마(`journal.SchemaVersion`)와 정확 일치**로 바꾼다. 순환 import 때문에 상수는 engine 이
  넘긴다. 주입이 0 이하면 원장을 열기 전에 거절한다. 불일치는 방향(더 새 원장 · 더 옛 원장)을 문구로 말한다. 거절 신원
  (`ErrProductionRiskSnapshotUnavailable` · `ErrProductionRouteUnavailable`)은 그대로 — 범위 국소가 아니라 결함이다.
- 리터럴 상수 둘을 지운다.
- 수락 시험을 실제 원장으로: a112 트립와이어를 양성 시험으로 뒤집고 시험 다리를 지운다(a112 소유 파일 — Manager 가 a112 소유자에 통지).
  축소 픽스처 셋은 주입 값과 같은 `user_version` 을 쓰게 바꾼다(행 · 단언 불변).
- 판독 SQL · 사용량 판정 · owner 재구성 · 원장 스키마는 바꾸지 않는다.

## Impact

- 코드: `internal/riskbucket/production_snapshot_authority.go`, `internal/strategyrouter/production.go`, `internal/app/engine/strategy_risk_authority.go`,
  `internal/app/engine/strategy_route_authority.go` + 시험.
- 동작: 서명 활성화 0 인 오늘은 변화 0. 활성화 뒤에는 두 권한이 실제 원장에서 ready 가 될 수 있다(진입 판단 자체는 불변 — design D6).
- a112: 시험 다리 제거 · L6 의 형식상 선행(실제 원장 권한) 해소.
</content>
</invoke>
<invoke name="Bash">
<parameter name="command">cd /mnt/D/Axipient/workspace/TossOS/openspec/changes/a127-strategy-authorities-read-the-current-ledger && sed -i '1s/.*/# a127 design/' design.md && head -3 design.md && mkdir -p specs/strategy-runtime && cat > specs/strategy-runtime/spec.md <<'EOF'
# strategy-runtime — a127 delta

## ADDED Requirements

### Requirement: 생산 전략 권한 적재기는 엔진이 연 현재 원장 스키마만 읽는다

Engine 이 원장을 읽기 전용으로 여는 생산 전략 권한 적재기(위험 snapshot 권한 · route 권한)는 원장의 `user_version` 이 engine 이 주입한 현재
journal 스키마 버전과 정확히 같을 때만 원장을 판독해야 한다(SHALL). 판독 조건은 마이그레이션과 함께 움직이지 않는 고정 리터럴이어서는
안 되며(MUST NOT), 주입 값이 없거나 0 이하이면 원장을 열기 전에 거절해야 한다(SHALL). 더 새 원장과 더 옛 원장은 둘 다 거절하되 거절 사유가
방향을 구분해야 하고(SHALL), 거절은 범위 국소 거절이 아니라 그 시장 권한의 결함으로 남아야 한다(SHALL). 적재기의 수락은 engine 이 실제로 연
현재 원장으로 시험되어야 한다(SHALL).

#### Scenario: 엔진이 연 현재 원장

- **WHEN** engine 이 `journal.Open` 으로 연(마이그레이션된) 원장 경로와 현재 스키마 버전을 두 적재기에 넘긴다
- **THEN** 스키마 조건은 통과하고 적재기는 기존 판독 · 판정을 그대로 수행한다

#### Scenario: 이 빌드보다 새 원장

- **WHEN** 원장 `user_version` 이 주입된 현재 버전보다 크다
- **THEN** 적재기는 원장 행을 판독하지 않고 더 새 원장임을 말하는 사유로 거절하며 exposure-raising 요청은 0건이다

#### Scenario: 마이그레이션되지 않은 원장

- **WHEN** 원장 `user_version` 이 주입된 현재 버전보다 작다
- **THEN** 표와 열이 모두 있더라도 적재기는 더 옛 원장임을 말하는 사유로 거절하며 exposure-raising 요청은 0건이다

#### Scenario: 주입 누락

- **WHEN** 호출자가 현재 스키마 버전을 넘기지 않았거나 0 이하를 넘긴다
- **THEN** 적재기는 원장 파일을 열기 전에 거절한다

#### Scenario: 읽기 집합의 열이 없는 원장

- **WHEN** 버전은 같지만 적재기가 읽는 열 하나가 원장에 없다
- **THEN** 판독 질의가 실패하고 적재기는 fail-closed 로 거절한다
EOF
cd /mnt/D/Axipient/workspace/TossOS && openspec validate a127-strategy-authorities-read-the-current-ledger --strict 2>&1 | tail -3