# a169 design 의 분기 근거 — AST 열거 색인

- 2026-10-28 review §4 반영 때 2 개 추가(`probe.armedAttempt` · `parseDecimal`, 같은 HEAD `0c9bfd272` · 같은 명령) — 합계 15.
- 2026-10-29 review §6 반영 때 5 개 추가(`LoadCredentials` · `apppaths.OpenAPI` · `tokenManager.token` · `quoteRead` · `probe.gap`, 같은 HEAD `0c9bfd272` ·
  같은 명령, design §7.5 를 쓰기 전) — 합계 20. 다섯 다 이 change 가 **바꾸지 않는** 함수다(근거로만).
- 2026-10-29 review §8 반영 때 6 개 추가(`engineJournalDir` · `spawnDetachedEngine` · `startEngine` · `classifyStatus` · `pacedSource.pause` ·
  `tokenManager.loadCache`, 같은 HEAD `0c9bfd272` · 같은 명령, design §7.5 감시기 · G6 를 쓰기 전) — 합계 26. 여섯 다 바꾸지 않는 함수다.

- 생성: 2026-10-28, HEAD `0c9bfd272`, `GOFLAGS=-trimpath go run ./tools/logic-map --file <f> --func <fn>`(규칙 5 — design 을 쓰기 **전에** 만듦).
- 각 JSON 머리의 `source_sha256` 이 판정 대상 바이트다. `tools/a140-breakout-signal-shadow/shadow.go` 는 `62c2fe84…` 로
  정의 단계 산출물(`../ast-evidence/`, HEAD `10377cb4b`)과 같다 — a140 도구는 그 뒤 바뀌지 않았다.
- B-id 는 위치 번호다(편집하면 뒤 행이 밀린다). 구현 단계 FLM 은 편집 전 이 파일을 `revision: base` 로 다시 뽑아 대조한다.

| 파일 | 함수 | 줄 | 분기 · 반환 | design 이 쓰는 주장 |
|---|---|---|---|---|
| `tools-a140-breakout-signal-shadow-shadow.go--probe.run.json` | `probe.run` | 539–566 | B 8 · R 7 | B3@547 개장 + 30 초 대기 → B4@550 `startSession`, B5@553 루프는 분마다 `:05`(554) step, B6@555 마감 + 라벨 유예에서 끝, 565 `finish` |
| `…--probe.startSession.json` | `probe.startSession` | 164–261 | B 26 · R 10 | B3@182 · B6@189 원장의 마지막 universe 행이 종목 집합(196), B9@200 없을 때만 `selectUniverse`(201), B22@249 · B23@250 끝나지 않은 종목마다 시작 호가(251) |
| `…--probe.selectUniverse.json` | `probe.selectUniverse` | 264–285 | B 5 · R 3 | B3@271 `Top` 차면 멈춤, B4@275 6 자리 · KRW · NaN · ≤0 · 상한 초과 건너뜀, B5@281 0 개만 거부 |
| `…--probe.step.json` | `probe.step` | 314–336 | B 6 · R 4 | B1@317 `p.symbols` 를 순서대로 돎, B3@321 끝났고 라벨도 끝난 종목 건너뜀, B4@324 마감 뒤 라벨 대기만 |
| `…--probe.stepSymbol.json` | `probe.stepSymbol` | 356–429 | B 21 · R 13 | B12@396 끝난 setup · 마감 뒤 · 앞부분 < 15 봉이면 평가 안 함, B13@399 호가가 없으면 그때 읽음(400), B17@414 ARMED/PROPOSED → `armedAttempt`, B18@417 단계 · 거절이 직전과 같으면 행 없이 반환(418) — 신호 행은 단계가 바뀔 때만 |
| `…--probe.armedAttempt.json` | `probe.armedAttempt` | 435–474 | B 8 · R 4 | (review §4 반영 때 추가) B2@439 창 만료 → `signalRow` 행(440) 반환 443, B3@446 호가 읽기 실패 → 결손 행만(447, 신호 행 없음), 그 밖은 `signalRow`(458) 행 473 — 신호 행은 모두 `signalRow` 라 `bars` 를 실음 |
| `tools-a140-breakout-signal-shadow-evaluate.go--evaluateSetup.json` | `evaluateSetup` | 66–162 | B 15 · R 13 | B9@124 ARMED 일 때 B10@126 앞부분 길이를 늘려 처음 ARMED 인 n 을 찾음(선례 — GET 0) |
| `tools-a140-breakout-signal-shadow-policy.go--effectivePolicyFor.json` | `effectivePolicyFor` | 54–69 | B 1 · R 1 | 분기 없는 리터럴 — 판(lot) digest 는 리터럴 칸의 JSON 바이트(`digest` 72–77) |
| `tools-a140-breakout-signal-shadow-main.go--runCommand.json` | `runCommand` | 68–165 | B 19 · R 19 | 브로커 호출 전 검사 전부(B1@83 … B8@111), 133 `official.New` 를 `marketSource` 로만, 150 `newProbe` |
| `tools-a140-breakout-signal-shadow-main.go--dispatch.json` | `dispatch` | 53–65 | B 5 · R 4 | B2@57 `run` · `report` 두 갈래, 그 밖 거부 |
| `tools-a140-breakout-signal-shadow-source.go--retrying.json` | `retrying` | 85–104 | B 5 · R 3 | 88~94 모든 GET 앞 `gap`(250 ms) 대기(B2@89 재시도면 지연 교체 · B3@92 대기), rate limit 만 재시도(B4@97), B5@100 지연 셋을 다 쓰면 포기 |
| `internal-official-market_reads.go--parseDecimal.json` | `parseDecimal` | 15–24 | B 2 · R 3 | (review §4 반영 때 추가) B1@16 빈 문자열 → 0(17), B2@20 파싱 실패 → 0(21) — 결측과 0 이 같은 값 |
| `internal-official-credentials.go--LoadCredentials.json` | `LoadCredentials` | 29–49 | B 4 · R 5 | (review §6) B1@32 환경 변수 둘 다 있으면 파일을 안 읽고 반환(33), B3@38 파일 없음 → (nil, nil)(39) — §7.5 G3 의 근거 |
| `internal-app-paths-paths.go--OpenAPI.json` | `apppaths.OpenAPI` | 48–61 | B 2 · R 3 | (review §6) B1@51 `configDir` 가 있으면 자격증명 · 토큰 둘 다 그 안(52), 아니면 기본 경로(60 — 토큰은 캐시 디렉터리) |
| `internal-official-token.go--tokenManager.token.json` | `tokenManager.token` | 60–79 | B 2 · R 3 | (review §6) B1@67 메모리 토큰, B2@72 디스크 토큰이 유효하면 그대로 채택(74) — 키 비교 없음, §7.5 G6 의 근거 |
| `tools-a140-breakout-signal-shadow-source.go--quoteRead.json` | `quoteRead` | 107–126 | B 3 · R 3 | (review §6) B1@112 정지 중이면 GET 없이 `errQuoteSuspended`(113), B3@120 rate limit 이면 정지 시작 — 도구 MARKET_DATA 429 의 흔적 |
| `tools-a140-breakout-signal-shadow-shadow.go--probe.gap.json` | `probe.gap` | 140–153 | B 2 · R 2 | (review §6) 141 사유 = `err.Error()`(B1@142 길이 상한), B2@148 같은 분 같은 결손은 접음(149) |
| `cmd-tossctl-engine.go--engineJournalDir.json` | `engineJournalDir` | 185–194 | B 2 · R 3 | (review §8) B1@186 `--config-dir` 가 있으면 그 디렉터리가 엔진 디렉터리(187), 아니면 journal 기본 경로의 디렉터리(193) |
| `cmd-tossctl-engineproc.go--startEngine.json` | `startEngine` | 177–227 | B 7 · R 8 | (review §8) 208 `logPath := engineLogPath(dir)` → 209 `engineSpawnDetached` — 엔진 로그 = 엔진 디렉터리의 `engine.log`(152 · 67) |
| `cmd-tossctl-engineproc.go--spawnDetachedEngine.json` | `spawnDetachedEngine` | 436–461 | B 3 · R 4 | (review §8) 440 `O_APPEND\|O_CREATE\|O_WRONLY` 로 열고(B2@441 실패 → 442) 446~447 stdout · stderr 를 그 파일에 붙임 — 회전 없음 |
| `internal-official-errors.go--classifyStatus.json` | `classifyStatus` | 39–61 | B 6 · R 5 | (review §8) B4@54 429 → `ErrRateLimited`(55, HTTP 코드 없음), 401 · 403 만 `(HTTP %d)`(B3@50 → 51 · 53) — 감시기 패턴의 근거 |
| `tools-a140-breakout-signal-shadow-source.go--pacedSource.pause.json` | `pacedSource.pause` | 77–82 | B 1 · R 2 | (review §8) B1@78 지연 0 이어도 `ctx.Err()`(79), 아니면 `sleep(ctx, d)`(81) — 취소 뒤 새 GET 0 의 근거 |
| `internal-official-token.go--tokenManager.loadCache.json` | `tokenManager.loadCache` | 178–194 | B 3 · R 4 | (review §8) B1@180 읽기 실패 → 없음이면 (nil, nil)(182) · 그 밖 오류(187), B3@190 JSON 실패 → 오류(191) — 파싱 못 하는 도구 토큰은 채택되지 않음 |
| `internal-breakoutlane-machine.go--evaluateFresh.json` | `evaluateFresh` | 36–128 | B 20 · R 9 | B4@52 opening range 뒤 봉을 순서대로, B6@57 처음 돌파 조건을 만족한 봉에서 `break`(66) — **세션당 종목 하나에 돌파 봉 하나** |
| `internal-breakoutsignal-signal.go--Derive.json` | `Derive` | 202–256 | B 14 · R 11 | B11@238 봉마다 지표(윗꼬리는 그 봉만), B13@245 i ≥ 15 일 때만 `RVOLPPM(bars, i)`(246, 직전 15 봉) — 앞부분에 대해 안정 |
| `internal-breakoutsignal-signal.go--OpeningRangeATR.json` | `OpeningRangeATR` | 120–149 | B 7 · R 6 | B2@125 첫 15 봉만 읽음 — ATR 은 앞부분 길이와 무관(주석 117–118) |

## 이 change 가 손대지 않을 함수(대조용)

`evaluateSetup` · `evaluateFresh` · `Derive` · `OpeningRangeATR` · `retrying` 은 design 이 **바꾸지 않는다**. 근거로만 쓴다.
