# a210 review

## 3.2 적대적 리뷰 (2026-11-28)

- 보이스: 구현자(Opus)와 별도 컨텍스트의 적대적 리뷰 에이전트 1. 처분: Manager(Fable). 처분 구현 · 기록: 구현자.
- 대상: 커밋 `06e6a3e30` · `04d025520` · `adfd5e80b`(base `8db0463b7`).
- 리뷰 원문은 Manager 가 보관. 아래 「발견」은 Manager 가 옮긴 요지, 「처분」은 Manager 처분 원문 그대로.

| # | 발견(요지) | 심각도 | 처분(원문) | 구현 · 증거 |
|---|---|---|---|---|
| F1 | 표지 `A210_TEST_HOME_BOX` 가 있으면 값 검증 없이 `m.Run()` — 위조 · 잔류 표지면 기본 경로가 실제 HOME 으로 떨어짐(리뷰어가 probe 로 재현: marker=forged 에서 DataDir=<HOME>/.local/share/tossos 에 씀) | should-fix | 수리: 자식 갈래에서 표지 디렉터리가 존재하고 `TOSSOS_DATA_DIR == filepath.Join(marker,"tossos")`(XDG 셋도 같은 상자 안인지) 확인, 아니면 stderr 한 줄 + `return 1`(fail-closed). 두 사본 모두. RED: 위조 표지 시험(별도 하위 프로세스로 띄워 rc≠0 과 HOME 무변화 단언, 또는 검증 함수를 분리해 단위 시험) 먼저. | 새 시험 `a210_box_marker_test.go`(엔진 · `cmd/tossctl`): 위조(없는 경로) · 잔류(있지만 경로 env 불일치) 표지로 자식을 띄워 rc≠0 · 거절 문구 · 자식 HOME 아래 파일 0 을 단언. RED(수리 전) 두 패키지 모두 FAIL — 자식 rc 0, `<자식 HOME>/.local/share/tossos/a210-probe` 생성(`analysis/review-f1/red/`). 수리: `a210CheckInheritedBox`(표지 디렉터리 존재 + 경로 env 넷이 `filepath.Join(marker, sub)` 와 같음) 실패 시 `a210: 물려받은 상자 표지 거절: …` + rc 1. GREEN 두 패키지 PASS(`analysis/review-f1/green/`, `-v` 로 하위 시험 둘 PASS 기록). 한계: 부모 시험이 경로 env 를 바꾼 뒤 자식을 띄우면 그 자식도 거절됨(지금 그런 자리 0 — a135 · a166 · mcp 자식은 경로 env 를 안 바꿈). |
| F2 | panic · 시한 초과 시 `defer RemoveAll` 이 안 돌아 `/tmp` 에 상자가 남음(재현됨) | — | 수리 안 함 — 주장을 「정상 종료 시 무잔여, 비정상 종료 시 /tmp 에 시험 자료만 남을 수 있음(실제 사용자 경로 아님)」으로 좁힘. | `analysis/measurement.md` 「잔여물 주장의 범위」. 코드 변경 없음. |
| F3 | 「누수원은 하나」는 HEAD 범위만. 컨테이너 이후 35 묶음 중 23 은 그 시험 도입(06066149b, 10-31) 이전이고 모양(13·15 설정)도 다름 | 문구 정정 + 측정 1 건 | 측정 추가: base 에서 `cmd/tossctl` 을 `-skip '^TestA166EngineDecoratorCannotRemoveRequiredBudget$'` 로 두 태그 상자 측정 → 0 이면 「HEAD 에서 그 시험 외 출처 0(측정)」. measurement.md · design 문구를 「HEAD 범위 유일 출처, 10-31 이전 23 묶음 출처 미상」으로. | 측정(`analysis/review-f3/`, base 를 임시 worktree 로 꺼내 측정 뒤 제거): 무태그 **0**, **태그 1** — 0 이 아님. 가려내기(`review-f3/bisect/`)로 둘째 출처 `engine_account_seq_recovery_test.go`(태그 전용, 2026-07-30 도입) 5 시험 각 1 을 찾음. 여섯 모두 뺀 태그 측정 **0**(`review-f3/skip-both/`). 그래서 문구는 「HEAD 범위 출처 **6 시험**(1 + 태그 5), 10-31 이전 출처 미상」으로 고침. 계수: 06066149b(2026-10-30T18:46Z) 이전은 컨테이너 이후 22 + 컨테이너 이전 판정 불가 1 = 23, 행 수 10·13·15(이후 15·16·17). 수리 뒤 GREEN 태그 0 — TestMain 이 여섯을 함께 덮어 추가 수리 없음. 고친 곳: `analysis/measurement.md` · `design.md` §6(〔구현자 정정〕 표시) · `analysis/code-context/facts.md` · `tasks.md` 2.1 주. |
| F4 | Manager 의 design.md §6 정정이 미커밋 | — | 위 문구 정정과 함께 커밋. §6 은 내 문장이니 사실이 틀린 곳만 고치고 표시. | §6 의 두 곳(누수원 문장 · 이전 묶음 문장)에 〔구현자 정정 2026-11-28, 리뷰 F3〕 표시와 취소선. 나머지 문장은 그대로. |
| F5 | 상자 env 목록이 `internal/testenv.Env`(internal/testenv/testenv.go:163)와 갈라짐 — 거기엔 `TOSSCTL_OPENAPI_KEY/SECRET` 비우기가 있음 | should-fix | search-first: testenv 를 **고치지 않고** 쓸 수 있으면(공개 목록·함수) 시험 파일에서 그것을 써서 사본을 없앨 것. 고쳐야만 쓸 수 있으면 고치지 말고 상자에 두 키 비우기만 더하고 이유 한 줄. 어느 쪽이든 기존 시험이 그 키를 기대하는 자리가 있는지 rg 로 먼저 세서 기록. | `testenv.Env` 는 공개 변수라 **고치지 않고 씀**: TestMain 이 `testenv.Env` 를 돌며 경로 이름 넷은 상자 경로, 나머지(`TOSSCTL_OPENAPI_KEY` · `TOSSCTL_OPENAPI_SECRET` · `TOSSCTL_LANG`)는 빈 값(`testenv.Isolate` 와 같은 규칙). 이름 목록 사본은 없어지고 경로 → 하위 디렉터리 표(`a210BoxedPaths`)만 남음 — `shortcut:` 주석으로 「testenv.Env 에 경로 env 가 더해지면 이 표에도」 명시. census(`analysis/review-f5-census.txt`): 세 이름을 `t.Setenv` 로 직접 주는 자리 15(엔진 2 · cmd 13), 실제 프로세스 값을 `os.Getenv`/`LookupEnv` 로 읽는 시험 자리 0, 나머지 언급은 가짜 getenv 표 · 주석 — 물려받은 값에 기대는 시험 0. `internal/testenv` 변경 0. |
| F6 | `a210PinGoDirs` 가 `go env -w GOCACHE` 설정(GOENV 파일)을 무시. 고정 자체는 필요(리뷰어 확인) | nit | 코드에 `// shortcut: GOENV 파일의 GOCACHE 설정은 무시함, 그런 설정이 생기면 go env 로 읽도록 바꿈` 류 한 줄. 그리고 주장 1 문구의 「캐시 디렉터리에도 쓰지 않음」은 거짓(자식 go 가 실제 go-build 캐시에 씀) → 「데이터·설정·tossctl 캐시」로 정정. | 두 사본의 `a210PinGoDirs` 주석에 `shortcut:` 한 줄 + 「상자로 가는 것은 데이터 · 설정 · tossctl 캐시 기본 경로뿐이고, 자식 go 는 이 고정값(실제 go-build 캐시)에 씀」. 주장 정정: 이 change 가 막는 것은 **데이터 · 설정 · tossctl 캐시 기본 경로**이며 go 빌드 캐시에는 씀. |
| F7 | FLM `function-logic-map.md` Calls 표 줄 번호가 편집 전(:18) · ast.json 은 편집 뒤(:19) | nit | 수리 — 일치시킬 것. TestMain 한 줄을 더 고치게 되면 번들 재생성. | Calls 표 · 머리말에 편집 전 :18 → 편집 뒤 `ast.json` :19 를 함께 적음. TestMain 은 더 안 고쳐 번들 재생성 불필요(`ast.json` sha 그대로, `check_analysis.py` evidence complete). |
| F8 | 하네스 사각 미기재 | 증거 문구 | measurement.md 에 하네스 사각 명시 — 가짜 HOME 안만 셈(절대경로·작업트리·/tmp 못 봄), env -i 가 GOCACHE 를 줘서 pin 갈래 미실행, 파일 수 지표라 출처 수 못 셈, 위조 표지 갈래 미측정(이제 1 의 시험이 덮음). | `analysis/measurement.md` 「하네스의 사각」 네 항목. F3 의 둘째 출처가 바로 「파일 수 지표라 출처 수 못 셈」의 실례. |

### 처분 뒤 검증

- 회귀(고친 두 패키지 전체, 두 태그, 가짜 HOME 상자): `analysis/green/summary.txt`(GREEN 3 회차) — 네 실행 모두 rc 0 · 상자 파일 0(cmd/tossctl 71.8s · 69.7s, 엔진 749.1s · 840.8s), 실행 뒤 `/tmp` 의 `a210-` 0(`green/tmp-leftovers.txt`).
- `make lint`: `analysis/verify/make-lint.log` — rc 0.
- `tools/logic-map/check_analysis.py --change a210-…`: evidence complete(required 1).

### 비례 원칙

- 생산 diff 0, High-risk 경로 아님. 변이 원장 · 다중 적대 리뷰: `not-applicable` — 시험 전용 change(design §4). 적대적 리뷰 1 은 사용자 지시로 수행(이 절).
- Function Logic Map: 기존 함수 1(`cmd/tossctl` `TestMain`) 번들 있음, 그 밖은 새 함수.

## 3.3 gstack review

- (미실행 — Manager 가 별도로 붙임)
