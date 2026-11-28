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
| `cmd/tossctl` | 없음 | **1** (`.local/share/tossos/audit.log`, 기동 설정 21 개 각 1 행) | **0** | 수리함 |
| `cmd/tossctl` | `tossos_testseams` | **1** (같음) | **0** | 수리함 |
| `internal/app/engine` | 없음 | 0 | 0 | D1 TestMain(방어) |
| `internal/app/engine` | `tossos_testseams` | 0 | 0 | D1 TestMain(방어) |
| `internal/journal` | 없음 | 0 | — | 수리 안 함 |
| `internal/journal` | `tossos_testseams` | 0 | — | 수리 안 함 |
| `internal/app` · `internal/app/paths` · `internal/candidate` · `internal/config` · `internal/auth` · `internal/doctor` · `internal/console` · `internal/testenv` | 없음 | 0 (각각) | — | 수리 안 함 |
| `tools/a087-market-sell-probe` · `a112-l1c-quote-probe` · `a128-d2-receipts` · `a128-jev-shadow-probe` · `a130-outcome-backfill` · `a140-breakout-signal-shadow` | 없음 | 0 (각각) | — | 수리 안 함 |

모든 줄 rc 0(RED 첫 판의 journal rc 1 은 기본 10 분 시한 초과 — 45 분으로 다시 재서 rc 0).

### GREEN 1 회차 → 2 회차

- 1 회차(`green-r1/`)도 네 줄 모두 상자 파일 0 · rc 0 이었으나, 실행 뒤 `/tmp` 에 빈 `a210-*-home-*` 디렉터리 4 개(패키지 전체 실행마다 1 개)가 남음.
  원인: 시험이 자기 바이너리를 자식으로 다시 띄우는 자리(엔진 `a135` panic 자식, `cmd/tossctl` `TestA166MonitorMainExitOne` exit 1 자식)에서
  자식의 TestMain 이 상자를 새로 만들고 panic · `os.Exit` 로 끝나 `defer os.RemoveAll` 이 안 돔. 남은 것은 빈 디렉터리(파일 0)였고 정확한 경로로 지움.
- 수리: 부모가 `A210_TEST_HOME_BOX` 표지를 env 로 남기고, 표지를 물려받은 자식은 부모 상자를 그대로 씀(새 상자 안 만듦).
- 2 회차(`green/`)는 이 판으로 다시 잼 — 위 표의 GREEN 열. 실행 뒤 `/tmp` 의 `a210-` 항목 0(`green/tmp-leftovers.txt`).

## 누수 시험 가려내기(`red-bisect/` · `green-bisect/`)

- `cmd/tossctl -run '^TestA166'` → 1. 엔진을 조립하는 A166 시험 넷을 하나씩 → `TestA166EngineDecoratorCannotRemoveRequiredBudget` 만 1,
  나머지 셋(`CanceledRequiredStartupNoWire` · `EngineAndFlattenRequiredMissingFailBeforeBootstrap` · `DirectAppBuilderMissingContextFailsClosed`) 0.
- 같은 시험을 수리 뒤(1 회차 판) → 0. 2 회차 판은 패키지 전체 GREEN 으로 덮음.
- 기전: 그 시험은 `configDir: t.TempDir()` + env 가짜 키로 `assembleEngine` → `engine.NewContext`. `AuditFile` 이 비어 `openAuditLog` 가
  `journal.DataDir()` = `$HOME/.local/share/tossos` 로 가고, `recordGateSettings`(engine.go:487)가 계좌 해석(:507)보다 먼저 돌아서
  시험이 기대한 대로 기동이 실패해도 기동 설정 행은 이미 써짐. 시험 config 는 기본값(꺼짐)이라 운영 audit 의 마지막 값(켜짐) 위에
  true→false 로 기록됨 — `contamination.md` 의 묶음 모양과 같음. 이 시험은 2026-10-31(06066149b, a166)에 들어옴 — 그 전 묶음의
  출처는 이 측정으로 가르지 못함.

## 대조 — 엔진 TestMain 이 실제로 가두는지(`verify/control-probe-*`)

base 엔진 시험은 누수가 0 이라 「수리 뒤 0」만으로는 엔진 TestMain 이 일하는지 알 수 없음. 그래서 `isolate(t)` 없이
`engineOptions(t.TempDir(), 스텁)` + env 가짜 키로 기동하는 **임시 시험**(`a210_probe_test.go`, 스크래치패드에 두고 측정 때만 복사 · 측정 뒤 삭제,
커밋 안 함)을 하나 넣어 잼:

| 조건 | `files_tossos` |
|---|---|
| 임시 시험 + `a210_isolated_home_test.go` 있음 | 0 |
| 임시 시험 + `a210_isolated_home_test.go` 를 잠시 치움 | **1** (`.local/share/tossos/audit.log`, 기동 설정 21 행) |

`cmd/tossctl` 쪽 대조는 RED 자체(base = 상자 없음 → 1)와 GREEN(→ 0)이 같은 일을 함.
