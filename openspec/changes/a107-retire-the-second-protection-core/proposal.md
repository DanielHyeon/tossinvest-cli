# a107 — 두 번째 보호 core를 퇴역시킨다

> **상태: 등록만 먼저 했다(2026-08-13).** a100 proposal-freeze 리뷰(2026-08-11)의 두 보이스가
> `internal/protection.Controller`·`Repository` 제거의 **선행**을 권고했고, a100 design D1이
> "등록만이라도 앞당기는 것이 절충"으로 결론냈다(a100 tasks 6.5). 이 문서가 그 등록이다.
> **착수는 a100 land 이후다** — 두 change가 같은 봉인 파일(`internal/protection/dormant_test.go`,
> `a071_security_review_test.go`)의 import walk·허용 목록을 편집하므로 순서를 겹치지 않는다.

## Why

저장소에 브로커 상주 보호의 상태 core가 **두 개** 있다. a100 design D1이 표로 고정한 사실:

| | `internal/protection` | `internal/protectionlifecycle` |
| --- | --- | --- |
| 생성 | 2026-08-01 (a045) | 2026-08-04 (a071) |
| 크기 | `controller.go` 824 + `repository.go` 540 + `domain.go` 531 | `lifecycle.go` 18.6K + `state.go` 8.8K |
| 상태 저장 | `database/sql` — `protection_sagas`, `protection_mutation_attempts` | 없음 (순수 전이) |
| non-test importer | `execgw`, `app/engine/gateway.go` — **domain 타입만** | 0 (a100이 조립한다) |

프로덕션 core는 `protectionlifecycle`이다(D1 확정). `NewController`는 `*Repository`를 요구하고
`NewRepository`는 `*sql.DB`를 요구하므로, Controller를 프로덕션에 넣으면 **두 번째 SQLite가
기동한다** — a100이 정확히 배제한 그것이다.

죽은 core는 공짜가 아니다. 봉인 테스트가 배선 사고를 막고 있지만, 봉인은 그 코드가 존재하는 한
계속 유지 비용을 낸다(a100 4.1~4.3이 그 봉인을 다시 조정해야 했다). **제거가 봉인보다 싸고
확실하다.** 리뷰 두 보이스의 권고가 이것이다.

## What Changes

- **제거**: `internal/protection/controller.go`(824줄), `repository.go`(540줄), 이 둘만 검증하는
  테스트, `protection_sagas`·`protection_mutation_attempts` 스키마 생성 경로.
- **유지**: `internal/protection/domain.go`(531줄) — `execgw`·`app/engine`이 도메인 타입을
  import한다. `readiness_adapter.go`·`reconcile.go` 등 나머지 파일은 착수 시 CodeGraph 소비자
  추적으로 개별 판정한다(등록 시점에 선언하지 않는다 — 손으로 읽은 목록은 볼 곳을 고른다).
- **봉인 재조정**: 죽은 심볼을 참조하던 봉인만 줄이고, a100 4.2가 넓힌 import walk
  (`protectionofficial`·`protectionlifecycle` 포함)는 유지한다. "보호 core는 하나다"를
  정적 가드로 고정한다.

## 번호에 관한 기록

a104(사이징 역산)·a105(`Wired` 생산·레인 활성화)·a106(전략적 실계좌 주문 1회)은 a100의 동결
문서(proposal·design·review)가 예약한 번호이므로 비워 둔다. 이 change가 a107인 이유다.
a100 문서의 "threshold 승인(a101)·라이브 평가(a103)" 참조는 이후 다른 내용으로 land된
a101(soak autostart)·a103(rollback pin)과 충돌하는 낡은 참조다 — 동결 리뷰 문서는 수정하지
않고 여기 기록으로 남긴다.

## Impact

- Affected specs: `protection-orders` (ADDED 1 — 프로덕션 보호 core는 하나다)
- Affected code: `internal/protection` 축소, 봉인 테스트 재조정
- 위험: High-risk 영역(보호 경로) 소속이나 제거 대상은 unwired 코드다. FLM·Pre-Edit 선언은
  착수 시 적용한다(등록은 코드를 건드리지 않는다).
