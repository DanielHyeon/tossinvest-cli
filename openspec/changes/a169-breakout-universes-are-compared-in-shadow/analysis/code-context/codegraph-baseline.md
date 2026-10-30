# CodeGraph baseline — a169 2.1 (`securityType` 노출)

- 도구: codegraph 1.6.0, 색인 = `make sdd-sync` 로 `/tmp/tossos-a169-20261023` 에서 재생성(HEAD `566e1cf7c`, 2026-10-23).
- `codegraph callers adaptStocks` → 4: `Client.Stocks`(stock_reads.go:44) · `TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime` ·
  `TestAdaptStocksUnit` · `TestAdaptStocksEmpty`.
- `codegraph callers Stocks` → 3: `readOperations`(internal/ops/read_operations.go:78 — `ops call stocks`) · `cmd/tossctl/instrument_names.go:36`
  (콘솔 종목명, `Name` 만 읽음) · `tools/a128-jev-shadow-probe/collect.go:29`(메타, `Name` · `MarketCode` · `Currency` 만 PublicState 로 복사).
- `codegraph affected internal/official/stock_reads.go internal/domain/models.go --filter '*_test.go'` → 1,224 건(stock_reads.go 단독 1,052).
  `domain` 은 거의 모든 패키지가 import 하므로 파일 단위 영향은 과대다 — 심볼 단위(위 callers)로 좁혀 시험을 골랐다.

## 0.3 — design 단계 hard evidence (2026-10-28)

- 도구: codegraph **1.6.0**, 색인 = 주 체크아웃 `/mnt/D/Axipient/workspace/TossOS/.codegraph`(파일 3,022 · 노드 52,296 · 엣지 185,436,
  DB mtime 2026-10-28 05:46). **`make sdd-sync` 는 돌리지 않았다**(Manager 지시 — GBrain 단계의 `pull --ff-only` 가 worktree 를 전진시킴).
  `.sdd/index-state.json` 의 codegraph 지문은 HEAD `d4dca006`(2026-10-25)에 머물러 있어 **sdd-sync 기준으로는 stale** 이다. 색인 자체는
  `PostToolUse(Write|Edit) → codegraph sync` 훅이 갱신한다. `git log d4dca006..0c9bfd272 -- <아래 대상 파일>` = 0 커밋이라 대상 파일의 색인
  내용은 현재 HEAD 와 같다고 보고, 줄 번호는 아래처럼 현재 파일에서 다시 확인했다.
- 판정 HEAD: `0c9bfd272`(worktree `TossOS.worktrees/a169-docs`, 브랜치 `docs/a169-design`).
- **호출자 수가 아니라 호출 자리**를 셌다. CodeGraph 는 호출 함수의 선언 줄을 내므로 자리는 `grep -nF` 로 현재 파일에서 확정했다.

### `adaptStocks` · `Client.Stocks`(유형 영수증의 출처)

| 질의 | CodeGraph 결과 | 호출 자리(현재 파일) |
|---|---|---|
| `codegraph callers adaptStocks` | 6 — `Client.Stocks` + 시험 5(`TestAdaptStocksUnit` · `TestAdaptStocksEmpty` · `TestA163…NoBrokerTime` · `TestA169UnknownSecurityTypeIsPassedThroughUnjudged` · `TestA169TheStocksQuoteJSONWithASecurityTypeIsPinned`) | 생산 자리 1: `internal/official/stock_reads.go:51` |
| `codegraph callers Stocks` | 4 — `instrument_names.go`(콘솔 종목명) · `tools/a128-jev-shadow-probe/collect.go` · **`tools/a128-d2-receipts/types.go`(2.1 이후 새 소비자)** · `readOperations` | 생산 자리 4: `cmd/tossctl/instrument_names.go:91` · `tools/a128-jev-shadow-probe/collect.go:392` · `tools/a128-d2-receipts/types.go:89` · `internal/ops/read_operations.go:198` — 함수마다 자리 1 |
| `Quote.SecurityType` 를 읽는 자리 | — | 생산 1: `tools/a128-d2-receipts/types.go:98~99`(원문 그대로 영수증에 적음). a140 도구 0 |

### a140 도구 심볼(Q6 (a) — 종목군 입력을 더할 자리)

CodeGraph 1.6.0 은 `p.selectUniverse(ctx)` 같은 수신자 호출을 `callers` 로 못 찾았다(「No callers found」 — `selectUniverse` · `startSession`).
그래서 `codegraph_explore` 와 `grep -nF` 로 자리를 셌다.

| 심볼 | 생산 호출 자리 | 시험 호출 자리 |
|---|---|---|
| `probe.selectUniverse` | `shadow.go:201`(1) | 0 |
| `probe.startSession` | `shadow.go:550`(1) | `step_test.go` 16 자리(78 … 488) |
| `probe.run` | `main.go:164`(1) | 0 |
| `probe.step` | `shadow.go:561`(1) | `step_test.go:56` |
| `probe.stepSymbol` | `shadow.go:327`(1) | 0 |
| `effectivePolicyFor` | `shadow.go:170`(1) | `policy_test.go` 6 자리 · `step_test.go:475` — CodeGraph 는 「5 callers」(함수 단위: 시험 함수 4 + `startSession`) |
| `evaluateSetup` | `shadow.go:406` · `:453`(2) | `evaluate_test.go` 6 자리 |
| `signalRow` | `shadow.go:420` · `:440` · `:458`(3) | — |
| `renderReport` | `main.go:277`(1) | `evaluate_test.go:269` · `:291` · `policy_test.go:84` |
| `dispatch` | `main.go:47`(1) | 0 |
| `newProbe` | `main.go:150`(1) | `step_test.go:48` |

### 새 읽기 표면(U2 · 유형)

| 질의 | 결과 |
|---|---|
| `codegraph callers Candles` | 5 — `AdaptOfficialMinutePage`(다른 `Candles`) · `bars_test.officialPage` · `tools/a128-jev-shadow-probe/collect.go` · `hybrid.GetChart` · `readOperations` |
| 일봉 호출 자리(`grep -nF '.Candles('`) | `tools/a128-jev-shadow-probe/collect.go:568` `Candles(ctx, symbol, "1d", dailyFetchCount, "", true)` — **이미 있는 선례**(일봉 · 수정주가 · 하루 캐시), `internal/hybrid/client.go:149`, `internal/ops/read_operations.go:263` |
| 20 일 평균 거래량 선례 | `tools/a128-jev-shadow-probe/features.go:216~247` `dailyFeatures` — 오늘 개장 전 봉만 · 마지막 20 봉 거래량 합 / 20(246), `dailyFetchCount = 25`(37). `package main` 이라 import 불가 — 식만 같게 쓴다 |
| 유형 갈래 선례 | `openspec/changes/a128-jev-shadow-judgment-probe/analysis/prereg_eval.py:372~373` `D2_TYPE_GROUP`(9 값) |

- `codegraph affected` 는 이번에 돌리지 않았다 — 구현 전 단계라 편집 파일이 없다. 구현 착수 때 `--filter '*_test.go'` 로 돌린다(하네스 주석).
