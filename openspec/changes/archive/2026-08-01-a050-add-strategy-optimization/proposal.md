# a050 · 전략 최적화 수명주기 추가

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-002`
- **Epic**: `EPIC-TOS-004`
- **Feature**: `FEAT-TOS-011`
- **Story**: `STORY-TOS-a050`

## Why

현재 `/optimization`은 세 가지 공통 exit policy ID 선택기이며 성과 근거·설정 snapshot·rollback이 없다. a049의 결정적 성과를 이용해 설정 변경을 preview하고 감사·복구할 수 있는 운영 수명주기가 필요하다.

## What Changes

- parameter registry, 설정 snapshot, 후보 설정, preview, apply, history와 rollback을 추가한다.
- `/optimization`을 개요, 익절·보호, 종목별 관리, 후보 필터, 전략·실행, 성과·이력의 여섯 카테고리로 구성한다.
- StockOS 최신 lane-console의 화면 단위 탐색·부분 저장·desired/effective 재검증 패턴을 재사용한다.
- 모든 조작에 설명, 단위, registry 기본값, desired/effective 값, 허용범위와 적용 시점을 표시하고 모바일에서도 같은 정보구조를 유지한다.
- 자유 텍스트·숫자 직접 입력·typed confirmation 없이 server-defined preset/선택/토글만 제공한다.
- 모든 쓰기는 version/CAS와 before/after/actor/reason 감사를 요구한다.
- 표본·귀속이 부족하면 추천을 만들지 않는다.
- apply는 lane 또는 LIVE trading을 자동으로 켜지 않는다.
- paper/shadow/canary 승격 단계는 만들지 않고 승인된 설정은 다음 LIVE 기동 계약에 직접 사용한다.
- StockOS 구형 optimization의 긴 slider matrix, 자유 reason/symbol 입력과 임의 값 편집은 만들지 않는다.

## Capabilities

### New Capabilities

- `strategy-optimization`: versioned 설정 후보·적용·이력·rollback 계약.

### Modified Capabilities

- `operator-console`: `/optimization`을 정책 선택기에서 증거 기반 운영 화면으로 확장한다.

## Impact

- optimization config store, audit/journal, `internal/console/optimization*`, a049 performance read model.
- 설정은 high-risk이며 기존 포지션 snapshot과 LIVE 토글 불변을 유지해야 한다.
