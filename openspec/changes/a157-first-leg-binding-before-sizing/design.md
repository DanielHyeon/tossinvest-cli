# a157 design — 첫 다리 결속 판정을 사이징 앞으로

근거 표기:
- 「AST:」는 `analysis/ast-evidence/<번들>.ast.json`(`go run ./tools/logic-map`, HEAD `7b1748481`, 편집 전)이다. 분기 번호는 그 열거의 `B<n>`(줄)이다.
  번들의 `source_sha256` 은 **파일** sha 다(함수 본문 sha 가 아님 — freeze 리뷰 P2-1).
- 「원문:」은 같은 HEAD 의 소스를 직접 읽은 것이다. AST 번들의 assignments 는 줄 · 종류만 담고 대입 대상을 담지 않으므로,
  칸 단위 주장(F3 · F4)은 원문 읽기이며 freeze 리뷰(Opus)가 따로 대조해 참으로 확인했다.
- 「raw:」는 `analysis/code-context/raw/`(codegraph 1.6.0 원문)이다.
- 「리뷰:」는 `analysis/review/freeze-{opus,codex}.md` 이다.

## 사람 결정

- 2026-10-18(a146 design 「사람 결정」 2): F10 은 a146 에서 한계로 수용했다. 결속 판정을 사이징 앞으로 옮기는 후속이 **P2 S5/S6 전에** 착지해야 한다.
- 2026-10-18(인수인계 §1): A1 「판정 내용은 그대로, 순서만」. A2 「생산 감독자와 같은 두 옵션을 조립 감독자에도」. 둘을 한 change 로 묶는다.
- **2026-10-18(freeze 리뷰 뒤, 사용자 「1-1, 2-1」):**
  - **1-1 일일 손실 우선.** 결속 결함은 상한 0 · 버킷 고갈 등 다른 진입 거절보다 먼저 보고한다. 단, 일일 손실(한도 도달 · 측정 불가)보다는 뒤다.
    a137 의 강화(`ENTRY_BLOCKED` 영속)는 겹친 요청에서도 지금과 같다(D2).
  - **2-1 조립 감독자에 `EntryGate` 도 배선한다.** `c.Entry` 가 nil 이 아닐 때만 넘긴다(D5).

## F. 편집 전 사실 (HEAD `7b1748481`)

| # | 사실 | 근거 |
|---|---|---|
| F1 | `PrecheckQFinalCampaignFirstLeg`(49-112)의 분기는 아홉 개다. B1(51) `PrecheckQFinalEntry` 오류를 그대로 반환 → B2(57) 정책 버전 → B3(61) 소유자 prospective 세대 비어 있음 → B4(67) 봉인 가격 단위 → B5(71) 계보 · q_candidate · 소유자 · 캠페인 결속 → B6(80) 포지션 세대 = 현재+1 → B7(85) attempt · revision · 활성화 digest → B8(91) 주간 예약 결속 → B9(105) 계보 투영. 이 함수가 직접 읽는 시계는 95행 `g.clk.Now()`(B8 뒤)뿐이다. 호출하는 `PrecheckQFinalEntry` 는 118 · 127 · 132 · 165 에서 시계를 읽는다 | AST `execgw--riskguardian.precheckqfinalcampaignfirstleg` branches · calls; AST `execgw--riskguardian.precheckqfinalentry` calls |
| F2 | B2~B8 이 부르는 함수는 `journal.QFinalPolicyVersion`(분기 1, 문자열 검사와 연결만 함) · `MajorDecimal` ×3 · `strings.*` · `guardianWeeklyFirstLegBinding`(분기 5, nil 안전, 복제를 반환) · `qFinalRefusal` 이다. 원장 · 브로커 · 파일 · 시계 호출이 없다 | AST 같은 번들 calls 56-92 · `journal--qfinalpolicyversion` · `execgw--guardianweeklyfirstlegbinding` |
| F3 | B2 · B3 · B5 는 `entry.admission.TransactionID` · `entry.admission.Owner.{Key.ProspectiveGeneration, LaneID, CampaignID}` 를 읽는다. `PrecheckQFinalEntry` 는 `admission` 을 `request.Admission` 의 값 복제로 만들고(113), 그 뒤 `admission` 에 쓰는 칸은 `Admission.QCandidate`(114) · `Admission.Policy.EvaluatedAt`(118) · `Owner.Key.AccountID` 공백 제거(119) · `Admission.Policy`(131) · `Admission.QExistingGuardian`(154)뿐이다. 그러니 첫 다리가 읽는 칸은 `request.Entry.Admission` 의 같은 칸과 값이 같다 | 원문 riskguardian_qfinal.go:113-175(AST assignments 줄 113 · 114 · 118 · 119 · 131 · 154 · 175 와 일치) |
| F4 | B5 는 `request.Result.Quantity != entry.decision.QCandidate` 를 본다. `CalculateAdmission` 이 `QCandidate` 를 쓰는 자리는 66행 `AdmissionDecision{QCandidate: request.QCandidate, …}` 하나이고, 그 값은 114행이 넣은 `request.Entry.QCandidate` 다. 그래서 `entry.decision.QCandidate == request.Entry.QCandidate` 다 | 원문 riskbucket/admission.go:66-124 · riskguardian_qfinal.go:114 |
| F5 | B9 `ProjectAcceptedStrategyflowLineage` 는 `entry.decision.QFinal` · `entry.intent` 로 만든 `RiskIntent` 와 `issuedAt` 을 쓴다. q_final 없이 돌 수 없다 | 원문 96-104 |
| F6 | `PrecheckQFinalCampaignFirstLeg` 의 생산 호출자는 `strategyFirstLegAdmissionBridge.admit`(:105) 하나다. `PrecheckQFinalEntry` 의 호출자는 둘이다: 첫 다리(:50)와 단독 발급 `IssueQFinalEntry`(:323) | raw `callers-*.txt`(codegraph 는 인터페이스 선언 줄 :65 를 가리킴 — reconciliation), 리뷰 grep |
| F7 | `admit` 은 precheck 오류를 타입으로 가른다. 일일 손실 측정 불가(B6-B7, 110-116: `DailyLossUnmeasurable` 표지 + `IssueRefusal` 체인 단계 · `ReasonInputUnavailable`) → 일일 손실 한도(B8, 118: 체인 단계 · `ReasonDailyLossLimitReached`, `EscalateDailyLossLimit` 호출) → 상한 0 스킵(a146, B9) → 버킷 고갈 범위 거절(B11) 순이다. 그 밖은 `AUTHORITY_MISMATCH` 다. B2~B8 의 거절(코드 `RISK_CALCULATION_INVALID` · `OWNER_CONFLICT`, B2 는 `journal.ErrInvalidRequest`)은 어느 표지에도 맞지 않는다 | AST `engine--strategyfirstlegadmissionbridge.admit` B5~B11 |
| F8 | 엔진의 `validateStrategyFirstLegAuthority`(admit B4, :100)는 precheck · 일일 손실 측정(:104-105) **앞**에서 B3~B7 꼴의 결함을 이미 막는다. 그래서 엔진 경로에서 결속 결함이 precheck 까지 오는 것은 B2 · B8 뿐이다 | 리뷰 Opus P1-3 · P2-4(admission.go:206-227 원문) |
| F9 | 조립 감독자는 `(*Context).NewPairedStrategyEntryProductionAssembly`(325-406)가 만든다. 394행 `NewStrategyEntrySupervisor` 옵션은 `Clock · CycleLimit · Workers` 뿐이다. 생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`, 414-443)는 433-434행에서 `EntryGate: c.Entry · DailyLossNotice · EntrySkipNotice` 를 넘긴다. 조립 감독자의 worker 는 `Effective: true` 일 수 있다(:541). 이 감독자는 투영에만 쓰이고 Run 되지 않으며, 권한 갱신 파도마다 새로 만들어진다(:611-631) | AST `engine--context.newpairedstrategyentryproductionassembly` calls(394) · `engine--context.newrefreshingpairedstrategyentrysupervisor`; raw `callers-NewPairedStrategyEntryProductionAssembly.txt`; 리뷰 Opus P2-2 · P3-1 |
| F10 | `c.Entry` 의 타입은 `*execgw.EntryGate` 다. 옵션 칸 `EntryGate` 는 인터페이스 `StrategyEntryBlocker`(:126)다. nil 포인터를 그대로 넣으면 nil 이 아닌 인터페이스가 된다. 그러면 `blockEntryOnCentralIntegrity` 의 `s.entry == nil` 검사(:1046)를 지나 `(*EntryGate).Block` 이 nil 수신자에서 `g.mu.Lock()`(execgw/retry.go:543-545)으로 panic 한다 | 리뷰 Opus P1-2 |

## D1. 결속 판정의 자리와 보고 우선순위

새 함수 모양(판정식 · 코드 · field · 문구는 그대로):

```text
1. g == nil → 거절(RISK_CALCULATION_INVALID, field "guardian", "Guardian is nil") — PrecheckQFinalEntry B1 과 같은 거절
2. B2~B8 을 편집 전 상대 순서대로 요청 값에 대해 판정하고 첫 결함을 bindErr 로 잡는다(policyVersion 은 여기서 한 번 계산)
3. entry, err := g.PrecheckQFinalEntry(request.Entry)          ← 호출은 한 번
4. err != nil 이고 (bindErr == nil 또는 err 가 일일 손실 체인 거절) → err 반환
5. bindErr != nil → bindErr 반환
6. B9 계보 투영(issuedAt 은 여기서) → 봉인 반환
```

입력 출처 변경(F3 · F4):

| 분기 | 편집 전 | 편집 뒤 |
|---|---|---|
| B2 정책 버전 | `entry.admission.TransactionID` | `request.Entry.Admission.TransactionID` |
| B3 prospective 세대 · B5 소유자 | `entry.admission.Owner` | `request.Entry.Admission.Owner`(B3 · B5 는 공백 제거되는 `AccountID` 를 읽지 않음) |
| B5 q_candidate | `entry.decision.QCandidate` | `request.Entry.QCandidate` |
| B4 · B6 · B7 · B8 | 요청 값 | 같음 |

**「일일 손실 체인 거절」 판별은 한 자리에 둔다.** execgw 에 판별 함수 하나(`execgw.DailyLossRefusalKindOf(err) → DailyLossRefusalKindNone | …Unmeasurable | …LimitReached` — 구현 리뷰 P3-2 로 이름 확정, a137 의 `dailyLossChainRefusal` 과 대소문자만 다른 충돌 회피)를 두고,
조건은 오늘 `admit` 의 B6-B8 조건(F7)을 **글자 그대로 옮긴 것**이다. 순서도 측정 불가 먼저, 한도 다음이다. `admit` 도 이 함수로 가른다(조건을 옮길 뿐 분기 결과는 같다).
같은 판별이 두 자리에 있으면, 변이가 한쪽을 지워도 다른 쪽이 시험을 통과시킨다(기억 「판정이 둘이면 반증이 죽는다」). 그래서 판별은 하나로 둔다.

**nil 가드.** 옮긴 판정은 `g.accountRef` · `g.policyVersion` 을 읽는다. 가드가 없으면 nil Guardian 에서 panic 한다.

**기각한 대안.**
- `PrecheckQFinalEntry` 안에 첫 다리 판정을 넣는 안: 단독 발급 `IssueQFinalEntry` 도 그 함수를 쓰므로(F6) 그 경로의 거절 집합이 바뀐다.
- 결속 결함을 무조건 우선하는 안(설계 원안): 겹친 요청에서 a137 강화가 사라진다(리뷰 codex P1). 사람이 1-1 로 기각했다.
- 결속 판정을 사이징 뒤에도 남기는 안: 판정이 두 자리에 생긴다.
- B2~B8 을 헬퍼로 추출하는 안: 목적(순서)에 필요하지 않다. 범위가 한 함수 안에 있어야 FLM/BTM 과 변이가 한 번들로 닫힌다.

## D2. 바뀌는 것 — 둘 이상의 판정이 거는 요청의 보고

받아들이는 집합은 같다. 이유는 셋이다.
- B1~B8 은 부작용 없는 판정의 논리곱이다(F2).
- B2~B8 은 시계를 읽지 않는다.
- `PrecheckQFinalEntry` 는 여전히 한 번 불린다.

단, 「같다」는 **같은 입력 · 같은 시계 관측**에서다. 결속 판정이 앞에 서면서 `PrecheckQFinalEntry` 의 시계 읽기(F1)가 그만큼 늦어진다.
신선도 경계 바로 위의 요청은 벽시계에서 다른 판정을 받을 수 있다. 이 차이는 편집 전에도 있던 실행 시간 지터와 같은 종류다(리뷰 codex P2-3).

바뀌는 것은 **결속 판정 X(B2~B8 중 첫 결함)와 진입 판정 거절 Y 가 함께 걸리는** 요청의 보고뿐이다. `admit` 분류(F7)로 열거하면 다음과 같다.

| Y(편집 전 보고) | 편집 전 결과 | 편집 뒤 결과 |
|---|---|---|
| (a) Guardian 상한 0 (a146) | 스킵 · 주기 오류 없음 · 뒤 범위 계속 | **X** → `AUTHORITY_MISMATCH`: effective 잠김 / refreshOnly 삼킴 계수 + 주기 멈춤. **이 change 의 목적** |
| (b) 버킷 고갈 | 범위 거절(건너뛰고 모아서 주기 오류) | **X** → 결함 · 주기 멈춤 |
| (c) 일일 손실 한도 도달 | 강화(`ENTRY_BLOCKED` 영속) · 성공 시 `DailyLossEntryBlock` 으로 주기 오류 없음, 실패 시 `errors.Join(거절, 강화 오류)` 결함 | **Y 그대로**(1-1) — 강화 · 그 성공/실패 갈래 · 모인 범위 오류 보존 모두 같음 |
| (d) 일일 손실 측정 불가 | 계좌 단위 진입 차단 결과 · `engine.daily_loss_unmeasurable` 기록 | **Y 그대로**(1-1) |
| (e) 그 밖(스냅숏 · FX · 소유자 범위 …) | `AUTHORITY_MISMATCH` | **X** → `AUTHORITY_MISMATCH`(원인 코드 · field · 문구만 X 로 바뀜) |
| (f) **입력 하나가 X 와 Y 를 함께 거는 경우**(리뷰 Opus P1-1, 실측) | Y | X — 결과 분류는 그대로 `AUTHORITY_MISMATCH` |

(f) 의 실측 · 점검 목록이다. 엔진 경로에서는 F8 의 `validateStrategyFirstLegAuthority` 가 이것들을 precheck 앞에서 먼저 막는다.
- `Entry.QCandidate = 0`: `ZERO_QUANTITY/q_candidate` → `RISK_CALCULATION_INVALID/first_leg_lineage`(실측)
- 비정규 `Entry.Symbol`(계보는 정규): `UNKNOWN_SYMBOL/symbol` → `first_leg_lineage`(실측)
- `lineage.AccountRef ≠ g.accountRef`(소유자 AccountID 도 같은 값): `OWNER_CONFLICT/owner` → `first_leg_lineage`(원문)
- 봉인 가격과 다른 Entry/Stop/Target: `EXISTING_GUARDIAN_CAP` → `first_leg_lineage`(원문). 상한이 0 이 되는 경우면 (a) 와 같은 의도된 뒤집힘이다.

어느 갈래에서도 편집 뒤가 편집 전보다 더 많은 발급 · 주문을 만들지 않는다. 일일 손실 강화도 줄지 않는다(1-1).

## D3. 바뀌지 않는 것 (시험으로 고정)

- 결속 판정 일곱을 모두 통과하는 요청은 진입 거절의 코드 · field · 문구 · 원인이 편집 전과 같다.
  `PrecheckQFinalEntry` 를 통과하는 요청은 결속 거절의 코드 · field · 문구 · 원인이 편집 전과 같다.
- 일일 손실 한도 · 측정 불가 거절은 결속 결함이 함께 있어도 편집 전과 같다.
- 통과한 precheck 의 봉인 값(`QCandidate` · `QFinal` · strategy · campaign · weekly · issuedAt)과 그 뒤의 발급 결과도 같다.
- `PrecheckQFinalEntry` · `IssueQFinalEntry` 는 riskguardian_qfinal.go 를 편집하지 않으므로 파일 sha 로 확인한다.
  `IssuePrecheckedQFinalCampaignFirstLeg` 는 편집하는 파일 안에 있다. 그래서 **함수 본문(gofmt 정규화한 FuncDecl 구간) sha** 로 편집 전후를 대조한다(리뷰 Opus P2-1).
- 엔진 경로(`admit` 이하)에서 뒤집을 수 있는 결속 결함은 B2 · B8 뿐이다(F8). 엔진 시험 행은 이 둘로만 만든다. B3~B7 은 execgw 시험에서 다룬다.

## D4. A2 — 조립 감독자 옵션

394행 옵션에 다음을 더한다.

```go
opts := StrategyEntrySupervisorOptions{Clock: clk, CycleLimit: MaximumStrategyCycleLimit, Workers: workers,
	DailyLossNotice: c.strategyDailyLossNotice(), EntrySkipNotice: c.strategyEntrySkipNotice()}
if c.Entry != nil { // 타입 있는 nil 을 인터페이스에 넣지 않음(F10)
	opts.EntryGate = c.Entry
}
```

두 알림 함수는 로그가 없으면 nil 을 돌려준다(a137_daily_loss_notice.go:58 · a146_entry_skip.go:114). 그러면 감독자가 알림을 생략하므로 투영 전용 조립에서도 안전하다.

**시험(리뷰 Opus P2-2 — 이 감독자의 worker `Cycle` 은 생산 권한 갱신 전체라 Run 으로는 닿지 않음):**
- 최소 `Context{Log: …}` 로 조립 감독자를 만들고 두 알림 칸이 nil 이 아님을 단언한다.
- 패키지 내부 기록자(a146 스킵 기록 · a137 측정 불가 기록)를 그 감독자에 직접 불러 `obs` 이벤트가 `c.Log` 까지 가는지 본다.
  생산 감독자를 대조군으로 둔다.
- `c.Entry == nil` 이면 감독자의 `entry` 가 nil 인터페이스인지 본다. 실제 게이트면 중앙 무결성 갈래가 진입을 닫는지(`Block` 경로) 본다.

**남는 한계(조립 감독자를 Run 하게 되는 change 가 닫아야 함):** dedup 상태는 감독자마다 있고, 조립 감독자는 파도마다 새로 만들어진다(F9).
그래서 Run 하게 되면 「한 프로세스 안에서 일 키마다 한 번」이 파도마다 다시 시작된다. 생산 감독자와 함께 돌면 줄이 두 번 나온다.
그 change 는 dedup 상태를 `Context` 로 올려야 한다. tasks 「후속」에 적는다.

## D5. 조립 감독자의 `EntryGate` (사람 결정 2-1 — 배선)

D4 의 nil 가드로 배선한다. 오늘 동작은 바뀌지 않는다(Run 안 함, 투영은 `entry` 를 읽지 않음). 조립 감독자를 Run 하게 될 때
「중앙 무결성 고장 → 프로세스 전체 fail-closed → 엔진 정지 → (protection UNWIRED 면) 손절 없음」 갈래가 남지 않는다.

## D6. 변이 계획 (원장 `analysis/mutation/`)

| id | 변이 | 죽여야 하는 시험 |
|---|---|---|
| M1 | 결속 블록을 다시 `PrecheckQFinalEntry` 뒤로(편집 전 순서) | 상한 0 + Bn 일곱 행 |
| M2~M8 | B2~B8 각각을 하나만 사이징 뒤로(부분 이동) | 그 Bn 의 상한 0 + Bn 행 |
| M9 | B2~B8 각 판정 삭제(일곱 변이) | 그 결함 단독 행 |
| M10 | nil 가드 삭제 | nil Guardian 행(panic 이 아니라 거절) |
| M11 | B5 q_candidate 입력을 `request.Result.Quantity` 자기 비교로 | q_candidate 불일치 행 |
| M12 | B5 소유자 입력(LaneID · CampaignID)을 계보 자기 비교로 | 소유자 불일치 행 |
| M13 | 4단계의 일일 손실 예외 삭제(결속 결함 무조건 우선) | 일일 손실 한도 + B2 · B8, 측정 불가 + B2 · B8 행(execgw · 엔진) |
| M14 | 판별 함수의 측정 불가 갈래 · 한도 갈래 각각 삭제 | 위 행 + a137 기존 분류 시험 |
| M15 | 결속 블록을 앞에 두고 뒤에도 복제 | 구조 시험(아래) |
| M16 | `opts.EntryGate` nil 가드 삭제 · 두 알림 칸 각각 삭제 | D4 시험 셋 |

**구조 시험(AST, M15 를 죽임 — 리뷰 Opus P1-3):** `PrecheckQFinalCampaignFirstLeg` 안에 다음이 성립하는지 단언한다.
- `g.PrecheckQFinalEntry` 호출이 정확히 하나다.
- 결속 거절을 만드는 field 문자(`prospective_generation` · `first_leg_price_unit` · `first_leg_lineage` · `position_generation` · `first_leg_strategy` · `weekly_reservation`)와
  `journal.QFinalPolicyVersion` 호출이 각각 정확히 한 번이고, 그 호출보다 앞에 있다.

비동등 변이는 전부 CAUGHT 여야 한다. 하네스는 저장소 밖 사본 · pid 접미 · 무변이 대조군 GREEN · 시작 sha 단언을 지킨다.
「B9 를 사이징 앞으로」는 컴파일이 안 되므로 변이가 아니라 빌드 메모로만 적는다(리뷰 Opus P3-6).
