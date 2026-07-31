## Context

trade outcomes는 close 단위 성과를 계산하지만 lane/candidate linkage와 intra-trade time series가 없다.

## Goals / Non-Goals

**Goals:** 결정적 attribution과 명시적 missing state를 가진 성과 read model을 제공한다.

**Non-Goals:** 추천, 설정 적용, 자본 배분, 거래 권한.

## Decisions

1. 기존 candidate/decision/attempt/order/fill/position IDs를 lineage key로 사용하고 symbol/time 근사 join을 금지한다.
2. high-volume observation은 journal과 분리한 derived `performance.db` append-only table에 저장한다. markout은 a046 `internal/markout`의 5/15/30분 +60초 tolerance 계약을 재사용하고 MFE/MAE는 기존 position lifetime 관측만으로 파생한다. a048가 별도 budget을 승인하기 전 추가 quote poll은 0건이다.
3. 수집 누락은 `not_measured`, 링크 누락은 `link_missing`, 표본 부족은 `insufficient_sample`이다.
4. 집계는 raw rows에서 파생하고 versioned query semantics를 가진다.
5. UI 소유 카테고리는 `performance-history`이며 mutation control을 두지 않는다. 기본 query는 최근 30일, all markets, all lanes, complete lineage only다.
6. 각 metric은 정의, 단위, sample count, period, source/version을 제공하고 `link_missing`, `not_measured`, `insufficient_sample`을 서로 다른 설명 상태로 표시한다.
7. raw derived observations는 90일 보존한다. 24시간마다 최대 500 rows/transaction을 prune하고 각 transaction의 writer lock 목표는 100ms 미만이다. authoritative journal lineage/outcome와 audit은 prune하지 않는다. aggregate/query semantics는 versioned이며 default 30일 query의 p95 목표는 250ms 이하(1,000,000 raw row fixture)다.

## Risks / Trade-offs

- [관측 API budget 증가] → 기존 가격 관측을 재사용하고 추가 poll은 예산 승인 전 금지한다.
- [DB 증가] → retention과 pruning을 명시하고 raw outcome 보존과 분리한다.
- [필터 기본값을 정책 추천으로 오인] → 조회 기본값임을 label에 명시하고 적용/save action을 제공하지 않는다.

## Migration Plan

journal에는 nullable lineage만 additive migration으로 추가하고 high-volume observation은 별도 `performance.db`에 둔 뒤 read-only 집계부터 노출한다. rollback은 collector를 중지하고 derived DB를 이동/재구축할 수 있지만 journal lineage는 보존한다. 기존 주문·exit에 영향이 없다.

## Open Questions

없음. 최초 cadence는 기존 관측 재사용으로 고정하며 별도 polling은 후속 budgeted change다.
