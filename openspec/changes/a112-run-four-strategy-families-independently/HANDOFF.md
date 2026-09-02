# a112 5.5 인계 — 다음 세션이 먼저 읽는 문서

작성 2026-09-01, 갱신 2026-09-02. 대상은 이 change 를 이어받는 다음 세션과, 아래
남은 셋을 정해야 하는 사람이다.

한 줄: **5.5 의 backstop 굳히기는 적대 리뷰 13 라운드 끝에 APPROVE(P0/P1/P2 = 0)로
끝났고, 2026-09-02 에 세 커밋으로 랜딩했다.** 사람이 정해야 하는 것 다섯 중 셋
(커밋 단위·a117 스텁·CI 배선)은 답을 받아 처리했고, 둘이 남았다.

상세는 `review.md` 의 `## 5.5 현재 상태` 블록(fix 절들보다 **앞**에 있다)과
`analysis/function-logic/internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority/branch-test-map.md`
의 B3 절에 있다. fix3~fix13 열한 절은 **방법 기록**이지 현재 상태가 아니다 —
뒷 절이 앞 절을 정정하므로 diff 해서 상태를 재구성하지 말 것.

---

## 1. 랜딩한 것 (2026-09-02)

세 커밋으로 나눠 올렸다. 사람이 "3 커밋 분할"을 골랐다.

| 커밋 | 무엇 |
|---|---|
| `14b76963` | 신규 시험·seam 5 + 수정 시험 4. 행동(시험 함수 넷/하위 여섯) · 구조 단언 하나 · 필드 완전성 32+8 |
| `3183e22b` | `tools/logic-map/{role_check,check_analysis,test_role_check,test_check_analysis}.py` — 좌표 역할 대조와 열거 강제 |
| `80d3931d` | 병행 세션 스텁 a117 → a119 renumber + PM 레코드 (아래 2 절 (4)) |
| `364190c8` | `review.md` · `tasks.md` · a112 분석 번들 54 · 이 문서 |
| 이 커밋 | CI 배선 — `sdd-check-ci` 타깃 · `sdd-checks` job · 갈라짐을 막는 시험 (아래 2 절 (2)) |

랜딩 직전 실측: `make lint` PASS · `make test` 98 ok/0 FAIL · `make test-seams`
99 ok/0 FAIL · python 77 OK · `check_analysis --change a112` evidence complete ·
`openspec validate a112 --strict` valid · `gofmt -l internal/ tools/` 무출력.

## 2. 사람이 정해야 하는 것 — 리뷰로는 못 정한다

(2)·(3)·(4) 는 2026-09-02 에 답을 받아 닫았다. (1)·(5) 가 남아 있다.

### (1) `dispatchHandoff` 위조를 연 채로 5.5 를 내보낼 것인가 ★ 가장 무거움

13 라운드가 굳힌 것은 **backstop 이지 seam 이 아니다.** `dispatchHandoff` 는 패키지
사설 구조체의 메서드라, 엔진 안의 아무 함수나 자기가 고른 entries 로 수신자를 만들어
봉투를 주조할 수 있다. 4 차 리뷰가 그런 우회 넷을 컴파일해 보였다.

그것들이 주문으로 안 이어지는 이유는 경계가 아니라 `strategy_account_first_leg_authority.go`
의 재유도·identity 대조(`:217`, `:221`–`:222`, `:223`–`:225`)다. 그 다섯 줄이 유일한
방어라는 뜻이다. 진짜 해법(`Admit` 이 조정자만 만들 수 있는 봉인 값을 받는 것)은
이 task 의 파일 소유 밖이라 **6.2 의 차단 선결조건**으로 걸어 두었다.

**결정할 것:** 이 상태로 5.5 를 완료로 볼지, 아니면 6.2 전에 seam 을 먼저 닫을지.

### (2) CI 를 배선할 것인가 — **닫혔다 (2026-09-02), 절반은 배선하고 절반은 이름만 적었다**

배선했다. `.github/workflows/ci.yml` 에 `sdd-checks` job 이 생겼고 `make sdd-check-ci`
하나를 부른다. 그 타깃은 `sdd-check` 에서 **러너로 옮길 수 없는 둘만 뺀** 부분집합이다.

옮길 수 없다는 것은 측정한 것이다 — 외부 도구를 지운 PATH 와 depth-1 클론에서 각각
exit 1: `sdd-doctor` 는 로컬 설치 CLI(rtk·openspec·codegraph·gbrain)를 보고,
`check_index_freshness.py` 는 codegraph 실행 파일과 gitignore 된
`.sdd/index-state.json` 을 본다. 나머지는 같은 환경에서 전부 exit 0 이었다.

이제 CI 가 도는 것: 파이썬 시험 180개(그중 `tools/logic-map` 77개가 열거표 감사 —
`check_analysis.py`·`role_check.py`) + `go test ./tools/logic-map` + 에이전트 부트스트랩
동기화 + 기억 원장 + PM 번호 계약 + `compileall`.

목록이 두 곳으로 갈라지는 것은 시험이 막는다
(`tools/sdd/test_ci_runs_portable_sdd_checks.py`, 뮤테이션 5종으로 반증 확인).
`sdd-check` 에 옮길 수 있는 검사를 직접 한 줄 더하면 실패한다.

**안 배선한 것 — `check_analysis.py --change <id>` 자체.** 그 검사는 번들을 유도 당시
소스에 묶으므로 change 하나의 **완료** 게이트(`make gate` 5/10)에서만 참이다. 측정:
활성 change 31개 중 통과 1개(a112). 나머지 30개는 **AST 소스 해시 stale 15 · 넓어진
수정 집합의 FLM 누락 11 · base-commit 누락/무효 4**. 저장소 전체로 켜면 첫날부터
빨갛고, PR 이 건드린 change 만 골라 켜도 a113~a115 를 잡는 세션이 남의 빚으로
막힌다. 이것은 (5) 의 구조적 빈틈과 같은 뿌리다.

### (3) 커밋·푸시 시점 — **닫혔다 (2026-09-02)**

세 커밋으로 나눴다(1 절). 푸시는 아직 안 했다.

### (4) 외부 a117 스텁의 주인 — **닫혔다 (2026-09-02), 다만 절반만**

`a117-codex-session-handoff-and-gbrain-startup` → `a119-…` 로 renumber 했다.
번호만 바꾸면 `active change has no Story` 가 남으므로 PM 레코드도 함께 만들었다:
`docs/pm/portfolio/stories/STORY-TOS-a119.yaml`(FEAT-TOS-001), `_registry.yaml`,
`FEAT-TOS-001.yaml`, 그리고 재생성한 `docs/pm/generated/` 3 파일.
Story 의 acceptance 는 **그 세션의 proposal.md 에서 그대로 옮겨 적은 것**이고,
`found_by` 에 누가 왜 만들었는지 적었다. 내용의 주인은 그 세션이다.

`make sdd-check` 는 이제 통과한다. **남은 것:** 그 스텁은 `specs/` 델타가 없어
`openspec validate --strict` 를 통과하지 못한다(`make validate --all` 이 1 건
실패). 그 세션이 델타를 채워야 한다. 그리고 그 세션이 스스로 다른 번호로
renumber 하면 디렉터리가 둘로 갈릴 수 있다.

### (5) 닫지 않고 이름만 적은 둘을 받아들일 것인가

- `revision: base` 번들 **15 개**가 소스 해시에 묶이지 않는다(13 개는 이 change 의
  `base-commit.txt` 와도 불일치). "133/133 이 열거한다"는 **표 모양의 참**이지 좌표가
  옳다는 뜻이 아니다. 독립 `go/parser` 재유도로 확인된 것은 118 개다.
- `internal/strategyflow` 에 exported 표면 census 가 없다. 이제 봉인된 `Result` 를
  만드는 함수가 태그 아래 **둘**이다(`AcceptedResultForAuthorityTest`,
  `ResultWithRestatedStopProvenanceForTest`). `strategyhandoff` 가 4 차 전에 있던 상태와 같다.

## 3. 다음 세션이 할 일 — 이 순서

1. ~~위 (3)·(4) 에 대한 사람의 답을 먼저 받는다.~~ **끝났다** — 3 커밋 + a119 renumber.
2. ~~커밋 → `make sdd-sync` 재실행 → `make sdd-check`.~~ **끝났다.** fingerprint 는
   sync 시점 HEAD 를 기록하므로, 이 뒤로 커밋이 붙으면 다시 `make sdd-sync` 를 돌린다.
3. 그다음 열려 있는 태스크로 간다. 미완료 44 / 완료 30. 5 절에서 남은 것은
   **5.1 · 5.1.1 · 5.2 · 5.3 · 5.6 · 5.7**, 그다음이 6 절이다.
   - 5.1.1 이 5.5 가 `not-applicable` 로 미룬 것을 갖고 있다: 워커의 `Cycle` 이 broker
     mutation 에 못 닿는다는 증명. 지금 production 워커의 cycle 은 `*Context` 클로저라
     Journal/Gateway/Guardian 을 들고 있다.
4. **6.2 를 여는 사람은 tasks.md 6.2 본문을 먼저 읽는다.** 아래 4 절이 그 요약이다.

## 4. 6.2 를 여는 사람에게

6.2 는 `strategy_account_first_leg_authority.go` 의 다섯 줄을 지우거나 바꾸게 된다.
**세 가지가 그 자리를 잡고 있고, 셋 다 초록으로 유지해야 한다.**

| 무엇을 증명 | 어디 |
|---|---|
| 이 가드가 돌고 거절한다 | `strategy_first_leg_identity_backstop_test.go` — 시험 함수 넷(하위 여섯) |
| 이 가드가 봉인된 identity 를 **그대로** 비교한다 | `strategy_first_leg_backstop_shape_test.go` — 구조 단언 하나 |
| 그 identity 가 **모든 필드**를 담는다 | `strategyflow/execution_terms_identity_fields_test.go` (32) + `weeklyvaluelane/execution_policy_identity_fields_test.go` (8) |

**구조 단언은 6.2 첫날에 정당하게 걸린다.** 그것이 `proposal.entries[0].authority` 를
못 박고 있는데, 6.2 의 일이 한 종목에 네 가족을 받는 것이라 `entries[0]` 이 정당하게
틀려진다. 그때:

- **기대 문자열만 고쳐서 통과시키지 말 것.** 실패 메시지에 그 이유를 적어 두었다.
- 여러 항목 중 하나를 고르게 바꾼다면 **그 선택이 `accepted` 에 기대면 안 된다.**
  accepted 와 맞는 항목을 골라 놓고 그것을 accepted 와 비교하면 가드가 **자기 참조**가
  되고, 행동 시험도 census 도 전부 초록으로 남는다(12 차 리뷰가 재현).
- 오늘 그 편집이 무해한 이유는 `:217` 의 `len(entries) != 1` 이 항목을 하나로 묶기
  때문인데, **6.2 가 바로 그것을 바꾼다.**

## 5. 검증 명령 (그대로 복사)

```
make lint                                   # PASS
make test                                   # 98 ok / 0 FAIL
make test-seams                             # 99 ok / 0 FAIL
python3 -m unittest discover -s tools/logic-map -p 'test_*.py'   # 77 OK
python3 tools/logic-map/check_analysis.py --change a112-run-four-strategy-families-independently
openspec validate a112-run-four-strategy-families-independently --strict
$(go env GOROOT)/bin/gofmt -l internal/ tools/    # 출력 없음
make sdd-sync && make sdd-check              # PASS (a119 renumber 뒤)
make sdd-check-ci                            # CI 가 도는 부분집합 — 로컬 도구 없이도 PASS
```

`make gate CHANGE=…` 는 **not-applicable** 이다 — change **완료** 게이트라 미완료
태스크가 있으면 2/10 에서 멈춘다. 진행 중 로트의 판정에 쓰지 말 것.

## 6. 함정 (여기서 실제로 밟은 것들)

1. **`go test -overlay` 는 소스를 읽는 시험에 안 보인다.** 구조 단언은
   `parser.ParseFile` 로 디스크에서 읽으므로 overlay 뮤테이션이 전부 GREEN 으로 나온다.
   그것은 "가드가 강하다"가 아니라 **"안 쟀다"** 이다. 파일을 진짜로 바꾸고 해시
   백업에서 복원할 것. `git checkout` 은 커밋 안 된 이 로트 전체를 지우므로 금지.
2. **뮤테이션 범위에 `_test.go` 를 넣지 말 것.** 패키지 전체 리네임을 돌렸더니 시험
   파일까지 함께 바뀌어, 뮤테이션이 자기를 잡아야 할 시험을 **수리**하고 초록을 냈다.
3. **커밋하면 `make sdd-sync` 를 다시 돌린다.** fingerprint 는 sync 시점 HEAD 를
   기록한다.
4. **`rtk` 가 출력을 요약한다.** 정확한 개수를 세려면 `rtk proxy <cmd>`.
   `rg` 는 PATH 에 없다(`grep -rn` 사용), `gofmt` 는 `$(go env GOROOT)/bin/gofmt`.
5. **`make sdd-check` 는 이제 통과해야 한다.** 실패 줄이 생기면 그것은 이 로트나
   병행 세션의 문제다. `make validate --all` 은 a119 스텁의 델타 부재로 1 건 실패가
   남아 있는데, 그것은 그 세션의 것이다.
