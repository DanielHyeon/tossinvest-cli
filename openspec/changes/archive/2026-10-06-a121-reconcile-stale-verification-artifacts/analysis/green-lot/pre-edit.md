# a121 GREEN 로트 — Pre-Edit 선언 (tasks 3.1~3.3)

- 작성: 2026-10-05, Terra 팀메이트(GREEN), 고정 사본 `a99a9059`(worktree `/tmp/a121-green`)
- 형식: `docs/WORKFLOW.md` 「Pre-Edit 선언」. 편집한 기존 함수는 FLM 번들이 있는 여섯뿐이다. 그 밖의 기존 파일 변경은
  함수 편집이 아니라 타입 필드 추가·새 함수 추가·시험 집합 갱신이며 아래 「그 밖의 변경 표면」 에 적는다.
- 공통: CodeGraph 근거는 로트 1 `analysis/code-context/codegraph-baseline.md`(CodeGraph 1.6.0)와 각 FLM 번들의 호출자 표.
  CodeGraphContext 는 advisory — 불일치 없음(`evidence-reconciliation.md` §3·§6). 실패 시험 선행 작성: 전부 yes(착지
  `ffa8eb0d`, RED 82). 안전 불변식 §0: 주문·조건주문 변이 경로 추가 0, 토글·설정 변경 0 — 통과.

```text
Pre-Edit Gate:
- change id / task id: a121-reconcile-stale-verification-artifacts / 3.1
- 대상 심볼(패키지.함수): verifylive.Artifact.terminal (record.go)
- CodeGraph definition/callers/callees/impact: 정의 record.go:575 → 편집 뒤 같은 한 줄 메서드. 호출자 outstandingLines
  (record.go:552·562) · Runner.joinChain(runner.go:1052, 진행 중 단계 artifacts 만 — 영향 없음, 로트 1 R5). 소비자 전부
  (PendingCleanup·liveCount·report·status·abort·redo)가 Outstanding 을 통해 셋째 종결을 함께 따른다(design G2).
- CodeGraphContext 후보와 evidence reconciliation: 직접 Cancelled/Filled 를 읽는 셋(m0_manual.go:11·runner.go:1033·
  steps.go:1152)은 원본·진행 중 줄만 본다(design P2-5 정확화) — 편집 대상 아님.
- 기존 동작 파악 근거: FLM internal-verifylive--artifact.terminal(0 분기), record_replay_test.go:109(실기록 terminal==Cancelled)
- Function Logic Map / Branch Test Map: analysis/function-logic/internal-verifylive--artifact.terminal/ (재추출 완료)
- upstream 상속 테스트 영향: no — 새 필드 false 인 기존 기록은 판정 동일(record_replay_test 회귀 감시 GREEN)
- 실패 테스트 선행 작성: yes — TestReconcileOnlyRemovesItsExactArtifactFromCleanupPlanning·TestReconcilePinsRedoSet…·
  TestReconcileLeavesTheReportAttributes…·TestReconcileTextLabels…
- 설정·DB·journal 변경과 rollback: 기록(JSONL) 필드 추가 reconciled_absent(omitempty)·reconciled_at(omitzero) —
  FormatVersion 1 유지. 구 바이너리는 필드를 버리고 그 줄을 비-종결로 읽어 artifact 를 다시 보유(안전 방향,
  TestAnOlderBinaryStillSeesTheReconciledArtifactAsOutstandingAndHeld). 기존 줄 바이트 불변(TestUnreconciledArtifactsSerialiseExactlyAsTheBase).
- 안전 불변식 §0 위반 여부 검토: 통과 — 종결은 대사 줄(ReconciledAbsent=true)만 만들고 그 줄은 대사 명령만 씀.
```

```text
Pre-Edit Gate:
- change id / task id: a121 / 3.2
- 대상 심볼: main.newVerifyCmd (cmd/tossctl/verify.go)
- CodeGraph: 호출자 newRootCmd. 편집 = AddCommand 인자에 newVerifyReconcileCmd(root) 한 줄 추가(분기 0 유지).
- CodeGraphContext/reconciliation: 불일치 없음.
- 기존 동작 파악 근거: FLM cmd-tossctl--newverifycmd(0 분기), verify_test.go:234 TestVerifyCommandsAreRegisteredAndAnnotated,
  help_convention_test.go:96 TestMutatingAnnotationOnTradeCommands·TestLeafCommandsHaveSourceAnnotation
- FLM/BTM: analysis/function-logic/cmd-tossctl--newverifycmd/ (재추출 완료)
- upstream 상속 테스트 영향: yes — 잎 명령 집합이 하나 늘어 mutating 고정 집합에 "tossctl verify reconcile" 한 줄 추가
  (Manager 허용 — 로트 1 S2). 다른 명령 불변.
- 실패 테스트 선행 작성: yes — TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand
- 설정·DB·journal 변경과 rollback: 없음(명령 등록). 되돌림 = 이 한 줄 제거.
- 안전 불변식 §0: 통과 — mutating=true 로 대화형 에이전트 자동 실행 금지가 따라온다.
```

```text
Pre-Edit Gate:
- change id / task id: a121 / 3.1
- 대상 심볼: verifylive.BuildReport (report.go)
- CodeGraph: 호출자 runVerifyReport(verify.go:728)·콘솔 c.report(internal/console)·시험. 편집 = 끝에
  rep.Reconciled = ReconciledArtifacts(entries) 한 줄(분기 추가 0).
- CodeGraphContext/reconciliation: 콘솔은 Report.Text 를 렌더 — 새 칸은 JSON omitempty, 콘솔 템플릿 비편집(S1).
- 기존 동작 파악 근거: FLM internal-verifylive--buildreport(10 분기), report_test.go 기준선 BTM
- FLM/BTM: analysis/function-logic/internal-verifylive--buildreport/ (재추출 완료, 분기 번호 불변)
- upstream 상속 테스트 영향: no — 속성·Unverified·Steps·Replay 계산 불변(TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged)
- 실패 테스트 선행 작성: yes — TestReconcileLeavesTheReportAttributes…·TestReconcileTextLabels…
- 설정·DB·journal 변경과 rollback: 없음(읽기 투영). verify report --output json 에 reconciled 칸(omitempty) 추가.
- 안전 불변식 §0: 통과.
```

```text
Pre-Edit Gate:
- change id / task id: a121 / 3.1
- 대상 심볼: verifylive.BuildProgress (report.go)
- CodeGraph: 호출자 runVerifyStatus(verify.go:709)·internal/console/data.go:375. 편집 = p.Reconciled 한 줄.
- CodeGraphContext/reconciliation: 콘솔은 Progress.Outstanding 을 그림 — 대사된 artifact 는 그 칸에서 빠지는 것으로 충분(S1).
- 기존 동작 파악 근거: FLM internal-verifylive--buildprogress(7 분기)
- FLM/BTM: analysis/function-logic/internal-verifylive--buildprogress/ (재추출 완료, 분기 번호 불변)
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — TestReconcileTextLabels…
- 설정·DB·journal 변경과 rollback: 없음.
- 안전 불변식 §0: 통과.
```

```text
Pre-Edit Gate:
- change id / task id: a121 / 3.1 (S1)
- 대상 심볼: verifylive.Report.WriteText (report.go)
- CodeGraph: 호출자 runVerifyReport(verify.go:734)·콘솔 pages.go:357. 편집 = 함수 끝에 writeReconciled(w, rep.Reconciled)
  호출 한 줄(새 함수 writeReconciled 는 report.go 의 새 잎). B1 조기 반환 뒤라 단계 0 기록에선 출력 없음(FLM 기록대로).
- CodeGraphContext/reconciliation: 불일치 없음.
- 기존 동작 파악 근거: FLM internal-verifylive--report.writetext(11 분기)
- FLM/BTM: 같은 번들(재추출 완료, 기존 분기 번호 불변 — 새 호출만 추가)
- upstream 상속 테스트 영향: no — 대사 줄 없는 기록은 출력 바이트 동일(writeReconciled 는 빈 목록에서 무출력)
- 실패 테스트 선행 작성: yes — TestReconcileTextLabelsReconciledAbsentInReportAndStatus
- 설정·DB·journal 변경과 rollback: 없음.
- 안전 불변식 §0: 통과 — 대사된 artifact 를 "살아 있다"·취소·체결로 쓰지 않는다.
```

```text
Pre-Edit Gate:
- change id / task id: a121 / 3.1 (S1)
- 대상 심볼: verifylive.Progress.WriteText (report.go)
- CodeGraph: 호출자 runVerifyStatus(verify.go:715). 편집 = Outstanding 절 뒤 writeReconciled(w, p.Reconciled) 한 줄.
- CodeGraphContext/reconciliation: 불일치 없음.
- 기존 동작 파악 근거: FLM internal-verifylive--progress.writetext(8 분기)
- FLM/BTM: 같은 번들(재추출 완료, 기존 분기 번호 불변)
- upstream 상속 테스트 영향: no — 대사 줄 없는 기록은 출력 동일
- 실패 테스트 선행 작성: yes — TestReconcileTextLabelsReconciledAbsentInReportAndStatus
- 설정·DB·journal 변경과 rollback: 없음.
- 안전 불변식 §0: 통과 — B5 의 취소 권고 꼬리줄은 Outstanding 에 남은 것에만 붙는다.
```

## 그 밖의 변경 표면 (함수 편집 아님)

| 파일 | 변경 | 근거 |
|---|---|---|
| `internal/verifylive/record.go` | `Artifact` 필드 2(`ReconciledAbsent` omitempty·`ReconciledAt` omitzero), 새 잎 함수 `readRecordRaw` | design G2 「결정(골격)」; 기록 파일 열기는 record.go 소유(`TestNoAutomationBypassExists`) |
| `internal/verifylive/report.go` | `Report.Reconciled`·`Progress.Reconciled` 필드, 새 잎 `writeReconciled` | 로트 1 S1 |
| `internal/verifylive/projection_reconciled.go` (새) | `ReconciledArtifacts` 투영 | S1 — 대사 경로 밖(종결 술어 한 곳 재사용) |
| `internal/verifylive/reconcile*.go` | 골격 → 구현, 새 파일 `reconcile_check.go`·`reconcile_record.go` | tasks 3.1·3.2 |
| `internal/official/reconcile_reads.go` | 골격 → 해독·읽기 구현 | RED 처분 ①④ |
| `cmd/tossctl/verify_reconcile.go` | 골격 → 명령·preflight·좁은 생성자·기본 승인(x/term) | tasks 3.2 |
| `cmd/tossctl/help_convention_test.go` | mutating 고정 집합에 `tossctl verify reconcile` | Manager 허용(S2) |
| `go.mod` | `golang.org/x/term v0.43.0` indirect → direct(go.sum 불변) | Manager 허용(새 외부 의존 0) |

## 3.3 롤백 — 코드 쪽 사실(문서화는 Manager 분담)

1. FormatVersion 은 1 그대로다. 대사 줄은 필드 추가뿐이라 구 바이너리의 `decodeEntry` 가 거절하지 않는다(AST B2 `:442`).
2. 구 바이너리는 `reconciled_absent` 를 버리므로 대사 줄을 비-종결 최신 줄로 읽는다 → 그 artifact 는 outstanding 이고,
   `holdGate`(조건주문 → conditional-cancel) + `heldAfter`(대사 줄이 gate 의 마지막 줄 뒤) 때문에 정리 대상이 아니다(다시
   보유). 다음 `conditional-cancel` 판정이 대사 줄 뒤에 기록되면 정리 대상으로 돌아온다 — 그때 DELETE 재시도가 계획될
   수 있으므로 롤백 후 `verify run --resume` 은 사람 판단 항목이다. 시험: `TestAnOlderBinaryStillSeesTheReconciledArtifactAsOutstandingAndHeld`.
3. 대사 줄은 Calls 를 싣지 않아 구 바이너리에서도 성공 endpoint 증거가 되지 않는다.
4. 기존 줄의 직렬화 바이트는 바뀌지 않는다(`Artifact.MarshalJSON` 은 `ReconciledAbsent` 인 artifact 만 확정형으로 씀).
5. 되돌림은 증거를 지우지 않는다 — 대사 줄은 append-only 이고 새 바이너리로 돌아오면 다시 종결로 읽힌다.
