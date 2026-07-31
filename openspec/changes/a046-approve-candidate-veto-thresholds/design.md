## Context

candidate store는 shadow bands와 crossings를 보존하지만 두 veto threshold가 비어 있어 모든 verdict가 unmeasured다. threshold는 코드 상수보다 versioned 승인 자료여야 한다.

## Goals / Non-Goals

**Goals:** 관측 근거를 재현할 수 있는 시장별 threshold set과 fail-closed 로더를 정의한다.

**Non-Goals:** 후보 source 확대, 전략 점수, 주문 소비, LIVE 토글.

## Decisions

1. threshold set은 market, session, metric definitions, values, sample window/count, evidence digest와 approved-at/by를 가진다.
2. runtime은 완전하고 승인된 한 version만 읽는다. 일부 field fallback이나 시장 간 공유를 금지한다.
3. markout은 판단 시점 이후 5/15/30분 관측으로 정의하되 데이터 누락을 0수익으로 처리하지 않는다.
4. verdict output에 threshold version을 포함해 이후 성과 귀속이 가능하게 한다.
5. UI 소유 카테고리는 `candidate-filters`다. descriptor는 market/session/metric label, 판정 방향, 단위, value, valid range, sample window/count, missing rate, evidence digest, help와 provenance를 포함한다.
6. 승인 전 상태는 숫자 default가 아니라 `unapproved`다. 근거 자료와 registry가 완전해질 때까지 field는 read-only이며 `passed=0`이 성과가 아니라 구조적 비활성임을 설명한다.
7. 이 change는 transport나 주문에 의존하지 않는 `internal/markout` 순수 계약을 소유한다. 판단 시각 이후
   5/15/30분 target 각각에 대해 기존 관측 stream에서 target 이상인 첫 관측을 최대 60초 tolerance 안에서
   선택하고, 없으면 `not_measured`로 남긴다. 새로운 quote poll은 만들지 않는다. a049는 이 계약을 재사용한다.
8. 기존 `near_high=2.0`은 자동 승인값으로 승격하지 않고 `legacy-unapproved` provenance로 명시한다.
   최초 numeric threshold set은 별도 사람 evidence review가 immutable version과 digest를 선택할 때만 active가 된다.

## Risks / Trade-offs

- [표본 편향] → 시장·세션별 표본 수와 누락률을 같이 공개한다.
- [임계값 변경으로 재현 불가] → immutable version과 evidence digest를 저장한다.
- [0을 실제 임계값으로 오인] → unknown/unapproved를 nullable typed state로 제공하고 승인 전에는 read-only로 유지한다. 승인 후에도 registry discrete option만 제공하고 숫자 직접 입력을 금지한다.

## Migration Plan

분석 보고서, 순수 markout 계약과 loader를 추가한 뒤 승인 전 구조적 0 상태를 유지한다. 구현 change는 `unapproved/passed=0`으로 완료할 수 있고, 최초 numeric approval은 근거가 생겼을 때 별도 human activation record로 수행한다. 승인 file이 명시적으로 저장된 다음부터 verdict만 활성화하며 주문 경로는 없다.

## Open Questions

최초 승인 수치는 구현 전 별도 evidence review에서 확정한다.
