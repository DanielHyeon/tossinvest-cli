# a203 분기 주장의 AST 근거

proposal · design 이 근거로 쓰는 **기존 함수 내부 분기 · 호출**은 이 디렉터리의 `tools/logic-map` 산출물로 열거함
(.claude/CLAUDE.md 필수 진입 5). 이 단계는 기존 함수를 편집하지 않음 — 여기 있는 것은 FLM 번들이 아니라 주장의 근거이고,
그래서 `analysis/function-logic/` 이 아니라 이 자리에 둠(선례 a140 `analysis/ast-evidence/`).

- 기준: worktree HEAD `2b86ea27d`(origin/main), go1.26.5, 2026-11-26.
- 생성: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수> > <패키지>--<함수>.ast.json`
- 각 파일의 `source_sha256` 이 그 시점 소스 지문임. 소스가 바뀌면 이 주장은 다시 뽑아야 함.

| 산출물 | 파일 | source_sha256(앞 12) | 분기 | return |
|---|---|---|---|---|
| `a140--evaluateSetup` | tools/a140-breakout-signal-shadow/evaluate.go:66-162 | ff6fed3c1af3 | 15 | 13 |
| `a140--contiguousPrefix` | evaluate.go:51-62 | ff6fed3c1af3 | 2 | 2 |
| `a140--quoteMeasurements` | evaluate.go:165-181 | ff6fed3c1af3 | 4 | 1 |
| `a140--labelProposal` | label.go:18-36 | b9bff19197b7 | 5 | 4 |
| `a140--planSizing` | sizing.go:77-93 | f2db104aa2b5 | 3 | 4 |
| `a140--rMultiplePPM` | sizing.go:96-114 | f2db104aa2b5 | 3 | 4 |
| `a140--krTickMinor` | sizing.go:17-34 | f2db104aa2b5 | 8 | — |
| `a140--probe.selectUniverse` | shadow.go:264-285 | 62c2fe84d040 | 5 | 3 |
| `a140--dispatch` | main.go:55-67 | 56fab140558f | 5 | 4 |
| `breakoutsignal--Derive` | internal/breakoutsignal/signal.go:202-256 | 51dbfe509601 | 14 | 11 |
| `breakoutsignal--RVOLPPM` | signal.go:153-173 | 51dbfe509601 | 5 | 5 |
| `breakoutsignal--OpeningRangeATR` | signal.go:120-149 | 51dbfe509601 | 7 | 6 |
| `breakoutlane--validateQuote` | internal/breakoutlane/rules.go:32-59 | e1e2651c5472 | 7 | 8 |
| `breakoutlane--NewQuoteSeal` | internal/breakoutlane/types.go:126-131 | 138a1b686e25 | 1 | 2 |

## 주장

- **C1 — `evaluateSetup` 은 입출력이 없음.** 호출 열거(`calls`)가 `breakoutsignal.Derive` · `metrics.LaneBars` · `breakoutlane.*`
  생성자 · `breakoutlane.Evaluate` · `krTickMinor` · `planSizing` · `quoteMeasurements` · `errors.New` · 내장 함수뿐이고
  `go_statements` 0 · `defers` 0. 브로커 · 파일 · 시계 호출 0 — 시각은 인자 `nowMS` 와 `sess` 로만 들어옴. 그래서 과거 봉 묶음 ·
  가짜 호가 · 과거 시각을 넣어 그대로 부를 수 있음. (호출되는 `breakoutlane.Evaluate` 내부는 여기서 재열거하지 않음 — a140 archive
  `analysis/ast-evidence/claims.md` 가 그 근거이고, 그 뒤 판본 변화는 미검증.)
- **C2 — 호가는 ARMED 갈래에서만 판정에 쓰임.** B5@94 는 호가 봉인 생성 실패만 거절(값의 신선도 · 스프레드 · 괴리는 안 봄).
  B9@124(`Phase()=="ARMED"`) 안에서만 B10@126 앞부분 탐색 → 진입가 = 되찾은 봉 종가(139) · 손절 = 레인지 저점(140) ·
  `planSizing(entry, q.Ask, …)`(141) → B14@142 사이징을 넣어 다시 판정(144) → `quoteMeasurements`(152). 그래서 ARMED 전 단계는
  호가 값과 무관하고, 재생은 ARMED 시도마다 호가 하나만 꾸미면 됨.
- **C3 — 꾸민 호가가 봉인을 통과하는 조건.** `NewQuoteSeal` B1@127 은 bid · ask · last 중 0, bid > ask, 비정규 통화, digest 불일치만 거절.
  design D3 의 대리 호가(bid = ask − 1 틱, last = ask, 통화 KRW, digest 계산)는 이 조건을 모두 피함.
- **C4 — 대리 호가에서 살아 있는 관문은 괴리 하나.** `validateQuote` B2@37(관측 > 수신 또는 수신 > 평가 → QUOTE_STALE) ·
  B3@40(평가 − 관측 > MaxQuoteAgeMS → QUOTE_STALE)은 관측 = 수신 = 평가 시각이면 서지 않음. B5@48 스프레드 =
  ceil((ask − bid) × 1e6 / mid) 는 1 틱 · 가격 1,000 원 이상이면 `krTickMinor` 표(B2~B8@19~31)상 최대 약 2,504 ppm 으로 MaxSpreadPPM
  3,000 을 넘지 않음(구간 하단 가격 / 틱: 2,000 / 5 · 20,000 / 50 · 200,000 / 500 이 최대). B7@55 괴리 = ceil(|ask − 진입가| × 1e6 / 진입가)
  > MaxEntryDriftPPM → ENTRY_DRIFT_EXCEEDED 만 대리 ask 에 따라 달라짐.
- **C5 — 가상 체결가.** `planSizing` 은 worst = max(진입가, ask)(78), B1@79 손절 ≥ worst 또는 ≥ 진입가면 거절, 왕복 =
  ceil(worst × 요율 / 1e6)(82), 목표 = worst + 왕복 + ceil(위험 × rr / 1e6)(86~92).
- **C6 — 라벨.** `labelProposal` B2@21 은 `fromOpenMS` 전에 연 봉을 건너뜀, B3@24(저가 ≤ 손절 → stop)가 B4@27(고가 ≥ 목표 → target)보다
  먼저라 한 봉에 둘 다면 stop, B5@31 은 `final` 이 아니면 라벨 없음, 끝까지 없으면 마지막 봉 종가로 close(34~35). 호출은
  `rMultiplePPM` · `len` 뿐(입출력 0).
- **C7 — 빈 분.** `contiguousPrefix` B2@54 는 처음 빈 분에서 앞부분을 자르고 그 분을 돌려줌. `Derive` B2@209 첫 봉 ≠ 개장 분 →
  NOT_AT_OPEN, B9@228 빈 분 → MINUTE_GAP.
- **C8 — 지표는 한 세션 안의 봉만 씀.** `RVOLPPM` B2@157 은 index < 15 를 거절하고 B3@161 은 같은 슬라이스의 index−15..index−1 만 더함.
  `OpeningRangeATR` 는 같은 슬라이스의 앞 15 봉만 봄(B2@125). `Derive` 는 인자 `series` 하나와 `regularOpen` 만 받음. 그래서 재생은
  세션마다 그 세션의 정규장 1분봉만 받으면 되고 전날 1분봉이 필요 없음.
- **C9 — a140 종목군.** `probe.selectUniverse` 는 `Rankings(rankingType, "KR", rankingDuration, false, rankingFetch)`(265; 상수는
  shadow.go:23~25 = MARKET_TRADING_AMOUNT · realtime · 100) 뒤 B4@275 에서 KR 6자리 · KRW · 가격 > 0 · 가격 ≤ NotionalCap 이 아니면 건너뛰고
  B3@271 에서 Top 개에서 멈춤. 투자유의 제외 = false. 과거 시점의 realtime 랭킹은 재현할 수 없음(design D4).
- **C10 — 하위 명령 추가는 `dispatch` 본문 편집.** `dispatch` B2@59 `switch` 의 case 는 run(B3@60) · report(B4@62) · default 오류(B5@64)
  셋뿐. `replay` 를 더하려면 이 본문에 case 를 넣어야 하고, 그 함수는 a166 census `shapes.json` 의
  `tools/a140-breakout-signal-shadow/main.go||dispatch` 로 본문이 고정돼 있음(design D2).

## 손으로 읽은 것(AST 열거 아님) — 미검증 표시

- a140 design §2.1-1 「평가기는 ARMED 에서 루프를 끝냄(machine.go:96~99)」 — a140 의 주장을 인용만 함. 재생 설계는 이 주장에 기대지 않도록
  ARMED 시도에 언제나 0..k 봉 앞부분을 넣음(design 「재생 절차」 4).
- `officialbars.PollClosedBars` 가 과거 시각 `PollAt` · 캐시 리더로 한 세션 전체를 적재하는지 — 미검증, tasks 2.2 RED 에서 확인.
- `breakoutlane.NewEvidenceSnapshot` 이 「평가 시각이 마지막 봉보다 몇 분 뒤」를 거절하는지 — 미검증, tasks 2.2 RED 에서 확인.
