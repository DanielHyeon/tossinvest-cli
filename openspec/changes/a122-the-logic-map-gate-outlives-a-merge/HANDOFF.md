# a122 인계 — 다음 세션이 먼저 읽는 문서

작성 2026-09-18. 대상은 이 change 를 이어받는 다음 세션과, 아래 **사람 결정 두 개**를
정해야 하는 사람이다.

**한 줄: 섹션 7 이 전부 닫혔다.** 남은 일은 **6.4 한 묶음**(그 안에 사람 결정 하나)과
**종결 3종**(1.4 · 3.2.3 · 4.5)이다. 6.1 은 자식 여덟이 다 닫혔고 부모 체크박스만 남았다.

---

## 0. 상태 (사실만)

- 브랜치 `feat/a112-four-family-runtime`.
- **푸시된 tip: `8091e6c4`.** 그 위에 **미푸시 커밋 둘**:
  - `e34e9e2c` `perf(a122): 착지 walk 가 blob 을 한 프로세스에 하나씩 읽지 않는다 [a122 7.5 · 7.9 · 7.3]`
  - `1764c833` `chore(codegraph): affected 의 Go 사각지대를 적고 색인 갱신자를 되살린다`
    — **이 change 것이 아니다.** 2026-09-18 **다른 세션**이 같은 워크트리에서 커밋했다.
  푸시는 사람이 요청할 때만 하고, 브랜치를 밀면 **둘 다** 올라간다. 이어받는 세션이 다른
  기계에 있으면 먼저 푸시를 요청할 것.
- **이 워크트리는 병행 세션이 쓴다.** 작업 전에 `git log --oneline -3` 으로 tip 이 움직였는지
  확인하고, 내 것이 아닌 변경은 커밋에 담지 말 것
  ([[tossos-parallel-session-gate-contention]]). 2026-09-18 한 시간 안에 실제로 두 번 겹쳤다.
- 마지막 검증(커밋 후 실행): `check_analysis --change a122` **rc=0** ·
  `openspec validate --all --strict` **58/58** · `make sdd-test`(logic-map **257**) ·
  `make lint` · `make test-seams` · `test_check_analysis` **182**.

## 1. 무엇을 읽나 (순서대로)

1. 이 문서.
2. `tasks.md` — 열린 줄만 본다. 닫힌 줄의 본문은 **그 task 가 실제로 한 일**이 적혀 있다.
3. `review.md` 의 **뒤에서부터**. 절이 앞 절을 정정하므로 diff 해서 상태를 재구성하지 말 것.
   가장 최근 셋: `## MEASURE · Pre-Edit Gate — task 7.5` · `## VERIFY — task 7.5` ·
   `## VERIFY — task 7.9`.
4. `analysis/harness/README.md` — 측정·변이를 재현하는 법.
5. `analysis/python-function-logic/` — 편집한 함수마다 `ast.before-<task>.json` /
   `ast.after-<task>.json` 과 map 둘.

## 2. 남은 일

### 6.1 — 부모 체크박스만 (작다)

자식 여덟(6.1.1 · 6.1.2 · 6.1.2.1~6.1.2.6)이 전부 `[x]` 다. 부모가 적은 P0(저자가 만든 값으로
저자가 고른 값을 검증하는 순환)은 6.1.2 가 값을 도구로 옮기고 7.2.2 · 7.3.1 이 더 조였다.
**7.3 부모를 2026-09-18 에 닫은 것과 같은 모양이다.** 닫기 전에 부모 본문의 위조 표 셋이
지금도 빨간지 `6.1.2.4` 의 기록과 대조할 것.

### 6.4 — P2 묶음 (a)~(f), 각각 작고 독립

**(a) 가 사람 결정이다.** 착지를 선언하면 커밋 안 된 Go 편집이 5단계에 안 보인다. 이관
경로가 이미 쓰는 `git diff --quiet` 를 착지 경로에도 둘지. 넣기 전에 **거부할 정상 입력을
먼저 열거**할 것 — 게이트를 더러운 트리에서 돌리는 것은 개발 중 정상이다
([[fail-closed-must-name-what-it-rejects]]; 7.7 이 같은 자리에서 dirty 를 **거절 순서 맨 뒤**로
옮긴 이유도 이것이다).

(b)~(f) 는 결정이 필요 없다: `base-commit.txt` 를 워킹트리에서 읽고 40-hex 검사가 없다 ·
`ARCHIVED_CHANGE` 의 `\d` 가 유니코드 숫자를 먹는데 `gate.sh` 의 `[0-9]` 는 안 먹는다(`re.ASCII`
한 글자) · `check` 의 `except ValueError` 가 `archive holds N copies` 를 삼킨다 · `gate.sh:110`
주석과 `test_check_analysis.py` docstring 이 지워진 코드를 현재형으로 인용한다 · 규칙이 두 집에 산다.

### 종결 3종

- **1.4** proposal-freeze 적대 리뷰 + gstack 리뷰 기록. (§6 로트 리뷰는 `## VERIFY — §6 로트
  독립 리뷰` 에 있고, 그 뒤 7 절 전체가 아직 독립 리뷰를 안 받았다.)
- **3.2.3** 인용된 테스트 이름의 탐색 범위 결정.
- **4.5** PM 동기화 후 `make gate CHANGE=a122-…`. **순서는 체크 → tracker 재생성 → sdd-sync →
  gate → archive → Story 경로**다.

### 5 절은 작업이 아니다

5.1~5.7 은 이 change 가 **열어 두는 한계 선언**이다. 구현할 것이 없고, 아카이브할 때 같이 닫는다.

## 3. 사람이 정해야 하는 것

1. **6.4(a)** — 착지 경로에 더러운 트리 거절을 둘지. (위 참조)
2. **푸시** — `e34e9e2c`(이 change)와 `1764c833`(다른 세션)이 아직 안 올라갔다.
   브랜치를 밀면 둘 다 간다.

## 4. 이어받는 세션이 밟기 쉬운 지뢰

- **`make lint` 은 Python 을 안 본다** (`go vet` 둘뿐). 이 change 가 만지는 것은 거의 전부
  Python 이다 — 판정은 `test_check_analysis`(182)와 `analysis/harness/75_mut.py` 뿐이다
  ([[missing-tool-reports-clean]]).
- **`make gate` 는 change **완료** 게이트라 진행 중 로트엔 못 쓴다.** 로트 검증은
  `lint` · `test` · `test-seams` · `sdd-test` · `check_analysis --change a122` ·
  `openspec validate --all --strict`.
- **Python 내부를 고치면 FLM 이 먼저다.** 게이트가 Go 만 요구해서 7.2.1 · 7.3.1 · 7.4 가
  **세 task 연속 조용히 빠졌고** 7.6 이 소급해 메웠다. `analysis/python-function-logic/enumerate.py`
  로 편집 **전** AST 를 먼저 뽑을 것 ([[flm-before-claiming-not-before-editing]]).
- **`_landing_refusal` 은 순서가 못이다.** 가드를 옮기면 뒤의 가드들이 못에서 빠진다
  ([[a-new-guard-unpins-the-guards-behind-it]]). 편집 전후 AST 의 분기·반환 수가 같은지 반드시 대조할 것.
- **변이가 살아남으면 그것이 발견이다.** 7.5 에서 첫 판 생존 넷은 전부 "시험이 그 갈래에 안
  닿았다"였다. "동등 변이"로 넘기기 전에 **무엇이 덮고 있는지를 변이로 증명**할 것 — 7.5 의
  T6 은 T18(사전 채움 삭제 → 59개 빨강)로 증명했다 ([[surviving-mutant-may-mean-accidental-safety]]).
- **`75_ab.py` 의 비교 기준을 `HEAD` 로 바꾸지 말 것.** 수리가 랜딩한 뒤엔 대조군이 오염된다.

## 5. 7.5 가 남긴 성능 사실 (다시 재지 말 것, 필요하면 하네스로 재현)

전수 A/B **SAME 116 · DIFFERENT 0**, 합계 **2023.0s → 162.1s (12.5×)**, 최대 36.9×.
비용은 **아무 후보도 안 받는 walk** 에만 있었고 spawn 의 97% 가 blob fetch 였다.
남은 병목은 없다 — 프로파일에서 `compute_landing` 이 a071 3.69s · spawn 697 이다.
`--ancestry-path`(후보 **집합**을 바꾼다)와 조기 종료(거절 문장이 고칠 자리를 **전부** 대야 한다)는
**의도적으로 안 했고** 사유는 `## VERIFY — task 7.5` 의 "안 한 것과 사유" 에 있다.
