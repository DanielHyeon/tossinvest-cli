# 후속 leg 원장 복원과 오프라인 atomic admission: 사전 조사

2026-10-19, 격리 HEAD `5f3506599fd9cbd40b888aa954ae880244bedf9b`. 읽기 전용 조사. `/tmp/tossos-a156-direct-recheck.md` 정정을 적용한다. 기존 정본 continuation 8:4:2, reversal 2:4:8, weekly 최대 7 positive-fill ordinal과 immutable campaign budget, confirmation 및 미체결 잔여 미이월을 재설계하지 않는다. 이전 `/tmp/tossos-a156-nextleg-research.md`의 횟수·비중·예산 미정이라는 표현은 폐기한다.

## 현재 원장과 누락된 복원 근거

| 정보 | 현재 저장 근거 | 후속 처리의 제한 |
|---|---|---|
| scope, lane, 최초 decision, expected/actual generation, version, entry latch, effective stop과 source/policy/time | `internal/journal/position_campaign.go:31-59`; position 버전 companion `:68-80` | 한 account/market/symbol/campaign/generation에서 읽고 version/CAS로 재검증. 값이 존재한다는 것만으로 수량 발급 권한 아님. |
| leg plan ID, 요청/체결/잔여 수량과 상태·버전 | `position_campaign.go:101-115`; order watermark·terminal·lineage `:117-145` | leg row는 개별 plan ID/수량만 포함. 후속 planned ceiling이나 원래 8:4:2 등 전체 계획을 이 row에서 복구했다고 주장 불가. |
| first-leg sealed lineage | `internal/strategyflow/canonical_projection.go:49-73`: ordinal, planned ceiling, risk budget digest, identity 등 | digest는 해시 식별자다. 이 payload의 필드는 campaign 전체 예산 원문·전체 planned quantity/각 ceiling·FX/fee preimage를 담지 않는다. digest만으로 이를 역산하거나 새 plan 생성 금지. 추가로 기존 다른 저장 근거를 확정하기 전 재구성 가능하다고 가정하지 않는다. |
| monetary cap의 snapshots와 decision·5개 예약·filled/held/overage/latch, 실제 체결 provenance | `internal/journal/risk_bucket.go:674-724`; 저장 `:240-254`, read `:476,589`; `risk_bucket_fill.go:1028-1093` | 각각 차원을 갖는 금액; 다섯 행을 합산하면 중복이다. monetary bucket 노출량을 entry-stop 손실 위험으로 동일시할 근거가 없다. original budget 및 실제 손실 위험에 필요한 entry/stop/FX/cost provenance를 따로 확인해야 한다. |
| first-leg receipt/lease | `internal/journal/strategy_first_leg_v26.sql:1-24`의 sequence=1 binding, `strategy_dispatch_issuance.go:134-155`; generic issuer는 `strategy_dispatch_runtime.go:253-258`에서 dormant | 새 leg 권한을 기존 first-leg receipt로 위장하거나 dormant issuer를 켜면 안 된다. 최초 경로와 이미 발급된 receipt/lease 바이트 및 의미 보존. |

세 lane의 `NewRiskState`는 시작 filled/held를 0으로 만든다 (`continuationlane/risk.go:68-70`, `reversallane/risk.go:82`, `weeklyvaluelane/risk.go:19-44`). 생산 adapter도 그 초기값을 사용한다 (`continuationlane/production_proposal.go:226`, `reversallane/production_proposal.go:231`, `weeklyvaluelane/production_proposal.go:180`). 따라서 과거 campaign에 그대로 호출해 복원하면 위험을 과소계상한다. weekly state는 seal/private fields를 가지며 외부에서 숫자를 주입하는 가짜 복원 seam을 추가해서는 안 된다.

## 최소 안전 단계 제안 — 설계 승인 전 구현 아님

1. **durable plan 근거 확보**: 실제 기존 원장에 완전 immutable plan preimage를 제공하는 경로를 추가로 찾고 검증한다. 없으면 새로 admission하는 campaign에 한해 versioned/additive immutable plan preimage와 digest 결속이 필요할 수 있다. 기존 first-leg payload/decision/receipt를 재해석하거나 변경하지 않는다. 과거 plan이 복구 불가하면 후속 admission을 unknown/refuse로 남긴다. 실제 손절 기준 filled risk, held risk의 authoritative 매핑과 비용·frozen FX·미결/unknown 해제 근거는 기존 정본에 맞춰 별도 증명한다. 계획 없이 bucket의 filled 금액을 stop-risk 값으로 전환하지 않는다.
2. **일관된 스냅샷**: 같은 transaction/read snapshot에서 scope·generation·campaign/position 버전·latch/stop·각 leg/watermark·위험 원장 및 immutable plan을 취득한다. 외부 authoritative evidence와 snapshot freshness/version은 기존 guardian/authority 규칙으로 검증한다. 선행 order 불명/중복/terminal 불확실·누락된 plan/risk evidence는 fail-closed이며 이미 존재하는 보호 exit는 계속된다. 단순 PlanCampaignLeg는 caller quantity를 저장하는 명령이며 이 스냅샷을 산정하지 않는다 (`position_campaign.go:522-600`).
3. **offline 전용 transaction 경계**: 기존 `RecordQFinalCampaignFirstLeg`의 tx 구성 (`strategy_first_leg_atomic.go:103-121,168-212`)과 `CalculateAdmission`/`commitFreshRiskBucketAdmissionTx`를 참고해, 기존 안전 검사와 replay/digest/CAS를 그대로 활용하는 후속 전용 journal writer를 설계한다. fresh evidence 재검증·campaign/position/version 확인·계획별 ceiling과 해당 leg 실제 fill 차감·nonretreating stop·campaign budget에 filled+held+proposed를 검사하고, qFinal decision/reservation/leg/immutable binding을 전부 commit 또는 전부 rollback. 구현 전에 기존 internal tx helpers 재사용 가능성과 risk-state seal 생성 권한을 세 lane별로 증명해야 한다. 함수명을 새로 만들거나 migration을 확정한 제안은 아니다.
4. **실행 권한 별도 단계**: 위 결과는 오프라인 admission 검증까지만. first-leg receipt/lease 조회 및 경로는 유지하고 후속 transport/lease를 만들지 않는다. 향후 별도 명세에서 신규 후속 binding의 진위·이중발급 방지·race/불확실 transport 복구까지 증명한 뒤에만 기존 최종 fence에 연결한다. 실주문·활성화·설정·배포는 승인되지 않았다.

## 의미 있는 독립 테스트

기존 `continuationlane/allocation_risk_test.go:34,141`, `reversallane/allocation_risk_test.go:29,48`, `weeklyvaluelane/plan_risk_rr_test.go:9,28`의 분할·부분 체결·filled/held budget은 그대로 재사용한다. `strategy_first_leg_atomic_test.go:21,143,183,211,302,370`은 최초 leg 경쟁·replay·restart·rollback·binding을 보존하고, `risk_bucket_fill_test.go:436-512`는 2 decision 귀속/restart를 이미 확인한다. `a159_protection_restart_test.go:84-177`은 실제 3/8 체결과 watermark·차단을 시험한다.

새 admission 범위가 확정되면 동일 임시 DB에서 원래 plan의 각 ceiling, actual partial fill, 미확정 held, 새로운 qFinal을 독립 수치 oracle로 검증하고 no-double-count, 미이월, stale/generation/scope/plan digest·stop mismatch, unknown evidence, duplicate/restart, 경쟁자 한 명만 성공, late statement rollback(결정/leg/모든 reservations 0), first-leg receipt 및 보호 exit 불변을 검증한다. fixture가 provenance를 조작해 권위를 가장하면 그 결과는 생산 연결 증거가 아니다. 구체 수치는 정본 분할과 승인된 계획값에서 파생한다. 이 조사는 테스트 실행이 아니다.

## 아직 설계로 정해야 할 기술 결속

기존 비중·횟수·고정 budget은 결정 완료다. 남은 것은 해당 값의 durable preimage를 언제 확보할지(기존 캠페인 부재 시 refusal), exact snapshot/transaction/CAS 경계, lane risk seal을 신뢰 가능한 복원자가 어떻게 구성할지, entry-stop 기반 위험과 monetary bucket exposure의 별도 산정 및 fee/FX 증거 결속, 동시성·replay에서 receipt identity와 보류 주문 처리다. 새로운 정책 비율이나 live authority를 결정하는 질문이 아니다. 기존 a156의 live 실증 요구는 이 offline 단계로 완료되지 않는다.

## 증거와 한계

현 HEAD 기준 CodeGraph status/caller 결과 및 함수별 소스 SHA가 포함된 Go AST는 `/tmp/tossos-nextleg-plan-evidence/`, 앞선 plan/fill/admission/stop/floor AST는 `/tmp/tossos-a156-nextleg-evidence/`. CGC는 여러 작업트리 혼합 가능하므로 보조 조회다. 원장 plan 저장의 부재 판단은 명시된 campaign schema와 accepted projection 필드 범위의 관찰이며, repository 전체에서 모든 대체 저장 경로가 없다는 확정이 아니다. 코드 변경 전 exact 대상 fresh AST/FLM/BTM, design/review, SDD fresh sync가 추가로 필요하다. 기존 파일·운영·원본 checkout 수정 및 신규 테스트 실행 없음.
