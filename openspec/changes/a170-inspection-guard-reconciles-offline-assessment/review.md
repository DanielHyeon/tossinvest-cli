# a170 review

Function Logic Map: not-applicable — 생산 함수 편집 0. 바뀐 것은 시험 파일 `internal/campaignstateevidence/authority_guard_test.go` 의 가드 함수뿐이다
(docs/WORKFLOW.md 「비례 원칙」: 시험 전용 · 게이트 판정 변경 0).

## 1. 원인 조사 · 판정 (2026-10-23)

- 사용자 결정(Manager 전달, 2026-10-23): a140 완료 gate 를 막는 이 실패를 작은 수리 change 로 고친다. 이 change 에 대해 `make test` 실행 승인.
- 판정: 의도된 검사 전용 · 권한 없는 사용(근거는 proposal 「Why」) → 허용 목록에 그 파일 하나를 이름으로 추가하고, 그 패키지의 비시험 import 를 거절.
  판단이 갈리지 않아 멈추지 않았다.

## 2. 경량 코드 리뷰 (2026-10-23)

- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only, base `2bbee345f` + 미커밋 diff). 판정: P1 0 · P2 0 · P3 2.
- 확인: 원인 서사 정확(a162 design:18 · review:34 · :118, `cd5f27219` 가 `3f8e34632` 의 조상, 비시험 importer 0). 조기 반환이 import 루프 **뒤**라
  이 파일에도 lane 점 import 거절 · `offlinecontinuation` import 거절이 그대로 적용. 생산 diff 0. 리뷰어 직접 실행 `go test ./internal/campaignstateevidence/` ok.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P3 | 조기 반환의 **위치**를 못 박는 시험이 없음 — 루프 앞으로 옮기면 평가 파일이 lane 점 import 로 캡처를 불러도 통과(추론) | **수용** — 시험 단언 추가 + 변이 M5(아래 영수증) |
| 2 | P3 | spec 문구가 가드 범위(`internal`·`cmd`·`tools`, 한정 selector)보다 넓음 | **수용** — 범위로 좁히고 a161 부터의 한계를 한 줄 명시 |

## 3. 검증 (2026-10-23)

- RED: `analysis/red/red.txt`(컴파일 RED → 상수만 더한 행동 RED: 기존 실패 + 새 시험 실패). GREEN: `analysis/red/green.txt`(두 패키지 · 태그 빌드 · gofmt, rc 0).
- 변이: `analysis/red/mutation.txt` — M1 · M3 · M4 · M5 CAUGHT, M2 는 1 판 생존(형제 파일이 evidence import 없이 캡처만 부르는 경우를 시험이 안 잼) →
  시험 추가 뒤 CAUGHT. 무변이 대조군 GREEN(두 번).
- `make test` rc=0(ok 115, FAIL 0) · `make lint` rc=0 — `analysis/red/verify.txt`. P3-1 단언은 그 뒤에 더했고 영향 패키지만 다시 돌렸다(green.txt receipt 3).
- 미실행: 완료 gate(`make gate CHANGE=a170-…`) — 승인 범위 밖(tasks 2.3).
