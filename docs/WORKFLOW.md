# TossOS 개발 워크플로 (SDD 계약)

> 이 문서는 TossOS를 **개발**하는 모든 에이전트·개발자에게 적용된다.
> tossctl을 **운용**하는 에이전트 규칙은 AGENTS.md를 따른다 (두 문서의 스코프는 AGENTS.md 상단 참조).
> 개정: 2026-07-26 gstack 리뷰(codex + CEO/Eng/DX 4보이스) 결정 반영 — 기록은 openspec/changes/add-tossos-foundation/review.md

## 역할 분리

- **Manager(총괄 아키텍트)**: 작업 분할, OpenSpec 변경 작성·검토, 구현 결과 리뷰·검증. 구현 코드를 직접 작성하지 않는다.
- **Teammate(구현 에이전트)**: tasks.md 단위 작업을 구현·테스트 (현재 운용 기본값: Opus 모델). 스펙 밖 임의 설계 변경 금지.
- 핵심은 모델명이 아니라 **작성자와 검증자의 분리**다: 구현을 만든 컨텍스트와 그것을 검증하는 컨텍스트는 항상 별도 세션이어야 한다. 사람 혼자 작업할 때도 구현 후 별도 리뷰 패스를 거친다.

## SDD 사이클

1. Manager가 `openspec/changes/<change-id>/`에 proposal.md, design.md(필요시), specs/ 델타, tasks.md 작성
2. `openspec validate <change-id> --strict` 통과
3. **리뷰 게이트(등급제)** — 아래 "리뷰 게이트" 절
4. Teammate가 tasks.md 항목 단위로 구현. 각 task의 완료 체크는 **그 산출물을 만드는 커밋과 같은 커밋**에서 수행한다
5. Manager가 diff 리뷰 + 독립 테스트 재실행으로 검증
6. `make gate CHANGE=<change-id>` 통과 후 change 완료 선언, `openspec archive <change-id>`

## 리뷰 게이트 (등급제)

모든 문서에 동일한 무게의 리뷰를 강제하면 게이트는 조용히 무시된다. 게이트는 두 지점에만 건다:

| 시점 | 대상 | 요구 |
|---|---|---|
| **proposal-freeze** | change의 첫 구현 task 착수 전 | proposal/design/spec 델타에 대한 gstack 리뷰(autoplan 4관점) 1회 |
| **requirement 변경** | 이후 spec의 Requirement 수준 수정 | 수정분에 대한 gstack 리뷰 재실행 |

- **면제**: tasks.md 체크박스·상태 갱신, 오탈자, 링크 수정, 리뷰 결정 반영 자체
- **리뷰 기록 필수**: 결과는 `openspec/changes/<change-id>/review.md`에 남긴다 — 날짜, 보이스 구성, 발견 요약, 수용/거절과 근거. 이 파일이 없으면 `make gate`가 실패한다
- **위험 등급 가중**: 주문 실행·위험관리·원장·reconciliation을 건드리는 change는 리뷰 보이스에 반드시 적대적 Eng 관점을 포함한다. UI·문서·도구 change는 경량 리뷰(validate + Manager 셀프리뷰 + 기록)로 충분하다

## 완료 게이트 (자동화)

`make gate CHANGE=<change-id>` = tasks.md 미완료 체크박스 0 + review.md 존재 + `make test` + `make vet` + `make validate` 전부 통과. 규율이 아니라 스크립트가 게이트다(tools/gate.sh).

## 불변 규칙

- 주문 실행은 토스 공식 Open API 경로만 사용. WTS는 조회 전용. 엔진 배선은 official-only 브로커임을 테스트로 증명해야 한다
- 토스 계좌가 포지션의 최종 권위. 불일치 시 신규 진입 차단, 청산 지속
- **실계좌 보호는 기계적으로**: 자동 테스트는 실계좌 주문 발생 금지. 테스트는 격리된 config 디렉터리(`t.Setenv`로 임시 경로)에서 실행하고, 실 endpoint 접근은 httptest 대체 없이는 금지. 문구가 아니라 테스트 인프라가 막는다
- upstream 테스트 회귀 금지 (베이스라인 650개 green 유지)
- push는 사용자 요청 시에만. upstream push URL은 DISABLED로 고정
- MIT LICENSE·원저작권 고지 유지, 시크릿·세션·로컬 DB(sidecar 포함) 커밋 금지
- 테스트 규율: 기능 커밋에는 해당 기능의 테스트가 같은 change 안에 존재하고 통과해야 한다(요구사항↔테스트 추적 가능). TDD(실패 테스트 선행)를 권장 절차로 하되, 검증은 이 추적성 기준으로 한다

## OpenSpec 적용 범위

- **change 필요**: 신규 기능, 동작 변경, 주문·위험·원장 등 안전 경로의 모든 수정
- **change 불필요**: 오탈자·주석·문서 수정, 리팩터링 없는 의존성 patch 업데이트, 테스트만 추가
- **긴급 경로**: 보안·자금 위험 수정은 즉시 구현 가능하되 24시간 내 사후 change 문서화

## 예외 경로

- **스펙 결함 발견 시 분류**: ① blocking(안전·동작 모순) → 구현 중단, `openspec/changes/<id>/issues.md`에 기록 후 Manager 호출 ② safe local(스펙 의도가 명백한 사소한 보완) → 구현하며 issues.md에 사후 기록 ③ editorial(오탈자) → 즉시 수정
- **막힌 task**: 3회 시도 실패 시 tasks.md에 `[blocked]` 표기 + issues.md 기록 후 다음 task로. WIP는 `wip/<task-id>` 사이드 브랜치에 보존(작업 브랜치에는 실패 상태 커밋 금지)
- **change 폐기**: changes/에서 삭제하고 후속 change proposal에 한 줄 사유 기록
- **동시 작업**: change당 활성 Teammate는 **1명**. 병렬이 필요하면 change를 파일 표면이 겹치지 않게 분할한다

## 브랜치·커밋 규칙

- `main`: TossOS 제품 안정 브랜치
- `upstream-sync`: upstream 선별 반영 전용 브랜치 — 여기서 충돌 해소 후 main으로 merge. 반영 내역은 `docs/upstream-sync-log.md`에 기록
- 작업 브랜치: `feat/p<N>-<change-id>` (예: feat/p1-harden-execution-base). ※ feat/p0-foundation은 규칙 제정 전 생성분으로 유지
- 커밋: upstream 관례를 따라 `type(scope): 제목` + 구현 커밋은 task id 참조 (예: `feat(trading): 주문 상태기계 추가 [T1.4]`)
