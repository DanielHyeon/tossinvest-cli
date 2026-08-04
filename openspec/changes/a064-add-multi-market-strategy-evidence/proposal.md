# Change: add multi-market strategy evidence

## Why

현재 전략 입력은 KRX Parker lane의 제한된 가격·세션 문맥에 머물러 있고, KR과 US 전략이 공유할 수 있는 권위 있는 증거 계약이 없다. 이 상태에서 시장별로 존재하지 않는 사실을 추정하거나 최신 수치를 과거 의사결정에 소급 적용하면 후보 승인과 백테스트가 모두 오염된다. KR과 US lane 개발을 동시에 진행하려면 먼저 시장·시계·통화·출처가 명시된 point-in-time 증거 경계가 필요하다.

## What Changes

- KR과 US 증거를 공통 envelope로 정규화하되 시장, 거래일/공시일, `source_available_at`, 관측·수집 시각, 통화, 출처, 신뢰도와 immutable digest를 보존한다.
- 원문·정정·projection은 거래 journal과 분리된 append-only `evidence.db`에 저장한다. 거래 journal은 실제 결정이 소비한 evidence snapshot ID와 digest만 보존하며 evidence payload를 복제하지 않는다.
- 거래 불가·공시 위험처럼 모든 lane에 적용되는 fatal veto와, 특정 lane만 소비하는 scoring evidence를 분리한다.
- 필수 증거가 누락·stale·상충하거나 시장에 존재하지 않으면 추정값을 만들지 않고 typed unavailable/refusal로 fail closed한다.
- OpenDART, KRX 및 SEC EDGAR의 공식 programmatic 계약만 사용하는 bounded adapter, source policy, rate budget, retry/staleness 및 revision 규칙을 정의한다. deployment source policy가 endpoint/version, absolute call window, page/bytes/concurrency, deadline, retryable status와 Retry-After를 모두 검증하지 못하면 adapter는 disabled이고 호출은 0건이다. KRX 공식 programmatic 계약을 증명할 수 없으면 `SOURCE_UNAVAILABLE`이며 HTML scraping·WTS fallback을 사용하지 않는다. OpenDART 자격증명은 외부 설정으로만 받고 로그·DB에 저장하지 않는다.
- 이 change는 증거 수집·정규화·재현 계약까지만 제공하며 주문, Guardian 승인, lane 활성화 또는 운영 토글을 수행하지 않는다.

## Capabilities

### New Capabilities

- `multi-market-strategy-evidence`: KR/US 시장별 point-in-time 증거 envelope, fatal-veto/scoring 분리, 공식 소스 adapter와 fail-closed 품질 계약

### Modified Capabilities

- 없음.

## Impact

- 신규 시장 증거 모델, 독립 append-only `evidence.db`, 공식 소스 policy/adapter와 deterministic fixture/test가 추가된다.
- trading journal에는 consumed evidence snapshot ID/digest만 additive lineage로 연결되고 원문 payload·revision·credential은 들어가지 않는다.
- 후속 KR/US continuation, reversal, weekly-value lane가 이 capability만 통해 외부 증거를 소비한다.
- 외부 API 장애·rate limit·공시 정정은 명시적 unavailable/revision 상태가 되며, 주문 경로에는 직접 영향이 없다.
