## 1. Lineage·수집 계약

- [x] 1.1 a047 provenance와 trade outcome schema의 CodeGraph impact, 추가 poll 0, raw 90일/24시간·500-row prune와 performance.db 계획을 작성한다.
- [x] 1.2 complete/link_missing/not_measured/insufficient_sample, a046 +60초 markout와 기존 관측 기반 MFE/MAE RED 테스트를 추가한다.

## 2. Persistence와 집계

- [x] 2.1 nullable lineage와 append-only observation schema를 구현한다.
- [x] 2.2 5/15/30분 markout, slippage, MFE/MAE와 cost-adjusted lane aggregate를 구현한다.
- [x] 2.3 기존 portfolio 승률·PF·MDD·R 결과가 동일함을 회귀 테스트로 고정한다.

## 3. Read-only 표면과 검증

- [x] 3.1 lane/policy performance query와 `performance-history` read-only view를 구현하고 metric help/unit/sample/period/provenance를 제공한다.
- [x] 3.2 최근 30일/all markets/all lanes/complete lineage only 조회 기본값과 link_missing/not_measured/insufficient_sample 설명을 fixture로 고정한다.
- [x] 3.3 mutation control·capability 부재, 모바일·접근성, pruning, migration/full test·vet·validate와 독립 리뷰를 통과한다.
- [ ] 3.4 `make gate CHANGE=a049-add-lane-performance`을 통과한다.
