# a157 design — 첫 다리 결속 판정을 사이징 앞으로

근거 표기:
- 「AST:」는 `analysis/ast-evidence/<번들>.ast.json`(`go run ./tools/logic-map`, HEAD `7b1748481`, 편집 전)이다. 분기 번호는 그 열거의 `B<n>`(줄)이다.
- 「raw:」는 `analysis/code-context/raw/`(codegraph 1.6.0 원문)이다.
- 줄 번호는 편집 전 HEAD 기준이다.

## 사람 결정

- 2026-10-18(a146 design 「사람 결정」 2): F10 은 a146 에서 한계로 수용했다. 결속 판정을 사이징 앞으로 옮기는 후속이 **P2 S5/S6 전에** 착지해야 한다.
- 2026-10-18(인수인계 §1): A1 「판정 내용은 그대로, 순서만」. A2 「생산 감독자와 같은 두 옵션을 조립 감독자에도」. 둘을 한 change 로 묶는다.

## F. 편집 전 사실 (HEAD `7b1748481`)

| # | 사실 | 근거 |
|---|---|---|
| F1 | `PrecheckQFinalCampaignFirstLeg`(49-112)의 분기는 아홉 개다. B1(51) `PrecheckQFinalEntry` 오류를 그대로 반환 → B2(57) 정책 버전 → B3(61) 소유자 prospective 세대 비어 있음 → B4(67) 봉인 가격 단위 → B5(71) 계보 · q_candidate · 소유자 · 캠페인 결속 → B6(80) 포지션 세대 = 현재+1 → B7(85) attempt · revision · 활성화 digest → B8(91) 주간 예약 결속 → B9(105) 계보 투영. 시계는 95행 `g.clk.Now()` 한 번만 읽고, 그 자리는 B8 뒤 · B9 앞이다 | AST `execgw--riskguardian.precheckqfinalcampaignfirstleg` branches · calls(95 `g.clk.Now`) |
| F2 | B2~B8 이 부르는 함수는 `journal.QFinalPolicyVersion`(분기 1, 문자열 검사와 연결만 함) · `MajorDecimal` ×3 · `strings.*` · `guardianWeeklyFirstLegBinding`(분기 5, 인자 검사와 복제만 함) · `qFinalRefusal` 이다. 원장 · 브로커 · 파일 · 시계 호출은 없다 | AST 같은 번들 calls 56-92 · `journal--qfinalpolicyversion` · `execgw--guardianweeklyfirstlegbinding` |
| F3 | B2 · B3 · B5 는 `entry.admission.TransactionID` · `entry.admission.Owner.{Key.ProspectiveGeneration, LaneID, CampaignID}` 를 읽는다. `PrecheckQFinalEntry` 는 `admission` 을 `request.Admission` 의 복제로 만든다(113). 그 뒤 `admission` 에 하는 대입은 `Admission.QCandidate`(114) · `Admission.Policy.EvaluatedAt`(118) · `Owner.Key.AccountID` 공백 제거(119) · `Admission.Policy`(131) · `Admission.QExistingGuardian`(154)뿐이다. 그러니 첫 다리가 읽는 칸은 `request.Entry.Admission` 의 같은 칸과 값이 같다 | AST `execgw--riskguardian.precheckqfinalentry` assignments(113 · 114 · 118 · 119 · 131 · 154 · 175) |
| F4 | B5 는 `request.Result.Quantity != entry.decision.QCandidate` 를 본다. `CalculateAdmission` 이 `decision.QCandidate` 를 쓰는 자리는 66행 `AdmissionDecision{QCandidate: request.QCandidate, …}` 하나다. 그 `request.QCandidate` 는 114행에서 대입한 `request.Entry.QCandidate` 다. 그래서 성공 경로에서 `entry.decision.QCandidate == request.Entry.QCandidate` 이다 | AST `riskbucket--calculateadmission` assignments(66 만 `QCandidate` 를 씀 — 원문 대조 `admission.go:65-125`) · precheckqfinalentry 114 |
| F5 | B9 `ProjectAcceptedStrategyflowLineage` 는 `entry.decision.QFinal` · `entry.intent` 로 만든 `RiskIntent` 와 `issuedAt` 을 쓴다. 그래서 q_final 없이는 돌 수 없다 | 원문 96-104 |
| F6 | `PrecheckQFinalCampaignFirstLeg` 의 생산 호출자는 `strategyFirstLegAdmissionBridge.admit`(guardian 인터페이스 경유) 하나다. `PrecheckQFinalEntry` 의 호출자는 둘이다: 첫 다리와 단독 발급 `IssueQFinalEntry`(:322) | raw `callers-PrecheckQFinalCampaignFirstLeg.txt` · `callers-PrecheckQFinalEntry.txt`(reconciliation: codegraph 는 인터페이스 선언 줄 :65 를 가리킴. 실제 호출은 `admit` 의 :105) |
| F7 | `admit` 은 precheck 오류를 타입으로 가른다. 일일 손실 측정 불가(a137, B6-B7) → 일일 손실 한도 + 강화(B8) → Guardian 상한 0 스킵(a146, B9) → 버킷 고갈 범위 거절(B11) 순이다. 그 밖은 `AUTHORITY_MISMATCH` 다. B2~B8 의 거절(코드 `RISK_CALCULATION_INVALID` · `OWNER_CONFLICT`, B2 는 `journal.ErrInvalidRequest`)은 이 표지들 중 어느 것에도 맞지 않는다. 그래서 오늘도 A1 뒤에도 `AUTHORITY_MISMATCH` 다 | AST `engine--strategyfirstlegadmissionbridge.admit` B5~B11 |
| F8 | 조립 감독자는 `(*Context).NewPairedStrategyEntryProductionAssembly`(325-406)가 만든다. 392행 옵션은 `Clock · CycleLimit · Workers` 뿐이다. 생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`, 414-443)는 433-434행에서 `EntryGate: c.Entry · DailyLossNotice · EntrySkipNotice` 를 넘긴다. 조립 감독자의 worker 는 `Effective: true` 일 수 있다(:541). 투영에만 쓰이고 Run 되지 않는다(a146 F7) | AST `engine--context.newpairedstrategyentryproductionassembly` B6(390 range) · `engine--context.newrefreshingpairedstrategyentrysupervisor`; raw `callers-NewPairedStrategyEntryProductionAssembly.txt` |

## D1. 옮기는 것과 남기는 것

B2~B8 을 그 상대 순서 그대로 `PrecheckQFinalEntry` 호출 앞으로 옮긴다. 판정식 · 코드 · field · 문구는 그대로 둔다.

| 분기 | 입력 출처 변경 | 근거 |
|---|---|---|
| B2 정책 버전 | `entry.admission.TransactionID` → `request.Entry.Admission.TransactionID` | F3 |
| B3 prospective 세대 | `entry.admission.Owner` → `request.Entry.Admission.Owner` | F3(바뀌는 칸은 `AccountID` 공백 제거뿐이고 B3 · B5 는 그 칸을 읽지 않음) |
| B4 가격 단위 | 없음(요청만 읽음) | — |
| B5 계보 결속 | `entry.decision.QCandidate` → `request.Entry.QCandidate`, owner 는 B3 과 같음 | F3 · F4 |
| B6 포지션 세대 · B7 attempt · B8 주간 결속 | 없음(요청만 읽음) | — |

B9 와 `issuedAt` 은 사이징 뒤에 남는다(F5). B2 가 만든 `policyVersion` 은 B9 의 `RiskIntent` 에서 다시 쓴다. 한 번 계산해 변수로 넘기고, 판정을 두 번 하지 않는다.

**nil 가드.** 옮긴 판정은 `g.accountRef` · `g.policyVersion` 을 읽는다. 지금은 `PrecheckQFinalEntry` B1(90)이 nil Guardian 을 먼저 거절한다.
그러니 첫 다리 맨 앞에 같은 거절(`RISK_CALCULATION_INVALID`, field `guardian`, 「Guardian is nil」)을 둔다. 이 가드를 빼면 nil Guardian 에서 panic 이 난다.

**기각한 대안.**
- `PrecheckQFinalEntry` 안에 첫 다리 판정을 넣는 안: 단독 발급 `IssueQFinalEntry` 도 그 함수를 쓰므로(F6) 그 경로의 거절 집합이 바뀐다.
- 옮긴 판정을 사이징 뒤에도 한 번 더 남기는 안: 같은 판정이 두 자리에 있으면 변이가 한쪽을 지워도 다른 쪽이 시험을 통과시킨다(기억 「판정이 둘이면 반증이 죽는다」).
- B2~B8 을 별도 헬퍼로 추출하는 안: 이번 change 의 목적(순서)에 필요하지 않다. 판정 범위가 함수 하나 안에 있어야 FLM/BTM 과 변이가 한 번들로 닫힌다.

## D2. 바뀌는 것 — 결함이 둘 이상인 요청의 보고

받아들이는 집합은 같다. B1~B8 은 부작용 없는 판정의 논리곱이고(F2), 시계는 그 뒤에서 읽는다(F1). 바뀌는 것은 **결속 결함 X(B2~B8 중 하나)와
진입 판정 거절 Y(`PrecheckQFinalEntry`)가 같이 있는** 요청의 보고뿐이다. 편집 전에는 Y 가, 편집 뒤에는 X 가 보고된다. `admit` 의 분류(F7)로 열거하면 다섯 갈래다.

| Y(편집 전 보고) | 편집 전 결과 | 편집 뒤 결과(X → `AUTHORITY_MISMATCH`) |
|---|---|---|
| (a) Guardian 상한 0 (a146) | 스킵 · 주기 오류 없음 · 뒤 범위 계속 | 결함 → effective 잠김 / refreshOnly 삼킴 계수 + 주기 멈춤 — **이 change 의 목적** |
| (b) 버킷 고갈 | 범위 거절(건너뛰고 모아서 주기 오류) | 결함 → 주기 멈춤 |
| (c) 일일 손실 한도 도달 | `EscalateDailyLossLimit`(계좌 진입 차단 강화) · 주기 오류 없음 | 결함 → 그 주기에 **강화 없음**(D2-c) |
| (d) 일일 손실 측정 불가 | 계좌 단위 진입 차단 결과 | 결함 |
| (e) 그 밖(스냅숏 · FX · 소유자 범위 · 0 수량 …) | `AUTHORITY_MISMATCH` | `AUTHORITY_MISMATCH`(원인 문구만 X 로 바뀜) |

어느 갈래에서도 편집 뒤가 편집 전보다 **더 많은 발급 · 주문을 만들지 않는다.** 결함은 그 범위의 발급을 막고, 주기를 멈추며(뒤 범위 미평가), effective 면 시장을 잠근다.

**D2-c 일일 손실 강화와 겹칠 때(사람 판정 요청 — 권고: 수용).** (c) 에서 그 주기의 강화(운영 모드 전환)가 일어나지 않는다. 수용 근거는 넷이다.
1. 그 범위는 결속 결함으로 발급이 막힌다.
2. 같은 시장의 뒤 범위는 그 주기에 평가되지 않는다(주기 멈춤).
3. 다른 시장의 주기와 수동 주문 경로는 각자 체인에서 일일 손실을 판정하고 강화한다.
4. 상한 0 은 오늘도 일일 손실 판정(체인 B16)을 가린다(a146 F10). 강화는 이미 「진입 판정까지 간 요청」에서만 일어나는 결과다.
대안 — 결속 결함이 나도 진입 판정을 마저 돌려 일일 손실이면 강화를 우선 — 은 판정 자리를 둘로 만들고 순서 의존을 되살린다(D1 기각 2와 같은 이유).

## D3. 바뀌지 않는 것 (시험으로 고정)

- 결함이 하나뿐이거나 없는 요청의 결과는 편집 전과 같아야 한다. 거절 코드 · field · 문구 · `Unwrap` 원인이 모두 같아야 한다.
  특성 시험 행: B2~B8 각 결함 단독, 진입 판정 거절 단독(대표 (a)~(e)), 통과.
- 통과한 precheck 의 봉인 값(`QCandidate` · `QFinal` · strategy · campaign · weekly · issuedAt)과 그 뒤의 발급 결과도 같아야 한다.
- `PrecheckQFinalEntry` · `IssueQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg` 의 소스 sha 는 편집 전과 같아야 한다(AST 번들 대조).

## D4. A2 — 조립 감독자 알림

392행 옵션에 `DailyLossNotice: c.strategyDailyLossNotice(), EntrySkipNotice: c.strategyEntrySkipNotice()` 를 더한다. 두 함수는 로그가 없으면 nil 을
돌려준다(a137_daily_loss_notice.go:58 · a146_entry_skip.go:114 주석). 그러면 감독자가 알림을 생략하므로 투영 전용 조립에서도 안전하다.
시험: 조립 감독자를 effective worker 로 시험 seam 에서 Run 한다. 일일 손실 측정 불가 · 상한 0 스킵이 각각 알림 함수까지 가는지 본다.
같은 시험을 생산 감독자에 대조군으로 두고, 옵션을 지운 변이가 시험을 죽이는지 확인한다.

## D5. 조립 감독자의 `EntryGate` (별도 발견 — 사람 결정)

조립 감독자에는 `EntryGate` 도 없다(F8). 생산 감독자는 진입 게이트가 없으면 아예 만들지 않는다(:417-419 — 「그 조립에서는 중앙 무결성 고장이 진입이
아니라 프로세스를 닫는다」). 엔진 정지 = 손절 없음(protection UNWIRED 일 때)이다. 그래서 조립 감독자를 Run 하게 되는 change 는 이것을 반드시 닫아야 한다.
이번 change 는 사람이 정한 두 옵션만 배선한다. 선택지:
1. **이번에 함께 배선(`EntryGate: c.Entry`)** — nil 이면 지금과 같고, 있으면 생산과 같아진다. 투영 경로의 동작은 바뀌지 않는다(Run 안 함).
2. 그 감독자를 Run 하는 change 로 미룬다(tasks 후속 메모).

## D6. 변이 계획 (원장 `analysis/mutation/`)

| id | 변이 | 죽여야 하는 시험 |
|---|---|---|
| M1 | 옮긴 블록을 다시 `PrecheckQFinalEntry` 뒤로(편집 전 순서) | 상한 0 + 결속 결함 B2 · B5 · B8 각 행 |
| M2~M8 | B2~B8 각 판정을 하나씩 삭제 | 그 결함 단독 행 |
| M9 | nil 가드 삭제 | nil Guardian 행(panic 이 아니라 거절) |
| M10 | B5 입력을 `request.Result.Quantity` 자기 비교로 | q_candidate 불일치 행 |
| M11 | B2 의 `TransactionID` 출처를 빈 값으로 | 정책 버전 행 |
| M12 | A2 옵션 둘 각각 삭제 | 조립 감독자 알림 행 둘 |
| M13 | B9 를 사이징 앞으로(컴파일 불가 기대 — `entry` 미정의) | build |

비동등 변이는 전부 CAUGHT 여야 한다. 하네스는 저장소 밖 사본 · pid 접미 · 무변이 대조군 GREEN · 시작 sha 단언을 지킨다
(기억 「뮤테이션 원복은 baseline 이 맞아야 한다」 · 「뮤테이션은 대상에 닿아야 한다」).
