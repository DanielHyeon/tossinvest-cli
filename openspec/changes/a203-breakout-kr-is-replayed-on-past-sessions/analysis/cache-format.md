# a203 캐시 디스크 형식 — a204 어댑터가 읽는 법

구현: `tools/a140-breakout-signal-shadow/a203_cache.go` · `a203_replay.go`(커밋 시점 기준). 형식이 바뀌면 `schema` 값이 바뀜.
값(종목 코드 · 가격)은 저장소 밖에만 있음 — 이 문서는 모양만 적음.

## 자리

상태 디렉터리 `S` = `~/.local/state/tossos-a203/`(`--state-dir` 로 바꿀 수 있음). 디렉터리 0700, 파일 0600(`sessions.txt` 를 `--out` 으로 저장소에 쓸 때만 0644).

```
S/universe/pa-kr-2026-11-02_2026-11-26.txt   P(a) — 사람이 고정(design §5), 도구는 읽기만(sha256 검증, M9)
S/sessions.txt                                `replay sessions` 기본 출력(커밋본은 analysis/sample/sessions.txt)
S/universe/uc-<sample>.json                   세션별 U-C 종목군(sample = control · main)
S/cache/calendar/<YYYY-MM-DD>.json            공식 달력 한 날짜
S/cache/ranking/pb-6mo-100.json               P(b) 랭킹(1 회)
S/cache/daily/<종목>.json                     일봉 200 개(종목당 1 회)
S/cache/minute/<종목>/<before>-c<count>.json  엄격 1분봉 한 페이지
S/a203-evidence.sqlite                        재생 전용 evidence 저장소(report 가 만듦 — 지워도 캐시에서 다시 만들어짐)
S/ledger/<sample>-<UTC 시각>.jsonl            재생 원장(report 실행마다 새 파일, 덮어쓰지 않음)
S/runs.jsonl · S/checkpoint.json              수집 실행 기록 · 마지막 정지 자리
S/control1-mismatches.txt                     대조 1 불일치 목록(종목 코드 포함 — 커밋 금지)
```

## 캐시 파일 한 개(`schema: "a203-cache/v1"`)

```json
{
 "schema": "a203-cache/v1",
 "kind": "minute | daily | ranking | calendar",
 "request": { "market": "KR", "symbol": "…", "count": 200, "before": "…" },
 "fetched_at": "RFC3339 UTC — 보낸 순간",
 "response": { … 응답 값 그대로 … },
 "rate_budget": { … 1분봉만: 응답의 rate 예산 칸 … }
}
```

- `minute` 의 `response` 는 `official.StrictMinutePage` 의 JSON(필드 이름은 Go 이름 그대로: `Market` · `Symbol` · `Candles` · `Terminal` ·
  `NextBefore` · `ReadAt` · `StatusCode` · `BodyDigest`). `Budget` 칸만 빼서 `rate_budget` 에 둠 — 남은 몫 0 을 실은 페이지가 캐시에 그대로 있으면
  재생 · 이어하기가 그 페이지에서 페이지 넘김을 멈추기 때문(코드 주석 `a203Fetcher.minute`).
- `Candles[i]` = `{"Timestamp","Open","High","Low","Close","Volume","Currency"}` — 전부 **응답 글자 그대로의 문자열**, 최신이 앞.
- `daily` 의 `response` 는 `domain.Chart`(`candles[i].time` 은 그 날짜 00:00 KST, `close` · `volume` 은 실수 — 표준 해독기).
- `calendar` 의 `response` 는 `official.MarketCalendarResponse`(`today` · `previousBusinessDay` · `nextBusinessDay`, KR 은 `integrated.regularMarket`).

## ⚠ 1분봉 시각 — 이름표는 **닫는** 시각

`Timestamp` 는 봉이 **닫는** 순간임(2026-08-18 실측, `internal/officialbars/producer.go` `adoptPage` 주석). 09:00~09:01 봉의 이름표는
`…T09:01:00.000+09:00`. 여는 시각 = 이름표 − 1 분. 이 빼기를 빼먹으면 모든 봉이 1 분 밀림.

## 한 세션의 페이지 열(a204 가 d−1 봉을 모으는 법)

수집은 `officialbars.PollClosedBars` 가 정한 요청 열을 그대로 캐시함(재생도 같은 열을 다시 읽음).

1. 첫 페이지 `before` = (그 세션 정규장 마감 + 30 분)을 초 단위로 자르고 KST 로 쓴 글자, 형식 `2006-01-02T15:04:05.000-07:00`
   (예: 마감 15:30 → `YYYY-MM-DDT16:00:00.000+09:00`). `count` = 200.
2. 다음 페이지 `before` = 앞 페이지의 `NextBefore` 글자 그대로(RFC3339). `Terminal` 이 참이거나, 페이지가 비었거나, 페이지의 가장 오래된 봉의
   **여는** 시각이 정규장 개장보다 이르면 멈춤. 페이지는 최대 4 장.
3. 파일 이름의 `<before>` = `before` 글자에서 `:` 만 지운 것. 예: `2026-06-24T16:00:00.000+09:00` → `2026-06-24T160000.000+0900-c200.json`.
4. 페이지를 이어 붙일 때 두 페이지에 같은 분의 봉이 하나 겹칠 수 있음(커서가 포함 상한) — 같은 분이면 하나만 씀.
5. 정규장 봉 = 여는 시각이 [개장, 마감) 안인 봉. 프리 · 애프터마켓 봉도 페이지에 섞여 있음 — 버릴 것.
   15:20~15:30 단일가 구간 봉은 정규장 창 안에 있음(a204 는 15:20~15:29 를 버린다고 함 — 그쪽 규칙).
6. officialbars 는 「다음 봉이 관측된 봉」만 증거로 적재함(가장 새 봉 하나는 버림). a204 가 캐시를 직접 읽으면 이 규칙을 따를지 스스로 정해야 함
   — 마감 뒤 봉이 없는 날은 마지막 정규장 봉이 a203 재생에서 빠짐(시험 `TestA203PastPollLoadsAWholeSession`).

## 종목군 파일(`format: "a203-universe/v1"`)

```json
{
 "format": "a203-universe/v1",
 "sample": "main",
 "sessions_sha256": "커밋된 sessions.txt 의 sha256",
 "universe": { "YYYY-MM-DD": [ {"symbol": "…", "share": "A | B | F"} ] },
 "prior_session": { "YYYY-MM-DD": "그 세션의 d−1 날짜(공식 달력)" }
}
```

- 키는 표본 세션 d(그 sample 의 세션만). `share`: A = d−1 거래대금 상위, B = 거래대금 급증 R 상위, F = B 부족분 채움(design §5).
- **a204 d−1 1분봉**: `universe[d]` 의 종목 전원에 대해 `prior_session[d]` 세션의 1분봉 페이지가 캐시에 있음(design §8 · M10 — 대조 1 의
  a140 종목군은 넣지 않음). d−2 는 수집하지 않음(design 범위 밖) — d−1 에도 그 종목이 종목군이었던 경우만 우연히 있음.
- `prior_close`(a204 입력)는 캐시에 따로 없음 — d−1 일봉 종가(`daily`, 날짜 = `prior_session[d]`) 또는 d−1 마지막 정규장 1분봉 종가 중
  a204 가 고름.

## 세션 목록(`a203-sessions/v1`)

`# ` 로 시작하는 머리 세 줄 뒤에 줄마다 `날짜 역할 개장 마감`(KST `HH:MM`), 오름차순. 역할 = `main` · `control` · `warmup`
(warmup 21 개 = 가장 이른 main 세션 앞 — R 창 d−2 … d−21 과 그 세션의 a204 d−1). collect · report 는 이 파일이 캐시 달력에서 다시 만든 목록과
**바이트로** 같아야 돎.
