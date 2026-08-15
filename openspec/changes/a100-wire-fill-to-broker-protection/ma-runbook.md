# M-A 실행 명령서 — 2026-08-14 역사본(만료, 실행 금지)

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

- `tossctl order conditional list --status OPEN` — 잔여 조건주문 목록·시각 기록
  (measurement-prereq 단계 0). 예상: 333430 한 건.
- 보유 확인: 콘솔 /positions 또는 read 명령. 대상 후보 046890·466100의 보유 수량 기록.

### 2. 👁 유동성 실측 → 종목 확정 (09:00 직후)

- `tossctl quote get 046890` / `tossctl quote get 466100` — 호가 스프레드.
- `tossctl quote trades <sym>` — 틱 간격(체결 빈도).
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
tossctl order conditional place --symbol <SYM> --type SINGLE --order-type MARKET \
  --qty 1 --first-side SELL --first-trigger <T> --expire 2026-08-14

# ② 사람이 preview 내용을 확인·승인한 뒤 실행
tossctl order conditional place --symbol <SYM> --type SINGLE --order-type MARKET \
  --qty 1 --first-side SELL --first-trigger <T> --expire 2026-08-14 \
  --execute --confirm <TOKEN>
```

- `<T>` = 현재가 1~2틱 **아래** (SELL stop은 가격이 trigger 이하로 내려오면 발동 —
  곧 발동하게 근처에 둔다). 등록 시각·요청 원문·응답 id를 기록.
- 가격이 올라 발동하지 않으면: trigger를 현재가 쪽으로 올리는 `modify`(사람 실행)로
  따라붙는다 — 이 정정 자체도 관측 대상이다(새 id 발급 여부).

### 6. 👁 관측 루프 (에이전트, read-only)

- 등록 직후: `tossctl order conditional get <id>` — trigger·수량이 요청과 비교 가능한
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

`"conditional": true` → `false` 되돌림. 잔여물 0 확인(`order conditional list --status OPEN`).

## 하지 않는 것

- 에이전트는 `mutating: true` 명령을 실행하지 않는다(preview 포함 전부 사람 터미널).
- attestation 재발급(`soak attest`)은 오늘 하지 않는다 — 0.10 (d)는 **a100 바이너리**로
  돌려야 M-A 증거가 실린다. 오늘은 증거 조달까지다(마감: 오늘 + 30일).
- US는 이 측정 범위 밖이다. KR 결과를 US로 옮겨 쓰지 않는다.
