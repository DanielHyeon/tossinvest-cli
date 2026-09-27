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

**막히는 기전(코드).** 기록에 outstanding artifact 가 남아 있는 동안 M0 trigger 검증은 시작되지 않는다 —
`validateM0TriggerMode` B9(`cmd/tossctl/verify.go:427`: "prior verification record has outstanding artifacts;
inspect with `tossctl verify status` and reconcile manually before any cleanup")와 `verifylive.New` B9
(`internal/verifylive/runner.go:346`). outstanding 은 노출 상한에도 셈해진다(`Runner.liveCount` B2·B3,
`mutate.go:679-681`). a063 의 엔진 진단(`analysis/adoption/task-4.2-engine-interlock-diagnosis-2026-09-06.md`)은
엔진이 "attest: capability-attestation coverage incomplete" 로 시작을 거절했고, 그 좁은 다음 행동이 사람의
`verify run` 으로 커버리지를 얻는 것이라고 적는다.

**a121 이 여는 것과 열지 않는 것.**

- 연다: stale 조건주문 artifact 를 기록에서 종결시키는 **유일한 비-변이 경로**. 그러면 위 두 거절(B9)이
  풀리고 사람이 승인하는 검증 재실행이 가능해진다. 그 재실행은 엔진 진단이 적은 "Human authorization required:
  a human may run `tossctl verify run`" 단계이고, 엔진 인터록이 요구하는 주문·취소 수명주기 커버리지를 채우는
  경로다 — a063 **4.4**(엔진 프로필에 새 attestation 이 쓰이는지 확인)가 확인할 커버리지가 거기서 온다.
  a121 은 그 재실행을 **허가하지 않는다**(사람 승인 몫). 막힌 문 하나를 여는 것뿐이다.
- 열지 않는다: a063 **4.2** 의 서비스 정의 설치·systemd reload, **4.3** 의 3일 연속 survey — 둘 다 사람 승인
  운영 작업이고 a121 과 무관하다. a063 의 도구 결함(execution-baseline 채택 경로의 "required commit ancestry
  is absent", source snapshot `c727ad12` 이 이 브랜치 역사에 없음)도 a121 과 무관하다.
- 그리고 G1 이 Q1 에 걸려 있는 한, a121 이 구현돼도 대사 명령은 **거절만** 할 수 있다(design Revision 1 G1-3).
  그 경우 a063 은 계속 막힌다 — 이것은 숨기지 않고 Q1 의 선택지 (c) 로 적었다.
