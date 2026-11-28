# a210 design

## 1. 사실 (구현자가 착수 때 재확인하고 `analysis/code-context/` 에 원문으로 남긴다)

| # | 사실 | 출처 |
|---|---|---|
| F1 | `engineOptions(dir, srv)` 는 `engine.Options{ConfigDir: dir}` 만 채움, `AuditFile` 비어 있음 | `internal/app/engine/engine_test.go:110` |
| F2 | `openAuditLog` 는 `AuditFile` 이 비면 `journal.DataDir()/audit.log` | `internal/app/engine/interlock.go:860-874` |
| F3 | `journal.DataDir()` 순서 `TOSSOS_DATA_DIR`(절대경로) → `XDG_DATA_HOME` → `$HOME/.local/share/tossos` | `internal/journal/paths.go:64-` |
| F4 | `internal/app/engine` 에 `TestMain` 없음 | `rg 'func TestMain' internal/app/engine` 0 건 |
| F5 | 운영 컨테이너는 호스트 `~/.local/share/tossos` 를 `/var/lib/tossos/data` 로 붙이고 audit 를 거기 씀 | compose · 메모리 기록(a197 배포) |
| F6 | 거짓 행 묶음: `subject=daniel`, 1 초 안 `recordGateSettings` 설정 여러 개, 켜짐 → 꺼짐 | 운영 audit 읽기 전용 분석(2026-11-28) |

## 2. 결정

- **D1 — 패키지 TestMain 으로 데이터 · 설정 디렉터리를 상자에 가둔다.** `internal/app/engine` 에 새 시험 파일 하나
  (`a210_isolated_home_test.go`)를 두고 `TestMain` 에서 임시 디렉터리를 만들어 `TOSSOS_DATA_DIR` 를 그 안으로,
  설정 기본 경로가 읽는 env(구현자가 `config.DefaultPaths` 코드로 확인 — `XDG_CONFIG_HOME` 등)도 그 안으로 돌린 뒤 `m.Run()`.
  이유: 기존 함수(`engineOptions` 20 호출 자리)를 고치지 않으므로 gate ⑤ FLM 번들이 0 개이고, 앞으로 이 패키지에
  더해지는 시험도 자동으로 갇힌다. console 패키지의 `TestMain` 가드(`internal/console/fake_broker_test.go:38`)가 선례.
  한 시험 바이너리에 `TestMain` 은 하나뿐 — 내부(`engine`) · 외부(`engine_test`) 패키지 어디에도 없음을 확인하고 둔다.
- **D2 — 이 패키지 안의 시험이 env 를 스스로 지우거나 바꾸는 자리를 센다.** `t.Setenv("TOSSOS_DATA_DIR", "")` 처럼
  상자를 벗어나는 시험이 있으면 D1 이 그 시험을 못 지킨다. census 결과(0 이어야 함, 아니면 그 시험도 상자 안 경로로)를 기록.
- **D3 — 다른 패키지는 실측으로 고른다.** 가짜 `HOME`(빈 임시 디렉터리)과 `XDG_*`·`TOSSOS_*` 를 비운 환경에서
  후보 패키지 시험을 돌리고, 끝난 뒤 가짜 HOME 아래 생긴 파일을 센다(`GOCACHE`·`GOMODCACHE`·`GOPATH` 는 실제 경로로
  명시해 빌드 캐시가 상자에 생기지 않게 한다). 후보는 시험이 `journal.DataDir`·`config.DefaultPaths`·`os.UserHomeDir`
  기본값에 닿을 수 있는 패키지(rg/CodeGraph 로 고른 목록을 기록). 파일이 생긴 패키지만 D1 과 같은 방식으로 고친다.
  **이 측정은 운영 파일을 건드리지 않는다** — 상자가 HOME 이므로.
- **D4 — 양성 대조.** 수리 전 커밋에서 같은 측정을 돌려 `internal/app/engine` 이 상자에 `.local/share/tossos/audit.log` 를
  만든다는 것을 먼저 보인다(RED). 수리 뒤 0 개(GREEN). 대조 없이 0 은 「측정이 못 봤다」일 수 있다.
- **D5 — 거짓 행 기록.** 운영 audit 를 읽기 전용으로 읽어 F6 규칙에 맞는 묶음의 **시각과 설정 이름만** `analysis/contamination.md`
  에 적는다. 값(한도 · 금액 · 토픽 · 계좌)은 적지 않는다. 파일은 고치지 않는다.

## 3. 시험 · 검증 계획

| 단계 | 내용 | 산출물 |
|---|---|---|
| RED | 수리 전 base 에서 D3 측정 — 엔진 패키지가 상자에 audit.log 를 만듦 | `analysis/red/` |
| GREEN | D1(+ D3 로 찾은 패키지) 뒤 같은 측정 — 상자에 생긴 파일 0 | `analysis/green/` |
| 회귀 | 고친 패키지 시험 전체(무태그 + `tossos_testseams` 가 그 패키지에 있으면) PASS, `make lint` | `analysis/verify/` |
| 대조 | TestMain 을 잠시 지운 사본에서 다시 측정해 audit.log 가 생기는지(변이 대신) | `analysis/verify/` |

## 4. 비례 원칙

생산 diff 0, High-risk 경로 아님 → FLM/BTM · 변이 원장 · 다중 적대 리뷰는 저장소 규칙상 `not-applicable` 사유로 갈음 가능.
그러나 **사용자 지시(2026-11-28)로 구현자와 별도의 적대적 리뷰 1 + gstack review 1 은 한다.** 새 함수를 만들 뿐 기존 함수 본문을
고치지 않으면 gate ⑤ 번들 0 — 기존 시험 함수 본문을 고치게 되면 그 함수마다 번들을 먼저 만든다.

## 5. 남는 위험

- 새 패키지가 같은 기본 경로 누수를 만들 수 있다(저장소 전역 가드 없음). 전역 가드(예: 모든 시험 바이너리가 상자 env 를
  요구)는 범위 밖 — 필요하면 별도 change.
- 운영 audit 에 이미 쌓인 거짓 행은 남는다. 읽는 사람은 `analysis/contamination.md` 의 규칙으로 가려 읽는다.

## 6. 정정 (Manager, 2026-11-28 — 구현 실측 반영)

- **누수원은 엔진 시험 패키지가 아니라 `cmd/tossctl` 의 ~~`TestA166EngineDecoratorCannotRemoveRequiredBudget` 이다~~ 시험이다.**
  〔구현자 정정 2026-11-28, 리뷰 F3〕 HEAD 시험 범위에서 출처는 6 시험: 무태그 · 태그 공통 `TestA166EngineDecoratorCannotRemoveRequiredBudget` 1 +
  `tossos_testseams` 전용 `engine_account_seq_recovery_test.go` 5(헬퍼가 데이터 경로를 안 돌리는 `cmd/tossctl` `isolate(t)` 사용, 2026-07-30 도입).
  그 6 을 뺀 base 측정 0(`analysis/review-f3/`). 운영 audit 의 06066149b 이전 묶음 출처는 미상(`analysis/measurement.md`).
  base 측정(`analysis/measurement.md`)에서 엔진 패키지는 두 태그 모두 0 — 엔진 기동 18 자리가 이미 `isolate(t)` 를 쓴다.
  `cmd/tossctl` 은 두 태그 각 1(`.local/share/tossos/audit.log`). 기전: `assembleEngine` → `engine.NewContext` 가
  `AuditFile` 빈 값으로 기본 경로에 가고, `recordGateSettings` 가 계좌 해석보다 먼저 돌아 기동이 실패해도 행이 남는다.
  F1 · F4 는 사실이지만 누수의 충분 조건이 아니었다(proposal 「원인」의 엔진 귀속은 이 절로 갈음).
- **D1 결정 유지 — 엔진 TestMain 은 방어책으로 둔다.** 이유: `engineOptions` 는 여전히 `AuditFile` 을 비우고, `isolate(t)` 를
  잊은 새 시험 하나면 같은 누수가 난다. 대조(`verify/control-probe-*`)로 TestMain 이 그 경우를 막는 것을 실측했다. 비용은 새 파일 하나 · FLM 0.
- **`cmd/tossctl` 은 기존 `TestMain` 한 줄을 고친다** — 「새 파일만」 불가(이미 TestMain 존재). FLM 번들 1(`analysis/function-logic/cmd-tossctl--testmain/`).
- **헬퍼 사본 2 개 수용.** `internal/testenv` 로 합치면 비시험 Go 파일 변경이 되어 생산 diff 0 목표에 어긋난다. 사본은 두 파일뿐.
- 호출 자리는 20 이 아니라 18(선언 2 줄 포함 수였음). 컨테이너 이후 거짓 행 묶음은 35(`analysis/contamination.md`) —
  proposal 의 「최소 9 번」은 11-16 이후 손 계수였다. 출처 시험이 2026-10-31(06066149b) 에 들어왔으므로 그 이전 묶음의 출처는 가르지 못했다.
  〔구현자 정정 2026-11-28, 리뷰 F3〕 이전 묶음은 컨테이너 이후 22(+ 컨테이너 이전 판정 불가 1 = 23), 행 수 10 · 13 · 15 로 이후(15 · 16 · 17)와 다름.
  태그 시험 5 개(2026-07-30~)가 후보지만 과거 판본을 다시 돌려 귀속하지는 않았다 — 「HEAD 범위 출처 6 시험, 10-31 이전 출처 미상」.
