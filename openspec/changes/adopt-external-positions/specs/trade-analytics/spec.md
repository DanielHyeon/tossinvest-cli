# trade-analytics Specification (delta)

## ADDED Requirements

### Requirement: 합성 R의 구분 집계
편입 포지션의 실현 R은 합성 분모(편입 시점 관측가 × default_stop_pct)에서 나오므로, 집계(승률·PF·ΣR 등)는 엔진 진입(실측 R)과 편입(합성 R)을 구분해 표기해야 한다(SHALL). 구분은 `positions.adoption_id IS NOT NULL` 명시 조인으로 하며(SHALL — trade_outcomes 스키마 무변경, 시간창 휴리스틱 매칭 금지 계약 준수), 혼합 집계를 낼 때는 두 모집단의 표본 수를 병기한다(SHALL).

#### Scenario: 혼합 계좌의 성과 집계
- **WHEN** 엔진 진입 3건과 편입 2건이 완결된 계좌의 집계를 조회하면
- **THEN** 실측 R 집계(n=3)와 합성 R 집계(n=2)가 구분 표기된다
