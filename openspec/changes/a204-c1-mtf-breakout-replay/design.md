# a204 설계 — 전략 C1 오프라인 재생 (StockOS 청산 원문 이식)

기준 HEAD `e2921fc00`(origin/main) · Python 3.12(표준 라이브러리만) · StockOS 복사 원본 커밋 **`fcad869ecc531a6de80777f59b6fb842a2891101`**.
StockOS 줄 번호는 모두 그 커밋의 원본 기준(`ES` = `packages/trading/stockos_trading/exit_strategy.py`, `AEE` =
`apps/api/stockos_api/auto_exit_execution.py`). 복사본은 머리 주석 1줄 때문에 원본보다 1줄 밀림.

## 0. 비례 원칙 · 적용 범위

- 위험 등급 Low. 파일만 읽는 오프라인 분석 도구. TossOS 주문 · 손절 · 원장 · 대사 · 인증 · 체결 경로와 게이트 판정 로직을 바꾸지 않음.
- **Function Logic Map: not-applicable** — 기존 Go 함수 본문 편집 0(새 Python 파일만, gate ⑤ required 0).
- **변이 원장 · 다중 적대 리뷰: not-applicable** — 비례 원칙(오프라인 도구, High-risk 경로 · 게이트 판정 변경 0). 시험 + 경량 gstack 리뷰 1회(최대 2회차).
- 기억 회고(2026-11-27): ai-memory `memory_query "StockOS exit port replay simulator offline python tool"` — 관련 결정 · 함정 0(세션 기록만).
  Hindsight `recall`(bank `tossos-dev-episodes`) — 관련 교훈 0. 둘 다 advisory.

## 1. 안전 경계

- 입력은 디스크 파일뿐(지금은 픽스처, 나중에 a203 캐시 — 읽기 전용). 출력은 `--output` 경로 하나(`open(…, "x")`, 덮어쓰기 거절).
- 금지: TossOS Go 코드 import, `.env` · 자격 증명 · `~/.config/tossctl` 읽기, 소켓 · HTTP · 하위 프로세스. 정적 시험(`tests/test_static.py`)이 도구
  소스의 import 를 허용 목록으로 묶고 금지 문자열을 셈.
- 설정 파일은 시뮬레이터 전용 — TossOS 런타임 설정 · 토글 · 엔진과 무관.

## 2. C1 진입 규칙 (사전등록 — 숫자 변경은 새 change)

시각은 모두 KST. 1분봉은 시작 시각 `start` 와 OHLCV 를 가지며 구간 `[start, start+1분)` 을 덮음. 정규장 봉 = `start` 가 09:00~15:19 인 연속매매 봉 + 15:30 종가 단일가 봉(15:20~15:29 동시호가 구간 봉은 체결 불가라 버림 — 리뷰 반영 2026-11-27). 「앞 20 봉」 · 「다음 10 봉」은 시각 칸이 아니라 **있는 닫힌 봉**을 셈(빈 칸이 있으면 창이 시간상 길어짐).

| 항목 | 규칙 |
|---|---|
| 종목군 | 세션마다 a203 design §5 규칙 U-A 의 30 종목 목록(어댑터가 줌, §5) |
| 봉 집계 | 1분봉을 09:00 기준 N분 칸(`(start−09:00)//N`)으로 묶음. 마지막 칸 번호로 자름(15:30 봉은 마지막 칸에 들어감). 칸 끝 시각 `end` 가 평가 시각 이하일 때만 **닫힌 봉**으로 씀. 칸 안 1분봉이 빠졌으면 있는 봉으로만 만듦 |
| 거시 필터 | 30분 닫힌 봉(전 세션 포함)의 **가장 최근 15개** 종가로 EMA5 · EMA15 계산 — α = 2/(n+1), 첫 값 = 15개 중 가장 오래된 종가, 재귀(pandas `adjust=False` 와 같음). EMA5 > EMA15 여야 함. 닫힌 30분 봉이 15개 미만이면 거래 없음. 평가 시각 = 진입 결정 시각(되돌림 봉 끝) |
| 돌파 봉 j | 3분 닫힌 봉, **당일 세션 봉만**. 앞 20개 닫힌 3분 봉이 있어야 함(첫 후보 = 10:00~10:03 봉). 조건 넷 모두: 종가 > 앞 20봉 최고가 L, 종가 > 시가, 윗꼬리(고가 − 종가) ≤ 몸통(종가 − 시가), 거래량 z = (v_j − 평균)/모표준편차(앞 20봉) ≥ 2.0. 모표준편차 0 → 신호 없음 |
| 되돌림 | 봉 j+1 … j+10(닫힌 3분 봉) 중 저가 ≤ L×1.003 인 **첫 봉**만 봄. 그 봉이 L×0.997 ≤ 저가 **이고** 종가 > L **이고** 종가 > 시가이면 되돌림 성공. 첫 되돌림이 이 조건을 못 채우면 setup 실패. 10봉 안에 저가 ≤ L×1.003 이 없으면 시한 초과 |
| 진입 결정 | 되돌림 봉 끝 시각 S 에서: (a) S ≤ 14:45, (b) 거시 필터 통과, (c) 되돌림 봉 종가 / 전 세션 종가 − 1 < +8%, (d) 시작이 S 인 1분봉이 있음. (a) 실패면 그 종목 그 세션의 탐색 종료. (b)~(d) 실패면 그 setup 만 종료 |
| 재무장 | setup 하나가 끝나면(시한 · 실패 · 거절) 그 끝 봉 **뒤**의 봉에서 다음 돌파를 찾음. setup 이 살아 있는 동안 새 돌파는 무시. 진입은 종목 · 세션당 1회, 물타기 · 불타기 없음 |
| 진입가 | 시작이 S 인 1분봉의 시가 = ask(a203 design §4 주 가정 P0 「그 분의 첫 거래가」). 매수 체결가 E = ask |
| 구조 손절 | 되돌림 3분 봉 저가 − 1틱(「한 틱 아래」는 `krTickMinor(가격 − 1)` — 20,000 원 아래는 19,990. 틱표는 a140 `sizing.go:17-34` `krTickMinor` — 2,000 · 5,000 · 20,000 · 50,000 · 200,000 · 500,000 원 경계에서 1 · 5 · 10 · 50 · 100 · 500 · 1,000 원). E 이상이면 AEE `_initial_stop_price`(AEE:2391-2404)대로 E − 3% |
| 진입 ATR% | StockOS 시그널 스냅샷 `atr_pct` 정의 재구현: 닫힌 **5분** 봉(전 세션 포함, 진입 시각 이하)의 마지막 15 봉으로 단순 평균 TR(14) / 마지막 종가 × 100 (strategy_signal_generator.py:443-488 · 6320-6345). 15 봉 미만이면 None(→ ATR breakeven 비활성, StockOS 와 같음) |

## 3. 체결 가정

- 매수: E = 진입 1분봉 시가(P0 ask). 매도: 그 틱 가격 p 의 1틱 아래 bid = p − `krTickMinor(p − 1)`(P0 의 1틱 스프레드, a203 design §4).
  편차: a203 은 bid = ask − `krTickMinor(ask)` 로 적었으나 가격대 경계(20,000 · 50,000 원 등)에서 그 식은 두 틱 이상 아래가 됨 — 한 틱 아래를 정확히 씀.
- 비용: StockOS `costs.py` 원문 `net_pnl`(FeeTaxModel 기본 — 매수 · 매도 수수료 0.015%, 매도세 0.20%, costs.py:73-90 · 287-300), 시장 `KOSPI`
  (KOSDAQ 와 요율 같음). 수량에 선형이므로 결과는 수량과 무관.
- 명목 수량 Q0 = 10,000 — 오직 StockOS 의 `int(수량 × 비율)` 반올림을 무시할 만큼 작게 하려는 값(ES:922 · 1072 · 1236-1241). `shortcut:` 주석으로 표시.
- 세션 마지막 정규장 봉까지 남은 포지션은 그 봉 종가의 bid 로 청산, 라벨 `EOD`. 장 마감 뒤 · 다음 날 모델링 없음(한계).

## 4. 청산 — StockOS 원문 + AEE 재구현

### 4.1 원문 복사 (`stockos_port/`)

| 복사본 | 원본(`packages/trading/stockos_trading/…`) | 이유 |
|---|---|---|
| `exit_strategy.py` | `exit_strategy.py` | 위임 목록 |
| `profit_ladder.py` | `profit_ladder.py` | 위임 목록 |
| `ladder/profile.py` | `ladder/profile.py` | 위임 목록 |
| `exit/stop_controls.py` | `exit/stop_controls.py` | 위임 목록 |
| `costs.py` | `costs.py` | 위임 목록 |
| `indicators/__init__.py` | `indicators/__init__.py` | `crossed_below` 가 사는 파일. 같은 이름의 `indicators.py` 도 있으나 패키지가 import 우선이므로 패키지 쪽이 실제 의존 |
| `models.py` | `models.py` | **추가 복사(편차)**: `costs.py` 의 `Side` · `indicators` 의 `Candle` 등이 의존. 표준 라이브러리만 씀 |
| `session_hours.py` | `session_hours.py` | **추가 복사(편차)**: `costs.py` 의 `ConfigurationError` 가 의존. 표준 라이브러리만 씀 |

- 모든 파일 첫 줄: `# stockos-port: source=<원본 경로> commit=fcad869e… — verbatim copy, do not edit`. 그 줄을 뺀 바이트의 sha256 = 원본 sha256
  (`stockos_port/MANIFEST.json`, 시험 `tests/test_port.py` — StockOS 저장소가 있으면 `git show <SHA>:<경로>` 와도 대조).
- **import 줄 수정: 0.** 원본 패키지 배치를 그대로 재현해 상대 import(`.costs` · `..profit_ladder` · `..models` 등)가 그대로 풀림.
- **편차(복사 아님)**: `stockos_port/__init__.py` · `exit/__init__.py` · `ladder/__init__.py` 는 자작 한 줄 docstring. 원본 `exit/__init__.py` ·
  `ladder/__init__.py` 는 evaluator · lifecycle · registry 등 무관 모듈을 import 하므로 복사하지 않음. 원본 `stockos_trading/__init__.py` 도 복사 안 함.

### 4.2 AEE 재구현 (`aee_exits.py`) — 한 틱의 순서

AEE 의 포지션 한 건 평가(AEE:1459-1680)를 같은 순서로 따름.

1. **고점 추적**(AEE:2464-2516): `high = max(entry, high, p)`(settle 창은 adoption 전용이라 해당 없음).
2. **floor ratchet**(AEE:2518-2542, 값은 `exit_trail_stop_pct` AEE:2456-2461): `trail = high × (1 − exit_trail_stop_pct/100)`,
   `trail > max(current_floor, initial_stop)` 일 때만 `current_floor = trail` · `protected_floor_enabled = True` — 올리기만 함.
3. **문맥**(AEE:2647-2781): `support_price = max(initial_stop, current_floor)`(AEE:2669-2673 · 2725), `minutes_held = ⌊(틱 시각 − 진입 시각)/60초⌋`(AEE:2698-2706),
   `max_hold_exit_enabled = False`(AEE:2724 — 라이브는 시간 청산 끔), `day_change_pct = (p − E)/E × 100`(AEE:2727 — 전일 대비가 아니라 진입 대비),
   `atr_pct = 0.8`(AEE:1956 · 2733-2739 — `exit_use_entry_atr_pct` 라이브 기본 false), `breakeven_atr_pct = 진입 ATR%`(AEE:2742-2746),
   trend_slope · rvol 미전달(→ ES 기본 0 · 1.0 → adaptive **neutral**), support 이탈 지속 시간(AEE:2585-2628).
4. **정책**: 설정 → `ExitPolicy`(AEE:1993-2053 + KRX 익절 AEE:2055-2091) → 세션 조임(AEE:2105-2176) → TP2 해제 `second_take_profit_pct = 10000`
   (AEE:286 · 6876-6884, StockOS 소유 포지션). 순서가 중요함 — 조임 뒤 TP2 해제라 마감 버킷에서도 TP2 는 꺼진 채(AEE:1487-1495).
   세션 버킷: **LATE_SESSION_PROTECT = 14:50 ≤ t < 15:20, CLOSE_MANAGEMENT = 15:20 ≤ t ≤ 15:30**(StockOS `gates.py:232-235` · `session_strategy.py:139-157`).
5. **EOD flatten**(AEE:1553-1564 · 1707-1755): `eod_flatten_enabled` 이고 CLOSE_MANAGEMENT 이면 전량 청산(보유 기간은 기본 intraday 로 가정).
6. **보호 floor 이탈**(AEE:1589-1603 · 2811-2866): `protected_floor_enabled` 이고 `p ≤ current_floor` 이고 확인 통과(AEE:2630-2645)면 전량 청산,
   라벨은 net 손익 > 0 이면 TAKE_FULL_PROFIT, 아니면 STOP_LOSS(AEE:2843-2847).
7. **`evaluate_exit`**(AEE:1616-1621, 원문): 분기 순서 ES:729-741, 우선순위 ES:1274-1314. 사다리 선택 시 상태 다음 값 보존(AEE:1622-1629).
8. **floor 올림**(AEE:3222-3256): MOVE_STOP_TO_BREAKEVEN · RAISE_PROFIT_LOCK · early-surge 부분익절의 `stop_price` 가 현재 floor 보다 높을 때만.
9. **체결**: 주문 가능 결정(ES:74-82)이면 `quantity_to_sell` 을 bid 로 체결. TAKE_PARTIAL_PROFIT 이면 `partial_tp_done = True`(AEE:6735-6744). 한 틱에 결정 하나.

기본 설정에서 실제로 살아 있는 값(neutral, ES:403-436): 하드 손절 −1.2 − 0.8×0.25 = **−1.4%**, 트레일 갭 4.5 + 0.2 = 4.7%, TP1 1.2% · 비율 0.5,
TP2 꺼짐, grace 1분(catastrophe −1.5%, ES:834-867), ATR breakeven k = 1.0(AEE:1943, ES:1093-1138), rapid lock(ES:439-493), 트레일 익절(ES:1002-1036).

재구현하지 않은 것(해당 없음): adoption settle 창 · 외부 보유 stop controls(AEE:1801-1914 — C1 은 StockOS 소유 진입), onset scalp(AEE:1217 —
레인 플래그 없음), 라이브 허용 목록 차단(AEE:1427-1453 — 고른 사다리는 허용됐다고 가정), `exit_use_entry_atr_pct` · `exit_consume_signal_tp`(라이브 기본 false 고정),
주문 재호가 · 보호 주문. **`COMMON_LADDER_HYBRID_50` 은 거절**(종료 코드 2) — runner 인계(AEE:1455-1458 · ES:653-680)의 체결 상태 동기화를 재구현하지 않음.

### 4.3 두 모드

- **주 모드 `live-effective`(판정)**: 1분봉마다 틱 경로 시가 → 저가 → 고가 → 종가를 **틱 단위로 걸으며** 각 틱을 시세 하나로 평가
  (StockOS 라이브는 시세마다 평가하고 사다리에 봉 고가 · 저가를 안 넘김, AEE:2717-2781 의 ExitContext 생성에 `bar_high` · `bar_low` 없음). 저가를 고가보다 먼저 걸으므로
  한 봉이 손절과 익절을 다 건드리면 **손절이 먼저**(profit_ladder `STOP_FIRST` 기본과 같은 보수 방향). Phase G 입력(EMA9 · VWAP) 미전달 → 꺼짐(라이브와 같음).
  한 봉 안 모든 틱의 시각 = 그 봉 `start`(grace · 세션 버킷 · 확인 초 계산용 — 1분 해상도 근사).
- **민감도 `as-designed`(보고만)**: 봉마다 종가 한 번 평가, `bar_high` · `bar_low` 를 사다리에 넘김(STOP_FIRST), 고점 추적에 봉 고가 포함,
  Phase G 입력을 당일 1분봉 EMA9 · 세션 VWAP(전형가 × 거래량)으로 채움. 청산 체결 = 종가 bid. 판정에 쓰지 않음.

## 5. 세션 자료 어댑터 (a203 의존 — 열린 항목 O1)

a203 은 캐시 디스크 형식을 아직 정하지 않았음(a203 design §3 · §8: 페이지마다 요청 인자 · `StrictMinutePage` · 본문 지문). 그래서 도구는 인터페이스만 고정함.

```
sessions()               -> 시간순 세션 날짜 전체(워밍업 포함)
sample()                 -> 판정 표본 세션(시간순, sessions 의 부분집합)
universe(d)              -> 세션 d 의 U-A 30 종목
minute_bars(sym, d)      -> 1분봉 목록 또는 None(결손)
prior_close(sym, d)      -> 전 세션 종가 또는 None
```

지금 구현은 픽스처 형식 `a204-fixture/v1` 하나: `manifest.json`(`format` · `sessions` · `sample` · `universe`) +
`bars/<종목>/<세션>.json`(`prior_close` · `bars: [["HH:MM", o, h, l, c, v], …]`). a203 캐시 어댑터는 O1(tasks 4.1).
**결손** = 종목-세션의 1분봉이 없거나 비었거나 `prior_close` 가 없거나, **직전 세션(d−1)의 1분봉이 없음**(30분 EMA 워밍업 15 봉은 당일만으로 14:45 전에 채울 수 없음 — 거시 거절로 숨기지 않고 결손으로 셈, 리뷰 반영). 세션 안 빠진 분은 결손이 아님(집계가 있는 봉으로 함).
**O1 계약 요구**: a203 캐시 어댑터는 표본 세션 d 의 종목군 전원에 대해 d−1 (가능하면 d−2) 1분봉을 줘야 함 — a203 은 세션마다 종목군을 다시 뽑으므로 d−1 에 종목군이 아니었던 종목의 d−1 봉이 캐시에 없을 수 있음(a203 쪽 수집 범위 확인 필요).
**입력 검증(경계 한 곳)**: sample 중복 · 표본 세션의 종목군 없음 · 종목 중복 · 종목 코드가 `[0-9A-Za-z]{1,12}` 가 아님 · NaN/Infinity · OHLC 불일치(고가 < max(시가, 종가) 등) · 0 이하 가격 · 분 시각 중복/역순 · 6칸이 아닌 행은 InputError(종료 코드 2).

## 6. 1R 과 결과 단위

- R% = (E − 보호선)/E × 100, 보호선 = max(구조 손절, E × (1 + 유효 하드 손절/100)). 유효 하드 손절 = `compute_effective_hard_stop_pct(adapt_exit_policy(진입 문맥, 정책).hard_stop_loss_pct, 진입 ATR%, exit_stop_atr_multiplier)`(ES:177-196 · 314-436, 진입 시각의 세션 정책).
- 거래 결과: net% = Σ net_pnl(E, 체결가, 수량) / (E × Q0) × 100(부분 체결 합), **R = net% / R%**. 매 거래 위험은 1R 고정(수량 무관).

## 7. 판정 사전등록 (실데이터 전 커밋 — 숫자를 바꾸면 새 change)

- **분할**: 표본 세션을 날짜순으로 놓고 **마지막 50 세션 = hold-out(판정)**, 그 앞 전부 = 개발(들여다봐도 됨). 표본 100 세션이면 50/50.
- 지표는 모두 **hold-out · 주 모드 · 비용 뒤 R**. 거래 순서 = (진입 시각, 종목 코드).
  1. 거래당 평균 net R ≥ **+0.10R**
  2. 거래 수 ≥ **40**
  3. 거래가 있는 세션 중 일 평균 R > 0 인 세션 비율 ≥ **55%**
  4. 최대 연속 손실 거래(R < 0) ≤ **6** **이고** 누적 R 곡선의 최대 낙폭 ≤ **8R**(고점은 0 에서 시작)
  5. 개발 절반의 거래당 평균 R 과 hold-out 평균의 **부호가 같음**(둘 다 > 0 또는 둘 다 < 0; 0 이나 개발 거래 0 건은 같지 않음)
- 표본 날짜는 중복을 지운 뒤 셈. **INSUFFICIENT**(먼저 봄): hold-out 거래 < 40, 또는 hold-out 결손 종목-세션 > 20%(hold-out 종목군 항목 수 대비), 또는 **표본 세션 ≤ 50**
  (개발 절반이 비어 (5)를 잴 수 없음 — 위임 원문에 없던 조건, O2 Manager 확인). **PASS** = INSUFFICIENT 가 아니고 1~5 모두. 그 밖 **FAIL**.
- **100점 루브릭 E 축(40점)**: (1) 16 · (2)+(3) 10 · (4) 8 · (5) 6 — 조건별 통과 시 전부, 실패 시 0(부분 점수 없음). INSUFFICIENT 는 E 축 0 으로 보고.
  (2)+(3) 은 둘 다 통과해야 10.
- 민감도 모드 · 개발 절반의 같은 지표는 「판정 아님」으로 나란히 보고.
- **FAIL 을 같은 세션에서 숫자를 고쳐 「수리」하지 않음.** 새 숫자 = 새 change · 새 표본.
- PASS 는 「실거래 전환 Story 를 제안할 근거」일 뿐 — 엔진 배선 · 서명 · 토글은 별도 Story 와 사람 승인.

## 8. 설정 파일 (`sim_config.py`, `config.default.json`)

JSON 하나. 알 수 없는 키 · 타입 불일치(bool 을 숫자로 쓰기 포함)는 거절. 범위 밖이면 필드 이름이 든 오류로 **종료 코드 2**.
이름 · 기본값은 StockOS `runtime_config.py:370-404`, 범위는 `routes/settings.py:772-826` 과 `runtime_config.py:865-1104`(교차 규칙), `exit/stop_controls.py:9-14 · 34-67`.

| 이름 | 기본 | 허용 | 쓰임 |
|---|---|---|---|
| `exit_hard_stop_loss_pct` | −1.2 | [−5, −0.1] | 하드 손절 |
| `exit_weak_trend_stop_loss_pct` | −0.8 | [−5, −0.1], ≥ hard | weak 분기에서만(neutral 이면 무효) |
| `exit_trailing_stop_gap_pct` | 4.5 | [0.2, 10] | 트레일 익절 갭 |
| `exit_trail_stop_pct` | 6.5 | [0.5, 15] | floor ratchet |
| `max_hold_minutes` | 90 | 정수 [1, 390] | 라이브가 시간 청산을 끄므로 무효(검증만) |
| `krx_first_take_profit_pct` | 1.2 | [0.1, 20] | TP1 |
| `krx_second_take_profit_pct` | 2.4 | [0.1, 20], ≥ first | TP2(소유 포지션은 10000 으로 꺼짐) |
| `krx_partial_exit_ratio` | 0.5 | [0.05, 1] | TP1 비율 |
| `rapid_profit_lock_base_capture_ratio` | 0.60 | [0.60, 0.95] | base ≤ strong ≤ tight |
| `rapid_profit_lock_strong_capture_ratio` | 0.65 | [0.65, 0.95] | 〃 |
| `rapid_profit_lock_tight_capture_ratio` | 0.70 | [0.70, 0.95] | 〃 |
| `exit_hard_stop_grace_minutes` | 1 | 정수 [0, 30] | grace |
| `exit_hard_stop_grace_catastrophe_pct` | −1.5 | [−5, −0.5], ≤ hard | grace 중 즉시 손절선 |
| `exit_breakeven_atr_multiplier` | 1.0 | [0, 3] | ATR breakeven |
| `exit_stop_confirm_seconds` | 0 | [0, 30] | 기술적 · floor 손절 확인 |
| `exit_stop_atr_multiplier` | 0 | [0, 3] | 하드 손절 ATR 확장 |
| `eod_flatten_enabled` | false | bool | CLOSE_MANAGEMENT 전량 청산 |
| `common_exit_policy` | null | null · `COMMON_LADDER_{BALANCED,CONSERVATIVE,DEFENSIVE,TREND,RUNNER}` | null = 레거시 2단(라이브 기본, `common_exit_policy` NULL) |

- rapid lock 하한은 settings 라우트의 0.0 이 아니라 런타임 검증의 안전 하한(runtime_config.py:860-864 · 946-987)을 씀 — 둘 중 엄격한 쪽.
- **digest** = 검증 · 기본값 채움 뒤 설정을 키 정렬 · 공백 없는 JSON 으로 만든 바이트의 sha256. 모든 보고서에 적음.

## 9. 보고서

`replay.py --source <dir> --config <json> --output <새 파일>` → JSON: 도구 판 · 설정 digest · StockOS 커밋 · 표본 지문(세션 목록 sha256) · 입력 지문(읽은 manifest · 봉 파일 sha256 목록의 sha256) ·
두 모드의 퍼널(돌파 · 시한 · 되돌림 실패 · 거시 거절 · +8% 거절 · 14:45 · 진입 봉 없음 · 진입) · 거래 목록 · hold-out / 개발 지표(판정용 R 지표 + 보고용 거래당 평균 net% · EOD 청산 건수) · 판정 · E 축 점수.
첫 `limitations` 항목: 「과거 호가 없음 — 진입은 그 분 시가(P0), 매도는 1틱 아래 bid. 체결 가능성 · 장 마감 뒤 미모델링.」

## 10. 한계

1. 호가 · 잔량 없음(P0 가정) · 1분 해상도 틱 경로(봉 안 순서는 가정).
2. 종목군은 a203 U-A 그대로 — 생존 편향 · 모집단 미래 정보(a203 design §5).
3. 진입 ATR% 는 StockOS 정의를 재구현했고 StockOS 실제 스냅샷과 대조하지 않음.
4. EOD 강제 청산(다음 날 보유 없음) · 사다리 HYBRID_50 미지원 · 라이브 허용 목록 차단 미모델링.
5. 한 시장 국면 표본일 수 있음.
6. 낙관 쪽 가정: 진입은 되돌림 봉 직후 분의 시가에 즉시 체결, 청산은 조건 틱에서 즉시 체결(라이브는 시세 주기 · 주문 대기만큼 늦음).
7. 라이브 기본(eod_flatten 꺼짐)은 다음 날까지 보유하지만 재생은 EOD 청산 — 갭 위험 · 연속 손실이 덜 잡힐 수 있어 EOD 건수를 함께 보고.
8. R% 가 작은 거래(구조 손절이 가까움)가 평균 R 을 크게 움직일 수 있음 — 판정은 사전등록대로 R, 거래당 평균 net% 를 나란히 보고.
9. 개장 · 마감이 늦춰진 특수일(수능일 등)은 09:00 칸 기준 · 15:30 마감 자르기를 그대로 씀.
10. `breakeven_atr_min_gap_pct` 는 StockOS 기본 0.0(hardening_runtime_config.py:27)으로 고정 — 운영 KV/env 값은 읽지 않음.
