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

## Manager 결정 대기

(design §6 과 같은 목록 — 리뷰 회차가 더한 것은 아래 회차 기록에 붙인다)

- **Q-A** IN_DOUBT 발생 즉시 critical(정본 「등급화된 알림」 첫 문장 대 exit 제출의 로그 줄) — (a) 비목표로 이름만 / (b) 실행자가 처음 볼 때 critical 1회. 추천 (b).
- **Q-B** CANCEL·AMEND IN_DOUBT 의 10분 알림 — 추천: 비목표 유지.
- **Q-C** 해소 읽기 예산 우선순위 Normal 대 Auxiliary — 추천: Normal.
- **Q-D** Story 의 Feature — FEAT-TOS-009 로 둠(대안 FEAT-TOS-005).
