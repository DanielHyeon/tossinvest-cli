# a046 · 후보 추격 방지 임계값 승인

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-004`
- **Feature**: `FEAT-TOS-006`
- **Story**: `STORY-TOS-a046`

## Why

후보 발굴은 구현됐지만 `seen_late`와 `extended`의 승인 임계값이 없어 `passed`가 구조적으로 0이다. 관측 근거와 시장 범위를 가진 versioned threshold set을 승인해 주문 권한 없이 실제 후보 verdict를 만들 수 있게 한다.

## What Changes

- `seen_late`, `extended`, `near_high` threshold set의 출처·시장·세션·버전을 정의한다.
- 관측·분포·markout 근거가 없는 임계값을 거부한다.
- 미승인 또는 불완전한 set은 fail-closed를 유지한다.
- 승인된 verdict는 후보 소비 계약까지만 제공하고 주문을 만들지 않는다.
- a050의 `candidate-filters` 카테고리에 시장별 `seen_late`, `extended`, `near_high`의 방향·단위·값·표본·근거·설명을 제공한다.
- 최초 기본 상태는 `미승인 / passed 구조적 0 / verdict 비활성`이다. 근거 없이 숫자 0을 threshold 기본값으로 표시하지 않고, 승인 요건이 불완전하면 입력을 읽기 전용으로 유지한다.
- **비목표**: 전략 진입, 자동 주문, 신규 후보 source 확대.

## Capabilities

### New Capabilities

- `candidate-veto-thresholds`: 근거 기반 임계값 승인과 후보 pass 계약.

### Modified Capabilities

- `operator-console`: 후보 필터의 근거·기본 상태·승인 preview 표면을 추가한다.

## Impact

- `internal/candidate`, `cmd/tossctl/candidate*`, candidate store/metrics와 `/signals`.
- 읽기·분석 경로이며 account mutation capability를 갖지 않는다.
