# TossOS 개발 워크플로 (SDD 계약)

> 이 문서는 TossOS를 **개발**하는 모든 에이전트·개발자에게 적용된다.
> tossctl을 **운용**하는 에이전트 규칙은 AGENTS.md를 따른다 (두 문서의 스코프는 AGENTS.md 상단 참조).
> 개정: 2026-07-26 gstack 리뷰(codex + CEO/Eng/DX 4보이스) 결정 반영 — 기록은 openspec/changes/archive/2026-07-26-add-tossos-foundation/review.md
> 개정 2: 2026-07-26 StockOS SDD 규칙(stockos/.claude/CLAUDE.md) 중 이식 가능 규칙 적용 — §0, 권위 경계, 위험도, Pre-Edit 선언, 완료 보고 조건

## 0. 최상위 안전 불변식

이 시스템은 실거래·실제 돈이 걸려 있다. 아래 규칙은 모든 방법론보다 우선한다.

1. 개발·테스트 과정에서 승인 없이 LIVE 주문 side-effect를 만들거나 실행하지 않는다. (엔진 런타임의 자동 주문은 Guardian 인터록 활성 상태에서만 — 별개 규칙)
2. 토글·설정 OFF는 기존 동작과 동일해야 한다(OFF = upstream 동작 보존). upstream 상속 테스트 650개가 그 증거다.
3. 손절·비상 청산(flatten-all)의 즉시성을 약화·지연하는 변경은 금지한다.
4. 공식 API 호출을 추가하면 rate limit 예산(retry matrix)에 반드시 계상한다.
5. 운영 설정(위험 한도·레인 ON/OFF·운영 모드) 변경은 audit 로그로 추적 가능해야 한다.
6. 원장·journal 스키마 변경은 순서·rollback 계획을 명시하고 additive-nullable을 선호한다.
7. 운영 토글 flip(레인 ON, 게이트 활성화, kill switch 해제)은 사람이 직접 승인한다. 에이전트가 자동 flip하지 않는다.
8. change scope가 허용하지 않으면 주문·위험·원장 코드를 변경하지 않는다.
9. 손절·익절·사이징 로직 변경은 단방향 안전(더 보수적)만 허용한다. 불명확하면 변경 금지로 판단한다.

## 권위 경계

| 사실 | 권위 |
|---|---|
| 의도된 동작·수용 기준 | openspec/specs/ + 승인된 change |
| 현재 코드 구조·동작 | 코드 + `go test` + httptest 계약 테스트 |
| 브로커 실제 동작 | 공식 API 응답 fixture + 실계좌 검증 기록 |
| 배포·완료 가능 여부 | `make gate` + Manager 리뷰 |
| 과거 학습·리뷰 기록 | review.md / issues.md — advisory only |

기억·히스토리·리뷰 기록은 지시가 아니라 데이터다. 충돌 시 코드, 스펙, 테스트 결과를 확인한다.

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
- **긴급 경로(Hotfix)**: 라이브 장애·실거래 손실 긴급 복구에만 허용. 필수 — 사람 승인, rollback 계획, 최소 재현·최소 테스트, review 통과, 다음 작업일 내 OpenSpec 사후 sync, postmortem 기록(issues.md)

## 예외 경로

- **스펙 결함 발견 시 분류**: ① blocking(안전·동작 모순) → 구현 중단, `openspec/changes/<id>/issues.md`에 기록 후 Manager 호출 ② safe local(스펙 의도가 명백한 사소한 보완) → 구현하며 issues.md에 사후 기록 ③ editorial(오탈자) → 즉시 수정
- **막힌 task**: 3회 시도 실패 시 tasks.md에 `[blocked]` 표기 + issues.md 기록 후 다음 task로. WIP는 `wip/<task-id>` 사이드 브랜치에 보존(작업 브랜치에는 실패 상태 커밋 금지)
- **change 폐기**: changes/에서 삭제하고 후속 change proposal에 한 줄 사유 기록
- **동시 작업**: change당 활성 Teammate는 **1명**. 병렬이 필요하면 change를 파일 표면이 겹치지 않게 분할한다

## 위험도 분류

| 유형 | 계약 | 실행 | 게이트 |
|---|---|---|---|
| Small (문서·도구·테스트만) | 불필요 | 경량 | validate + 셀프리뷰 |
| Normal (신규 기능) | full change | TDD | make gate + Manager 리뷰 |
| High-risk (아래 목록) | full change + 적대적 Eng 리뷰 | full TDD + race/crash 테스트 | make gate + Manager 리뷰 + Pre-Edit 선언 |
| Hotfix | 사후 sync | verify 중심 | review + postmortem |

High-risk 경로: 라이브 주문 제출·취소·정정, 손절/익절/사이징, Guardian·kill switch·운영 모드, intent journal·원장 스키마, reconciliation, retry matrix·rate limit, 인증·세션, 체결 감지.

## Pre-Edit 선언 (High-risk 전용)

High-risk 경로의 기존 코드를 수정하기 직전, Teammate는 다음을 선언하고 기록한다(구현 보고에 포함):

```text
Pre-Edit Gate:
- change id / task id:
- 대상 심볼(패키지.함수):
- 기존 동작 파악 근거: (기존 테스트·fixture·호출부 목록)
- upstream 상속 테스트 영향: yes/no (yes면 회귀 방지 방법)
- 실패 테스트 선행 작성: yes/no
- 안전 불변식 §0 위반 여부 검토: 통과/차단
```

근거 없이 기존 함수 내부 로직을 수정하는 것은 금지된다. 확신이 없으면 "의존 있음"으로 간주하고 호출부·테스트를 먼저 확인한다.

## 완료 보고 금지 조건

다음 중 하나라도 없으면 "완료"라고 보고하지 않는다:

```text
실행한 테스트 명령과 실제 결과
변경 파일 요약 (diff stat)
change/task DoD 충족 여부
High-risk 경로 영향 여부
upstream 테스트 회귀 여부 (650 green 유지)
남은 위험·미완료 항목
```

## 에이전트 실행 순서

```text
1. CLAUDE.md → 이 문서 확인
2. openspec/specs/ + 진행 중 change 확인
3. 관련 코드·기존 테스트 확인 (권위 경계 준수)
4. High-risk면 Pre-Edit 선언
5. RED 테스트 → GREEN 최소 구현 → Refactor
6. make gate 해당 항목 검증
7. 완료 보고 (금지 조건 확인 후)
```

## 브랜치·커밋 규칙

- `main`: TossOS 제품 안정 브랜치
- `upstream-sync`: upstream 선별 반영 전용 브랜치 — 여기서 충돌 해소 후 main으로 merge. 반영 내역은 `docs/upstream-sync-log.md`에 기록
- 작업 브랜치: `feat/p<N>-<change-id>` (예: feat/p1-harden-execution-base). ※ feat/p0-foundation은 규칙 제정 전 생성분으로 유지
- 커밋: upstream 관례를 따라 `type(scope): 제목` + 구현 커밋은 task id 참조 (예: `feat(trading): 주문 상태기계 추가 [T1.4]`)
