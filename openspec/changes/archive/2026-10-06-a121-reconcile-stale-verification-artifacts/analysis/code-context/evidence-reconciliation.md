# a121 evidence reconciliation — tasks 1.1 · 1.2

- 작성: 2026-10-05, Terra 팀메이트(로트 1, 증거만 — 생산 코드·시험 변경 0)
- 구현 base: `de147cc285c5274cad6d6ab7b208513027a70b40` (옛 base `9408fc957fbb8687fdf75fd8a58cded50bf4f687`)
- 질의 영수증: `codegraph-baseline.md`(CodeGraph 1.6.0), `codegraphcontext-context.md`(CGC, advisory)

## 1. base 재고정 영수증 (task 1.1)

`tools/sdd/capture_change_base.py` 는 파일이 있으면 거절하므로(`change base already captured`) 기존 파일을 지운 뒤
de147cc2 고정 worktree 에서 다시 캡처했다. **효력은 Manager 의 단독 커밋 뒤에만 생긴다**(WORKFLOW 「사람 승인 base 재고정」 3).

| 조건 | 실측 |
|---|---|
| 옛 base 가 새 base 의 조상 | 예 (`git merge-base --is-ancestor`), 사이 582 커밋 |
| 귀속 실측 — 옛 base 이후 이 change 디렉터리를 만진 비병합 커밋 중 `.go` 수정 | **0** / 6 (`bc03c4d4`·`3cf64d5d`·`f680b137`·`ddd39a83`·`5220646d`·`de147cc2`, 전부 문서). 이 change 는 renumber 이력 없음 |
| 대상 소스 변경(옛→새) | `internal/verifylive/` · `internal/official/` · `internal/attest/` · `cmd/tossctl/verify.go` **0 파일**. `cmd/tossctl/` 의 다른 25 파일(`console.go`·`engine*.go`·`help_convention_test.go` 등)만 바뀜 |
| 옛 base 에서의 required (5단계) | **242** 함수 — 전부 형제 change 착지 몫(`internal/app/engine`·`internal/journal`·`internal/strategy*` 등; `verifylive`·`verify.go` 0 건) |
| 새 base 에서의 required | **0** — 스크래치 clone 에 이 로트 산출물을 합성 커밋(`589cfb1a`, 저장소 밖)해 `check_analysis.py --change a121-…` 실행: `required 0 function(s)` · `evidence complete or diff-proven exempt` · rc 0 |
| strict validate | `openspec validate a121-reconcile-stale-verification-artifacts --strict` → `Change '…' is valid` |

승인 기록(WORKFLOW 재고정 2)은 아직 없다 — Manager 가 review.md·tasks.md 에 남긴다.

## 2. AST 증거 재생성 (task 1.2)

`analysis/ast-evidence/` 16 개를 base de147cc2 에서 같은 추출기(`tools/logic-map`)로 다시 뽑았다 — **16/16 바이트 동일**.
분기 근거 2 개(`heldAfter`·`LastEntry`)를 더했다(README 표).

## 3. design 인용 전수 대조 (Revision 1 G1·G2·G3 + proposal)

새 AST 와 HEAD 원문으로 file:line·분기 ID 를 하나씩 대조했다.

**일치** (내용·줄·분기 ID 모두): `conditional_reads.go:44-48`·`:105-134`·`:111`·`:129-131`·`:158-164`(B1) ·
ConditionalOrdersRaw B6 `:181`·B7 `:189` · `record.go:236-241`·`:185-187`·`:85-88`·`:265-266`·`:65-88`·`:55-60`·
`:247-279`·`:552`(B4)·`:562`(B6)·`:571-574`·`:575`·`:442`(decodeEntry B2)·B1 `:439`·`:583-595` ·
`steps.go:789`(stepConditionalCancel B3)·`:61` · `steps_trigger.go:517-520`·`:553` · m0RecoverPending B2 `:134`·B3 `:137`·
B4 `:138`·B6 `:146`·B12 `:163`·B13 `:166`·B14 `:170`, limit 100 `:145` · `protection_reads.go:24-47`(+ `OrderSide`·
`ClientOrderID` 보존 `:51-53`, ConditionalOrdersRaw 의 행 복사 `:190-208` 에는 둘 다 없음) · `orders_raw.go:62-92` ·
`conditional_writes.go:52-60` · `endpoints.go:82-104`(B1 `:82`·B3 `:86`·B4 `:88`, Kind 필터 없음) · `cleanup.go:119-121`·
holdGate B1 `:156`·B2 `:159`·cleanupFrom B3 `:131` · `m0_manual.go:45` · `receipt.go:353-361` · `mutate.go:679`(liveCount B2)·
`:84`·`:657-667`(B2 `:662`)·`:650` · `m0_recovery.go:23` · `cmd/tossctl/verify.go:439`·`:796-807`·`:941-945`(B5·B6)·`:947`(B7)·
`:894`(buildVerifyBroker B5)·`:911-914`·`:765`(B1)·`:768`·`:769`(B2)·`:427`(B9)·`:406`(B1)·`:414-425` · `runner.go:278`·`:688`·
`:890`·`:346`(New B9)·`:341`(B7) · `reads.go:93` · `attest.go:281-291` · `credentials.go:30-34` · `cleanup.go:103-108`.

**어긋남 목록** (문서 수정은 Manager 몫):

| # | 위치 | design 의 주장 | 새 AST / HEAD | 등급 |
|---|---|---|---|---|
| D1 | design G2 첫 항 · README `outstandinglines` 행 | "투영의 키는 `Kind + "\x00" + ID` 다(AST B3, `record.go:548`)" | 키는 **547행의 대입**(분기 아님)에서 만들어진다. B3 `:548` 은 `if _, seen := latest[key]; !seen` — 처음 본 키의 순서를 적는 분기다. 내용(키 규칙)은 참, 분기 귀속이 한 줄·역할 어긋남 | P2 |
| D2 | design G1 Q1 원 근거 | "`Artifact` 에는 방향도 수량도 없고(`record.go:182-243`)" | 구조체는 182-**244**(243 = `Note`, 244 = 닫는 괄호). 내용 참 | P3 |
| D3 | design G3 「코드가 주는 것」·G3-1 | "`attest.Mask`, 끝 4자리만 남긴다" | 빈 값 → `"(none)"`(`attest.go:283-285`), 4자 이하 → 전부 `*`(`:287-289`). 4자 이하 참조끼리의 대조는 **길이 비교**로 퇴화한다. 인용 오류는 아니고 RED(2.2.2) 경계값 누락 | P3 |
| D4 | design G3 「코드가 주는 것」·G3-3 | "`TOSSCTL_OPENAPI_KEY`·`SECRET` 이 있으면 파일을 읽지 않는다" → "환경 변수 자격 증명이 설정돼 있으면 거절" | 파일을 건너뛰는 것은 **둘 다** 비어 있지 않을 때뿐(`credentials.go:32`). 하나만 있으면 파일을 읽는다. 거절 조건이 "하나라도" 인지 "둘 다" 인지 미정 — 보수 방향은 "하나라도" | P2 |

그 밖의 분기 주장은 전부 일치했다(분기 ID 인용 중 어긋남은 D1 하나).

## 4. `outstandingLines` 비편집 확인 (Q2 = (a))

`verifylive.outstandingLines`(`record.go:542-567`, AST 6 분기, `source_sha256` df526d2c… — 옛 증거와 바이트 동일)는 이
change 의 편집 대상이 아니다: Q2 = (a) 로 대사 줄이 대상 artifact 의 원문 키(`Kind`·`ID`)를 그대로 재사용하므로 B3 의 키
규칙과 B4·B6 의 종결 처리는 바뀌지 않고, 셋째 종결은 `Artifact.terminal` 한 곳의 편집으로 이 함수에 전달된다. FLM 번들을
만들지 않는다. 구현 diff 가 이 함수를 건드리면 게이트 5단계가 번들을 요구하므로 그때 이 줄은 무효다.

## 5. 범위 차이 — task 1.2 목록 밖의 편집이 필요해 보이는 자리 (Manager 판단)

| # | 자리 | 왜 |
|---|---|---|
| S1 | `Report.WriteText`(`report.go:239-291`) · `Progress.WriteText`(`report.go:346-379`) | `BuildReport`·`BuildProgress` 는 `Outstanding` 만 담고 종결 artifact 를 담는 칸이 없다. "reconciled absent" 를 **보이게** 하려면 새 필드 + 두 렌더러 편집이 필요하다. 렌더러는 기존 함수 내부 편집이므로 FLM 대상이 된다. 콘솔도 `progress.Outstanding` 을 그린다(`internal/console/data.go:383` → `templates.go:620`·`:700`) — 콘솔에도 표시하려면 UI 범위가 열린다 |
| S2 | `TestMutatingAnnotationOnTradeCommands`(`help_convention_test.go:96`) | 잎 명령의 `mutating=true` 집합을 **정확히** 고정한다. 새 대사 명령의 `mutating` 주석 값을 design·spec 이 정하지 않았다(브로커 변이 0 · 로컬 기록 추가 1; 원장을 쓰는 운영자 명령 선례 `engine mode-release`·`engine alerts ack` 는 true). 어느 쪽이든 이 시험과 `TestLeafCommandsHaveSourceAnnotation` 이 판정한다 |

## 6. 구현 위험 — 이 로트가 AST 로 확인한 것

| # | 근거 (AST) | 위험 | 요구 |
|---|---|---|---|
| R1 | `LastEntry` B1·B2 · `heldAfter` B1·B2 — 둘 다 `StepID` 만 비교, `Kind` 무시 | 대사 줄이 카탈로그 단계 ID(특히 `conditional-cancel` = 조건주문의 기본 보유 게이트)를 달면 `Settled`·`Passed`·`BuildProgress` 단계 판정·`heldAfter` 해제 판정이 움직인다 | 대사 줄의 `StepID` 는 카탈로그 밖(빈 값 또는 전용 값)이어야 함 — design 에 값이 없다. RED 후보 |
| R2 | `BuildReport` B2 · `BuildProgress` B2 — 첫 줄의 `AccountRef` 를 종류 무관 채택 | 대사 줄은 기존 줄 뒤에 붙으므로 실제 영향 없음 | 확인만 |
| R3 | `BuildReport` B3 — `!isStepEntry` 는 건너뜀 | `KindReconcile` 줄의 관측이 속성·Unverified·ReplayEnabled 에 들어가지 않음 — design 「Safety and attestation boundary」 와 일치 | 유지 시험(2.4) |
| R4 | `heldAfter` B2 — gate 단계의 **마지막** 줄 위치만 봄 | design G2 「구 바이너리 → 다시 보유」 주장을 코드가 뒷받침: 구 바이너리에서 대사 줄이 비-종결 최신 줄(`at` = 대사 줄)이 되고, 그 뒤에 `conditional-cancel` 줄이 없으므로 `cleanupFrom` B3 가 거짓 → 정리 대상 아님 | 2.4.1 의 구 판본 시뮬레이션이 이 경로를 재야 함 |
| R5 | `Artifact.terminal` 호출자 `Runner.joinChain`(`runner.go:1052`) — design 소비자 목록에 없음 | 진행 중 단계의 `sr.artifacts` 만 보므로 대사 줄 영향 없음 | 확인만 |
| R6 | `record_replay_test.go:109` — 실기록에서 `terminal() == Cancelled` 단언 | 셋째 항 추가 뒤에도 기존 실기록에서는 성립해야 함(새 필드 false) | 기존 시험이 회귀 감시 |

## 7. 판정

- 1.1: Story 존재·내용 모순 없음(아래), base 재캡처 완료(커밋 대기), strict validate 통과.
- 1.2: 편집 예정 4 함수 FLM/BTM 번들 완성(`check_analysis` 새 base 합성 커밋에서 rc 0), 호출만 하는 2 함수 호출자 증거
  기록, AST 증거 16 재생성(동일) + 2 추가, design 인용 대조 어긋남 4(P2 2 · P3 2), 범위 차이 2, 구현 위험 6.

Story `STORY-TOS-a121`(`docs/pm/portfolio/stories/STORY-TOS-a121.yaml`): intent `active`, change_id·경로 일치, acceptance 4 항은
Revision 1·Q1~Q6 결정과 모순 없음. 주의 하나: acceptance 2 「Only fresh complete … absence evidence appends …」 는 Q1 (a)
측정 전 잠정 거절 상태에서 "제약"(그 외에는 쓰지 않는다)으로 읽으면 충족되고, "능력"(그 증거가 있으면 쓴다)으로 읽으면
측정 완료 전에는 충족될 수 없다. 어느 읽기로 수락할지는 Manager 판단 — Story 문구 수정은 하지 않았다.
