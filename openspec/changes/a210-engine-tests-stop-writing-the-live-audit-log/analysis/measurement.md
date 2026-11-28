# a210 D3 · D4 측정 — 가짜 HOME 상자

하네스 `measure.sh`(같은 디렉터리). 패키지마다 새 빈 상자를 `HOME` 으로 주고 `env -i` 로 `XDG_*` · `TOSSOS_*` 를 뺀 채
`go test -count=1 -timeout 45m` 을 돌린 뒤 상자 안 파일을 셈. `GOCACHE` · `GOMODCACHE` · `GOPATH` 는 실제 경로로 명시,
`GOTOOLCHAIN=local` · `GOPROXY=off`. 상자 부모는 세션 스크래치패드, 상자는 측정마다 정확한 경로로 지움.

- `files` = 상자 안 모든 파일. `files_tossos` = go 툴체인 자신이 쓰는 `.config/go/telemetry/` 를 뺀 수(시험 코드가 만든 파일).
  텔레메트리 6~8 개는 모든 패키지 · 수리 전후에 똑같이 생김(`go` 명령 자체, 시험 무관).
- RED 의 첫 7 개 패키지는 `files_tossos` 열이 생기기 전 판본으로 쟀다 — `*.after` 목록에서 `.config` 밖 항목 0 으로 확인(`red/summary.txt` 주석).
- 각 줄의 `# head=… dirty=N` 은 그 실행 시작 때 워킹트리의 미커밋 `.go` 수. 수리 대상 두 패키지의 RED(`cmd/tossctl` 두 태그 ·
  엔진 두 태그)는 전부 `dirty=0`(base 그대로). `dirty=3` 인 RED 줄(journal 이후)은 수리 파일이 다른 패키지(`cmd/tossctl` · 엔진 시험)에만
  있어 그 패키지 빌드에 들어가지 않음.

## RED(base `8db0463b7`) / GREEN(수리 뒤) — `files_tossos`

| 패키지 | 태그 | RED | GREEN | 비고 |
|---|---|---|---|---|
| `cmd/tossctl` | 없음 | **1** (`.local/share/tossos/audit.log`, 기동 설정 21 개 각 1 행 — 출처 1 시험) | **0** | 수리함 |
| `cmd/tossctl` | `tossos_testseams` | **1** (같은 파일 — 출처 6 시험, 아래 F3) | **0** | 수리함 |
| `internal/app/engine` | 없음 | 0 | 0 | D1 TestMain(방어) |
| `internal/app/engine` | `tossos_testseams` | 0 | 0 | D1 TestMain(방어) |
| `internal/journal` | 없음 | 0 | — | 수리 안 함 |
| `internal/journal` | `tossos_testseams` | 0 | — | 수리 안 함 |
| `internal/app` · `internal/app/paths` · `internal/candidate` · `internal/config` · `internal/auth` · `internal/doctor` · `internal/console` · `internal/testenv` | 없음 | 0 (각각) | — | 수리 안 함 |
| `tools/a087-market-sell-probe` · `a112-l1c-quote-probe` · `a128-d2-receipts` · `a128-jev-shadow-probe` · `a130-outcome-backfill` · `a140-breakout-signal-shadow` | 없음 | 0 (각각) | — | 수리 안 함 |

모든 줄 rc 0(RED 첫 판의 journal rc 1 은 기본 10 분 시한 초과 — 45 분으로 다시 재서 rc 0).

### GREEN 1 · 2 · 3 회차

- 1 회차(`green-r1/`)도 네 줄 모두 상자 파일 0 · rc 0 이었으나, 실행 뒤 `/tmp` 에 빈 `a210-*-home-*` 디렉터리 4 개(패키지 전체 실행마다 1 개)가 남음.
  원인: 시험이 자기 바이너리를 자식으로 다시 띄우는 자리(엔진 `a135` panic 자식, `cmd/tossctl` `TestA166MonitorMainExitOne` exit 1 자식)에서
  자식의 TestMain 이 상자를 새로 만들고 panic · `os.Exit` 로 끝나 `defer os.RemoveAll` 이 안 돔. 남은 것은 빈 디렉터리(파일 0)였고 정확한 경로로 지움.
  수리: 부모가 `A210_TEST_HOME_BOX` 표지를 env 로 남기고, 표지를 물려받은 자식은 부모 상자를 그대로 씀.
- 2 회차(`green-r2/`): 그 판으로 네 줄 0 · rc 0, 실행 뒤 `/tmp` 의 `a210-` 0.
- 3 회차(`green/`, 위 표의 GREEN 열): 적대적 리뷰 F1 · F5 수리 판(표지 검증 · `testenv.Env` 사용 — `review.md`). 결과는 표와 `green/summary.txt`.

### 잔여물 주장의 범위 (리뷰 F2)

- **정상 종료(rc 0 이든 시험 실패 rc 1 이든 `m.Run` 이 돌아온 경우)에는 `/tmp` 에 상자가 남지 않음** — 2 · 3 회차 실측.
- **비정상 종료(부모 프로세스의 panic · `-timeout` 초과 · kill)에는 `defer os.RemoveAll` 이 안 돌아 `/tmp/a210-*-home-*` 에 시험 자료가 남을 수 있음**
  (리뷰어가 재현). 남는 것은 시험이 쓴 상자 안 파일뿐이고 실제 사용자 경로(`~/.local/share/tossos` 등)가 아님. 수리하지 않음(처분 F2).

## 누수 시험 가려내기(`red-bisect/` · `green-bisect/` · `review-f3/`)

- `cmd/tossctl -run '^TestA166'` → 1. 엔진을 조립하는 A166 시험 넷을 하나씩 → `TestA166EngineDecoratorCannotRemoveRequiredBudget` 만 1,
  나머지 셋(`CanceledRequiredStartupNoWire` · `EngineAndFlattenRequiredMissingFailBeforeBootstrap` · `DirectAppBuilderMissingContextFailsClosed`) 0.
- 같은 시험을 수리 뒤(1 회차 판) → 0. 이후 판은 패키지 전체 GREEN 으로 덮음.
- 기전: 그 시험은 `configDir: t.TempDir()` + env 가짜 키로 `assembleEngine` → `engine.NewContext`. `AuditFile` 이 비어 `openAuditLog` 가
  `journal.DataDir()` = `$HOME/.local/share/tossos` 로 가고, `recordGateSettings`(engine.go:487)가 계좌 해석(:507)보다 먼저 돌아서
  시험이 기대한 대로 기동이 실패해도 기동 설정 행은 이미 써짐. 시험 config 는 기본값(꺼짐)이라 운영 audit 의 마지막 값(켜짐) 위에
  true→false 로 기록됨. 이 시험은 2026-10-31(06066149b, a166)에 들어옴.
- **그 시험을 뺀 base 측정(리뷰 F3, `review-f3/`)** — base `8db0463b7` 를 임시 연결 worktree 로 꺼내(측정 뒤 제거)
  `-skip '^TestA166EngineDecoratorCannotRemoveRequiredBudget$'`:
  - 무태그 **0** → 무태그 `cmd/tossctl` 에서 그 시험 외 출처 0(측정).
  - `tossos_testseams` **1** → 태그 시험에 둘째 출처가 있음. 하나씩 가려냄(`review-f3/bisect/`):
    `engine_account_seq_recovery_test.go`(`//go:build tossos_testseams`, 2026-07-30 93165f969 도입)의 5 개
    — `TestActualEngineRecoveryReusesTheStartupAccountSequence` · `…AcceptsAMatchingExplicitAccountSequence` · `…StillFailsClosedOnASnapshot429` ·
    `TestEngineRefusesAnIncompleteFirstAccountRecordBeforeOpeningTheJournal` · `TestEngineRefusesAnExplicitSequenceThatDoesNotMatchTheFirstRecord` 가 각 1,
    다른 태그 파일 `engine_assembly_test.go` 의 `TestAssembleEngineWires…`(`testenv.Isolate` 사용) 0.
    기전: 공통 헬퍼 `assembleAccountSequenceTestEngine` 이 `cmd/tossctl` 의 `isolate(t)` 를 쓰는데 그 `isolate` 는 `XDG_CONFIG_HOME` · `XDG_CACHE_HOME` 만
    돌리고 `TOSSOS_DATA_DIR` · `XDG_DATA_HOME` 은 안 돌림 → audit 가 `$HOME/.local/share/tossos` 로 감.
  - 여섯 시험을 모두 뺀 태그 측정 **0**(`review-f3/skip-both/`) → HEAD 의 `cmd/tossctl` 출처는 이 6 시험(측정).
  - 수리 뒤 GREEN(태그)은 0 — TestMain 상자가 여섯 모두를 덮음(시험별 수리 없음).
- **출처 귀속의 범위**: 위 결론은 **HEAD 시험 범위**뿐임. 운영 audit 의 컨테이너 이후 35 묶음 중 22 는 06066149b(2026-10-30T18:46Z) 이전이고
  (컨테이너 이전 판정 불가 1 을 더하면 23), 행 수도 10 · 13 · 15(이후 13 묶음은 15 · 16 · 17)로 다름 — 이전 묶음의 출처는 **미상**.
  태그 시험 5 개는 2026-07-30 부터 있었으므로 후보이지만, 과거 판본을 다시 돌려 귀속하지는 않음(과거 이력 재현은 범위 밖).

## 하네스의 사각 (리뷰 F8)

- **가짜 HOME 아래만 셈.** 시험이 절대경로 · 작업트리 · `/tmp` 등 HOME 밖에 쓰는 것은 못 봄.
- **`env -i` 가 `GOCACHE` 를 실제 값으로 주므로** TestMain 의 `a210PinGoDirs` 「비어 있으면 채움」 갈래는 측정 중 한 번도 안 돎
  (기본값과 같다는 것은 `verify/go-dir-pin.txt` 의 계산식 비교뿐).
- **파일 수 지표**라 같은 파일에 쓰는 출처가 여럿이면 1 로 보임 — 출처 수는 `-run` / `-skip` 가려내기로만 셈(위 F3 측정이 그 예).
- **위조 · 잔류 표지 갈래는 이 측정이 안 봄** — 이제 `a210_box_marker_test.go`(두 패키지) 시험이 덮음(`review-f1/`).

## 대조 — 엔진 TestMain 이 실제로 가두는지(`verify/control-probe-*`)

base 엔진 시험은 누수가 0 이라 「수리 뒤 0」만으로는 엔진 TestMain 이 일하는지 알 수 없음. 그래서 `isolate(t)` 없이
`engineOptions(t.TempDir(), 스텁)` + env 가짜 키로 기동하는 **임시 시험**(`a210_probe_test.go`, 스크래치패드에 두고 측정 때만 복사 · 측정 뒤 삭제,
커밋 안 함)을 하나 넣어 잼:

| 조건 | `files_tossos` |
|---|---|
| 임시 시험 + `a210_isolated_home_test.go` 있음 | 0 |
| 임시 시험 + `a210_isolated_home_test.go` 를 잠시 치움 | **1** (`.local/share/tossos/audit.log`, 기동 설정 21 행) |

`cmd/tossctl` 쪽 대조는 RED 자체(base = 상자 없음 → 1)와 GREEN(→ 0)이 같은 일을 함.
