## ADDED Requirements

### Requirement: 시장별 증거 envelope는 출처와 시점을 보존한다

모든 KR/US 전략 증거는 market-qualified issuer/symbol, source authority와 record identity, evidence kind/schema version, source event·filing·trade time, market-local effective date, `source_available_at`, observed-at, ingested-at, revision identity, currency/unit, availability/confidence 및 canonical payload digest를 가진 immutable envelope로 저장되어야 한다(SHALL). `source_available_at`을 filing/effective time으로 근거 없이 대체해서는 안 된다(MUST NOT). KR과 US에 공통으로 존재하지 않는 field는 합성·추정하거나 다른 시장의 값으로 채워서는 안 된다(MUST NOT).

#### Scenario: 시장에 존재하지 않는 필드
- **WHEN** US evidence consumer가 KR 공식 출처에만 정의된 field를 요청한다
- **THEN** 시스템은 typed unsupported evidence를 반환하고 0·false·빈 문자열 또는 추정값을 반환하지 않는다

#### Scenario: 동일 원문 재수집
- **WHEN** 같은 authority, source record, revision과 canonical payload가 다시 수집된다
- **THEN** 기존 evidence identity를 멱등 반환하고 중복 immutable row를 만들지 않는다

#### Scenario: 동일 revision identity의 payload 충돌
- **WHEN** 같은 authority, source record와 revision identity에 기존 row와 다른 canonical digest가 수집된다
- **THEN** SOURCE_REVISION_CONFLICT로 새 payload를 quarantine하고 두 payload를 모두 유효 evidence row로 노출하지 않는다

### Requirement: point-in-time 조회는 미래 정보와 정정을 누출하지 않는다

증거 저장소는 정정을 append-only revision으로 기록하고 이전 payload를 변경해서는 안 된다(MUST NOT). 모든 평가 조회는 `evaluation_at`과 `ingestion_cutoff`를 명시해야 하며(SHALL), `source_available_at > evaluation_at` 또는 `ingested_at > ingestion_cutoff`인 최초 자료·거래·정정 revision을 결과에 포함해서는 안 된다(MUST NOT). source availability를 확인할 수 없는 required evidence는 unavailable로 fail closed해야 한다(SHALL). 결과는 선택된 evidence ID와 digest를 보존해 동일 fixture와 두 cutoff에서 결정적으로 재현되어야 한다(SHALL).

#### Scenario: 평가 뒤 도착한 정정
- **WHEN** 평가 cutoff 이후 공시 정정 revision이 수집된다
- **THEN** 과거 as-of 조회는 이전 revision을 유지하고 현재 조회만 정정 revision과 supersedes lineage를 반환한다

#### Scenario: 미래 거래일 데이터
- **WHEN** 요청한 market-local evaluation date보다 뒤의 거래 데이터가 저장소에 존재한다
- **THEN** 해당 데이터는 snapshot에서 제외되고 digest에도 포함되지 않는다

#### Scenario: 공개는 늦고 수집은 빠른 자료
- **WHEN** evidence의 effective date는 evaluation date 이전이지만 source_available_at이 evaluation_at 뒤다
- **THEN** evidence는 snapshot에서 제외되고 effective date를 공개 시각으로 간주하지 않는다

#### Scenario: 공개됐지만 아직 수집되지 않은 자료
- **WHEN** source_available_at은 evaluation_at 이전이지만 ingested_at이 ingestion_cutoff 뒤다
- **THEN** evidence는 snapshot에서 제외되고 historical query에 소급 포함되지 않는다

### Requirement: Evidence payload 권위는 trading journal과 분리된다

원문 response, normalized payload, revision과 projection은 독립 append-only `evidence.db`에만 저장되어야 한다(SHALL). trading journal은 실제 결정이 소비한 immutable `snapshot_id`와 `snapshot_digest`만 저장해야 하며(SHALL), evidence payload, source response, revision table, API credential 또는 source header를 저장해서는 안 된다(MUST NOT). snapshot ID 부재 또는 digest 불일치는 신규 exposure-raising 결정을 거부해야 한다(SHALL).

#### Scenario: 결정의 evidence lineage 기록
- **WHEN** lane decision이 봉인된 evidence snapshot을 소비한다
- **THEN** trading journal에는 snapshot ID/digest만 기록되고 payload와 source response는 evidence.db에 남는다

#### Scenario: Evidence DB를 열 수 없다
- **WHEN** snapshot ID/digest를 evidence.db에서 검증할 수 없다
- **THEN** 신규 진입은 EVIDENCE_SNAPSHOT_UNAVAILABLE로 거부되고 trading journal payload fallback은 시도하지 않는다

### Requirement: fatal veto와 lane scoring evidence는 분리된다

동일한 as-of snapshot에서 시장 공통의 거래 불가·공시 위험을 표현하는 fatal assessment와 lane별 scoring evidence projection은 별도 typed 결과로 산출되어야 한다(SHALL). 명시적 fatal fact는 score로 상쇄되어서는 안 되며(MUST NOT), scoring evidence의 가중치나 필수 여부는 fatal 판정을 변경해서는 안 된다(MUST NOT).

#### Scenario: 높은 점수와 거래정지 증거
- **WHEN** lane scoring evidence가 높은 점수를 만들지만 authoritative snapshot에 거래정지 fatal fact가 있다
- **THEN** fatal assessment가 진입 불가를 반환하고 scoring 결과는 이를 상쇄하지 못한다

#### Scenario: 선택 evidence 누락
- **WHEN** fatal evidence는 완전하지만 특정 lane가 선택 사항으로 선언한 scoring evidence가 없다
- **THEN** fatal assessment는 그대로 보존되고 lane projection은 해당 evidence를 unavailable로 표시한다

### Requirement: required evidence 품질은 fail closed다

lane가 required로 선언한 evidence가 missing, stale, ambiguous, identity-mismatched, unit/currency-unresolved 또는 authoritative source 간 conflict이면 시스템은 안정적인 typed refusal을 반환해야 한다(SHALL). 이 상태를 중립 점수나 이전 시장일의 값으로 묵시적 대체해서는 안 된다(MUST NOT). freshness 기준과 source 우선순위는 evidence kind·market·version별 감사 가능한 정책이어야 한다(SHALL).

#### Scenario: 상충하는 필수 공시
- **WHEN** 같은 as-of 시점의 authoritative records가 required fact에 대해 해소되지 않은 상충 값을 제공한다
- **THEN** projection은 EVIDENCE_CONFLICT로 거부하고 lane decision input을 만들지 않는다

#### Scenario: stale 필수 증거
- **WHEN** required evidence의 observed-at 또는 effective date가 versioned freshness 한도를 넘는다
- **THEN** projection은 EVIDENCE_STALE로 거부하고 최근 값이라고 추정하지 않는다

### Requirement: 공식 adapter는 bounded access 계약을 지킨다

OpenDART, KRX 및 SEC EDGAR source는 deployment config의 versioned policy가 공식 programmatic endpoint/version, method/schema/access, request identity, credential requirement, absolute call-window limit, max pages, max response bytes, max concurrency, request/operation deadline, retryable status, max retries와 `Retry-After` 해석을 모두 고정·검증한 경우에만 호출할 수 있다(SHALL). 어느 필드라도 미설정·invalid·미검증이면 source는 disabled이고 network call은 0건이어야 한다(SHALL). 활성 adapter는 이 bound를 강제하고 runtime input으로 완화해서는 안 된다(MUST NOT). credential, access token 또는 개인 식별 header는 payload, digest, log, evidence.db와 trading journal에 기록해서는 안 된다(MUST NOT). 비공식 endpoint나 다른 시장 source를 fallback으로 사용해서는 안 된다(MUST NOT).

#### Scenario: rate budget 소진
- **WHEN** source adapter의 현재 rate budget이 소진된다
- **THEN** 추가 호출을 수행하지 않고 SOURCE_RATE_LIMITED를 반환하며 기존 evidence를 fresh로 가장하지 않는다

#### Scenario: OpenDART 키 부재
- **WHEN** OpenDART 요청에 필요한 credential이 설정되지 않았다
- **THEN** adapter는 SOURCE_CREDENTIAL_UNAVAILABLE을 반환하고 credential을 로그에 노출하거나 대체 source를 호출하지 않는다

#### Scenario: 불완전 pagination
- **WHEN** 공식 응답의 일부 page만 deadline 안에 수집된다
- **THEN** adapter는 완전한 evidence snapshot을 commit하지 않고 SOURCE_INCOMPLETE를 기록한다

#### Scenario: Source policy 미설정
- **WHEN** source policy의 endpoint/version, absolute call window, page/bytes/concurrency, deadline, retryable status 또는 Retry-After 규칙 중 하나라도 설정·검증되지 않았다
- **THEN** source는 SOURCE_DISABLED를 반환하고 HTTP 요청은 0건이다

#### Scenario: Adapter가 deployment bound를 넘으려 한다
- **WHEN** runtime 요청이 source policy보다 많은 page/bytes/concurrency/retry 또는 긴 deadline을 요구한다
- **THEN** adapter는 요청을 거부하거나 policy bound로 제한하며 deployment limit을 확대하지 않는다

#### Scenario: KRX programmatic 계약 부재
- **WHEN** KRX 웹 화면은 존재하지만 공식 programmatic endpoint/schema/access 계약이 동결되지 않았다
- **THEN** source는 SOURCE_UNAVAILABLE과 호출 0건을 반환하고 HTML scraping, WTS 또는 비공식 endpoint를 사용하지 않는다
