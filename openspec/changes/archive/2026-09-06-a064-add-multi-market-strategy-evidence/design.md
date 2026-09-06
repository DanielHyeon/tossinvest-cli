## Context

현재 `strategy-engine`의 첫 lane는 KRX 정규장 5분봉과 제한된 종목 상태만 소비한다. 후속 KR/US continuation·reversal·weekly-value lane는 가격뿐 아니라 공시, 재무, 수급과 거래 가능성 증거를 필요로 하지만 두 시장은 출처, 거래일, 통화, 정정 방식이 다르다. 최신 레코드를 과거 평가에 덮어쓰거나 한 시장에 없는 필드를 대체 추정하면 재현성과 안전성이 깨진다.

권위 경계는 공식 OpenDART, KRX 정보데이터시스템 및 SEC EDGAR 계약이다. 네트워크 결과는 외부 사실이지 주문 권한이 아니며, 이 change는 어떤 broker mutation이나 운영 토글도 연결하지 않는다.

## Goals / Non-Goals

**Goals:**

- KR과 US를 동일한 소비 인터페이스로 제공하면서 시장별 의미와 부재를 보존한다.
- 임의 평가 시점에 당시 이용 가능했던 증거 집합을 결정적으로 재구성한다.
- fatal tradability/disclosure veto를 lane scoring과 분리한다.
- 공식 adapter 장애, rate limit, 정정과 상충을 typed 상태로 남기고 fail closed한다.

**Non-Goals:**

- lane 점수, 진입/청산 결정, Guardian 수량 또는 주문 제출
- 공식 출처에 없는 데이터의 합성·추정·제3자 fallback
- OpenDART 키 발급이나 운영 자격증명 배포
- 과거 공시의 경제적 의미를 자동 재해석하는 정정 정책

## Decisions

### D1. 공통 envelope와 시장별 typed payload를 분리한다

공통 `EvidenceEnvelope`는 evidence ID, market, symbol/issuer identity, evidence kind, source authority/record ID, market-local effective date, source event/filing/trade time, `source_available_at`, observed-at, ingested-at, revision identity, currency, confidence/availability, schema version 및 canonical payload digest를 가진다. `source_available_at`은 해당 revision을 외부 소비자가 실제로 취득할 수 있게 된 최초 시각이며 filing/trade effective time과 같다고 추정하지 않는다. 실제 필드는 evidence kind별 typed payload로 둔다.

단일 거대 nullable 구조보다 이 형태를 선택한 이유는 KR에만 있는 필드와 US에만 있는 필드를 서로 존재하는 것처럼 보이게 하지 않기 위해서다. 소비자는 지원하지 않는 kind/field를 typed unavailable로 받는다.

### D2. 별도 evidence.db가 append-only 증거 권위를 가진다

원문, revision과 projection payload는 trading journal과 별도 파일·연결·migration 계보를 가진 append-only `evidence.db`에 저장한다. evidence identity unique key는 `(authority, source_record_id, revision_identity)`다. 같은 identity와 같은 canonical digest의 재수집만 멱등 성공이며, 같은 identity에 다른 digest가 도착하면 두 번째 payload를 유효 row로 append하지 않고 `SOURCE_REVISION_CONFLICT` quarantine record로 격리한다. 정정은 반드시 새 revision identity와 supersedes 참조로 추가하며 이전 행을 갱신하지 않는다. trading journal은 주문 결정이 실제로 소비한 immutable snapshot의 `snapshot_id`와 `snapshot_digest`만 저장하고 evidence payload, source response, revision table 또는 credential을 저장하지 않는다.

별도 DB 대신 trading journal에 원문을 넣는 대안은 거래 원장의 크래시·migration·보존 범위를 외부 데이터 ingestion에 결합하므로 기각한다. 두 DB를 한 transaction으로 묶지 않는다. evidence snapshot을 먼저 봉인한 뒤 거래 결정이 ID/digest를 참조하며, 참조 대상 부재나 digest 불일치는 신규 진입을 fail closed한다.

### D3. point-in-time 조회는 source와 ingestion의 이중 시점을 검사한다

조회는 `evaluation_at`과 `ingestion_cutoff`를 모두 필수로 받아 `source_available_at <= evaluation_at`이면서 `ingested_at <= ingestion_cutoff`인 revision만 선택한다. source availability를 알 수 없으면 filing/effective time으로 대체하지 않고 unavailable로 둔다. 이 이중 조건은 늦게 공개된 자료와 늦게 수집된 자료를 각각 차단한다.

최신값 materialization만 저장하는 대안은 간단하지만 당시 알 수 없던 정정을 과거 결정에 누출하므로 기각한다. payload digest는 canonical encoding에 대해 계산하고 원문 시각·통화·단위 변환 전후 provenance를 함께 보존한다.

### D4. fatal veto와 lane evidence projection은 별도 결과다

수집 계층은 증거를 저장하고, 결정적 projection 계층은 동일 as-of snapshot에서 `FatalAssessment`와 `LaneEvidenceSet`을 별도로 만든다. 거래정지 등 명시된 fatal fact는 모든 진입 lane에 적용되지만, 수급·재무 점수의 누락은 해당 evidence를 필수로 선언한 lane만 거부한다.

하나의 총점으로 합치는 대안은 fatal 상태가 가중치에 묻힐 수 있어 기각한다. 상충하는 authoritative fatal fact, required evidence 누락/stale 또는 identity 불일치는 typed refusal이며 0·false·중립 점수로 대체하지 않는다.

### D5. 시장 시계와 identity를 명시적으로 보존한다

KR은 `Asia/Seoul`, US는 해당 거래 캘린더가 정한 IANA timezone으로 market-local date를 계산하고 원본 timestamp와 UTC 정규값을 모두 보존한다. DST는 고정 offset으로 계산하지 않는다. ticker 문자열만으로 issuer를 결합하지 않고 market-qualified symbol과 source issuer identity mapping을 버전 관리한다.

### D6. source policy가 공식 adapter의 호출 권한을 선행한다

각 source는 deployment config에서 공식 programmatic endpoint identity/version, 허용 method/schema, request identification, credential requirement, absolute call-window limit, max pages, max response bytes, max concurrency, per-request/whole-operation deadline, retryable status 집합, 최대 retry 수와 `Retry-After` 해석을 고정한 versioned source policy가 있어야 한다. 필드 하나라도 미설정·invalid·공식 계약과 불일치이면 `SOURCE_DISABLED`이고 network call은 0건이다. runtime caller가 이 bound를 넓힐 수 없다. adapter는 policy가 mint한 request scope 안에서만 page/cursor, conditional retry, response metadata와 typed error를 반환하는 좁은 port를 구현한다. OpenDART 키는 환경/secret provider에서만 읽고 로그·digest payload에 포함하지 않는다. SEC 요청 식별 정보와 공식 access policy를 지키며, adapter별 absolute window/token budget과 bounded backoff를 공유 scheduler가 집행한다.

KRX 웹 화면의 존재는 programmatic contract 증거가 아니다. endpoint/schema/rate/access 계약이 공식 자료로 동결되지 않으면 KRX source는 `SOURCE_UNAVAILABLE`, 호출 0건을 반환한다. 임의 HTTP client, HTML scraping, WTS 또는 비공식 fallback은 금지한다. rate limit, schema drift, authentication failure 또는 page 불완전은 partial success로 숨기지 않고 source health와 unavailable evidence로 기록한다. 공식 fixture와 canonical digest golden test로 parser drift를 탐지한다.

## Risks / Trade-offs

- [공식 응답 스키마나 접근 정책 변경] → fixture contract test, schema version, source-health refusal과 보수적 retry budget으로 격리한다.
- [공시 정정으로 최신 화면과 과거 평가가 다름] → append-only revision과 `source_available_at/evaluation_at`, `ingested_at/ingestion_cutoff` 이중 조건으로 두 관점을 모두 재현한다.
- [issuer mapping 오류] → market-qualified immutable identity와 mapping version을 요구하고 ambiguity는 fail closed한다.
- [외부 API 지연이 lane loop를 지연] → ingestion과 lane evaluation을 분리하고, evaluation은 bounded local snapshot 조회만 수행한다.
- [OpenDART 자격증명 부재] → source policy는 disabled이고 호출 0건을 유지하며 다른 시장 사실이나 추정값으로 대체하지 않는다.
- [KRX 공식 programmatic 계약 부재] → `SOURCE_UNAVAILABLE`로 유지하고 웹 화면 scraping이나 WTS를 도입하지 않는다.

## Migration Plan

1. 독립 `evidence.db` schema를 생성하고 trading journal에는 nullable consumed snapshot ID/digest lineage만 additive migration으로 배포하되 consumer와 runtime binding은 OFF로 둔다.
2. recorded official fixtures로 adapter, canonicalization과 as-of replay를 검증한다.
3. 자격증명이 필요 없는 source부터 shadow ingestion을 시작하고 source health/digest만 관측한다.
4. OpenDART 키가 사람이 제공된 환경에서만 해당 adapter를 활성화한다.
5. 후속 lane는 이 capability의 immutable snapshot ID만 참조한다.

Rollback은 consumer를 비활성화하고 신규 ingestion을 중지하는 방식이다. append-only `evidence.db`와 journal의 snapshot 참조는 보존하며 각 DB의 더 새로운 schema는 해당 구버전 reader가 fail closed로 거부한다.

## Open Questions

- 운영 환경별 공식 source rate budget 수치는 배포 전 각 서비스의 현재 정책과 계정 한도로 확정하고 감사 가능한 source policy로 남긴다. 미확정 source는 disabled/0 calls다.
