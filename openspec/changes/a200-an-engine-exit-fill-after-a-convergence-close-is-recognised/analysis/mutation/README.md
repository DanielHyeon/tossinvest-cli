# a200 2.5 — 변이 원장

- 하네스 `mutate.py`(커밋됨). 사본은 이 브랜치 커밋을 detached 로 뽑은 별도 worktree(`/home/daniel/tossos-work/a200-mutation`), 시작 sha 단언 · `git status` 깨끗함 단언 · 무변이 대조군 GREEN 확인 · 변이마다 원문 1 회 단언 → 적용 → `go test ./internal/journal -run TestA200` + `./internal/reconcile -run TestA200` → `git checkout` 되돌림 + sha256 단언. 끝나면 사본 `git worktree remove`(원장 끝 줄과 실행 로그 「mutation worktree removed」).
- 닿음: 변이 자리는 전부 편집 뒤 커버 측정에서 실행됨(`../coverage/after-ProjectPosition-blocks.txt` — 판정 함수 각 조건 줄 hit=1, `ProjectPosition` 104.116 · 107.3 hit=1). M3 · M9 처럼 「항상 거짓/참」 으로 바꾸는 변이는 미사용 변수 빌드 실패를 피하는 모양으로 씀.

| 판 | 사본 HEAD | 결과 | 원문 |
|---|---|---|---|
| 1 회차 | `8560c2622` | 16 중 CAUGHT 14 · **SURVIVED 1(M2 「0 조정 정확히 하나」 삭제)** · 빌드 실패 1(M3 — 미사용 변수, 가짜 CAUGHT) | `ledger-run1.md` |
| 처분 | `8d0655294` | M2: 판정 시험 「0 조정 둘」 사례의 둘째 행 prev 가 0 이라 (라)(감소분 0)가 대신 거절하고 있었음 — prev 31 로 고쳐 「정확히 하나」 자체를 잼(범위 안 시험). M3: 변이를 `(!lastZero && false)` 로 다시 씀 | 커밋 메시지 |
| 2 회차 | `8d0655294` | **16 / 16 CAUGHT**, 무변이 대조군 GREEN, 생존 0 | `ledger-run2.md` |
| 3 회차(impl-r1 처분 뒤) | `d939e8795` | M1 앵커를 `recognisableRefusal` 로 갱신, M17(조기 반환 삭제) · M18(「> 0」 삭제) 추가 — 18 중 CAUGHT 17 · **M18 은 빌드 실패(미사용 변수 `positive`) = 가짜 CAUGHT** | `ledger-run3.md` |
| 4 회차(gstack r1 보강 뒤) | `25d9aea2d` | M18 을 `(positive <= 0 && false)` 로 다시 씀, M19(조정 행 정렬 뒤집기 — 두 단계 수렴 시험이 잡음) 추가 — **19 / 19 CAUGHT, 빌드 실패 0**, 무변이 대조군 GREEN, 사본 worktree 제거 | `ledger.md` |

- M1 주석: 조건 (가)는 도달 가능한 상태에서 (나)에 함의된다(CLOSED 에서 거절되는 매도는 `SELL_ON_CLOSED` 뿐 — 전이표 X58~X60). 그래서 (가) 삭제는 원장 시험이 아니라 순수 판정 시험(`TestA200TheVerdictRecognisesTheIncidentShapeAndRefusesEachBrokenCondition`)만 잡는다 — 순수 판정의 조건별 시험이 그 조건을 따로 고정함.
