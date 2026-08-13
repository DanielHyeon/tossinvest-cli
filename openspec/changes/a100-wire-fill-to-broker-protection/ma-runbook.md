# M-A 실행 명령서 — 2026-08-14(금) KR 장중 09:00–15:30 KST

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
- 이후 **15초 간격** `get <id>` 폴링(429 회피 — order walk는 하지 않는다):
  상태 전이, `triggeredOrderId` 노출 시각(등록→노출 지연), child 주문 id.
- child 체결: `tossctl order get <child-id>` — 체결 시각·수량·가격.
- 발동 후 원 조건주문 최종 상태 문자열. **`PAUSED`가 관측되면 0.11 결정의 입력이다.**

### 7. 기록·판정

- 관측 전량을 `measurement-prereq.md` M-A 표 형식으로 이 change에 기입(에이전트),
  판정은 4갈래(통과/부분 통과/4b 실패/실패) 중 하나를 명시.
- **실패면 a100을 멈춘다**(design D10). 통과·부분 통과면 tasks 0.2 체크, 0.11 결정 기록
  → 1절부터 Teammate/적대 리뷰 파이프라인 가동.

### 8. ⌨ 원복 (사람)

`"conditional": true` → `false` 되돌림. 잔여물 0 확인(`order conditional list --status OPEN`).

## 하지 않는 것

- 에이전트는 `mutating: true` 명령을 실행하지 않는다(preview 포함 전부 사람 터미널).
- attestation 재발급(`soak attest`)은 오늘 하지 않는다 — 0.10 (d)는 **a100 바이너리**로
  돌려야 M-A 증거가 실린다. 오늘은 증거 조달까지다(마감: 오늘 + 30일).
- US는 이 측정 범위 밖이다. KR 결과를 US로 옮겨 쓰지 않는다.
