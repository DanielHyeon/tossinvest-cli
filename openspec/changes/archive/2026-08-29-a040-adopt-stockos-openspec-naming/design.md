## Context

TossOS PM은 파일 기반 hierarchy와 생성기로 Story/change 1:1을 검증하지만 신규 change 번호 규칙이 없다. StockOS는 `aNNN-kebab-intent`와 같은 번호의 Story를 사용한다.

## Goals / Non-Goals

**Goals:** 신규 번호형 change의 형식·연속성·1:1 번호 일치를 PM gate에서 검증하고 문서 템플릿을 통일한다.

**Non-Goals:** 기존 39개 Story와 archive를 rename하거나 StockOS의 프로젝트 prefix를 그대로 사용하지 않는다.

## Decisions

1. 신규 기준선은 `a040`이며 이후 번호를 단조 증가시킨다. 기존 최대 Story 039 다음 번호라 충돌하지 않는다.
2. change는 `aNNN-kebab-intent`, Story ID는 `STORY-TOS-aNNN`이다. intent는 Story 제목이 아니라 change slug에만 둔다.
3. PM validator는 legacy ID를 허용하고 `aNNN` Story에만 신규 강제를 적용한다. 전체 rename보다 이력과 링크를 보존한다.
4. 같은 `aNNN` change가 둘 이상이면 거부한다. StockOS의 과거 a044 충돌은 이식하지 않는다.

## Risks / Trade-offs

- [두 명이 같은 다음 번호를 선택] → PM validation에서 중복 번호를 거부하고 merge 전에 번호를 재배정한다.
- [legacy와 신규 규칙 혼재] → WORKFLOW에 cutoff `a040`을 명시하고 generated map에 Story/change를 함께 표시한다.

## Migration Plan

WORKFLOW·템플릿·validator test를 먼저 갱신하고 신규 Story 12개를 규칙에 맞춰 등록한다. rollback은 validator의 신규 검사를 제거하되 생성된 ID는 변경하지 않는다.

## Open Questions

없음.
