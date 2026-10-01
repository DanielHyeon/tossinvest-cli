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
| route 권한 `strategyrouter.LoadProductionRouteAuthorityBatch` | `internal/strategyrouter/production.go:38` `productionRouteJournalV = 27` | `:609` 같은 비교 → `:612` 거절 |

원장은 v35 다(`internal/journal/schema.go:6`). 두 리터럴은 같은 커밋 `8022f578`(2026-08-04)에서 태어났는데 그때 `journal.SchemaVersion` 은 이미
29 였다 — **핀은 출하된 어느 스키마와도 맞은 적이 없다**(ROADMAP R1 행의 「2026-08-05 a084 부터 28 이상」은 부정확). supervisor 는 두 적재기에 엔진이 연 **실제 원장 경로**를 넘긴다
(`internal/app/engine/strategy_entry_supervisor.go:300-305` · `:328`). 그래서 **생산 route 권한과 위험 권한은 어느 범위에서도 ready 가 될 수
없다** — 서명 매니페스트와 활성화를 갖춰도 1차 레그 0. 오늘 생산에는 서명 매니페스트가 없다고 기록돼 있다(a112 8.7.1 실측, 2026-09-04 —
배포 전 재실측은 사람 항목 H1, design D6) — 동작 변화가 없다.

**왜 여태 안 보였나 — 가린 픽스처.** 두 적재기의 시험은 전부 `PRAGMA user_version=27` 로 만든 **축소 원장**을 읽었다
(`riskbucket/production_snapshot_authority_test.go:219`, `strategyrouter/production_test.go:216`, `app/engine/strategy_risk_authority_test.go:224`).
픽스처가 핀과 같은 값을 써서 결함이 자기 시험을 통과했다. a112 는 실제 원장 흐름을 세우다 위험 적재기 쪽만 마주쳤고, 시험 다리
`a112MirrorLedgerIntoRiskStub`(원장 행을 v27 stub 로 복사)와 트립와이어 `TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal`
로 표시해 두었다. route 쪽은 a126 · a112 문서 어디에도 기록이 없었다(grep 0) — 착수 실측이 찾았다. a126 의 snapshot digest replay 시험도
같은 이유로 v27 축소 원장에서 생성기를 돌렸다(a126 review R3 절).

**전수(census).** 세는 규칙: 비시험 `.go` 파일 중 DSN 에 `mode=ro` 또는 `query_only` 를 **코드로** 쓰는 것(주석만 걸리는
`cmd/tossctl/engine_risk_relaxation.go` 는 `journal.OpenReadOnly` 를 쓰므로 제외). 그런 파일은 6 개다 — `internal/journal/readonly.go`(엔진 원장의
엔진 밖 열람 — 콘솔 등, 범위 수락 `checkSchema`), candidate · performance · strategyevidence(각자 자기 저장소), 그리고 이 둘. 즉 **journal
패키지 밖에서 엔진 원장을 직접 읽기 전용으로 여는 비시험 파일은 이 둘뿐**이다. 원장 스키마를 리터럴로 고정한 비시험 상수도 이 둘뿐이다.

**구조가 결함이다.** 마이그레이션 규칙(`schema.go:8-21`)상 원장은 가산적으로만 자라고, 그때마다 리터럴 핀은 깨진다. 고칠 것은 값 27 이
아니라 「마이그레이션과 함께 움직이지 않는 핀」이다.

## What Changes

- 두 적재기의 스키마 조건을 **호출자가 주입한 현재 스키마(`journal.SchemaVersion`)와 정확 일치**로 바꾼다(Manager 재판정 — design D1). 순환 import 때문에 상수는 engine 이
  넘긴다. 주입이 0 이하면 원장을 열기 전에 거절한다. 불일치는 방향(더 새 원장 · 더 옛 원장)을 문구로 말한다. 거절 신원
  (`ErrProductionRiskSnapshotUnavailable` · `ErrProductionRouteUnavailable`)은 그대로 — 범위 국소가 아니라 결함이다.
- 리터럴 상수 둘을 지운다. route 의 방향 문구가 공개 경계(`LoadProductionRouteAuthorityBatch` `:352` 감싸기)까지 오도록 원인을 보존한다.
- 두 적재기의 판독을 각각 읽기 트랜잭션 하나로 묶고(risk 는 새로, route 는 기존), 버전 확인 직후 · 첫 판독 전에 자기 SQL 상수 전부를 prepare 해
  조건부 질의의 열 부재도 결함 신원으로 먼저 드러나게 한다(design D7).
- 수락 시험을 실제 원장으로: a112 트립와이어를 양성 시험으로 뒤집고 시험 다리를 지운다(a112 소유 파일 — Manager 가 a112 소유자에 통지).
  축소 픽스처 셋은 주입 값과 같은 `user_version` 을 쓰게 바꾼다(행 · 단언 불변).
- 판독 SQL · 사용량 판정 · owner 재구성 · 원장 스키마는 바꾸지 않는다.

## Impact

- 코드: `internal/riskbucket/production_snapshot_authority.go`, `internal/strategyrouter/production.go`, `internal/app/engine/strategy_risk_authority.go`,
  `internal/app/engine/strategy_route_authority.go` + 시험.
- 동작: 오늘(생산 서명 매니페스트 0 — 기록) 변화 0. **착지 뒤에는 핀이라는 우연한 차단이 사라지고 설계된 조건 사슬만 남는다**(design D6 의
  1~12 — 엔진 기동의 automation gate · attestation, scheduler 활성화, candidate · route · FX · proposal · risk · account 의 서명 매니페스트와 환경값,
  4-가족 활성화가 없으면 유효 제안 정확히 하나, 보호 readiness 와 배선, 진입 관문, 1차 레그 admission). 자동 판정을 뺀 나머지는 전부 사람 서명 ·
  운영자 설정이다(불변식 3 · 7). 주의: 시장 **승격**(화면)은 주문 관문이 아니다 — 주문은 refresh 사이클이 승격과 무관하게 내보낸다(design D6).
- 시험 기반: route 실원장 양성 시험을 위한 서명 매니페스트 작성 seam(`tossos_testseams`) + 외부 시험 패키지(design D4).
- a112: 시험 다리 제거 · 트립와이어 반전 · a112 의 동결 증거(FLM 번들 · 하네스)가 그 시험을 인용하는 자리의 재기준(tasks 1.3) · L6 의 형식상 선행 해소.
