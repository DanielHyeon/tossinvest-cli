# M-A 재발행 런북 — 2026-10-06 발행 (실행일 = 다음 KR 정규장 사람 세션)

> 절차 정본은 `measurement-prereq.md`다. 이 문서는 그날의 손 순서이며 **실행 권위가 아니다**:
> 등록·정정·취소·매도 등 모든 `mutating` 명령은 실행 직전 그 명령에 대해 사람이 별도 승인한다
> (§0-1·§0-7). 이 문서에 대한 동의는 그 승인이 아니다. 아래 2026-08-14 역사본의 날짜·종목·
> expire·id 는 전부 만료 상태고 복사 실행 금지가 유지된다.
>
> **이번 재발행에 반영된 그간의 결정**: D8 계약 1 = (다)(2026-10-05 사용자 — 취소-선행 유지,
> 6.4.4 창 운영 명기), T1 종결(도달 불가 3 증명 + 2.7 D3 실측 결함 편입 — M-A 완전 통과 시
> T2-A 가 열리고 2.7 수리가 배선 선행), 동승 3건 큐 편성(사용자 2026-10-05 3항 + a121 Q1 결정
> (a) 2026-09-28 + a112 L1c 결정 46 잔여).
>
> **2026-10-07 사용자 동의(권장안 3건 — 날짜 외 사람 결정 종결)**: ① R0-2 = (a) 8-15 설치본
> 그대로 측정, ② R0-3 잔여 조건주문 = 기본 retain(아래 개정 문구), ③ 대상 심볼 = 당일 §2
> 유동성 실측으로 확정. 남은 입력은 세션 날짜뿐.
>
> **2026-10-07 §3 앞 HOLD**: 대상 심볼 **010170 대한광통신 확정** — 사용자가 장중 6주 매수
> (의도 확인, 평단 18,485 · sellable 6 · 09:19 영수증), 호가 스프레드 1틱(18,410/18,420)·깊이
> 충분으로 §2 충족. 그러나 사용자가 콘솔·터미널 원격 접속 불가 상태라 §3(콘솔 verify 승인
> 클릭)부터의 κ·⌨ 단계를 진행할 수 없다. 에이전트 대행은 안전 불변식(1·7 — live 검증·승인
> 클릭·mutating 실행 = 사람 직접)과 하네스 안전 분류기 양쪽이 차단 — 대행하지 않는 것이
> 맞다(전부 대행이면 "감독하 실행"이라는 측정 의미 자체가 깨짐). **재개 조건 = 사람이 콘솔·
> 터미널에 닿는 KR 정규장**(§3→§8 소요 약 1h, 15:30 전 복귀 시 당일 완주 가능). 보유 6주는
> 세션을 넘겨도 유효하며 재개일에 R0-1(시장·expire)·보유 재동결만 다시 한다.
>
> **2026-10-07 세션 개시(사용자 "오늘 진행") — R0 실측과 재선정**: R0 전 항목 read-only 동결
> 완료(03:27~03:32 KST). R0-2 SHA 일치, R0-3 잔여 5건 전수·KR 108490뿐이라 충돌 없음·전건
> retain·333430 고아 소멸, R0-5 토글 8-14와 동일, R0-6 soak ready=false(68일)·attestation 만료
> 기록. **R0-4 후보 공멸**: official 보유 전체 = TSLA 0.000154주 — 046890·466100 보유 0.
> 재선정 사람 결정(선택지 1) = **당일 장중 사람이 KR 유동성 종목 1개 3주 매수로 표본 조달**
> (M-A 발동 1 + a087 5.1 매도 1 + 5.2 경계 시도 1 — 보유 없는 시도는 거부 사유가 섞여 관측
> 오염). 종목은 09:00 직후 호가·틱 실측으로 확정, 제외 = 엔진 adoption 272210·333430 +
> 잔여 108490. 매수는 mutating — 전부 사람 실행. 참고: WTS 세션 만료 관측(auth extend 필요
> 상태) — M-A CLI 명령은 openapi 표면이라 무관, §3 콘솔 verify에서 세션 상태 재확인.

**역할 표기**: κ = 사람 콘솔 클릭, ⌨ = 사람 터미널(mutating 전부), 👁 = 에이전트 read-only 관측·기록.

> **2026-11-10 준비 착수 정정(사용자: 준비만 승인, live 세션 승인 아님).** blocker 원장은
> `measurement-prereq.md` 「2026-11-10 M-A 준비 착수 — blocker 원장」이다. 이 런북에서 결정 없이
> 고친 것은 셋이다. ① 모든 `tossctl` 명령과 측정 도구 명령에 `--config-dir ~/.config/tossctl`을
> 붙였다(2026-10-19 토큰 무효화·401 기록, `docs/handoff/2026-10-19-manager-handoff.md:124-127`).
> ② R1 §6 M0 명령에 코드가 요구하는 `--trigger-receipt-dir`를 적었다(`cmd/tossctl/verify.go:201-206`
> — `--include-trigger`에 필수). ③ R2 동승 A는 a087 5.1~5.3 완료로 소멸했다고 적었다.
> **verify 명령의 `--config-dir`는 기록 경로까지 바꾸므로 Q5가 정해지기 전에는 그 명령을 쓰지 않는다.**
> 아래 R0-2의 SHA, R0-4의 후보, 2026-10-07 표본 문단은 2026-11-10 현재 사실이 아니다(원장 B1·B3).

## Manager 결정 필요 (2026-11-10 — 답은 고르지 않았다)

각 질문은 서로 맞지 않는 두 문장과 고르는 데 필요한 코드 증거만 적는다. 분기를 근거로 쓴 함수는
`analysis/ma-prep-2026-11-10/citation-ast/`의 `tools/logic-map` AST(HEAD `6dbb752f6`)로 열거했다.
「Teammate 의견」으로 표시한 줄만 이 문서 작성자의 의견이고 나머지는 사실 기록이다.

### Q1. M-A의 parent 조건주문은 누가 만드는가 — CLI `order conditional place`인가, M0 verify인가

- 충돌하는 문장:
  - `ma-runbook.md:63-65`(6dbb752f6 판)(R1): "§5 ⌨ 등록(preview → 사람 확인 → `--execute --confirm`, trigger 는
    현재가 1~2틱 아래 SELL stop) → §6 👁 M0 관측 루프(`--include-trigger --confirm-each --resume
    --redo conditional-trigger` 만, …)". 역사본 §5(6dbb752f6 판 `:197-213`)는 그 등록을 `tossctl order conditional
    place … --execute --confirm <TOKEN>`으로 적는다.
  - `tasks.md:90-92`(0.2a.6): "human gate 승인 뒤 broker create **전에** unique client ID, run ID와 approved
    market/symbol/SINGLE/MARKET/SELL/qty=1/trigger/expiry를 owner-only verify record의 pending intent로
    append+fsync한다." `design.md:581-583`(D10-A)도 같다.
- 코드 증거(AST 열거 기준):
  - `verifylive.Runner.stepConditionalTrigger`(`internal/verifylive/steps_trigger.go:91-166`, 분기 6)는
    L144에서 `r.createConditional`을 불러 **자기 parent를 새로 만든다.** 그 전에 L97 매도가능수량,
    L110 호가, L114 `NearStopTrigger`를 읽고, 본문은 `SINGLE`·`MARKET`·수량 `MinQuantity`·SELL·
    `r.expireDate()`로 고정된다(L124-132).
  - `verifylive.Runner.createConditional`(`internal/verifylive/mutate.go:503-556`, 분기 11)은 L508 `r.gate`
    (사람 확인) → L515 `appendM0Checkpoint(Kind "pending-create")` → L532 `r.broker.CreateConditionalOrder`
    → L545 `appendM0Checkpoint(Kind "parent-created")` 순서로 호출한다. 0.2a.6의 pending-before-create는
    이 경로에만 있다(`"pending-create"` 쓰기 자리는 비시험 소스에서 `mutate.go:516` 하나, `rg` 기준).
  - CLI 경로 `trading.Service.ConditionalPlace`(`internal/trading/conditional.go:123-131`, 분기 2)는 L124
    `ConditionalGate`와 L130 `s.conditional.CreateConditionalOrder`만 부른다. verify 기록·M0 checkpoint
    호출이 없다. 이를 부르는 cobra 명령은 `newOrderConditionalCmd`(`cmd/tossctl/order.go:456-647`)의
    `place` 하위 명령이다.
  - `verify run`에는 이미 있는 parent id를 넘기는 flag가 없다(`cmd/tossctl/verify.go:185-215`의 flag 목록).
  - M0 parent의 trigger는 「현재가 1~2틱 아래」가 아니라 **최우선 매수호가 바로 위 첫 틱**이며
    마지막 체결가 이하여야 한다(`NearStopTrigger`, `internal/verifylive/pricing.go:221-256`, 분기 5 —
    B3 `bid >= last`·B4 틱 격자에 자리 없음이면 오류로 step skip). expire는 당일이 아니라 KST 기준
    **오늘+7일**이다(`internal/verifylive/steps.go:1253-1256`).
  - `cmd/tossctl/verify.go`와 `internal/verifylive`의 비시험 소스에 `trading.conditional` 토글 참조가
    `rg`로 0건이었다. M0 경로가 §4의 토글 flip을 요구하는지는 이 정적 검색만으로 확정하지 않는다
    (호출 그래프 미추적 — 미검증).
- 선택지:
  1. M0가 parent를 만든다. §5 CLI 등록을 런북에서 빼고, §6 한 번의 `verify run`이 등록·관측을
     함께 한다. 0.2a.6·D10-A와 맞는다. trigger·expire는 위 코드가 정한다.
  2. CLI `order conditional place`로 만든 parent를 관측한다. 이 parent에는 pending intent·checkpoint가
     없고, M0가 그것을 받을 입력이 현재 코드에 없다. 코드 변경이 필요하며 0.2a 재리뷰 대상이다.
- Teammate 의견: 선택지 1이 현재 코드·tasks·design과 맞는다. 2는 새 구현이다.

### Q2. soak/attestation은 M-A 실행의 필수 입력인가, 참고 기록인가

- 충돌하는 문장:
  - `measurement-prereq.md:68-72`: 다음 세션의 필수 fresh input 중 "(6) binary와 일치하는
    soak/attestation authority" — "하나라도 없으면 preview나 config flip을 열지 않는다."
  - `ma-runbook.md:56-57`(6dbb752f6 판)(R0-6): "👁 soak/attestation 상태 기록(참고용 — attestation 은 2026-08-29 만료.
    재발급은 오늘 하지 않는다 …)".
- 코드·문서 증거:
  - `cmd/tossctl/verify.go`와 `internal/verifylive`의 비시험 소스에 `attest` 참조가 `rg`로 0건이었다
    (정적 검색 — 미검증). 즉 verify 명령이 attestation을 읽어 거부한다는 근거는 찾지 못했다.
  - tasks 0.10(`tasks.md:203-205`)과 D11은 재발급 순서를 「verify DELETE·modify 증거 → a100
    바이너리로 `soak attest` → 배포」로 고정한다. 그 순서에서 attestation은 M-A **뒤**에 온다.
  - 2026-10-07 R0-6 기록은 `ready=false`·만료였다(`ma-runbook.md:29-30`(6dbb752f6 판)). 현재 상태는 미확인이다.
- 선택지:
  1. 필수 — soak ready와 측정 바이너리에 결속된 attestation이 생길 때까지 M-A를 열지 않는다.
  2. 참고 — 상태만 기록하고 M-A를 막지 않는다. measurement-prereq.md (6)을 정정한다.
- Teammate 의견: 없음(운영 정책 판단).

### Q3. 새 표본 종목을 엔진 편입에서 어떻게 떼어 두는가

- 낡은 문장:
  - `tasks.md:180-185`(0.9): "`include_symbols = [272210, 333430, IONQ, TSLA]` — **제외 목록이 아니라
    포함 목록이므로 a100의 대상은 정확히 이 4종목이다.**"
  - `ma-runbook.md:32-33`(6dbb752f6 판): "종목은 09:00 직후 호가·틱 실측으로 확정, 제외 = 엔진 adoption 272210·333430 +
    잔여 108490."
- 코드 증거(`ReconcileDriver.judgeHoldings`, `internal/app/engine/adoption.go:107-193`, 분기 16):
  - B7(L140) `p.ExitEligible()`이면 수량 증가 검사만 하고 `continue`한다. 이 분기가 B12보다 **앞**이므로
    이미 편입된 포지션은 exclude 목록에 나중에 넣어도 이 함수에서 미관리로 돌아가지 않는다.
  - B12(L163) `Adoption.Excludes(symbol)`가 참이면 미관리. B13(L171) `!Enabled && !Included(symbol)`이
    참일 때만 미관리. `enabled=true`이면 B13은 항상 거짓이라 exclude 밖의 보유가 전부 후보가 된다.
  - `internal/config/engine.go:115-125` 주석: exclude는 "never adopted", include는 "adoption candidates even
    with Enabled false".
  - 2026-10-07 표본 010170은 편입된 뒤 엔진 손절이 같은 날 전량 매도했다(STORY-TOS-a129·a130).
  - exclude 변경은 콘솔 설정 경로에서 audit 기록을 남긴다(`cmd/tossctl/adoptionsettings.go:88-91`).
    변경이 실행 중 엔진에 언제 반영되는지는 확인하지 않았다(미확인).
- 선택지:
  1. 표본 매수 **전에** 사람이 그 종목을 `adoption.exclude_symbols`에 넣고, 엔진이 그 값을 읽었음을
     확인한 뒤 매수한다(B7 순서 때문에 매수 뒤 추가는 늦다).
  2. 측정 세션 동안 `adoption.enabled`를 끈다. 모든 미편입 보유에 영향이 있다.
  3. 편입을 허용한다. 엔진 손절과 M-A 조건주문이 같은 주식을 두고 경쟁한다(10-07에 엔진이 먼저 팔았다).
- Teammate 의견: 선택지 1이 영향 범위가 가장 좁다. 설정 변경은 운영 설정이므로 사람 승인과 audit가
  필요하다(§0-5·§0-7).

### Q4. 어느 바이너리로 측정하는가

- 낡은 문장: `ma-runbook.md:41-45`(6dbb752f6 판)(R0-2) "설치본(`/home/daniel/.local/bin/tossctl`, SHA `899a74ac…e882`,
  commit `882a0b49` …)으로 측정한다."
- 증거: 현재 설치본은 SHA-256 `5287fe3c…3807`, `vcs.revision=488b370f`, `vcs.modified=true`다(원장 B1).
  `899a74ac` 판의 파일과 8/15 artifact·receipt 디렉터리는 2026-11-10 현재 없다. a144 TTL-edge 예외
  `6f06eee8e`는 `488b370f`와 `882a0b49`보다 뒤다(원장 B2). 예외 이전 판의 `M0ExactPrerequisites`
  (`internal/verifylive/m0_recovery.go`, 분기 3)는 trigger 외 **모든** step의 PASS를 요구하므로, 기록에
  TTL-edge PASS가 없으면 M0 진입을 거부한다.
- 선택지:
  1. 현재 main(≥`6f06eee8e`)에서 재현 가능한 후보를 새로 만들고, 8/15과 같은 별도 적대 검토와
     사람 승인 원자 설치를 거친다.
  2. 현재 설치본을 쓴다. dirty라 소스 결속이 없고 R0-2 신원 게이트와 맞지 않는다.
  3. `882a0b49` 판을 다시 만든다. 기록에 TTL-edge PASS가 있어야만 M0가 열린다.
- Teammate 의견: 없음(설치는 사람 승인 작업).

### Q5. `--config-dir`를 붙인 verify 명령은 어느 기록을 쓰는가

- 증거: `resolveVerifyRecordFor`(`cmd/tossctl/verify.go:765-778`, 분기 3) — B1 `--record`가 있으면 그 경로,
  B2 `--config-dir`가 있으면 `<config-dir>/capability-verify.jsonl`, 아니면 `journal.DataDir()` 아래 같은 이름.
  2026-10-19 이전 런북·역사 명령은 `--config-dir` 없이 쓰였다. 2026-10-20 기록은 KR 기록 파일이 두 곳에
  모두 있다고 적는다(`docs/remaining-work-reconciliation-2026-10-20.md:14`). M0 진입(이전 PASS·outstanding·
  M0 checkpoint)은 고른 기록으로 판정된다. 콘솔 `/verify`(§3)가 어느 기록을 쓰는지도 확인하지 않았다.
- 선택지:
  1. `--config-dir ~/.config/tossctl`만 붙인다 — 기록은 config 디렉터리 쪽이다.
  2. `--config-dir ~/.config/tossctl --record <data dir 기록 경로>`로 기존 기록을 명시한다.
- 필요한 실측(read-only): 두 KR 기록의 step별 최신 판정·outstanding·M0 checkpoint를 대조한다.
- Teammate 의견: 없음. 실측 전에는 판단할 근거가 없다.

## R0. 재발행 preflight — 당일 장 전/직후, 전부 read-only receipt 로 동결

1. 👁 시장: 당일이 KR 정규장(09:00~15:30 KST)인지, expire = 당일로 재동결.
2. **측정 바이너리 — 결정 종결 (a)(2026-10-07 사용자 동의).** 설치본
   (`/home/daniel/.local/bin/tossctl`, SHA `899a74ac…e882`, commit `882a0b49`,
   2026-08-15 설치 게이트 통과본)으로 측정한다. 근거: M-A 핵심 증거인 M0 transport 다섯
   플래그가 그 빌드에 완결, 설치 이후 커밋은 verifylive·문서뿐이라 conditional 주문 경로
   기여 0. 당일 신원 재확인만 남음: `sha256sum ~/.local/bin/tossctl` = 위 SHA 일치 확인.
   **2026-11-10: 일치하지 않는다**(현재 `5287fe3c…`, revision `488b370f`, dirty — 원장 B1·B2). 이 결정은
   그 설치본이 남아 있다는 전제였으므로 Q4가 정해지기 전에는 측정 바이너리가 없다.
3. 👁 잔여 조건주문: `tossctl --config-dir ~/.config/tossctl order conditional list --status OPEN` — 8-15 관측은 5건
   (PAUSED 2·WATCHING 3). **기본 retain (2026-10-07 사용자 동의)** — WATCHING 건은 실보유의
   손절 역할일 수 있어 cancel 은 보호 약화(안전 불변식 4). 섞임 방지는 취소가 아니라 분리:
   M-A 관측은 당일 등록한 주문 id 로 추적하고, 잔여 건과 심볼이 겹치면 그 심볼을 M-A
   대상에서 제외한다. 당일 재조회 값이 기준 — 5건이 그대로라는 보장은 없으니 재조회 후
   건별 확정(예외적 cancel 은 ⌨ 사람 실행).
4. 👁 보유·sellable: official holdings 재조회, 대상 후보(미관리 whole-share KR — 8월 기준
   046890·466100)의 보유·sellable 수량 완결. 후보가 비면 사람과 재선정.
5. 👁 토글: `trading.conditional=false` 확인(flip 은 §4 에서 사람). `allow_live_order_actions`,
   `sell/place/cancel` 상태 기록.
6. 👁 soak/attestation 상태 기록(참고용 — attestation 은 2026-08-29 만료. 재발급은 오늘 하지
   않는다: 0.10(d)는 a100 바이너리로 돌려야 하므로 오늘은 증거 조달까지, 유효기간은 오늘+30일).

## R1. M-A 본체 (역사본 §1~§8 순서 준용 — 명령·값은 당일 동결본으로)

- §1 👁 잔여물·보유 확인(개장 전 가능) → §2 👁 유동성 실측으로 1종목 확정(스프레드·틱 —
  얇으면 자동 결정 금지, 사람에게 재질의) → §3 κ verify 이어하기(persist 관측 →
  conditional-modify → conditional-cancel 증거 — 5분 창 내 승인 클릭까지) → §4 ⌨
  `trading.conditional` true 로 flip → §5 ⌨ 등록(preview → 사람 확인 → `--execute --confirm`,
  trigger 는 현재가 1~2틱 아래 SELL stop) → §6 👁 M0 관측 루프(`--include-trigger
  --confirm-each --resume --redo conditional-trigger` 만, raw 응답·receipt seq·monotonic 시각) →
  §7 기록·4갈래 판정 → §8 ⌨ 원복(false) + 잔여물 0 확인.
- §6 명령의 현재 코드 형태(2026-11-10 정정 — `--trigger-receipt-dir`는 `--include-trigger`에 필수,
  `cmd/tossctl/verify.go:201-206`; `--record`는 Q5, §5와의 관계는 Q1이 정하기 전 실행 금지):
  `tossctl --config-dir ~/.config/tossctl verify run --market KR --holding-symbol <SYM> --include-trigger
  --confirm-each --resume --redo conditional-trigger --trigger-receipt-dir <현재 uid 소유 0700 디렉터리>
  [--record <Q5 결정>]`
- **판정 효과**: 완전 통과만 tasks 0.2 체크 + 0.11 raw-status 판정표 동결 → T2-A 개방
  (단, 2.7 D3 실측 결함 수리가 배선 선행). 부분 통과·4b 실패·실패는 D2/D8/3.10 재설계로.

## R2. 동승 A — a087 §5 (5.1·5.2, 주문별 사람 승인)

> **2026-10-10 수단 확정(a087 2차 재리뷰 P1-1 + 사용자 결정):** `tossctl order place` 는
> `placeIntentSupported`(`internal/trading/service.go:281`) 가 비지정가를 로컬 거부해 **5.1·5.2 에
> 못 쓴다**. 실행 수단 = `tools/a087-market-sell-probe/`(2단 confirm·단일 전송·영수증 JSON,
> 사람 터미널 전용). 도구가 착지·검증되기 전에는 R2 는 큐에서 보류 — M-A 본체는 영향 없음.
>
> **2026-11-10 정정: R2는 끝났다.** `openspec/changes/a087-a-protective-exit-is-a-market-order/tasks.md`의
> 5.1·5.2·5.3이 모두 `[x]`다. M-A 세션에 동승하지 않는다. 아래 항목은 역사 기록이다.

- 5.1 ⌨ **KR MARKET 매도 1회**: 최소 수량·장중. 종목은 사용자와 합의(§2 유동성 실측 결과
  재사용 가능). 성공·실패 모두 기록.
- 5.2 👁/⌨ **세션 경계**: 같은 날 정규장 밖(마감 직후)에 KR MARKET 매도 시도 1회 — 응답 코드
  관측(`422 order-hours-closed` 인지). 시간외단일가에 시장가 없음 주의. 이것도 mutating 시도라
  사람 실행·사람 승인.
- 5.3 기록처: `openspec/changes/verify-execution-capability/measurements.md` M계열
  (`docs/trading/measurements.md` 는 존재하지 않는다).

## R3. 동승 B — a121 Q1 (조회 전용, 표본 = R1 의 발동 그 자체)

- **표본 금지**: a063 심볼 **333430 으로 발동을 만들지 않는다**(freeze P1-5 — 같은 심볼 발동
  이력이 CLOSED 에 남으면 G1-2 가 a063 artifact 를 영구 거절). R1 대상이 다른 심볼이면 자동 충족.
  §3 의 333430 고아 **cancel** 은 발동이 아니므로 무관.
- 👁 발동 직후부터 간격 재조회(당일 수 회 + 이후 일 단위)로: ① CLOSED 발동 잔존 여부
  ② 보존 기간 하한 ③ **축출 모형**(기간 기반인지 행 수 기반인지 — codex F5: 행 수 축출이면
  기간만으로 배제 불가) ④ **CLOSED status 어휘 영수증**(발동·취소·만료가 각각 어떤 raw status
  문자열인지 — a121 allowlist 상수의 유일 출처).
- 효과: 측정 영수증 → a121 활성화 상수 후속 change(리뷰 거쳐 Q1 상수 커밋) → `verify
  reconcile` 거절 전용 해제. 축출 모형이 행 수 기반·판별 불능이면 활성화 불가로 기록.

## R4. 동승 C — a112 L1c 프로브 (read-only, 1회로 수락 성립)

- ⌨ 장중 1회: `go run ./tools/a112-l1c-quote-probe --config-dir ~/.config/tossctl --market KR --symbol <R1 종목>` —
  결정 46 잔여(9/2 실행은 두 시장 마감이라 수락 아님). 산출 보고서를 a112 에 기록.

## 하지 않는 것 (역사본 승계 + 추가)

- 에이전트는 `mutating: true` 명령을 실행하지 않는다(preview 포함 전부 사람 터미널).
- attestation 재발급 금지(위 R0-6). US 는 범위 밖. WTS 를 official 증거로 쓰지 않는다.
- 333430 으로 발동 표본을 만들지 않는다(R3). 엔진 관리 4종목과 섞지 않는다.

---

# M-A 실행 명령서 — 2026-08-14 역사본(만료, 실행 금지)

> **2026-11-10 기계적 정정:** 이 역사본의 `tossctl` 명령에 `--config-dir ~/.config/tossctl`을 덧붙였다
> (복사 실행 금지는 그대로다). 다른 글자는 바꾸지 않았다. 원문은 `6dbb752f6` 판에 있다.

> **이 명령서는 실행 권위가 아니다.** 아래 날짜·expire·종목 후보·보유·잔여 주문·토글·attestation은
> 2026-08-14 세션 전제라 2026-08-15 이후 재사용할 수 없다. 새 세션은
> `measurement-prereq.md`의 2026-08-15 재발행 규칙에 따라 read-only receipt로 시장·보유·sellable·
> 잔여물·토글·capability·새 expire를 다시 동결해야 한다. 그 뒤에도 등록·정정·취소마다 사람이
> 즉시 승인하기 전에는 어떤 mutating 명령도 실행하지 않는다. 아래 명령은 형식 참고용 historical
> record일 뿐이며 복사 실행을 금지한다.
>
> **2026-08-15 추가 정정:** 당시의 `tossctl order get <child-id>` 표기는 실제 CLI 명령이 아니다.
> `order show`는 WTS surface라 M-A official causal evidence가 될 수 없다. 다음 세션은 tasks 0.2a의
> M0가 official transport의 body-read-complete에서 parent/child raw result와 모든 attempt를 기록하고,
> broker create 전 pending client intent, child GET 전 exact parent/child cleanup checkpoint를 owner-only
> verify record에 fsync한다. prior outstanding가 있으면 cleanup prologue도 실행하지 않는다. M0가
> GREEN·A-M0 accepted되기 전에는 이 역사본의 place 절차를 열지 않는다.
>
> **2026-08-15 artifact 갱신:** reviewed M0 source의 재현 가능한 설치 후보는
> `/tmp/a100-m0-artifact.M8EeNi`에 봉인됐고 별도 Terra reviewer가 `P0=0/P1=0/P2=0`으로
> 수용했다. 그러나 PATH binary 교체는 실행하지 않았으며 이 역사본의 날짜·종목·expire·잔여물은
> 여전히 만료 상태다. 후보 설치에는 별도 명시 승인, 설치 뒤에는 새 장중 read-only preflight와
> exact 주문별 승인이 필요하다. artifact acceptance를 아래 명령의 실행 승인으로 해석하지 않는다.
>
> **설치 후 정정:** 이후 사람 승인으로 candidate SHA `899a74ac…e882`를
> `/home/daniel/.local/bin/tossctl`에 원자 설치했고 기존 SHA `b0e805f3…96f9`는 같은 directory의
> 검증된 backup으로 남겼다. 설치 adversary 최종 판정은 `P0=0/P1=0/P2=1`이다. 이로써 binary
> identity gate만 닫혔으며, 이 역사본의 실행 금지와 새 세션·새 주문별 승인 요구는 변하지 않는다.
>
> **2026-08-15 fresh preflight 정정:** corrected receipt
> `/tmp/a100-ma-preflight-corrected.kSn14D`에서 official OPEN은 terminal 5건
> (`PAUSED` 2, `WATCHING` 3)으로 확인됐다. 하지만 KR 폐장, 각 residual의 사람 retain/cancel 결정
> 부재, official sellable·미관리 후보 미완결, soak/attestation unready·binary-unbound 때문에
> 운영 판정은 HOLD다. 이 역사본의 후보·expire·명령을 재사용하거나 preview·toggle·verify run을
> 시작하지 않는다.

`measurement-prereq.md`의 M-A 절차를 실행 가능한 명령으로 편성한 것이다. 절차의 정본은
measurement-prereq.md이고, 이 문서는 그날의 손 순서다. 사용자 승인(2026-08-14, "권장사항
대로 진행"): 일정 = 8/14 KR 장중, 대상 = **미관리 보유 중 유동성 실측으로 고르는 1종목**
(046890 또는 466100 — 엔진 exit 관리 대상 4종목과 섞지 않는다).

**역할 분담.** 표기 κ = 사람의 콘솔 클릭, ⌨ = 사람의 터미널 실행(`mutating: true` 명령과
config 편집은 전부 사람 몫), 👁 = 에이전트 read-only 관측·기록. **§0-1·§0-7: 실계좌 주문은
실행 직전 그 주문에 대해 사람이 별도 승인한다. 이 문서에 대한 동의는 그 승인이 아니다.**

## 사전 상태 (2026-08-14 00시대 확인)

- `trading.conditional = false` — **flip 대상은 이 키 하나다.** `allow_live_order_actions`는
  이미 true, `sell`·`place`·`cancel`도 true (`amend`는 false지만 조건주문 verb는
  `trading.conditional` 단일 게이트라 무관 — `internal/trading/conditional.go` 패키지 주석).
- 333430 고아 조건주문 `DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg` 생존
  (soak by-id probe ok, 2026-08-13T14:35Z 사이클).
- verify KR 기록: `conditional-persist = awaiting-restart` (2026-07-31 이후 무변화).
  콘솔은 그 뒤 여러 번 재시작됐으므로 [이어하기]가 바로 진행 가능해야 한다.
- attestation 만료 2026-08-29. 이 세션의 verify 단계가 `POST …/modify`·`DELETE` 증거를
  조달하고, **오늘 날짜가 그 증거의 30일 유효기간 기준일이 된다.**

## 순서

### 1. 👁 잔여물·보유 확인 (개장 전 가능)

- `tossctl --config-dir ~/.config/tossctl order conditional list --status OPEN` — 잔여 조건주문 목록·시각 기록
  (measurement-prereq 단계 0). 예상: 333430 한 건.
- 보유 확인: 콘솔 /positions 또는 read 명령. 대상 후보 046890·466100의 보유 수량 기록.

### 2. 👁 유동성 실측 → 종목 확정 (09:00 직후)

- `tossctl --config-dir ~/.config/tossctl quote get 046890` / `tossctl --config-dir ~/.config/tossctl quote get 466100` — 호가 스프레드.
- `tossctl --config-dir ~/.config/tossctl quote trades <sym>` — 틱 간격(체결 빈도).
- **판정: 스프레드가 좁고 틱이 조밀한 쪽 1종목.** 둘 다 얇으면 발동 대기가 길어지고
  INCONCLUSIVE 위험이 커진다 — 그 경우 관리 대상이지만 유동성이 확실한 종목으로의 변경을
  사람에게 다시 물어 결정한다(자동 결정 금지).

### 3. κ verify 이어하기 — 고아 정리 + modify·DELETE 증거 (한 번에)

1. 콘솔 `/verify` (KR 기본 화면) → **[이어하기]**.
2. 승인 목록에 남은 단계(persist 관측 → conditional-modify → conditional-cancel)가 보이는지
   확인 → **[위 목록을 승인하고 실행] 클릭** (5분 창 — 클릭까지 해야 실행이다).
3. 👁 완료 후 기록 확인: modify가 새 id를 발급하고 옛 id 404(M40 재확인), cancel 완료,
   잔여물 0. **이 두 호출이 supervised 증거가 된다** — M-A의 CLI 호출은 verify 기록에
   실리지 않으므로 증거 조달은 이 단계뿐이다.

### 4. ⌨ config flip (사람)

`~/.config/tossctl/config.json`에서 `"conditional": false` → `true` 한 키만 편집.
(CLI는 호출 시점에 읽으므로 구동 중 엔진·콘솔·soak에는 영향 없다. 엔진은 이 빌드에서
조건주문을 만들지 않는다.)

### 5. ⌨ M-A 본체 — 등록 (사람, 실행 직전 별도 승인)

```bash
# ① preview — confirm token을 받는다 (브로커 전송 없음)
tossctl --config-dir ~/.config/tossctl order conditional place --symbol <SYM> --type SINGLE --order-type MARKET \
  --qty 1 --first-side SELL --first-trigger <T> --expire 2026-08-14

# ② 사람이 preview 내용을 확인·승인한 뒤 실행
tossctl --config-dir ~/.config/tossctl order conditional place --symbol <SYM> --type SINGLE --order-type MARKET \
  --qty 1 --first-side SELL --first-trigger <T> --expire 2026-08-14 \
  --execute --confirm <TOKEN>
```

- `<T>` = 현재가 1~2틱 **아래** (SELL stop은 가격이 trigger 이하로 내려오면 발동 —
  곧 발동하게 근처에 둔다). 등록 시각·요청 원문·응답 id를 기록.
- 가격이 올라 발동하지 않으면: trigger를 현재가 쪽으로 올리는 `modify`(사람 실행)로
  따라붙는다 — 이 정정 자체도 관측 대상이다(새 id 발급 여부).

### 6. 👁 관측 루프 (에이전트, read-only)

- 등록 직후: `tossctl --config-dir ~/.config/tossctl order conditional get <id>` — trigger·수량이 요청과 비교 가능한
  형태로 돌아오는지(단계 4b, **실패 시 tasks 3.3 수렴 정의 재설계**). sellable-quantity
  재확인(M13 재확인 — 예약 없음 예상).
- 이후 M0의 기존 verify trigger poll만 사용한다. M0 mode는 exact
  `--include-trigger --confirm-each --resume --redo conditional-trigger`이며 `--include-ttl-edge`나 다른
  redo를 함께 쓰지 않는다. transport body-read 경계에서 raw 응답·모든 401/429 attempt,
  `triggeredOrderId` local receipt sequence/monotonic 시각을 기록한다.
- child 체결: 역사본의 `order get`/WTS `order show`를 사용하지 않는다. M0가 official child by-id
  raw GET을 호출해 raw payload digest와 local receipt sequence/monotonic 시각, 체결 수량·상태·가격을
  기록한다. parent child-id receipt의 fsync가 child request start보다 앞서고 그 receipt seq가 child
  first-observed-fill seq보다 엄격히 작아야 하며 broker server timestamp만으로 판정하지 않는다.
- parent child-ID receipt부터 child fill receipt까지 parent 404/read gap은 HOLD다. durable child fill 뒤
  parent terminal GET은 요구하지 않는다. **pre-trigger `PAUSED`가 관측되면 0.11 결정의 입력이다.**

### 7. 기록·판정

- 관측 전량을 `measurement-prereq.md` M-A 표 형식으로 이 change에 기입(에이전트),
  판정은 4갈래(통과/부분 통과/4b 실패/실패) 중 하나를 명시.
- **완전 통과가 아니면 a100을 멈춘다**(design D10). 부분 통과와 4b 실패도 tasks 0.2를 완료하지
  않으며, D2/D8/3.10 재설계와 별도 proposal-freeze 적대 리뷰 전에는 T1을 열지 않는다. 완전
  통과만 tasks 0.2 체크와 0.11 raw-status 판정표 동결 뒤 Teammate/적대 리뷰 파이프라인을 연다.

### 8. ⌨ 원복 (사람)

`"conditional": true` → `false` 되돌림. 잔여물 0 확인(`tossctl --config-dir ~/.config/tossctl order conditional list --status OPEN`).

## 하지 않는 것

- 에이전트는 `mutating: true` 명령을 실행하지 않는다(preview 포함 전부 사람 터미널).
- attestation 재발급(`soak attest`)은 오늘 하지 않는다 — 0.10 (d)는 **a100 바이너리**로
  돌려야 M-A 증거가 실린다. 오늘은 증거 조달까지다(마감: 오늘 + 30일).
- US는 이 측정 범위 밖이다. KR 결과를 US로 옮겨 쓰지 않는다.
