# a210 — design §1 사실 재확인 (구현자, 2026-11-28, base `8db0463b7`)

원문 출력은 `facts-raw.txt`(같은 디렉터리). CodeGraph 질의 원문은 `codegraph-callers.txt` · `codegraph-affected.txt`.

| # | design 의 사실 | 재확인 결과 | 근거 |
|---|---|---|---|
| F1 | `engineOptions` 는 `ConfigDir` 만, `AuditFile` 비어 있음 | 맞음. `engine_test.go:110` 함수 선언, `:111` `engine.Options{ConfigDir: dir}` | facts-raw.txt |
| F2 | `openAuditLog` 는 `AuditFile` 이 비면 `journal.DataDir()/audit.log` | 맞음. `interlock.go:860` 선언, `:862-868` | facts-raw.txt |
| F3 | `DataDir()` 순서 `TOSSOS_DATA_DIR`(절대) → `XDG_DATA_HOME` → `$HOME/.local/share/tossos` | 맞음. `paths.go:64-81`. 상대 `XDG_DATA_HOME` 은 무시, 상대 `TOSSOS_DATA_DIR` 은 오류 | facts-raw.txt |
| F4 | `internal/app/engine` 에 `TestMain` 없음 | 맞음. `rg 'func TestMain' internal/app/engine` 0 건(rc 1) | facts-raw.txt |
| F5 | 운영 컨테이너는 호스트 데이터 디렉터리를 `/var/lib/tossos/data` 로 붙임 | `compose.yaml:16,47-48` 에서 확인(`TOSSOS_DATA_DIR: /var/lib/tossos/data`, 호스트 source 는 `${TOSSOS_DATA_DIR}`). 호스트 쪽 값은 배포 env 파일에 있어 **읽지 않음**(.env 접근 금지). 운영 audit 에 subject `unknown` 행이 2026-07-30T19:31Z 부터 있다는 것이 간접 증거 | facts-raw.txt, `../contamination.md` |
| F6 | 거짓 행 묶음 모양 | 규칙과 묶음 목록은 `../contamination.md` | contamination_scan.py |

## 차이(사실 정정)

- proposal 의 「`engineOptions`·`startEngine` 호출 자리 20 곳」은 **선언 2 줄을 포함한 수**다. 선언을 뺀 호출 자리는 18 곳
  (`engine_test.go` 15, `a071_protection_failure_test.go` 1, `a071_readiness_assembly_test.go` 1, `a166_runtime_replay_wiring_test.go` 1).
  D1 은 이 자리를 고치지 않으므로 결론에 영향 없음.
- 이 패키지에는 이미 `isolate(t)`(`engine_test.go:35` → `testenv.Isolate`, `TOSSOS_DATA_DIR` 포함) 가 있고 90 곳에서 쓰인다.
  **`engineOptions`·`startEngine`·`NewContext` 를 부르는 18 자리는 전부 `dir := isolate(t)` 로 기동한다**(시험 함수별 `dir` 출처를
  awk 로 열거, 19 줄 전부 `isolate(t)`). 직접 `engine.New(engine.Options{…})` 를 부르는 자리(precheck · wts_isolation 등)도 본 곳은
  모두 `isolate(t)` 또는 `AuditFile` 지정. 즉 design 의 「engineOptions 가 AuditFile 을 비워 실제 audit 에 쓴다」는 **경로로는 맞지만
  base 의 엔진 시험은 그 경로에 닿지 않는다** — 실측도 0(`../red/summary.txt`). 실제 누수 시험은 `cmd/tossctl`
  `TestA166EngineDecoratorCannotRemoveRequiredBudget`(`configDir: t.TempDir()` + env 가짜 키로 `assembleEngine` → `engine.NewContext`;
  `recordGateSettings`(engine.go:487)가 계좌 해석(:507)보다 먼저라 계좌 단계에서 실패해도 기동 행이 이미 써짐) — `../red-bisect/`.
  **정정(리뷰 F3, 2026-11-28)**: 무태그에서는 그 시험 하나지만 `tossos_testseams` 에는 둘째 출처가 있음 — `engine_account_seq_recovery_test.go` 의
  5 시험(헬퍼가 `cmd/tossctl` `isolate(t)` 를 쓰는데 그 `isolate` 는 데이터 경로를 안 돌림). HEAD 출처는 이 6 시험(`../measurement.md`, `../review-f3/`).
- 설정 기본 경로: `config.DefaultPaths`(`internal/config/paths.go:20`)는 `os.UserConfigDir()`(`XDG_CONFIG_HOME` → `$HOME/.config`)와
  `os.UserCacheDir()`(`XDG_CACHE_HOME` → `$HOME/.cache`)를 읽는다. 엔진은 `ConfigDir` 가 비면 이 경로(실제 `~/.config/tossctl` — 운영 journal ·
  credential 자리)로 간다.

## CodeGraph(1.6.0, 색인 저장소 primary `/mnt/D/Axipient/workspace/TossOS`, HEAD `8db0463b7` = a210 base)

- 이 worktree 는 codegraph 미초기화 → 같은 커밋인 primary 색인에 질의(읽기 전용).
- `codegraph callers openAuditLog`: 엔진 쪽 호출자는 `NewContext`(`engine.go:452`) 하나. 나머지 8 개는 `cmd/tossctl` 의 **동명 다른 함수**
  (`cmd/tossctl` 자체 `openAuditLog`) 호출자 — 이름 단위 병합.
- `codegraph callers DataDir`: 생산 호출 파일 `cmd/tossctl/*`(7 파일), `internal/app/engine/interlock.go`, `internal/candidate/store.go`(자기 DataDir),
  `internal/journal/paths.go`; 시험 쪽 `cmd/tossctl`(3), `internal/journal/paths_test.go`, `internal/testenv/testenv_test.go`.
- `codegraph affected interlock.go paths.go --filter '*_test.go'`: 시험 파일 1,346 개(journal 을 import 하는 거의 전부) — 정적 도달은
  후보 선정에 변별력이 없음. 그래서 D3 후보는 **기본 경로 해석 함수의 직접 호출 패키지 + 그 직접 import 패키지**로 고르고, 실측으로 판정함:
  `internal/app/engine`, `internal/app`, `internal/app/paths`, `internal/journal`, `internal/candidate`, `internal/config`, `internal/auth`,
  `internal/doctor`, `internal/console`, `cmd/tossctl`, `internal/testenv`, 그리고 `internal/app/paths` 를 import 하는 `tools/` 6 개
  (`a087-market-sell-probe`, `a112-l1c-quote-probe`, `a128-d2-receipts`, `a128-jev-shadow-probe`, `a130-outcome-backfill`, `a140-breakout-signal-shadow`).

## a166 census

- `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 에 `internal/app/engine/*_test.go` 경로 0 건. 이 change 는 새 시험 파일 하나만 더하므로 census 영향 없음.
