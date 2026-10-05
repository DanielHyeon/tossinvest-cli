# a121 — Reconcile stale verification artifacts

## Why

An approved live verification cleanup received official
`conditional-order-not-found`, while the local append-only verification record
kept the conditional artifact outstanding. A subsequent operator observation
found no active orders, but it did not establish the artifact's identity,
terminal lifecycle, or causal relationship to the failed DELETE. The current
fail-closed behavior correctly avoids calling the 404 a successful cancel, but
it has no read-backed path to reconcile a stale local artifact.

This blocks a063 operational acceptance. It must be repaired without retrying
the live cleanup, creating a new verification record, weakening the engine
interlock, or treating a 404 as success.

## What changes

- Add a read-only, account-scoped reconciliation path for an artifact already
  owned by a verification record.
- Append a distinct, idempotent `reconciled-absent` local event only after a
  fresh, complete official conditional-order read proves the same record-owned
  artifact absent under explicit identity, account, object-type, and pagination
  rules.
- Preserve the failed cleanup call and its failure. Reconciliation is neither a
  cancel, fill, endpoint success, nor capability-attestation evidence.
- Keep absence ambiguous or unproven when the official read is stale, partial,
  wrong-account, wrong-type, identity-mismatched, or otherwise unverifiable.

## Non-goals

- Retry, cancel, amend, create, or otherwise mutate any live order.
- Infer a broker terminal lifecycle from DELETE 404 or a generic operator view.
- Reclassify failed verification steps or promote reconciliation into soak or
  engine-interlock capability coverage.
- Complete, archive, or relax a063. a063 remains blocked until its own
  operational evidence is independently satisfied.

## Impact

- `internal/verifylive` append-only record, outstanding-artifact projection,
  and verification CLI surface.
- Official conditional-order read adapter and focused contract tests.
- `order-execution` specification only; runtime trading, risk limits, and
  engine startup semantics remain unchanged.

## a063 과의 관계 — 무엇이 무엇을 막는가 (Revision 1, 2026-09-27)

**의존의 원문.** a063 쪽은 a121 을 이름으로 부르지 않는다. 의존은 a063 의 진단 기록이 다음 SDD 작업으로
적은 문장이고(`openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption/task-4.2-verify-cleanup-404-diagnosis-2026-09-06.md`
「Safe next action」):

> Do **not** retry, cancel, resume, abort, or start the engine from this state.
>
> The next SDD task is an OpenSpec-scoped reconciliation design and implementation before any operational
> retry. Its contract must require a read-only, account-scoped authoritative absence observation with explicit
> freshness and object-type semantics. Only that observation—not the DELETE 404—may support a new append-only
> `reconciled-absent` event.

그리고 이 문서의 Why 가 그것을 받는다: "This blocks a063 operational acceptance."

**막히는 기전(코드).** a063 이 필요로 하는 것은 **보통의** `verify run` 이다 — 엔진 진단
(`analysis/adoption/task-4.2-engine-interlock-diagnosis-2026-09-06.md`)은 엔진이 "attest: capability-attestation
coverage incomplete" 로 시작을 거절했고, 좁은 다음 행동이 사람의 `verify run` 으로 주문·취소 수명주기 커버리지를
얻는 것이라고 적는다. 그 경로에서 stale artifact 가 하는 일은 둘이다.

- 재개의 정리 prologue 가 그것을 다시 정리 대상으로 고른다 — `Runner.cleanupTargets`(`internal/verifylive/cleanup.go:103-108`)
  → hard-evidence 3 이 적듯 그 대상은 `runCleanup` 으로 가고 그것은 `CancelConditionalOrder` 를 부를 수 있다 — 같은
  DELETE 의 재시도다(이 분기는 이 개정에서 AST 로 열거하지 않았다; task 1.2 의 몫). a063 의 진단은 바로 그것을
  금지한다("Do **not** retry, cancel, resume …").
- 그것은 살아 있는 조건주문 하나로 셈해진다 — `Runner.liveCount`(AST B2·B3, `mutate.go:679-681`)가 조건주문
  상한 `MaxLiveConditionals = 1`(`mutate.go:84`)과 비교되므로(`checkConditionalCap`, `mutate.go:657-667`) 새 조건주문
  단계가 노출 상한에 걸린다. 일반 주문 상한은 `"order"` 만 센다(`mutate.go:650`).

`--include-trigger` 경로에서는 추가로 `validateM0TriggerMode` B9(`cmd/tossctl/verify.go:427`)와 `verifylive.New` B9
(`internal/verifylive/runner.go:346`)가 outstanding 이 있으면 시작을 거절한다. 두 B9 는 trigger 모드에만 선다
(`validateM0TriggerMode` B1 `:406`, `New` B7 `:341`). 그 경로에는 `M0Unsettled`·`M0ExactPrerequisites` 검사도 따로
있다(`verify.go:414-425`).

**a121 이 여는 것과 열지 않는 것.**

- 연다: stale 조건주문 artifact 를 기록에서 종결시키는 **유일한 비-변이 경로**. 그러면 재개 정리가 그 DELETE 를
  다시 고르지 않고, 조건주문 상한이 그것을 세지 않으며, trigger 모드의 B9 두 자리도 그 artifact 때문에는
  거절하지 않는다. 사람이 승인하는 검증 재실행이 그때 가능해진다. 그 재실행은 엔진 진단이 적은 "Human authorization required:
  a human may run `tossctl verify run`" 단계이고, 엔진 인터록이 요구하는 주문·취소 수명주기 커버리지를 채우는
  경로다 — a063 **4.4**(엔진 프로필에 새 attestation 이 쓰이는지 확인)가 확인할 커버리지가 거기서 온다.
  a121 은 그 재실행을 **허가하지 않는다**(사람 승인 몫). 막힌 문 하나를 여는 것뿐이다.
- 열지 않는다: a063 **4.2** 의 서비스 정의 설치·systemd reload, **4.3** 의 3일 연속 survey — 둘 다 사람 승인
  운영 작업이고 a121 과 무관하다. a063 의 도구 결함(execution-baseline 채택 경로의 "required commit
  ancestry is absent" — 당시 `tools/logic-map/execution_baseline.py:100`; 그 파일은 a125 `9e63b681` 로
  제거됐고 이 기록은 역사 인용이다, freeze P2-1)도 a121 과 무관하다.
- 그리고 G1 이 Q1 에 걸려 있는 한, a121 이 구현돼도 대사 명령은 **거절만** 할 수 있다(design Revision 1
  G1-4 — freeze P2-2 정정). 그 경우 a063 은 계속 막힌다 — 이것은 숨기지 않고 Q1 의 선택지 (c) 로 적었다.
- **freeze P1-5 (2026-10-05): Q1(a) 아래에서도 a063 의 artifact 는 사실상 영구 거절일 공산이 크다.**
  artifact 는 2026-09-06 이전 생성이라 보존 한도 측정 시점에 이미 그 한도보다 늙기 쉽고(G1-4 거절),
  한도 안의 같은 심볼 발동은 G1-2 가 거절한다. 그 경우 a063 잔여물의 처분은 a121 경로가 아니라
  **사용자 재결정 항목**이다(design 「치르는 값」 동일 기록). Q1 측정 표본은 a063 심볼로 만들지 않는다.
- **codex F1 후속 범위(2026-10-05):** verify 기록의 계좌 결속은 끝 4자리 마스크뿐이라 자격 교체 +
  접미 충돌을 신원으로 가르지 못한다. **기록 작성기**가 작성 시점에 내구 계좌 신원 digest 를 싣는
  것은 a121 범위 밖의 후속 change 다 — a121 은 대사 줄 digest + 사람 승인 표시로 좁히기만 한다
  (design G3 F1 처분).
