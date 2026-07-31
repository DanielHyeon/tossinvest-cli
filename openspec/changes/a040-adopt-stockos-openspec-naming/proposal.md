# a040 · StockOS OpenSpec 명명 규칙 채택

## 0. 백로그 계층 추적

- **Initiative**: `INIT-TOS-001`
- **Epic**: `EPIC-TOS-001`
- **Feature**: `FEAT-TOS-001`
- **Story**: `STORY-TOS-a040`

## Why

TossOS는 Story와 OpenSpec을 1:1로 관리하지만 change ID에 연속 번호가 없어 우선순위와 Story 대응을 이름만으로 검증하기 어렵다. StockOS의 검증된 `aNNN-kebab-intent` 규칙을 신규 change부터 적용해 PM 역추적을 기계적으로 고정한다.

## What Changes

- 신규 OpenSpec change는 `a<3자리 연속번호>-<kebab-intent>`를 사용한다.
- 대응 Story는 `STORY-TOS-a<NNN>`을 사용하고 같은 번호의 change 하나만 가리킨다.
- PM 검증기는 신규 번호형 Story/change의 번호 불일치, 중복 번호, 잘못된 slug를 거부한다.
- `docs/WORKFLOW.md`와 TossOS OpenSpec 템플릿에 명명 규칙을 기록한다.
- 기존 `STORY-TOS-001~039`와 무번호 change는 이력 보존을 위해 변경하지 않는다.
- **비목표**: 기존 change 일괄 rename, archive 경로 재작성, 제품 거래 동작 변경.

## Capabilities

### New Capabilities

- 없음.

### Modified Capabilities

- `sdd-workflow`: 신규 Story와 OpenSpec change의 번호·slug·1:1 추적 규칙을 추가한다.

## Impact

- `docs/WORKFLOW.md`, `openspec/templates/`, `tools/pm/generate_master_tracker.py`와 관련 테스트.
- PM 원본 및 generated tracker. 주문·위험·원장·LIVE 설정에는 영향이 없다.
