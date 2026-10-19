# 불변 계획·위험예산 복원 사전 조사

2026-10-19, 격리 작업트리 HEAD `67810038f21dffd519a3d44edb9e255238b5b74a`, status clean 확인. 소스 읽기 전용. 실행 시험·운영·원본 checkout 수정 없음. 선행 a160은 저장 사실 조회만 하며 신규 진입 권한을 부여하지 않는다.

## 확인된 근거

- `internal/continuationlane/allocation.go:46-82,92-157`: scope/generation, 원래 예산·per-share risk·planned quantity·8:4:2 ceiling·policy/config·currency/FX가 계획 preimage. digest는 길이 접두 순서와 frozen FX의 private seal까지 묶는다. `internal/reversallane/plan.go:10-32,34-103`은 2:4:8 계획과 서로 다른 NUL 구분 digest, private request/ceiling과 FX seal을 쓴다. 해당 digest를 서로 공통 포맷으로 임의 해석할 수 없다.
- `internal/weeklyvaluelane/plan.go:19-42,106-130`은 7개 ceiling과 plan digest/seal을 private 상태로 보유. `internal/strategyflow/adapters.go:117-127,143-148`에서 투영 digest는 plan digest 자체가 아니라 `weekly-risk-v1` + plan digest + budget을 다시 해시한다. 세 lane에서 동일한 역산 복원법은 성립하지 않는다.
- `internal/strategyflow/canonical_projection.go:45-73`의 admitted lineage는 단일 leg ceiling과 budget digest를 보존한다. `internal/journal/position_campaign.go:32-59,101-145`는 scope/decision/generation, stop/latch, 개별 leg 수량 및 order watermark를 저장한다. 이 행과 projection에는 전체 계획의 예산·계획수량·모든 ceiling·FX 원문이 보장되지 않는다. 전 저장소의 모든 다른 외부 자료 부재까지 단정한 것은 아니다.
- `internal/journal/risk_bucket.go:61-90,211,232-241,350-373`의 `request_preimage`는 qFinal admission과 bucket 증거·정책·caps preimage이며 lane 전체 계획 preimage와 동일하지 않다. 같은 파일 `:476,589,674-724`의 filled/held 및 allocation은 다섯 monetary bucket 차원이다. 이들을 합하거나 stop 기반 손실 위험으로 간주할 근거는 없다. 값 단위·시점과 권위가 다르다.
- `internal/continuationlane/risk.go:27-34,68-70`, `internal/reversallane/types.go:308-315`, `internal/reversallane/risk.go:82-84`는 filled/held와 latch·fill/cancel dedupe를 갖고 생성자는 0으로 시작한다. `internal/weeklyvaluelane/risk.go:19-44,246-274`의 private seal은 filled/held 외 latch·fill fingerprint도 묶는다. 액수만 있거나 `NewRiskState`를 호출한다고 과거 상태가 복원되는 것은 아니다. FX의 private seal도 JSON 직렬화 왕복으로 자동 복구된다고 주장할 수 없다.
- `internal/journal/campaign_continuation_snapshot.go:15-29,31-89`는 같은 read transaction에서 현재 scope/generation/position/decision을 검증하고 `FULL_PLAN_PREIMAGE_UNAVAILABLE`을 명시한다. 기존 API는 안정된 사실 기반으로 재사용 가능하지만 full plan이나 위험 잔액을 제공하지 않는다. Journal schema는 현재 v36 (`internal/journal/schema.go:6`); 기존 migration은 historical rows에 값을 발명하지 않는 사례 (`position_campaign.go:68-70`)를 갖는다.

## 권장 범위 및 경계

1. 기존 a160 조회를 기반으로 **오프라인 전용 versioned plan/risk evidence 저장·복원 계약**을 별도 설계한다. scope·generation·최초 decision과 원래 lane digest를 결속한 전체 preimage, 단위·통화·FX 출처, filled/held 및 latch와 replay/dedupe에 필요한 근거를 명시한다. 저장 증거와 현재 원장 사이의 binding을 같은 DB transaction에서 검증한다. 저장 값의 형식/완전성 검증과 그 출처의 실제 권위 인증은 별개다. 호출자 제공 자료나 새 해시만으로 기존 production admission을 증명하거나 신뢰 인장을 만들지 않는다.
2. DB 내부 durable 저장이 요구되면 additive schema/migration 및 좁은 journal writer/reader가 필요하다. 기존 production admission에 자동으로 완전 preimage가 생성되는 일은 없으므로, production wiring이 금지된 이번 단계는 임시 DB의 명시적 synthetic/offline fixture 저장·reopen과 일관된 읽기로 한정한다. 실제 캠페인에 동일한 보장을 제공하려면 향후 최초 admission과 같은 transaction에서 정당한 preimage 생성·결속이 필요하며 별도 승인/검증이 필요하다. public writer가 임의 입력을 authoritative state로 격상하는 통로가 되지 않도록 설계해야 한다.
3. 기존 DB/legacy/누락·손상·unknown schema·digest 불일치는 typed unavailable/error, partial result 없음. 이미 불변인 v2/v3/v4 decision/receipt를 수정하거나 historical budget을 backfill하지 않는다. 기존 거래 보호 및 최초 leg 경로는 유지. plan 복원과 risk state 복원은 별도 성공조건이다. fill/held/수수료·FX/중복기록/취소/latch 근거 불충분하면 risk unavailable; plan만 확인돼도 ready/order capability를 발급하지 않는다.
4. 실데이터 즉시 복원이 목표라면 먼저 authoritative preimage와 fill-risk provenance 존재를 별도 입증해야 한다. 오프라인 sidecar 파일은 caller-supplied 자료의 보존/검증에는 사용할 수 있으나 admission에 원자적으로 결속된 증거가 없으면 생산 원장의 과거 권위를 증명할 수 없다. 권장안은 synthetic DB 안에서 안전한 불변 binding/복원 계약을 검증하는 단계이며 실운영 활성화·주문·설정·UI/API/CLI 연결이나 후속 qFinal 발급은 포함하지 않는다.
5. 세 lane의 private FX/weekly plan 및 risk seal은 단순 필드 추출·재해싱으로 복원할 수 없다. 기존 생성 경로와 권위 경계를 세밀히 검토해 lane별 불변 DTO/복원 책임을 설계해야 하며, 현재 조사만으로 가능한 공개 복원 생성자나 seal 전달을 확정하지 않는다. source/state bytes만 신뢰하는 대안 역시 변조 방어/출처 책임을 별도 정의해야 한다.

## 테스트·guard 재사용 후보

- a160의 일관성·재시작·오류·scope/foreign/orphan 테스트 `internal/journal/campaign_continuation_snapshot_test.go:79-300` 및 `readonly_test.go:116` ReadOnly method allowlist. 동일 transaction의 동시 쓰기 전후 snapshot과 DB 변경 없음 재검증.
- `internal/journal/strategy_first_leg_atomic_test.go:21-370`: 실제 최초 admission, replay/reopen, 경쟁, rollback, immutable binding. `internal/journal/risk_bucket_fill_test.go:18-198,436-514`: partial/replay/held, 다결정 및 restart. 각 lane `allocation_risk_test.go` 및 `weeklyvaluelane/plan_risk_rr_test.go`의 기존 숫자와 sealed state validation.
- 새 범위에 대한 독립 oracle: 정상 full preimage와 모든 위험 증거, reopen 및 replay; missing/legacy/corrupt/digest·scope·generation·decision/FX mismatch; 0과 unknown 구분; fill/cancel ID 충돌 및 latch 보존; peer isolation; exact replay vs conflicting write; concurrent transactional read; migration failure rollback, v36 historical bytes 및 older-binary rejection. 승인된 사실이 없으면 결과를 unavailable로 검증. 변이 시험으로 해시 결속과 위험 누락 검출 확인.
- 변경 대상에 따라 `internal/position/eligibility_drift_test.go:34-36,135-139`, `internal/journal/readonly_test.go:116`, `internal/strategyengine/a155_wiring_guard_test.go` 및 schema census/migration/version guard를 사전 검토. 테스트 guard는 약화하지 않는다. 실제 변경 함수 선정 후 최신 Go AST/FLM/BTM 및 repo SDD freshness 필요.

## 검색 및 한계

기존 repo `rg` 및 이전 source-grounded 조사 `/tmp/tossos-nextleg-implementation-plan.md` 재사용. 기존 CodeGraph/AST 자료는 그 문서 기재의 선행 HEAD 자료이므로 이번 HEAD에서 최신 증거라고 주장하지 않는다. ai-memory 도구 목록에 memory 조회 capability가 없어 explicit-scope 조회 불가; 설정 변경/쓰기 없음. 현재 로컬 메모리는 경로 탐색 용도로만 참고했고 본 결론은 현재 소스에 근거한다. 외부 패키지·GitHub·네트워크 검색 생략: 문제는 이 저장소의 봉인/원장 정본과 결속되며 외부 라이브러리 도입 제안 없음. 이 조사는 테스트 실행 또는 완성 구현이 아니다.

## 추가 AST·그래프 확인과 보존 방식 비교

2026-10-19 현재 HEAD 기준 정확한 `BuildCampaignPlan`, `planDigest`, frozen FX 생성자, `NewRiskState`/weekly `mintRiskState`·seal, snapshot과 bucket digest, first-leg commit 함수의 Go AST 16건은 `/tmp/tossos-immutable-restore-evidence/*.ast.json`에 함수별 source SHA와 위치를 포함한다. 동일 폴더의 CodeGraph status는 up to date, plan callers·bucket context와 CGC stats는 보조 근거다. CodeGraph 이름 검색은 동명 lane 함수를 합칠 수 있으므로 실제 소스·AST로 구분했고 CGC stats는 공유 인덱스 가능성 때문에 단독 증거가 아니다. 이 조회는 sync나 source 변경을 수행하지 않았다.

| 선택 | 가능한 검증 | 제한 |
|---|---|---|
| 오프라인 durable artifact | versioned caller-supplied plan 기록, 정확한 내용과 scope/digest 확인, 파일 재개봉·변조 검출 | 최초 admission과 원자 결속되지 않으며 실제 원장 preimage나 실행 권위를 입증하지 못함. 개인키 없는 일반 hash는 출처 인증이 아님. 기존 sealed FX를 DTO에서 임의 mint 금지. |
| additive journal schema + 미연결 오프라인 writer | 테스트 DB에서 immutable record와 scope/decision 결속, 원장 snapshot 일관성·재시작·충돌·migration을 검증 가능 | schema가 운영 DB에 적용될 수 있으므로 migration 자체의 무손실/호환 검증 필요. 임의 writer의 정보는 caller assertion이고 기존 최초 admission과 사후 결속을 권위로 위장하면 안 됨. production admission을 바꾸지 않으면 실제 기존/신규 캠페인은 preimage unavailable. |

초기 예산 자체 복원은 원본 preimage·해시·scope 일치가 확인된 범위의 별도 결과다. **현재 filled/held risk**는 체결·취소·남은 예약·비용·FX 및 dedupe/latch의 완전한 동일 시점 권위가 독립적으로 입증되어야 한다. 입증되지 않을 때 unknown이며 초기 예산이 알려져도 zero나 사용 가능 잔액으로 만들 수 없다. synthetic fixture에서도 실제 lane seal을 외부 DTO로 임의 mint하지 않는다.
