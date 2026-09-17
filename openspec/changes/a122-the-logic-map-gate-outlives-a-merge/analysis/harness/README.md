# a122 측정·검증 하네스

**이 저장소의 도구가 아니라 이 change 의 일회용 하네스다.** 커버리지나 안전을 주장하지
않는다. 여기 있는 이유는 하나다 — `branch-test-map.md` 와 `review.md` 가 이 스크립트들의
출력을 **증거로 인용**하는데, 예전 task 들의 하네스(`722_mut.py` · `726_mut.py` · `76_mut.py`)는
세션 스크래치패드에 살다가 사라져서 **인용만 남고 재현이 불가능**해졌다.

루트는 세지 않고 **유도**한다 — 체크아웃 이름이 바뀌어도 고아가 안 된다
([[renamed-checkout-strands-absolute-path-state]]). 런타임 산출물은 `_work/` 이고 커밋하지 않는다.

| 파일 | 무엇을 재나 | 실행 |
|---|---|---|
| `75_census.py` | change 마다 후보 수 · 번들 수 · 예측 spawn (전수, 걷지 않는다) | `python3 75_census.py` |
| `75_levers.py` | 지렛대 넷의 단가 (`show` 하나씩 · `cat-file --batch` · `ls-tree` · 번들 재파싱) | `python3 75_levers.py` |
| `75_attrib.py` | spawn 을 **호출자 함수**로 귀속 + 한 개짜리 fetch 두 방식 | `python3 75_attrib.py <change-dir-name>…` |
| `75_time.py` | `compute_landing` 의 벽시계와 spawn 수 | `python3 75_time.py <change-dir-name>…` |
| `75_ab.py` | 판정 A/B 전수 (사본 대조군 + 이어 달리기) | `python3 75_ab.py [<before-sha>] [<초 예산>]` |
| `75_mut.py` | 변이 T1~T18 (사본 대상 · 무변이 대조군 선행) | `python3 75_mut.py` |

## 읽는 법 두 가지

**`75_ab.py` 의 비교 기준은 인자다.** 기본값은 7.5 직전 커밋(`8091e6c4`)이다. `HEAD` 로
굳혀 두면 수리가 랜딩한 다음 날 "before" 사본에 수리가 들어가 대조군이 조용히 오염된다 —
그래서 사본에 `_committed_many` 가 있으면 **멈춘다**
([[a-recorded-boundary-stops-being-rechecked]]). 한 번에 다 못 도므로(before 쪽 합계 약 2000초)
초 예산을 주고 여러 번 부르면 `_work/75_ab_done.json` 에 이어 달린다.

**하네스는 도구를 따라 낡는다.** `75_census.py` 는 `_walk_floor` 가 7.5 에서 값 셋을 돌려주게
되자 전수 126건을 전부 "못 걸음"으로 찍었다 — 빈 결과가 발견처럼 보였다
([[missing-tool-reports-clean]]: 0 은 "위반 0"이 아니라 "검사 0"이다). 여기 있는 것들을
쓰기 전에 **작은 표본 하나로 먼저 돌려 보고** 숫자가 말이 되는지 확인할 것.
