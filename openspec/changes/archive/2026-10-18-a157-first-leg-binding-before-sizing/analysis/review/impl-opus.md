# a157 구현 뒤 독립 적대 리뷰 (task 2.3) — Opus

- 대상: `git diff 7b1748481 HEAD -- internal/` (HEAD `10ab9f87a`). 생산 파일 넷, 시험 파일 다섯.
- 방법: 원문 읽기, 저장소 밖 사본 두 곳(`copy-3065099`, `copy-3474575-b`, `.git` 없음, `/tmp/.git` 없음 확인, 사본 전용 GOCACHE)에서 시험을 돌림.
  변이 셋을 직접 넣음(원장에 없는 것), 편집 전 생산 코드에 새 시험을 돌림(특성 행 정직성 확인), 끝에서 사본 파일이 HEAD 와 같은지 `cmp` 로 확인함.
- 실주문 · 브로커 · 네트워크 호출: 0. 실제 저장소 쓰기: 0.

## 판정: **APPROVE-WITH-FIXES**

P0 0 · P1 0 · P2 1 · P3 6. 생산 동작 결함은 찾지 못함. P2 하나는 시험이 계약(spec 「일곱 결속 판정 각각은 한 자리에만」, design D6 M15)을
주장만큼 못 박지 못한다는 것이고, 고칠 시험 코드를 사본에서 실측함.

---

## P2-1 구조 시험은 「이름 붙인 클로저를 두 번 부르는」 복제에 뚫림 — M15 의 유일한 못이 실제로는 리터럴 개수만 셈

- 자리: `internal/execgw/a157_first_leg_structure_test.go:36-78`. 이 시험은 field 문자열 BasicLit 과 `journal.QFinalPolicyVersion` 호출 **문법 위치**의
  개수 · 순서만 셈. 결속 판정이 **몇 번 평가되는지**는 보지 않음.
- 실측 변이(원장 밖, 사본 `copy-3065099`): `riskguardian_first_leg.go:66` 의 즉시 호출 클로저 `bindErr := func() error {…}()` 를
  `bind := func() error {…}; bindErr := bind()` 로 바꾸고, `issuedAt` 앞(:122 부근)에
  `if again := bind(); again != nil { return QFinalCampaignFirstLegPrecheck{}, again }` 를 더함. 결속 판정이 사이징 앞과 뒤 **두 자리에서** 평가되는 M15 모양임.
  - 결과: 원장과 같은 시험 명령 둘이 모두 `ok`(execgw 30.1s · engine 10.2s). **SURVIVED.**
  - 이유: 문자열 리터럴은 여전히 한 번, `PrecheckQFinalEntry` 호출보다 앞임. 행동은 같음(같은 판정을 두 번 함).
- 영향: 오늘의 생산 동작 결함은 아님. 그러나 원장 `analysis/mutation/round1.md` 의 「M15(복제)는 구조 시험만 잡음 … AST 셈이 유일한 못」은
  리터럴을 옮겨 적는 복제에만 참임. 「판정이 두 자리에 있으면 반증이 죽는다」 위험을 막는다는 이 시험의 목적이, 헬퍼 클로저 하나로 무너짐
  (기억 「존재 검사는 역할 검사가 아니다」 — 세는 범위의 문제).
- 수리(사본 `copy-3474575-b` 에서 실측함): 구조 시험에 「함수 본문의 FuncLit 은 정확히 하나이고, 그것은 즉시 호출됨(`CallExpr.Fun` 이 `*ast.FuncLit`)」을 더함.

  ```go
  lits, invoked := 0, 0
  ast.Inspect(target.Body, func(n ast.Node) bool {
      switch v := n.(type) {
      case *ast.FuncLit:
          lits++
      case *ast.CallExpr:
          if _, ok := v.Fun.(*ast.FuncLit); ok {
              invoked++
          }
      }
      return true
  })
  if lits != 1 || invoked != 1 { t.Fatalf(...) }
  ```
  - HEAD 에서는 통과함(`ok 0.007s`).
  - 위 변이에서는 실패함: `func literals=1 immediately-invoked=0 — want exactly one immediately-invoked binding closure`.
  - 원장에 이 변이를 M15b 로 올리고 CAUGHT 를 다시 잴 것. 원장의 「유일한 못」 문장도 고칠 것.

---

## P3

**P3-1 「판별은 여기 하나뿐」은 저장소 전체로는 거짓. 측정 불가를 표지만으로 가르는 셋째 자리가 이미 있음.**
- `internal/execgw/a157_daily_loss_kind.go:8` 은 「그래서 판별은 여기 하나뿐」이라고 씀.
- `internal/app/engine/strategy_market_handoff_delivery.go:46` 은 `errors.As(err, &*DailyLossUnmeasurable)` **표지만** 보고
  감독자 계수 · 알림을 남김. 체인 단계 · 입력 불가 조건은 보지 않음(a137 에서 온 코드이고 a157 편집 밖임).
- 오늘은 갈라지지 않음. 표지는 `dailyLossChainRefusal`(a137_daily_loss.go:204-210)이 체인 · 입력 불가에만 붙이고, 생성 자리는 그곳 하나임(grep).
  다만 판별 함수가 좁아지거나 넓어지면, 그 계수 · 알림이 승인 분류와 어긋날 수 있음.
- 수리: 주석을 「precheck 와 첫 다리 승인의 분류는 여기 하나」로 좁히거나, 기록 자리도
  `DailyLossChainRefusalOf(err) == DailyLossChainRefusalUnmeasurable` 로 바꾸는 후속을 tasks 「후속」에 적을 것.

**P3-2 같은 패키지에 이름이 대소문자만 다른 둘이 있음.** 타입 `DailyLossChainRefusal`(a157_daily_loss_kind.go:24)과
함수 `dailyLossChainRefusal(verdict, measureErr) *IssueRefusal`(a137_daily_loss.go:204)임. 이 둘은 서로 다른 것(분류 결과 대 거절 생성기)인데
읽는 사람이 혼동하기 쉬움. 수리 예: 타입을 `DailyLossRefusalKind` 로, 함수를 `DailyLossRefusalKindOf` 로 바꿈(design D1 의 예시 이름 `DailyLossChainRefusalKind` 와도 맞음).

**P3-3 엔진 행이 「강화 실패 → errors.Join」 갈래와 결속 결함이 겹친 경우를 직접 재지 않음.**
- `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect`(a157_first_leg_admit_testseam_test.go)는 강화 **성공** 갈래만 봄.
- 결론 자체는 참임. precheck 가 결속 결함과 무관하게 같은 `err` 객체를 돌려주고(riskguardian_first_leg.go:113),
  `admit` 의 강화 갈래는 그 `err` 만 읽음. 실패 갈래는 `TestA137AFailedEscalationIsJoinedNotLost` 가 결속 결함 없이 고정함.
- design D2(c) 가 「그 성공/실패 갈래 … 모두 같음」을 표로 주장하므로, 실패하는 강화 스텁 + B8 한 행을 더하면 주장과 시험이 맞음. 필수는 아님.

**P3-4 통과 봉인 시험이 design D3 가 열거한 봉인 값 일부만 단언함.**
- `TestA157AnAcceptedFirstLegSealsTheSameValues`(:320)는 `QCandidate` · `QFinal` · `DecisionID` 만 봄. D3 은 strategy · campaign · weekly · issuedAt 도 열거함.
- 사본에서 직접 잼:
  - 변이 X1 `policyVersion = version` 삭제 → CAUGHT. a157 시험 + 기존 `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes` ·
    `…WeeklyBindingPairedKRUS` · `TestA136…` · `TestA138…` · 엔진 `TestA146…` 둘이 실패함.
  - 변이 X2 `weekly = binding` 삭제 → CAUGHT. execgw 다섯이 실패하고, 엔진은 통과함.
- 그러니 커버리지 구멍은 아님. 다만 이 시험의 머리말 「봉인 값 … 편집 전과 같음」이 실제 단언보다 넓음. 머리말을 좁히거나, 기존 시험이 덮는다고 적을 것.

**P3-5 원장의 「동등 변이: 없음」과 M14c 의 설명이 서로 긴장함.**
- M14c(측정 불가 갈래의 체인 단계 조건 삭제)는 생산에서 만들 수 있는 모든 오류에서 행동이 같음. 원장 메모도 그렇게 씀.
  표지 생성 자리는 하나이고, `chainRefusal` 은 늘 `StageChain` 임(riskguardian.go:702-704).
- 합성 표 행으로만 잡히므로 「생산 입력 기준 동등, 합성 행으로 조건을 고정」이라고 적는 것이 정직함. 분류 자체를 숨긴 것은 아님(메모에 공개됨).
  표 행이 조건을 못 박는 것은 유지할 가치가 있음.

**P3-6 주석이 「편집 전과 같음」처럼 시점을 참조함.**
- `riskguardian_first_leg.go:53`(「… 받아들이는 요청의 집합은 편집 전과 같음(design D2)」)과 :63(「편집 전 상대 순서 그대로」)이 그렇게 씀.
- 코드 안 주석은 편집 이력이 아니라 불변식을 말해야 오래감. 예: 「B2~B8 은 q_final 과 무관한 순수 판정이라 순서를 바꿔도 받아들이는 집합이 같음」.
- 문체 규칙(~함/~음 · 명사형) 위반은 찾지 못함. `strategy_entry_supervisor.go:394-395` 의 「…진입 차단으로 가게.」는 연결형 종결이지만 의미는 분명함.

---

## 공격 항목별 확인 (verified correct)

1. **받아들이는 집합 불변.**
   - 두 판본 모두 `err == nil && bindErr == nil` 일 때만 성공함. :113 · :117 의 네 조합을 전수 열거했고, `err != nil` 이 성공 경로로 새는 조합은 없음.
   - 결속 입력의 출처 교체(F3 · F4)는 원문으로 확인함.
     - `cloneRiskBucketAdmissionPlan` 은 slice 만 복제함.
     - `PrecheckQFinalEntry` 가 `admission` 에서 바꾸는 칸 중 결속이 읽는 것은 없음(`Owner.Key.AccountID` 는 B3 · B5 가 안 읽음).
     - `riskbucket/admission.go:66` 은 `decision.QCandidate = request.QCandidate` 임.
   - 시계 읽기 수와 순서는 그대로임(`issuedAt` 은 여전히 진입 판정 뒤 한 번).
   - 성공 경로에서 `policyVersion` · `weekly` 가 초기화되지 않은 채 쓰이는 일은 없음. `bindErr == nil` 이면 둘 다 대입된 뒤임. X1 · X2 변이로 확인함.
2. **새 panic 경로 없음.**
   - 결속 블록은 이제 진입 판정이 거절할 요청에도 돎. 그래서 영값 입력을 점검함.
     - `MajorDecimal`(영값 → `canonicalExecutionMinor("")` 가 false).
     - `guardianWeeklyFirstLegBinding`(nil 안전).
     - `QFinalPolicyVersion`(문자열 검사).
     - 비교 연산은 모두 값 비교임.
   - nil Guardian 은 :56 가드가 `PrecheckQFinalEntry` B1 과 같은 거절을 냄(시험 행 + M10).
3. **우선순위 D2 (a)~(f).**
   - 편집 전 생산 코드(사본에 base 판본 셋을 넣음)에 새 시험을 돌림.
     - 뒤집힌다고 한 행은 정확히 실패함: `CapZeroAndABindingDefect` 9행 · `OtherEntryRefusals…` 4행 · `OneInputThatTripsBoth…` 2행 · 구조 시험 · 엔진 `CapZeroNoLongerHides…` · 조립 감독자 셋.
     - 그대로라고 한 행은 통과함: 결속 단독 · 진입 단독 · 일일 손실 + 결속(execgw · 엔진) · 봉인 값 · nil · 판별 실물 행.
   - 특성 분류는 정직함.
4. **판별 함수 = 옛 `admit` 조건의 글자 그대로 옮김.** 조건 · 순서 · `errors.As` 의미가 같음.
   - 결속 거절은 판별에 넣지 않음(:113 은 진입 `err` 만 판별함). 그러니 결속 결함이 일일 손실로 분류될 길이 없음.
   - `PrecheckQFinalEntry` 는 결합(Join) 오류를 만들지 않음. 일일 손실 체인 거절은 `IssueRefusal{cause: *DailyLossUnmeasurable}` 단일 사슬임.
5. **`admit`.**
   - `switch` 에 default 가 없어 None 이면 a146 상한 0 · 버킷 고갈 · AUTHORITY_MISMATCH 로 그대로 떨어짐. 표지는 있으나 체인 · 입력 불가가 아닌 경우도 옛 코드처럼 한도 판정으로 내려감.
   - 강화 호출과 그 `errors.Join` 실패 갈래는 무편집임.
   - precheck 와 `admit` 이 같은 함수를 부르므로 두 판정 위험은 없음(P3-1 의 기록 자리는 별개).
6. **조립 감독자.**
   - `c.Entry != nil` 가드는 타입 있는 nil 을 막음(시험 + M16a).
   - 조립 감독자는 Run 되지 않음. `entry` 는 `blockEntryOnCentralIntegrity`(Run 경로)만 읽고, 알림은 `runMarket` 이 ctx 에 단 기록자로만 닿음
     (strategy_entry_supervisor.go:968-970). 그래서 `publishStrategyRuntime` · `Read` 의 스냅숏 덧씌움에는 바뀌는 것이 없고, 오늘 이중 알림도 없음.
     (Run 하게 되면 이중 알림 위험은 tasks 「후속」에 이미 있음.)
7. **불변 대조.** `riskguardian_qfinal.go` · `a137_daily_loss.go` 는 base..HEAD diff 0 임.
8. **회귀.**
   - 사본에서 `go test ./internal/execgw/`(무태그 · 태그)와 `./internal/app/engine/`(무태그)을 돌림.
     - engine 은 `ok 411s` 임.
     - execgw 는 A066 둘이 실패했음. 사본에 `tools/` 가 없어서 난 실패(`lstat …/tools`)이고, 복사 뒤 `ok` 임.
   - 태그 엔진 전체는 사본 + `-trimpath` 에서 10분 기본 한도를 넘음. 하네스 한계이며 a157 귀속 아님. 대상 시험군은 양쪽 태그 모두 통과함.
9. **spec 대 코드.**
   - ADDED 의 SHALL(결속 앞 · 일일 손실 예외 · 판별 하나 공유 · nil 거절 · 단독 precheck 불변)과 MODIFIED 의 두 조립 · nil 게이트 금지는 코드와 맞음.
   - 「일곱 결속 판정 각각은 한 자리에만」은 P2-1 수리 전까지 리터럴 기준으로만 고정됨.

## 사본 · 영수증

- 사본: `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad/review-a157-impl/copy-3065099` (끝에서 생산 파일 셋 `cmp` = HEAD),
  `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad/review-a157-impl/copy-3474575-b`
  (구조 시험 보강 실측. `internal/execgw/zz_review_iife_test.go` 는 사본에만 있음).
- 직접 넣은 변이: X1(policyVersion 대입 삭제) CAUGHT · X2(weekly 대입 삭제) CAUGHT · **M15b(이름 붙인 클로저 이중 호출) SURVIVED → P2-1**.
