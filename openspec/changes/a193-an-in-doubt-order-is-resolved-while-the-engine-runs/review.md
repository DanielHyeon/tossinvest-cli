# Review

## 1단계 기록 (구현 팀메이트 Opus, 2026-11-18)

### 기억 회고

- ai-memory(workspace `default`, project `junghoonghae-tossinvest-cli`): 「IN_DOUBT 해소 park UNRESOLVED_IN_DOUBT」 1회. 걸린 것: `gotchas/toss-single-order-read-stays-pending-after-app-cancel.md`
  (a191 — 이 change 와 무관), 나머지는 세션 기록 잡음. 세션 중 해소·park·ctx 취소에 대한 기존 결정·함정 기록 없음.
- Hindsight(bank `tossos-dev-episodes`): 「IN_DOUBT resolver park session runtime resolve ctx cancel」 1회(budget low). 관련: a165 교훈 「FAILED_CONFIRMED latch writer 와
  진입 PLACE writer 의 TOCTOU 는 engine run 의 journal directory flock 단일 writer 로 닫혔고 flatten 은 definitive latch writer 가 아니다」 — D6 에서 flatten 이 별도 프로세스로
  같은 원장을 연다는 사실을 HEAD 로 재확인(cmd/tossctl/flatten.go 의 `journal.Open`). 나머지는 게이트·worktree 교훈.
- 파일 기억 중 적용한 것(현재 HEAD 로 재확인): `surviving-mutant-may-mean-accidental-safety`(D3 — park 쓰기의 ctx 결속이 지키는 안전을 시험·변이 m3 로 못 박음),
  `flm-before-claiming-not-before-editing`(인용 40개 AST 를 문서보다 먼저), `missing-tool-reports-clean`(`codegraph affected` 기본값 1 대 1,280),
  `a166-census-freezes-wiring-function-bodies`(배선을 census 밖 `engineRuntime` 에 둠), `tossos-engine-stop-removes-stoploss`(D3 — 배포마다 park 가 나면 손절 동결),
  `fail-closed-must-name-what-it-rejects`(design §4), `positional-branch-ids-break-hand-renumbering`(FLM 재번호는 정렬로), `existence-check-is-not-a-role-check`.

### 번호·Story·Feature

- 번호: Manager 배정 a193. `docs/pm/portfolio`·`openspec/changes`·archive 에 a193 change·Story 없음(STORY-TOS-a102 의 기록 안 부분문자열 1건뿐).
- `STORY-TOS-a193` → **FEAT-TOS-009**(Exit line truth and position policy lifecycle). 근거: 이 change 는 a094(park·발의 해제·park 원인 critical — FEAT-TOS-009)의
  세션 쪽 짝이고, 해악이 「손절 동결」이다. 지시의 「a094·a192 가 속한 Feature」는 둘(009·005)이라 하나를 골랐다 — **Manager 확인 대상(Q-D).**

### 증거

- CodeGraph 1.6.0: `analysis/code-context/`(worktree 색인, 원문 `codegraph-callers.txt`·`codegraph-affected.txt`).
- CodeGraphContext·GBrain: not-applicable — 그 채널은 안 봤음(`codegraphcontext-context.md`).
- AST: 인용 40개 `analysis/cited-ast/`(전부 rc 0), 편집 대상 2개 번들 `analysis/function-logic/`(pre-edit). 손으로 읽은 것은 `analysis/cited-ast/README.md` 표.
- 운영 저널·엔진 로그·브로커: 읽지 않음(지시상 금지). 운영 사실은 Manager 서술 인용.
- Go 시험: 1단계라 돌리지 않음(지시). `check_analysis` 는 required 0(Go 편집 0).

### Pre-Edit 선언 (초안 — 2단계 착수 전 Manager 결정 반영 뒤 확정, tasks 2.1)

```text
Pre-Edit Gate:
- change id / task id: a193-an-in-doubt-order-is-resolved-while-the-engine-runs / 3.11
- 대상 심볼(패키지.함수): main.engineRuntime(보조 실행자 슬라이스 원소 1), 시험 main.TestProductionRuntimeStartsExactlyOneAlertDeliverer(기대 목록 원소 1).
  새 파일 internal/app/engine/in_doubt_settle.go(Context.InDoubtSettler · inDoubtSettle.Run/sweep/noteStale/onStop — 새 함수만).
  무편집: execgw.Resolver 전체, reconcile.Recovery, ExitObserver(submit·noteHeldProposal), journal, internal/obs, a166 census 네 함수.
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md(Resolve production 호출자 2, PendingAttempts 3, engineRuntime production 1;
  affected --filter '*_test.go' 1,280 / -d 2 138)
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(사유 기록), evidence-reconciliation.md(불일치 0)
- 기존 동작 파악 근거: design §1(cited-ast 40), 기존 시험 TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess · a098 두 시험 · execgw indoubt 시험
- Function Logic Map / Branch Test Map: analysis/function-logic/cmd-tossctl--engineruntime, cmd-tossctl--testproductionruntimestartsexactlyonealertdeliverer
- upstream 상속 테스트 영향: no(엔진·cmd 엔진 명령은 TossOS 계층)
- 실패 테스트 선행 작성: yes(tasks 3.1~3.10)
- 설정·DB·journal 변경과 rollback: 스키마·설정·토글 변경 없음. 롤백 = 이전 이미지.
- 안전 불변식 §0 위반 여부 검토: 통과(예정) — 손절 경로 무변경·비대기(§0.3, 시험 3.6), 새 API 엔드포인트 없음·빈도 계상(§0.4, tasks 4.6), 스키마 무변경(§0.6),
  토글 없음(§0.7), 손절·사이징 로직 무변경(§0.9). 새 자동 동작은 조회만(주문 0).
```

### 리뷰 보이스 계획

- 작성자: Claude(Opus). 바깥 보이스: Codex(다른 계열), `codex exec -s read-only … </dev/null`, 적대적 Eng 관점 필수(High-risk). 회차 상한 2.
- 프롬프트에 `set -euo pipefail`·`git -C <worktree 절대경로>`·toplevel 단언을 넣는다(기억 「리뷰 픽스처가 실제 저장소에 떨어졌다」). 원문: `analysis/review/`.
- 리뷰가 제안한 새 가드·새 정책·범위 확대는 설계에 넣지 않고 아래 「Manager 결정 대기」에 적는다. 사실 오류(인용·기전)만 문서에 반영한다.

## proposal-freeze 리뷰 1회차 (2026-11-18)

- 보이스: Codex(codex-cli 0.165.0, `codex exec -s read-only -C <worktree> … </dev/null`), 적대적 Eng. 기준 HEAD `490e94f9d`. 원문 `analysis/review/codex-r1-output.md`.
  리뷰어 보고: 파일 쓰기·Go 시험·네트워크·브로커·운영 저널·docker·tossctl 실행 0. 리뷰 뒤 `git status` 로 확인한 변화: 출력 파일 하나(`analysis/review/`)뿐.
- 판정: **BLOCK**(P0 0 · P1 5 · P2 3).

| # | 지적 | 분류 | 처분 |
|---|---|---|---|
| R1-1 [P1] | 직렬 해소가 5초 스윕과 10분 검사를 함께 늦춘다; `PendingAttempts` 연속 실패는 침묵 | 사실(설계 문장의 과장) + 설계 판단 | **사실 정정 반영**: design D2 표 「스윕 간격」·직렬 대기 문단, D4 「원장 읽기 연속 실패의 침묵」, proposal What Changes 1·3, delta 「주기와 상한」(「직전 스윕 뒤 5초」). **구조 변경(10분 검사 분리·연속 실패 승격)은 채택 안 함 → Q-F** |
| R1-2 [P1] | Resolver 는 전체 호출 5분 상한을 강제하지 않는다(기한은 다음 대기 여부에서만) | 사실 오류 | **반영**: design D2 「해소 하나의 길이」 정정, delta 의 「5분을 넘어 계속되어서는 안 된다(SHALL NOT)」를 「기동과 같은 예산·기한 규칙(SHALL), 새 상한 없음」으로, tasks 3.9 에 벽시계 측정 추가. 강제 시한 추가는 채택 안 함 → Q-I |
| R1-3 [P1] | 「park 알림은 기존 것이 낸다」는 exit 발의의 attempt 에만 성립 | 사실 오류 | **반영**: design D4 「중복 없음」·D5 표, proposal 3, delta 의 「park 결과를 따로 알려서는 안 된다(SHALL NOT)」 문장 삭제(결정 전 금지 규범을 두지 않음). 새 통지 경로는 채택 안 함 → Q-E |
| R1-4 [P1] | Critical 우선은 이미 wire 임차를 쥔 해소 조회를 선점하지 못한다 | 사실 오류(과장) | **반영**: design D2 예산 문단 정정(「함께 대기 중이면 Critical 먼저; 쥔 요청 하나만큼 늦을 수 있음 ≤ 15초」), delta 는 「대기 우선순위」로 좁힘. 측정·선점 정책 → Q-G |
| R1-5 [P1] | 프로세스 사이(엔진·flatten) 「검사 → 기록」 경합이 있어 D6 의 절대 배제 주장이 성립하지 않음 | 사실 오류(주장 과장), 기존 공백 | **반영**: design D6 에 주장 철회와 경합 순서·해소 쪽 결과(오염 → park, 2건 → park, **같은 지문 1건이면 오귀속 CONFIRMED 가능**) 서술. 원자적 배제 → Q-H |
| R1-6 [P2] | D3 의 좁은 주장은 맞으나 커밋 진입 검사 뒤 취소 경계·`noteStale` 이 ctx 검사보다 먼저인 문제 | 경계 누락 + 설계 결함 | **반영**: D3 「남는 경계」 보강(원인 → 검사 순서 논증, 드라이버 원문 미확인 명시), **설계 수정** — 스윕이 `noteStale` 앞에서 ctx 검사(취소로 `ReasonAlertUndelivered` 래치가 서는 길을 닫음; 검사·기록 사이 경합은 기존 입구 성질로 잔여), delta 에 「취소된 컨텍스트로 이 알림을 기록하려 해서는 안 된다(SHALL NOT)」, tasks 3.4 하위 시험. 이 수정은 새 가드가 아니라 실행자 자기 순서 정정이다(Resolver·알림 입구 무편집) |
| R1-7 [P2] | 정본의 실측 조건이 delta 에서 사라짐, `>`/`>=` 불일치, 재생 순서 범위 | 규범 누락 | **반영**: delta 에 「실측」 항목(배포 전 SHALL), 「해소 순서」에 「기동의 해소 — 세션은 2~4만」, 10분 「이상」 통일, proposal 「실측은 계획」 명시 |
| R1-8 [P2] | `B11(default@216)` 은 AST 상 `case@216`; tasks 3.4 「다섯 자리」 는 네 자리 | 사실 오류 | **반영**: design·proposal 인용 정정, tasks 3.4 정정 |

리뷰어가 확인하고 문제없다고 한 것(원문 끝): 복구 뒤 보조 실행자 시작·drain, 패닉 회수, flatten 직접 Resolver 호출 대상 불겹침, 운영자 해소의 출발 상태 요구,
CONFIRMED 뒤 추적·발의 유지, park 래치의 계정 전역·프로세스 수명 공시, stale key·입구·창 0, 재생 미호출·재시작 규칙 미적용, 조회 해소 경로의 mutator 0.

## proposal-freeze 리뷰 2회차 (2026-11-18, 마지막)

- 보이스: Codex(같은 설정), 범위 = 1회차 처분의 사실성·남긴 결정의 안전성·delta 정합. 기준 HEAD `37af4867a`. 원문 `analysis/review/codex-r2-output.md`.
  리뷰어 보고: 파일 변경·시험·네트워크·브로커·운영 저널 0. AST 소스 해시 42개 대조 불일치 0. 리뷰 뒤 변화: 출력 파일 하나.
- 판정: **BLOCK**(P0 1 · P1 2 · P2 3). 1회차 처분 8건 중 사실 정정은 전부 「맞음」 확인(R1-1~R1-8).

| # | 지적 | 분류 | 처분 |
|---|---|---|---|
| R2-1 [P0] | Q-G(해소 조회가 쥔 wire 임차만큼 손절 지연)를 「측정 뒤 잔여」로 닫으면 불변식 4 위반 | 안전 판단 | **채택하지 않음(설계 무변경), freeze 차단 결정으로 격상**: design §6 Q-G 를 「freeze 차단」으로, 작성자 근거(같은 지연을 대사·체결 감지의 Normal 조회가 오늘도 만든다 — `rg WithPriority` 결과; a166 사용자 승인 「보호 요청 대기 허용」)와 선택지 (a)(b)(c)를 적음. **Manager 결정 전 READY 아님** |
| R2-2 [P1] | ctx 검사 이동은 delta 의 새 SHALL NOT 을 보장하지 못함(검사 뒤 취소) | 규범·설계 불일치 | **반영**: delta 를 「기록 직전 확인 의무(SHALL)」로 바꾸고 검사 뒤 취소는 입구의 기존 성질로 명시, design D3 「부분 해결」 명시, 래치가 재시작에 복원되지 않음(gateway.go:158-173) 기록 |
| R2-3 [P1] | Q-E 의 「통지 없음」은 보존된 「해소 불능」 시나리오(운영자 알림 발송)와 충돌 | 규범 정합 | **반영(서술)**: Q-E 에 정합 조건 명시 — (a) 는 기동·세션 둘 다 그 시나리오 미충족 상태를 잔여로 기록하는 것. 선택은 Manager |
| R2-4 [P2] | 「≤ 15초」 는 HTTP 시한에서 도출되지 않음(후처리 뒤 임차 해제) | 사실 오류 | **반영**: 숫자 상한 철회, 「점유 길이 미측정」 |
| R2-5 [P2] | 커밋 검사 뒤 취소 시 「park 는 남는다」 단정 과잉 | 사실 단정 | **반영**: 「남을 수 있다 — 커밋 성공에 따른다」 (design §1.4·D3 두 곳) |
| R2-6 [P2] | delta 실측 항목이 스윕 간격·동시 Resolve 수를 빠뜨림; proposal 「10분 넘게」 | 정정 누락 | **반영**: delta 실측에 실제 스윕 간격·동시 해소 최대 수·해소당 요청 수와 시간·동시 IN_DOUBT 최대 수, proposal 「10분 이상」 |

회차 상한(2)에 닿았다. 3회차는 Manager 결정 반영분에 대해서만(선례 a192).

## Manager 결정 대기

(design §6 과 같은 목록 — 리뷰 회차가 더한 것은 아래 회차 기록에 붙인다)

- **Q-A** IN_DOUBT 발생 즉시 critical(정본 「등급화된 알림」 첫 문장 대 exit 제출의 로그 줄) — (a) 비목표로 이름만 / (b) 실행자가 처음 볼 때 critical 1회. 추천 (b).
- **Q-B** CANCEL·AMEND IN_DOUBT 의 10분 알림 — 추천: 비목표 유지.
- **Q-C** 해소 읽기 예산 우선순위 Normal 대 Auxiliary — 추천: Normal.
- **Q-D** Story 의 Feature — FEAT-TOS-009 로 둠(대안 FEAT-TOS-005).
- **Q-E**(R1-3·R2-3) exit 발의가 아닌 PLACE 의 세션 park 통지 — (a) 잔여(정본 「해소 불능」 시나리오 미충족을 기록) / (b) intent 가 무장 exit 발의가 아닐 때만 critical 1회(key `order.unresolved_in_doubt|attempt:<id>`). 추천 (b).
- **Q-F**(R1-1) 10분 검사를 직렬 해소와 분리(같은 실행자 안 별도 goroutine, 원장 읽기만) + `PendingAttempts` 연속 실패 승격. 추천 둘 다 예.
- **Q-G — freeze 차단**(R1-4·R2-1 P0) 해소 조회가 쥔 wire 임차만큼의 손절 지연(점유 길이 미측정) — (a) a166 승인 범위의 기존 성질로 측정·공시 / (b) 해소 조회 점유를 시한으로 짧게 / (c) a166 운영 enforcing 이 아니면 지금은 없음. 추천: a166 운영 상태 먼저 확인 → 아니면 (a) 로 기록, enforcing 이면 Manager·사용자 판단.
- **Q-H**(R1-5) 프로세스 사이 「검사 → 기록」 경합(기존 공백) — 추천 이 change 밖 후속 후보.
- **Q-I**(R1-2) Resolver 전체 호출 강제 시한 — 추천 두지 않음(D3 와 충돌).
