# a130 검증 영수증 (구현 리뷰 P2-1)

GREEN 커밋 `dbb0681c`, 구현 리뷰 수리는 그 뒤 커밋(아래 「수리 커밋」). 각 줄은 명령 · rc · 순간(UTC) · 잰 트리.
운영 원장 · 계좌 데이터는 저장소에 없음(F-6) — 수 · 사유만.

## 시험 · 정적 검사

| 항목 | 명령 | rc | 순간 · 트리 |
|---|---|---|---|
| a130 시험(편집 패키지) | `go test -count=1 -run TestA130 ./internal/journal/ ./internal/app/engine/ ./internal/official/ ./tools/a130-outcome-backfill/ ./internal/console/` | 0 | 2026-10-13 08:10~08:20Z(초 미기록) · HEAD dbb0681c + 수리 워킹트리 |
| lint | `make lint`(gofmt · go vet · 태그 vet) | 0 | 2026-10-13 09:00~09:20Z(초 미기록) · HEAD 02660f3d + 수리 워킹트리 |
| test-seams | `make test-seams`(태그 · 무태그 전체) | 0 | 2026-10-13T09:50:53Z~10:50:24Z · HEAD 35936cc8 + 수리 워킹트리 |
| race(전체 6 패키지) | `go test -count=1 -race -timeout 45m ./internal/journal/ ./internal/app/engine/ ./internal/console/ ./internal/official/ ./tools/a130-outcome-backfill/ ./internal/position/` | 1 | 2026-10-13T10:50Z~11:35:41Z · HEAD 35936cc8 — 다섯 패키지 통과, `internal/journal` 전체가 -race 에서 45분 상한에 닿음(2700.0s; 병행 세션의 `make test-seams` 와 같은 시각). 실패 시험 이름 없음(시간 상한) |
| race(a130 범위) | `go test -count=1 -race -timeout 45m -run 'TestA130\|Outcome\|Backfill\|Adopt' ./internal/journal/ ./internal/app/engine/` | 0 | 2026-10-13T11:38:10Z~11:42:22Z · HEAD 2c2daae7 + 수리 워킹트리 (journal 240s · engine 185s) |
| race(저장소 정본) | `make test-race` | 0 | 2026-10-13T11:42Z~11:48:06Z · 같은 트리 |
| openspec | `openspec validate a130-an-adopted-trade-keeps-its-outcome --strict` | 0 | 2026-10-13 · 수리 워킹트리 |
| sdd-sync | `make sdd-sync` | 2 | 2026-10-13 12:5x~13:2xZ, 3회 — CodeGraph 갱신은 매번 성공, `codegraphcontext update` 만 300초 시한 초과(같은 시각 `cgc mcp start` 서버가 떠 있음, advisory). GBrain advisory busy |
| sdd-check | `make sdd-check` | 0 | 2026-10-13T13:25:12Z · HEAD 0e4b3fc9(a130 수리 커밋 727a841e 포함) — 직전 1회는 `codegraph status` 15초 시한 초과로 rc 2(부하, 재시도로 통과) |

`-skip TestA133`(BTM 커버리지 하네스의 전체 판): 병행 change a133 의 미커밋 RED 시험이 `internal/app/engine` 에 있어 전체 판을 그것으로 실패시킴(2026-10-13 실측, a130 무관).
첫 test-seams 판(dbb0681c)은 `TestExitEligibilityHasOneSpelling` 하나로 실패 — 새 파일 둘이 `adoption_id` 를 읽음; 허용 목록에 사유와 함께 추가(`internal/position/eligibility_drift_test.go`).

## 변이 (사본, 무변이 대조군 GREEN 먼저)

- `analysis/mutation/round1.log` — dbb0681c 사본, 31 중 CAUGHT 30 · M24 동등(쓴 분봉 종가가 결과 값을 정하고 결과 값이 토큰에 들어감).
- `analysis/mutation/round2.log` — 구현 리뷰 생존 후보 M32 · M33 · M34: M32 CAUGHT · M34 CAUGHT · M33 첫 판 SURVIVED → `TestA130AFailedAmendDoesNotDateALocalSell` 추가 뒤 CAUGHT.
- 하네스 `analysis/harness/a130_mutate.py`(`--only` · `--tree`), 사본은 시스템 임시 디렉터리(원본 무접촉).

## Branch Test Map 커버리지

`analysis/harness/a130_btm_coverage.py` → `analysis/verify/btm-coverage.md` 와 편집 묶음 9 의 `branch-test-map.md` 표. 시험 하나씩 커버리지 프로필로
분기 몸체를 실행한 시험을 적고, 미도달은 원천을 읽은 사유(`not-applicable`)가 있어야 rc 0. 결과(2026-10-13 여섯째 판, HEAD 2c2daae7 + 수리 워킹트리, rc 0): 편집 묶음 9 의 분기 110 — 이름 붙은 시험이 몸체 실행 38 · 후보 패턴 밖 패키지 전체 시험이 실행 25 · `not-applicable` 사유 47(대부분 SQLite 오류 반환 갈래; 무편집이고 시험 전체가 도달하지 않는 갈래 2 — ApplyExitFill B2 `!found` · RecordFill B4 NaN/Inf — 는 a130 이전부터의 시험 공백으로 기록). 첫 판들은 조건식 블록을 몸체로 세어 오류 갈래를 「덮임」 으로 적었음 — 원시 프로필로 확인 뒤 정정.

## 운영 원장 사본 (읽기 전용, 수 · 사유만)

- census: `census-journal-copy.md` — 판정 코드 dbb0681c, 동결 행 31/31 재계산 같음 · 새로 거절 0 · 판정 최대 2.694 ms.
- 분봉 판정: `intake-candles.md` — 묶음 23 · 그 분의 봉 있음 21 · 없음 2.
- 도구 미리보기(dbb0681c 도구, 스키마 32 사본, 사본 무변경): 모집단 12 → fill 9 · `no_intake_candle` 2 · `sell_unpriced` 1. 항목별 값은 완료 보고에만.
- 쓰기 잠금 시간(design D4): 이관한 사본(스키마 35)에서 도구 실행 — 항목당 9.8~22.17 ms. 리뷰 P3-4a 수리 **전** 측정이라 잠금 대기를 포함하나, 경쟁 없는
  사본이므로 대기 ≈ 0 — 상한으로 유효.
- R4-4 「READOPTED 훑기 판정 최대 20~27 ms」 는 미커밋 초안 판정에서 잰 값이라 커밋된 코드로 재현할 수 없음. 현재 판정의 값은 census(최대 2.694 ms).
- 사본 삭제: 스크래치패드의 운영 원장 사본 전부 2026-10-13T08:53:43Z 삭제(리뷰 P3-5).

## 토큰 캐시 사고 (리뷰 4차 R4-1)

2026-10-13 04:32Z~05:11Z 의 실측(`tossctl ops call` · 도구 미리보기)이 `--config-dir` 없이 기본값 토큰 캐시를 써 새 토큰을 발급했고, 05:11:33Z 캐시 토큰 ·
05:11:35Z 엔진 config 토큰이 2초 차로 다시 쓰였음(장 마감 뒤, 주문 없음). 이후 도구는 묶음이 있으면 `--config-dir` 없이는 client 를 열지 않고, 쓰는 토큰
파일 경로를 출력함. `--config-dir ~/.config/tossctl` 로 돈 뒤 판은 토큰 파일을 다시 쓰지 않았음(mtime 확인).
