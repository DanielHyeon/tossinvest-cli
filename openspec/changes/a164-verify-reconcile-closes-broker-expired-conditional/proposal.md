# a164 · 브로커가 만료로 확인한 검증 조건주문을 기록에서 닫는다

> **High-risk 인접.** 검증 기록의 미결(outstanding) 판정은 `--include-trigger` 실주문 측정의 HOLD 와 정리(cleanup) 프롤로그의
> 대상을 정한다. 「닫힘」을 잘못 쓰면 살아 있는 브로커 조건주문이 아무도 보지 않는 상태로 남는다. 그래서 이 change 의 핵심 계약은
> **브로커의 양성 증거로만 닫는다**이다(design D2). 주문을 내는 경로는 바꾸지 않는다.
> 사용자 결정: **2026-10-22 (a)** — 「만료된 333430 artifact 를 정리하는 경로를 리뷰된 코드 변경으로 연다」.

## 왜

엔진 프로필의 검증 기록(`capability-verify.jsonl`, KR)에 2026-07-31 verify 실행이 333430 에 등록한 조건주문
`DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg` 가 아직 outstanding 이다(식별자 출처: `openspec/changes/a100-wire-fill-to-broker-protection/ma-runbook.md:159`).
읽기 전용 감사(호출자 보고, 2026-10-22)에 따르면 2026-08-13 브로커 조회가 status `EXPIRED` 를 보였고, 그 뒤 정리 DELETE
(09-06 ×2 · 09-09 · 10-19)는 전부 404 `conditional-order-not-found` 로 실패했다. **이 감사 사실은 이 로트에서 재현하지 않았다** —
이 로트는 계좌 조회·`~/.config/tossctl` 접근을 하지 않는다(아래 「미검증」).

기록을 비울 수 있는 유일한 비-정리 경로 `tossctl verify reconcile`(a121)은 이 artifact 를 받을 수 없다. 이유 둘(AST 근거
`analysis/ast/enumerate-output.md`):

1. **Q1 정적 거절.** `Reconcile` B8(`internal/verifylive/reconcile.go:205`)이 사람 승인 **앞**에서 `checkReconcileRetention` 을 부르고,
   그 함수 B1(`reconcile_check.go:118`)은 `reconcileRetention == nil` 이면 `retention-unmeasured` 로 거절한다. 생산 빌드에서
   `reconcileRetention` 은 항상 nil 이다(`reconcile.go:346`, a121 G1-4 P0-1). 그래서 생산에서는 목록을 한 번도 읽지 않는다.
2. **Q6 영구 거절.** 읽기까지 갔더라도 `checkReconcileAbsence` B5(`reconcile_check.go:384`)가 대상이 CLOSED 에 `EXPIRED` 로 있으면
   `closed-target-expired` 로 거절한다. 원 결정(a121 `design.md` 「Q6」, archive `2026-10-06-a121-reconcile-stale-verification-artifacts`):

   > **Q6 — 대상이 CLOSED 에 `EXPIRED` 로 있을 때.** **결정(Manager, 2026-09-27): 거절. 귀결 — a063 의 artifact 가 EXPIRED 면
   > 이 경로로는 영구히 대사되지 않고, 그때 사용자 재결정 항목이 된다. 거절 메시지와 이 문서 양쪽에 적는다.**

   거절 문구 상수 `reconcileQ6Message = "만료된 artifact 는 이 경로로 영구히 대사되지 않는다"`(`reconcile_check.go:21`)와 정본 spec
   (`openspec/specs/order-execution/spec.md` 「verification cleanup reconciliation requires authoritative absence」: "an artifact found
   there as `EXPIRED` is refused, and the refusal SHALL state that an expired artifact cannot be reconciled through this operation at all")이
   같은 결정을 적는다.

**사용자 결정 2026-10-22 (a) 는 Q6 의 「사용자 재결정 항목」 그 재결정이다.** 이 change 는 Q6 를 **브로커가 종결을 양성으로 확인한
artifact 에 한해** 대체한다: 부재(absence) 대사는 여전히 만료 대상을 거절하지만(부재가 아니라 종결이므로 — Q6 의 사실 판단은 그대로
옳다), 별도의 **만료 대사** 모드가 그 종결을 받는다. 「영구히」는 지운다.

미결 artifact 하나가 남긴 비용(AST 근거):

- `--include-trigger` 경로의 HOLD — cmd `validateM0TriggerMode` B9(`cmd/tossctl/verify.go:428`)와 `verifylive.New` B9(`runner.go:346`).
- 정리 프롤로그가 같은 DELETE 를 계속 제안한다 — `PendingCleanup` → `cleanupFrom` B1(`outstandingLines`)이 이 artifact 를 대상으로 낸다.
  매번 404 로 실패한다.

## 무엇을

1. `tossctl verify reconcile` 에 **명시 모드 플래그 `--expired`** 를 더한다. 플래그가 없으면 동작은 지금과 바이트 단위로 같다
   (a121 경로 불변 — 끄면 기존 동작, 안전 불변식 3 과 같은 모양).
2. `--expired` 모드는 같은 사전 검사·계좌 결속·승인·두 번 읽기·행 검사·신선도·잠금·추가 경로를 그대로 쓰고, 판정 하나만 바꾼다:
   **대상 id 의 종결이 이 실행 안의 공식 GET 으로 `EXPIRED` 라고 양성 확인될 때만** 닫는다. 근거는 by-id 상세 GET 과 CLOSED 목록 행이며
   404 는 어느 경우에도 근거가 아니다(design D2·D3).
3. 닫힘은 기록 끝에 **추가 한 줄**로만 남는다 — 새 종결 필드 `reconciled_expired`(Artifact 의 넷째 종결) · 판정한 원 status · 출처별
   응답 구분 · 시각 · 근거 지문. 기존 줄은 고치지 않는다.
4. `terminal()` 이 넷째 종결을 존중해 `Outstanding`·`PendingCleanup`·HOLD·노출 상한이 한 번에 따른다. status·report 는
   「reconciled expired」를 취소·체결·부재와 구별해 보인다.
5. 부재 모드의 Q6 거절 문구를 「부재 대사로는 받지 않는다 — 브로커 확인 만료는 `--expired`」로 고친다(거절 자체는 그대로).

## 비목표

- `M0ExactPrerequisites`(`m0_recovery.go:98`)가 요구하는 이전 단계 PASS(감사상 sell-boundary · conditional-modify · conditional-cancel 의
  마지막 판정이 fail)를 다루지 않는다. 그것은 실주문 재측정(`verify run --redo …`, 사람 승인)이지 코드가 아니다 — 사람 결정 H2.
  **그래서 a164 만으로는 `--include-trigger` 가 풀리지 않는다**: cmd `validateM0TriggerMode` 는 B8 `M0ExactPrerequisites`(`verify.go:424`)를
  B9 미결 검사(`:428`)보다 **먼저** 본다.
- 부재(absence) 경로의 Q1 보존 측정 — 그대로 미측정·거절.
- 브로커 변이(취소·정정·등록) 경로 — 0. 대사 리더는 GET 만 노출한다(봉인 시험 갱신).
- 콘솔 UI 편집 — 닫힌 artifact 는 outstanding 투영에서 빠지는 것으로 충분(a121 S1 과 같은 처분).
- 과거 관측(08-13 조회, 감사 문서)을 근거로 쓰는 것 — 기본 거절, 사람 결정 H3.

## 영향

- 코드(구현 로트, 이 로트는 문서만): `internal/verifylive/{reconcile.go, reconcile_check.go, record.go, reconcile_record.go,
  projection_reconciled.go, report.go}` + 새 파일 `reconcile_expired.go`, `internal/official/` 새 by-id 대사 읽기(새 파일),
  `cmd/tossctl/verify_reconcile.go`.
- 공식 API: `--expired` 실행당 by-id GET 2회(`CONDITIONAL_ORDER_HISTORY` 그룹, 목록 읽기와 같은 그룹)를 더한다 — 같은 rate-budget lease
  안이고 design D6 에 계상한다(안전 불변식 §0-4).
- 기록 형식: 추가 필드(additive, omitempty). 형식 버전 불변. 구 바이너리는 필드를 모르고 그 줄을 비-종결로 읽어 artifact 를 다시
  outstanding 으로 보인다 — 안전 방향(a121 「구 바이너리」와 같음). DB·journal 스키마 변경 0.
- 명령 메타데이터: `verify reconcile` 은 `mutating: "true"`(`verify_reconcile.go:60` Annotations) — **대화형 에이전트는 실행하지 않는다**.
  실기록 대사는 사람이 터미널에서 돌린다(tasks 6).

## 미검증 (이 로트가 재현하지 않은 사실)

- 감사가 보고한 08-13 `EXPIRED` 관측과 10-19 까지의 404 정리 실패 — 기록 파일을 읽지 않았다. 저장소 커밋 문서에는 08-13 의 `EXPIRED`
  관측이 **없다**(`git grep EXPIRED` 결과 0, active change 범위). 반대로 `ma-runbook.md:159` 는 같은 날(2026-08-13T14:35Z) soak by-id
  probe 의 ok 를 「생존」으로 적었는데, OpenAPI 계약은 상세 GET 이 「진행 중 + 종료된 조건주문을 모두 조회」한다고 적으므로
  (`docs/migration/openapi.latest.json` `getConditionalOrder`) **by-id ok 는 생존의 근거가 아니다** — 두 기록이 모순되지 않는다.
- 지금 브로커가 이 id 를 by-id GET 이나 CLOSED 목록에서 **아직** 이름하는지 — 측정 없음. 둘 다 404·부재면 이 change 의 증거 규칙은
  닫지 못한다(fail-closed). 그래서 구현 전에 사람의 읽기 전용 확인(tasks 0.6)이 이 change 의 진행 여부를 정한다.
- 기록에서 이 artifact 의 `ChainID`·`HeldUntil`·M0 체크포인트 유무 — 후보 선택이 실행 때 판정한다.

## 착수 상태

정의만 한 change 다(production 편집 0). 동시 change 상한(4, 착수한 것만 셈)에는 아직 들어가지 않는다.
