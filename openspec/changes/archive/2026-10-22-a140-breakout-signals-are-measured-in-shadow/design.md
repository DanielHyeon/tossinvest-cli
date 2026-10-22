# a140 설계 — 돌파 신호의 그림자 측정

기준 HEAD `050d092f` · codegraph 1.6.0 · 분기 근거 `analysis/ast-evidence/claims.md` · 호출 근거 `analysis/code-context/`.

## 비례 원칙 적용

새 코드는 새 파일(`internal/breakoutsignal/` · `tools/a140-breakout-signal-shadow/`)에만 들어가고, 주문 · 손절 · 사이징 · Guardian ·
원장 · 대사 · 인증 · 체결 경로의 기존 함수를 편집하지 않는다. 그래서 FLM/BTM · 변이 원장 · 다중 적대 리뷰는
`Function Logic Map: not-applicable`(기존 함수 편집 0, 게이트 판정 변경 0 — docs/WORKFLOW.md 「비례 원칙」). 기본 = 시험 + lint +
경량 gstack 리뷰. 단 `internal/breakoutsignal` 은 다음 Story(S4)가 생산 증거로 배선하므로, 그 Story 에서 다시 시험할 필요가 없도록
경계값 · 결정성 · 거절 시험을 여기서 끝낸다.

## 1. 지금 무엇이 있고 무엇이 없는가 (측정)

| 가족 | 순수 평가기 | 필요한 증거 | 증거 생산자 | 공식 데이터로 가능? | 생산 입력 |
|---|---|---|---|---|---|
| continuation KR | 있음 | `kr_net_flow`(순매수 금액 / 거래대금) | 0 | **아니오** — 종목별 투자자 매매 없음(a128 1.7 영수증) | 서명 매니페스트 + 재생 |
| continuation US | 있음 | `us_participation`(참여 거래량 / 기준 거래량 · 가격 변화) | 0 | 1분봉으로 가능하나 「기준 거래량」 정의 없음 | 같음 |
| reversal KR | 있음 | `kr_net_flow`(kr-absorption-v1) + 구조 사건 | 0 | **아니오**(같은 이유) | 같음 |
| reversal US | 있음 | `us_participation`(us-dislocation-v1) + 구조 사건 | 0 | 정의 없음 | 같음 |
| weekly KR · US | 있음 | `disclosure` + 주간 예약 | 0 | 공시 어댑터 정책 먼저 | 같음 + 원장 예약 |
| breakout KR · US | 있음(`breakoutlane.Evaluate`) | 닫힌 1분봉 + L1 호가 + **파생 지표 4** | 봉 · 호가 생산자 **있음**(생산 호출 0) · 파생 지표 **0** | **예** | **무조건 거절**(buildLaneInput B1@439) |

공통: 생산 `evidence.db` 는 읽기 전용(B5@329), 쓰기 열기 생산 호출 0. 생산 신호의 신뢰 앵커는 환경 핀 서명 파일(B8@370 · B3@321)이라
주기 신호를 실을 수 없다. → breakout 이 「데이터는 있고 계산만 없는」 유일한 가족이다. 이것이 D1 권고의 근거다.

## 2. 이 change 의 모양

```text
[공식 랭킹 1회/세션] ─→ 종목 집합 N (D8, 세션 동안 고정)
        │
        ▼ 1분마다 종목별
officialbars.PollClosedBars ──→ 도구 전용 evidence DB (strategyevidence.Open — 생산 경로 거절)
        │
        ▼ SealBarSeries (정규장 · 닫힌 봉 · ingestion cutoff = 지금)
internal/breakoutsignal  ──→ 봉별 RVOL · 윗꼬리 ppm · 거래량 확장 + 세션 ATR   (D3~D6)
        │
        ▼ breakoutlane.NewEvidenceSnapshot + Evaluate(prior = nil)   (동결 v1 상수)
   단계 = RANGE_LOCKED / RETEST_WAIT / ARMED / INVALIDATED / TIMED_OUT / FIRST_TOUCH …
        │ ARMED 인 setup 만, 처음 본 분부터 최대 5 분(§2.1 의 1)
        ▼ officialbars.PollQuoteL1 (GET 2) → 같은 봉 + 새 호가로 다시 평가
   PROPOSED(가상 수량) 또는 QUOTE_STALE / SPREAD_TOO_WIDE / ENTRY_DRIFT_EXCEEDED / ZERO_QUANTITY …
        │
        ▼ append-only JSONL 원장 (judgment 아닌 signal 행 · gap 행)
라벨러: 같은 DB 의 뒤 봉으로 손절 / 목표 / 마감 결과 (추가 GET 0)  →  보고서
```

- **측정 방법의 근거**: ARMED 전 단계는 호가 신선도를 보지 않고 구조적으로 유효한 봉인만 요구한다(claims 6 —
  validStructuralEvidenceInput @162 · evaluateFresh B18@111 · B19@114). 그래서 세션 시작에 종목마다 호가 1 회를 봉인해 두고
  ARMED 동안만 새로 읽는다. 매 분 호가를 읽는 대안은 GET 이 3 배(봉 1 + 호가 2)인데 ARMED 전 판정은 같다.
- **setup 은 종목 · 세션당 하나**(`setupID` = 시장 · 종목 · 세션 · 달력 · 첫 봉 · 15 번째 봉 · 레인 · 설정). PROPOSED · `armed_expired`(§2.1 의 1) · INVALIDATED ·
  TIMED_OUT 을 적은 뒤 그 종목은 그 세션에 다시 평가하지 않는다(중복 제안 금지).
- `prior = nil` 로 매 분 전체 봉에서 다시 유도한다 — ARMED 전 단계는 재시작해도 같은 결과다. ARMED 뒤(호가 의존)는 원장의 setupID 로 중복을 막는다(§2.1 의 6).
- 도구는 `internal/official` 을 import 해 엔진과 토큰 캐시를 공유한다(a128 1.2 (c) 선례 — 토큰 전쟁 방지). 브로커 표면은 읽기 GET 만 담는
  좁은 인터페이스로 다루고, 주문 · 계좌 · 보유 이름의 import · 호출을 AST 시험이 금지한다(a128 static_test 선례).
- 도구 전용 evidence DB 경로는 플래그로 받고, 엔진의 생산 `evidence.db` 와 같을 수 있으면 시작을 거부한다(§2.1 의 9).

## 2.1 측정 규칙 (proposal-freeze 리뷰 반영, review §2)

1. **ARMED 뒤 호가 재시도(P1-1)**: 평가기는 ARMED 에서 루프를 끝내므로 setup 은 그 세션 내내 ARMED 로 남는다(machine.go:96~99). 엔진은 주기마다
   호가를 새로 읽을 것이므로 도구도 **ARMED 를 처음 본 분부터 최대 5 분, 분마다** 새 호가로 다시 판정한다. 처음 PROPOSED 가 나오면 그 setup 은 끝,
   5 번 모두 거절(QUOTE_STALE · SPREAD_TOO_WIDE · ENTRY_DRIFT_EXCEEDED · ZERO_QUANTITY …)이면 마지막 거절로 끝(`armed_expired`). 시도마다 호가 나이 ·
   스프레드 · 괴리 실측값을 원장에 적는다(D21 한계를 바꿔 재계산 가능). 5 분은 생산 증거 신선 창 상한(continuationlane `productionEvidence` 5 분)과 같은 값.
2. **종목 · 세션마다 고정하는 설정(P1-2)**: `TickMinor` = opening range 저항(첫 15 봉 최고가)의 KRX 호가 단위 — RANGE_LOCKED 에서 정해 세션 동안 고정
   (장중에 다시 계산하면 config digest → setupID 가 갈림). `SizingInput.FinalCap` = 최대(2⁶⁴−1) — 그림자에는 q_final 상한이 없으므로 수량 = 위험 ·
   명목 상한이 정한 후보 수량(0 이면 ZERO_QUANTITY 그대로). FX 봉인(KRW/KRW 1/1) `AsOfMS` = 세션 개장, `FreshUntilMS` = 세션 마감.
3. **빈 분(P1-3)**: `breakoutsignal.Derive` 는 엄격하게 빈 분을 거절한다(spec). 도구는 **개장부터 이어진 앞부분만** 평가한다 — 다음 폴이 늦게 들어온 봉을
   채우면 앞부분이 다시 길어진다. 앞부분이 끊긴 채 3 분 넘게 안 채워지면 gap 행(stage=`bars`, `MINUTE_GAP`, 빠진 분)을 한 번 적고, 세션 끝까지 안 채워지면
   그 종목의 퍼널 끝을 `GAP` 으로 센다. VI(변동성 완화장치) 단일가 구간이 봉을 비우는지는 실측된 적이 없다 — 이 계수가 그 첫 측정이다.
4. **폴 하한(P2)**: `LowerBound` = 이미 저장된 가장 새 봉의 여는 시각 − 5 분(없으면 정규장 개장). 그래서 평소 폴은 1 페이지(종목당 1 GET)이고, 놓친 분은
   5 분 안이면 다음 폴이 소급 적재한다.
5. **ARMED 봉의 위치(P2)**: `Decision` 이 ARMED 봉 번호를 내보내지 않는다(types.go:197~210). 도구는 앞부분 길이 n 을 늘려 가며 `Evaluate` 가 처음
   ARMED 를 내는 n 을 찾고, 그 n 번째 봉(되찾은 봉)의 종가를 진입 제안가로 쓴다(평가기는 순수 함수라 추가 GET 0).
6. **재시작(P2)**: PROPOSED 는 그 순간의 호가에 의존하므로 재시작 뒤 같은 결과가 보장되지 않는다. 시작할 때 원장에서 오늘 세션의 setupID 를 읽어
   끝난 setup(PROPOSED · armed_expired · INVALIDATED · TIMED_OUT)은 다시 평가하지 않는다.
7. **봉 ID(P2)**: setupID 는 첫 봉과 15 번째 봉의 ID 를 쓴다(machine.go:206). 판(revision)마다 바뀌는 `EvidenceID` 를 쓰면 봉 정정 하나로 setup 이 둘이
   된다. 그래서 lane 봉 ID = `세션 ID:여는 시각(ms)`(판 무관), 판은 `Revision` 칸으로만 넘긴다.
8. **왕복 비용 · 목표가(P2)**: 비용 요율(ppm) = `ceil((KR 매수 수수료 + 매도측 요율) × 1e6)`(`costs.DefaultModel`, 현재 5,000 ppm), 주당 왕복 비용 =
   `ceil(worst × 요율 / 1e6)`, worst = max(진입 제안가, ask). 목표 = `worst + 왕복 + ceil(위험 × 1.5)` — 평가기의 RR 검사(sizing.go:70~74)와 같은
   올림을 쓰므로 그 검사는 **구성상 통과**한다(보고서에 「RR 거절은 공허」로 밝힘).
9. **생산 DB 가드(P2)**: 거절 조건 셋 — ① 파일 이름이 `evidence.db`, ② 같은 디렉터리에 `journal.db` 가 있음, ③ 생산 후보(`journal.DefaultPath()` 의
   디렉터리 · 설정 디렉터리의 `evidence.db`)와 `os.SameFile`(하드링크 · 심링크 모두). 비교는 심링크를 푼 경로 · inode 둘 다.
10. **키 공유(P2)**: 호가 GET 은 엔진의 `/prices` · `/orderbook` 과 같은 MARKET_DATA 그룹이다. 그 그룹에서 rate limit 을 한 번이라도 받으면 **호가 읽기를
    5 분 멈추고**(그 사이 ARMED 시도는 gap 행), 봉 그룹(MARKET_DATA_CHART)에서 재시도를 다 쓰면 그 분의 나머지 봉 폴을 건너뛴다. 이 도구가 엔진의
    보호 경로 가격 읽기와 경합하지 않게 하려는 것이다(안전 불변식 4 인접).
11. **쓰기 목록(P2)**: 도구가 쓰는 것은 ① 도구 전용 evidence DB, ② 원장, ③ 공유 Open API 토큰 캐시 갱신(`official.New` — a128 선례, 엔진과의 토큰 전쟁
    방지) 셋뿐이다. 설정은 `tossconfig.NewService(...).Status` 로 읽기만 하고 저장 함수 이름은 AST 시험이 금지한다.
12. **마감가(P3)**: 생산자는 다음 봉이 보여야 그 봉을 적재하므로(producer.go:355~357) 정규장 마지막 봉은 장 뒤 봉이 successor 가 될 때만 들어온다.
    라벨러는 마감 뒤 30 분까지 기다린 다음 **적재된 마지막 정규장 봉**의 종가를 「마감가」로 쓰고 그 봉의 여는 시각을 라벨 행에 적는다. ARMED 감지가
    봉 종가보다 1~2 분 늦는 것도 같은 이유다(엔진도 같은 생산자를 쓰면 같은 지연).
14. **코드 리뷰(BLOCK) 수리 — 평가 · 라벨 시각(2026-10-15, review §3)**: ① ARMED 시도의 평가 시각 = 호가를 **읽은 뒤**의 지금(그 앞 시각이면
    received_at > evaluated 로 언제나 QUOTE_STALE). ② 라벨은 제안 시각이 든 분 **다음에 연** 봉부터(`label_from_ms` 를 행에 적음) — ARMED 봉은
    다음 봉이 보여야 적재되므로 제안 시점에 이미 닫힌 봉이 있다. ③ 봉 읽기 컷오프 = 폴 **뒤** 시각(방금 적재한 봉이 보이게). ④ ARMED 창은 시간(처음 본
    순간부터 5 분) — 호가가 실패 · 정지해도 창이 지나면 armed_expired. ⑤ 결손 행 = 종목 · 단계 · 분마다 한 행, 계수는 budget 행에도. ⑥ 봉 그룹
    재시도 포기 → 그 분의 나머지 종목 봉 폴 건너뜀(결손 행). ⑦ 랭킹 포기도 결손 행 뒤 시작 거부. ⑧ 라벨은 이어진 앞부분이 아니라 **전체 묶음**.
    ⑨ 마감 뒤(라벨 유예 30 분)에는 라벨 대기만 폴하고 새 setup 을 평가하지 않음. ⑩ 같은 세션에 다른 판(lot) 행이 있으면 섞지 않고 시작 거부.
    ⑪ 보고서는 세션별(보류가 기본)과 **판 합계**(같은 정의의 세션을 합쳐 n ≥ 30 이면 판정)를 함께 냄.
    리더 계약: 엄격 리더는 「모든 봉 ≤ before」라 자라는 봉을 주지 않고, 생산자는 받은 것 중 가장 새 봉도 적재하지 않으므로 봉 k 는 k+2 분에 적재된다
    (시험 가짜 브로커가 이 계약을 따름 — 실측 대조는 첫 장중 실행의 원장 시각으로).
13. **비용(P3)**: `SealBarSeries` 는 정규장 전체 봉 읽기에 종목당 약 47 ms(breakout_series.go:50~56) — 30 종목이면 분당 약 1.4 초(폴의 기존 읽기 포함 두 배).

## 3. 파생 지표 (`internal/breakoutsignal`)

정수 산술만(float 금지 — spec 「float … 은 snapshot seal 전에 거부」). 판본 문자열 `breakoutsignal-v1` 을 원장 행 · 지표 digest 에 박는다.
값을 만들 수 없으면(봉 부족 · 분모 0 · 해석 불가) **typed 거절**이고 0 을 넣지 않는다(spec required evidence fail-closed). 아래 정의의
숫자는 D3~D6 의 **권고안**이며 사람이 바꾸면 spec 델타의 해당 문장을 같이 고친다.

| 지표 | 권고 정의(D) | 근거 · 대안 |
|---|---|---|
| `UpperWickRangePPM` | `ceil((high − max(open, close)) × 1e6 / (high − low))`, high == low → 0 (D6) | 기계가 상한 35% 로 거부권에 씀 → 올림 = 보수 |
| `ATRMinor` | opening range 15 봉의 Wilder ATR(14) — 유리수로 계산 후 minor 내림, **세션 동안 고정** (D3 · 구현 정밀화 아래) | 1분봉 위의 기계라 단위를 맞춤. 대안: 일봉 ATR(일 단위 변동을 1분 버퍼에 씀 — 버퍼가 과대) |
| `RVOLPPM` | opening range 뒤 봉 i: `floor(vol_i × 1e6 × 15 / 직전 15 봉 거래량 합)`. opening range 15 봉은 정의 안 됨(요청 시 거절; lane 입력 칸은 0 — 평가기가 읽지 않음, evaluateFresh B4@52 `bars[15:]`) (D4) | 오프닝 레인지(15 봉)가 끝나야 돌파가 가능하므로 첫 평가 봉부터 값이 선다. 편향: 장 초반 거래량이 커서 RVOL 이 낮게 나온다 — 기계가 1.2×/2.0×/2.5× 반사실을 함께 적으므로 보고서가 그 영향을 보인다. 대안: 지난 5 거래일 같은 분 평균(과거 1분봉 깊이 미실측) |
| `VolumeExpanded` | `RVOLPPM ≥ 1,000,000` (D5) | 되돌림 중 「거래량 늘며 저항 아래 마감 = 실패」(B15@101). 대안: 직전 봉보다 큼 |

**구현 정밀화(2026-10-15, 승인된 D3 · D4 안에서 — Manager 확인 대상)**: ① ATR 의 창을 「opening range 15 봉」으로 고정했다. 승인 문구
「세션 정규장 1분봉 Wilder ATR(14)」를 매 분 전체 봉으로 계산하면 ATR 이 분마다 바뀌고, 평가기가 전체 봉에서 다시 유도하므로(prior = nil)
이미 지난 돌파 · 되돌림 판정이 소급해 뒤집힌다 — spec 「config/version/digest 변경은 진행 중 setup을 소급 재해석해서는 안 된다」와 같은 위험.
setup 은 RANGE_LOCKED(15 번째 봉)에서 정의되므로 그 순간의 ATR 이 자연스러운 고정점이다. ② 입력 봉은 개장 분부터 빈 분 없이 이어져야 한다 —
평가기는 봉 번호로 opening range · timeout 을 세므로 빈 분이 있으면 「15 봉」이 15 분이 아니다(producer 의 `MinuteGap` 보고와 짝).

## 4. 가상 사이징 · 라벨 (도구 안에서만)

breakoutlane 은 사이징 입력 없이는 PROPOSED 를 내지 않는다(B19 뒤 `size`). 그림자 측정이므로 아래 값은 **가상**이며 어떤 생산 값도 아니다.

- 손절 = 오프닝 레인지 저점(기계의 무효화 수준과 같은 값 — claims 7) (D7).
- 진입 제안가 = ARMED 봉의 종가, 괴리 검사는 기계가 ask 로 한다.
- 목표 = 비용 포함 순 RR 1.5 가 되는 가격(KRX — StockOS 출시 선례 1.5, 기억 「StockOS 순 RR 선례」) → `MinRiskRewardPPM = 1,500,000` (D7).
- 위험 예산 · 명목 상한 = 설정 파일에서 **읽기만**: 위험 예산 = `floor(engine.automation_gate.max_daily_loss_amount / 3)`(a139 사용자 결정과 같은 식 — a139 착지 뒤 `risk.PerTradeRiskBudget` 로 바꿀 수 있음), 명목 상한 = `max_order_notional`. `limit_currency` 가 KRW 가 아니거나 값이 양의 유한수가 아니면 시작 거부(D7). 플래그로 덮어쓸 수 있고 쓴 값의 출처를 원장 행에 적는다. 슬리피지 = 0(호가 ask 를 진입 기준으로 이미 씀).
- 왕복 비용 = `internal/costs.DefaultModel()`(읽기 전용).
- 라벨(D10): PROPOSED 뒤의 닫힌 봉을 순서대로 보며 low ≤ 손절 이면 `stop`, high ≥ 목표 이면 `target`(한 봉에서 둘 다면 `stop` — 보수),
  마감까지 둘 다 없으면 `close` 와 마감가 기준 순 R 배수. 같은 DB 를 쓰므로 추가 GET 0. 체결 가능성(호가 잔량)은 재지 않는다 — 보고서에 한계로 적는다.

## 5. 원장 · 보고서

- 위치: `openspec/changes/a140-…/analysis/shadow-ledger/ledger.jsonl`(장중 산출물 — 측정 종료 후 커밋; 계좌 데이터 없음). evidence DB 는 커밋하지
  않는다(크기) — 보고서와 원장이 그 digest 를 적는다.
- 행 3 종(a128 선례): `signal`(시각 · 종목 · setup · 단계 · 거절 코드 · 전이 · 지표 판본 · 스냅숏 digest · RVOL 반사실 · 가상 진입/손절/목표/수량) ·
  `label`(signal 참조 · 결과 · R) · `gap`(봉 · 호가 · 랭킹 읽기 포기, 세션 · 달력 결손 — 침묵 생략 금지).
- 행마다 `lot = breakoutsignal-v1@<설정 digest>` — 정의가 바뀌면 모집단이 갈린다(측정은 잰 순간을 달고 다닌다).
- 보고서: 세션별 퍼널(종목 N → RANGE_LOCKED → BREAKOUT_CLOSED → RETEST_WAIT → ARMED → PROPOSED) · 거절 분포 · RVOL 1.2/2.0/2.5 반사실 ·
  제안 수 · 결과 분포(n<30 이면 승률 · 평균 R 판정 보류) · 사이클 GET 수(그룹별) · 수집 시간.

## 6. 호출 예산 (D11)

- 봉: 종목 N 개 × 분당 1 GET(MARKET_DATA_CHART 그룹, a128 주석상 초당 5) → N = 30 이면 0.5 GET/s(그룹 한도의 10%). 폴 하한 규칙(§2.1 의 4) 없이는 200 분 뒤 종목당 2 GET(약 1 GET/s) — 하한이 이 예산의 전제다.
- 호가: 세션 시작 N × 2, ARMED 시도마다 2(setup 당 최대 5 시도 = 10). 랭킹: 세션 시작 1.
- 모든 GET 앞 `ReadGap` 250ms(a128 선례). 429 · limited 는 지수 백오프 후 그 분만 결손(gap 행).
- 엔진과 같은 키 — 엔진의 안전 루프(체결 · 대사 · 보호)는 다른 프로세스의 low/safety 예산이라 이 도구가 그 예산을 **소비**하지는 않지만,
  브로커 쪽 한도는 키 단위로 공유된다. 그래서 분당 GET 을 원장 · 로그에 남기고, 수집 시간이 간격 절반을 넘으면 WARNING.

## 7. 안전 경계

- 브로커 호출은 읽기 GET 뿐. 주문 생성 · 정정 · 취소 · 토글 · 활성화 쓰기 0(AST 시험이 import · 호출 이름을 금지).
- 생산 `evidence.db` · 저널 · 설정 파일에 쓰지 않는다. 설정은 읽기만.
- Jev · 외부 판단 API 호출 0(이 change 의 비목표). 계좌 데이터를 외부로 보내는 길이 없다.
- 손절 · 비상 청산 경로와 공유하는 코드 0 — 이 도구는 엔진 프로세스 밖이다.

## 8. P2 Story 분할 (Feature ↔ Story 다수, Story ↔ change 1:1)

| Story | 내용 | Feature | 위험 | 선행 | 만지는 파일 | 병렬 |
|---|---|---|---|---|---|---|
| **S1 = a140** (이 change) | KR breakout 신호를 그림자에서 잰다 | FEAT-TOS-013 | Low | 없음 | 새 파일만 | a138 · a139 와 병렬 OK |
| S2 | Jev 판단 코어를 `internal/` 로 **옮긴다** — a128 의 `PublicState` · 3 겹 가드(타입 · 바이트 · 시험) · 판단 클라이언트를 새 패키지로 이동, a128 도구가 그것을 import(포크 금지) | FEAT-TOS-006 | Low | **a128 측정(3.1·3.2) 종료** — 측정 중 계기를 바꾸지 않는다 | `tools/a128-*` · 새 `internal/jevjudge/` | S1 과 병렬 OK(겹치는 파일 0) |
| S3 | **Jev 진입 seam(DORMANT · 기본 OFF, 켜면 먼저 RECORD-ONLY)** — 판단 · 원장 기록만 하고 막지 않음. 거부 임계는 a128 보고서(시장마다 ≥5 세션) 뒤 사람이 KR · US 따로 정한 다음에만(D16). 아래 「S3 개요」 | FEAT-TOS-006 | **High**(주문 경로) | S2 · a138/a139 착지(설정 파일 공유) | `strategy_market_handoff_delivery.go` + 새 파일 + 설정 1 필드 | S4 와 병렬 OK |
| S4 | 엔진 안 증거 생산 — 봉 · 호가 생산자를 생산 `evidence.db` 에 배선(쓰기 열기 1 곳 · low 예산 · 안전 예산 무접촉) + `breakoutsignal` 로 breakout 증거 기록. 주문 경로 무접촉 | FEAT-TOS-013 | Medium(API 예산 공유) | S1 결과(신호가 실제로 나는가) | 엔진 조립 · 새 파일 | S3 과 병렬(파일 분리 확인 필요) |
| S5 | 살아 있는 제안 권위 — breakout 은 `buildLaneInput` B1 을 걷어 내고, 제안 입력을 「사람 서명 + 환경 핀 파일」에서 「사람이 서명한 활성화 아래 엔진이 봉인한 제안」으로(신뢰 앵커 변경, D13) | FEAT-TOS-013 | **High** | S4 · D13 | strategyproposal · collectMarket | 단독 |
| S6 | 서명 매니페스트 운영 — KR breakout 만, 먼저 SHADOW 매니페스트(digest 핀, a112 7.3)로 반사실 관측 → 그다음 활성화 · 경로 · 계좌 · 위험 매니페스트 | FEAT-TOS-013 | 사람 운영(코드 0~소) | S5 | 운영 파일 · `tools/a112-family-*` | — |
| (사람) | KR breakout first-leg ON 토글 · Jev ON 토글 | — | — | S6 · a128 결론 · 사람 승인 | — | — |

권고 순서: **S1 ∥ S2 → S3 · S4 → S5 → S6 → 사람 토글.** S3 는 S1 과 무관하게 착지할 수 있다(DORMANT) — 착지 뒤에도 S5 전에는 그 자리를
지나는 제안이 0 이다. a138 · a139 가 편집 중인 파일(`risk.Policy` · `guardian_wiring.go` · `interlock.go` · config · 콘솔 `settings_limits`)은
S1 · S2 가 만지지 않는다. S3 의 설정 필드는 config 를 공유하므로 a139 아카이브 뒤에 착수한다.

### S3 개요 — Jev seam 자리와 실패 동작 (이 change 의 구현 대상 아님 · 결정만)

- **자리**: `dispatchStrategyMarketHandoffs` 의 전달 몸통, 캠페인 CAS 확인(B2@41) **뒤** · `dispatcher.dispatch`(44 줄) **앞**.
  - 조정자가 소유자 범위마다 승자를 고른 **뒤**라 Jev 거부가 다른 종목을 대신 풀지 않는다(collectMarket B10@412 가 막는 「목록이 짧아져 다른
    종목이 풀림」과 같은 위험이 구조적으로 없음).
  - Guardian 첫 다리 승인(dispatch 151 줄) · 원장 쓰기 · 브로커 호출 **전**이라 되돌릴 것이 없다.
  - 이 자리를 지나는 의도는 `Side: "buy"` 리터럴뿐(strategyFirstLegPlaceIntent 268 줄). 매도 · 손절 · 비상 청산 · flatten 은 exitloop ·
    protection 경로라 이 함수를 지나지 않는다 — 금지가 규율이 아니라 구조다.
  - 대안(조정 앞에서 J1 로 순위): 조정 점수는 승인된 score 판본 · 보정 digest 로만 비교된다(four-family-strategy-runtime 「calibrated score」)
    — Jev p 를 점수에 섞으려면 별도 보정 change 가 필요하다. v1 에서 기각.
- **켜는 순서(D16)**: OFF → RECORD-ONLY(판단 · 원장만, 진입에 영향 0 — Jev 실패도 진입을 막지 않고 결손만 기록) → 사람이 시장별 임계를 정한 뒤 VETO.
  RECORD-ONLY 의 판단은 dispatch 를 **기다리게 하지 않는다**(S3 설계 과제: 같은 자리에서 판단 요청만 떼어 보내고 dispatch 는 그대로 진행 —
  지연 0 · 진입 결과 불변을 시험으로 고정). 아래 실패 동작은 VETO 모드의 것이다.
- **실패 동작**: Jev 시간 초과 · 오류 · 키 없음 · 예산 초과 = **그 범위는 이번 주기 진입 안 함**. 범위 거절 타입(`strategyScopeRefusal`)으로
  돌리면 refreshOnly 에서는 삼킨 오류 계수, effective 에서는 시장 잠금이 되므로(claims 5 — runMarket B12@956 · 984 줄), a137 의 일일 손실
  차단처럼 **오류가 아닌 타입 결과**로 끝내고 계수 · 알림을 남긴다. 레인을 잠그지 않는다. 다음 주기에 제안이 살아 있으면 다시 묻는다.
- **계좌 데이터**: S2 가 옮긴 `PublicState` 와 3 겹 가드를 그대로 쓴다(포크 금지). 제안의 수량 · 계좌 · 보유는 state 에 넣지 않는다.

## 9. 사람 결정

| # | 결정 | 권고 | 근거 |
|---|---|---|---|
| **D1** | P2 첫 가족 · 시장 | **breakout-retest · KR** | 공식 데이터만으로 증거가 서는 유일한 가족(§1). KR: 계좌 기준 통화라 FX 봉인이 1/1, 장이 KST 주간이라 감독 쉬움, a128 KR 자료와 같은 시장. KR continuation · reversal 은 종목별 투자자 매매 부재로 막힘 |
| **D2** | 첫 change 의 형태 | **도구 그림자 측정(a140)**, 엔진 SHADOW 배선은 S4 | 엔진 안 배선은 생산 `evidence.db` 쓰기 · API 예산 공유 · 엔진 조립 편집을 부른다. 신호가 하루 몇 건인지 모르는 채 그 비용을 치르지 않는다 |
| **D3** | ATR 정의 | **세션 정규장 1분봉 Wilder ATR(14)**(승인) — 구현에서 창을 **opening range 15 봉으로 고정**하도록 정밀화(§3 · review §1, **사람 확인 대기** — tasks 0.4) | §3. 기계의 버퍼 · 허용폭이 1분봉 단위. 매 분 전체 봉 ATR 은 이미 판정된 setup 을 소급 재해석 |
| **D4** | RVOL 기준 | **같은 세션 직전 15 봉 평균** | §3. 과거 1분봉 깊이 미실측 · 장 초반 편향은 반사실로 보고 |
| **D5** | 거래량 확장 | **RVOL ≥ 1.0×** | §3 |
| **D6** | 윗꼬리 비율 반올림 | **올림**(보수) | 거부권 쪽으로 틀림 |
| **D7** | 가상 사이징 | 손절 = 레인지 저점 · 목표 = 순 RR 1.5 · 위험 예산 · 주문 상한 = 설정 읽기(a139 값) | §4 · claims 7. 생산 값은 S5 에서 따로 정한다 |
| **D8** | 종목 집합 | **세션 시작 공식 랭킹(거래대금) 상위 30, 1 주 가격 ≤ 주문 상한, 세션 동안 고정** | breakout 은 개장부터 봉이 필요 — a128 의 「매 주기 무작위 30」은 오프닝 레인지를 못 가짐. 랭킹은 후보 발굴 용도로만(spec 「Toss rank … candidate discovery evidence로만」) |
| **D9** | 측정 기간 | **KR 5 세션** | setup 은 종목 · 세션당 하나라 세션당 제안 최대 30, 실제는 훨씬 적을 것. 2 세션(a128)으로는 결과 분포가 거의 0 |
| **D10** | 라벨 | 손절/목표 선착(한 봉에 둘 다면 손절) · 아니면 마감가 순 R | §4. 추가 GET 0 |
| **D11** | 호출 예산 · 동시 실행 | 봉 분당 N GET(N=30 → 0.5/s) · 호가는 시작 + ARMED 만 · a128 KR 프로브와 **같은 날 허용**하되 그룹별 계수 기록 | §6. MARKET_DATA_CHART 그룹 합계(a128 ≈ 32 GET/10 분 ≈ 0.05/s + a140 0.5/s)가 그 그룹 초당 5 의 11% |
| **D12** | a140 에서 Jev 를 부를까 | **부르지 않음** — 제안 행에 스냅숏 digest 만 남김 | Jev 개입은 S2 · S3 에서 엔진과 같은 자리 · 같은 상태로 잰다. 여기서 부르면 a128 과 다른 모집단의 세 번째 로트가 생긴다 |
| **D21** | (a140, 승인 뒤 발견) breakout v1 설정 중 범위만 정해진 값 · 양수만 요구되는 값 | 그림자 기본값: `RetestTolerancePPM = 250,000`(허용 범위 상단 — 신호 표본을 넓게) · `MaxQuoteAgeMS = 5,000` · `MaxSpreadPPM = 3,000`(KRX 1 틱 스프레드 대부분 통과) · `MaxEntryDriftPPM = 5,000`. 셋(호가 나이 · 스프레드 · 괴리)은 실측값을 원장에 함께 적어 **다른 한계로 재계산 가능**, 되돌림 허용폭은 재계산 불가라 판본(`lot`)에 박음. 플래그로 바꿀 수 있음 | spec 「v1 config … retest tolerance 100,000..250,000 · positive finite max_quote_age_ms · max_spread_ppm · max_entry_drift_ppm」 — 생산 값은 어디에도 없음(시험 픽스처뿐). 생산 값은 S5 에서 |
| **D13** | (S5) 살아 있는 제안의 신뢰 앵커 | **(a) 사람은 활성화(레인 · 기간 · 한도)에 서명하고, 제안은 엔진이 봉인** | (b) 별도 서명 데몬이 5 분마다 재서명 — 환경 digest 핀을 키 핀으로 바꿔야 하고 키가 엔진 옆에 있어 앵커 이득이 없음. (c) 사람이 매 제안 서명 — 신선 창 5 분 + 환경 핀이라 불가능(R3) |
| **D14** | (S3) Jev 질문 | **J2(a128 `j2_hold_off_red_flag`, question_rev a128-r1) 거부권만**. J1 은 a128 보고서가 예측력을 보이기 전 배선 금지 | a128 design 「후속 2」 · spec 「측정 영수증 없이 배선 금지」. 질문 · state 판본을 a128 과 같게 두어야 그 보정이 이전된다 |
| **D15** | (S3) Jev 결과와 전략 점수의 결합 | **AND 거부만** — Jev 는 점수를 올리거나 순위를 바꾸거나 대체 종목을 고르지 않는다 | 조정 점수는 보정 digest 로만 비교(§8 S3 개요). 순위 결합은 그 보정 change 뒤 |
| **D16** | (S3) 거부 임계 | **확정(사용자 2026-10-15, 권고에서 변경): 기본 거부 임계 없음.** S3 seam 을 처음 켜면 **RECORD-ONLY** — Jev 가 판단하고 원장에 적되 막지 않는다. 거부 임계는 a128 보고서(시장마다 5 세션 이상) 뒤 사람이 **KR · US 따로** 정한다 | a128-s2 원장 실측(시장마다 J2 990 건): p ≥ 0.5 면 KR 16.0% · US 20.2% 를 거부하고, KR 에서는 거부군이 더 나았다(60 분 +0.14% 대 −0.39%). 시장마다 하루 · 상관된 표본이라 결론은 아니지만 고정 0.5 의 근거가 없다는 증거 |
| **D17** | (S3) Jev 에게 주는 정보 | **확정(사용자 2026-10-15): v1 state = a128 보고서가 보정한 state 판본**(a128 은 1.8 에서 `a128-s3` — WTS 뉴스 · KR 종목별 투자자 매매 · US 지수 — 로 옮김. 그래서 s2 가 아니라 s3). 전략 맥락(가족 · 진입 · 손절 · 목표)은 **v2** 에서 별도 그림자 측정 뒤 | 사용자 지시 「충분한 정보」 — state 를 바꾸면 a128 보정이 이전되지 않는다(state_rev 가 모집단). 그래서 확장은 측정과 짝으로 |
| **D18** | (S3) 비용 상한 | 범위당 1 요청 · 같은 (종목, 제안 스냅숏 digest) 는 제안 신선 창 동안 판정 재사용 · **하루 300 요청 상한, 넘으면 Jev 불가 = 진입 안 함** · 요청당 입력 ≈ 3,509 토큰(a128 s2 median) | 상한 도달 시 Jev 를 건너뛰는 쪽(넓힘)이 아니라 막는 쪽. 상태 조립 GET 은 엔진 low 예산에 계상(안전 불변식 4), 안전 예약 무접촉 |
| **D19** | (S3) 시간 제한 | **한 번 시도 · 8 초**, 재시도 없음 | dispatch 경로에서 지연을 쌓지 않음. a128 은 시도당 10 초 + 재시도(측정 도구) |
| **D20** | (S3) 켜는 조건 | 설정 필드 기본 OFF · audit 기록 · 사람만 flip · a128 tasks 4.2(사람 결정) 기록이 있어야 켜짐 | 안전 불변식 2 · 5 · 7 |

D1~D12 가 이 change 의 READY 조건이다. D13~D20 은 S3 · S5 의 입력이다.

### 사람 결정 기록 (2026-10-15)

사용자 승인 「권장대로」(Manager 전달, 2026-10-15): **D1~D15 · D17~D20 은 위 권고 그대로.** 변경 하나와 후속 하나:

- **D16 변경** — 「no default veto threshold. When the S3 seam is first turned on it runs RECORD-ONLY (Jev judges, logs to ledger, never
  blocks). The veto threshold is set later by a human, per market (KR and US separately), from the a128 report after ≥5 sessions per market.」
  위 표의 D16 행과 §8 S3 행에 반영.
- **D17 후속** — 「a128 is moving to state rev a128-s3 (1.8: WTS news, KR per-stock investor flows, US indices) — v1 Jev state for S3 =
  whatever a128 state rev the report calibrates (s3), not s2.」 위 표의 D17 행에 반영.
- 부수 메모(D1 근거의 유효 범위): a128-s3 가 KR 종목별 투자자 매매를 읽게 되면 §1 의 「KR continuation · reversal 은 공식 데이터로 불가」는
  「그 출처의 권위(공식 · WTS)와 시점 보존을 따져 다시 볼 대상」이 된다. 이번 결정(D1 = breakout KR)은 바뀌지 않는다 — 그 가족들은
  `kr_net_flow` 증거 생산자 · 시점 보존 계약이 여전히 0 이고, WTS 는 조회 전용 출처라 spec 의 증거 권위 조항을 따로 거쳐야 한다.
