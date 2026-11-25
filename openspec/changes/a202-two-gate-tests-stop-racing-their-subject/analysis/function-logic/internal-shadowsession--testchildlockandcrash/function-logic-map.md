# Function Logic Map: `TestChildLockAndCrash`

- Source: `internal/shadowsession/session_linux_test.go` (210-275, base `7bc35e5d1`)
- AST evidence: `ast.json`(편집 전 추출 — 구현 뒤 재추출해 갱신한다)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: Low — 시험 함수. 생산 `Begin`(`session_linux.go:257`) 은 부르기만 하고 바꾸지 않는다.
- 생산 호출자: 없음(시험). 같은 바이너리를 자식으로 다시 띄운다(`exec.Command(os.Args[0], "-test.run=^TestChildLockAndCrash$")`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `SHADOWSESSION_CHILD` env | `"1"` 이면 자식 갈래, 아니면 부모 갈래 | 부모가 `cmd.Env` 로 넣음(225) | — |
| `SHADOWSESSION_ROOT` · `SHADOWSESSION_ID` | 부모 `fixture(t)` 의 경로 · dev/inode | 부모 | 자식 `Begin` 실패 → exit 3 |
| 자식 수명 | **부모가 `cmd.Process.Kill()`(246) 할 때까지** 잠금을 쥐어야 함 — 시험이 재는 「잠금 보유 중 Busy」의 전제 | 시험 설계 | 편집 전: 타이머 없는 `select {}`(221) → 런타임 교착 검출로 `READY` 직후 exit 2, 잠금 해제(`analysis/red/shadowsession-child-before.txt`: 0.10 s, exit 2) |
| state 파일 | 자식 `Begin` 이 `active` 로 씀 | 생산 `Begin` | 자식이 일찍 죽으면 부모 `Begin` 이 잠금을 얻고 `active` 를 보고 `ErrInvalid`(의도된 fail-closed) |

## Branches and early returns — 편집 전(base `7bc35e5d1`)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 211 `SHADOWSESSION_CHILD == "1"` — 자식 갈래 | 자식 `Begin` 으로 잠금 · state `active` | 갈래 끝은 B3 에서 막힘(반환 없음) | 이 시험 자신(부모가 자식으로 띄움) |
| B2 | if 215 자식 `Begin` 오류 | 없음 | `os.Exit(3)` | 이 시험(부모 B7 이 `READY` 아님으로 FAIL) |
| B3 | select 221 `select {}` — **편집 대상** | 없음(영원히 막힘 의도) | 실제: 타이머 없는 자식에서 교착 검출 fatal, exit 2 | 이 시험 + `analysis/red/` |
| B4 | if 227 `StdoutPipe` 오류 | 없음 | `t.Fatal` | 환경 결함 경로(시험 없음 — 시험 함수 자신의 가드) |
| B5 | if 230 `cmd.Start` 오류 | 없음 | `t.Fatal` | 환경 결함 경로 |
| B6 | if 234 (defer 안) `cmd.ProcessState == nil` | 자식 kill · wait | — | 정리 경로(B7 · B8 실패 시) |
| B7 | if 240 첫 줄 읽기 오류 또는 `READY` 아님 | 없음 | `t.Fatalf("child …")` | 이 시험 |
| B8 | if 243 부모 `Begin` 이 `ErrBusy` 아님 | 없음 | `t.Fatalf("lock: …")` — **게이트 1회차 실패 지점** | 이 시험 · `analysis/red/shadowsession-old-test-parent-delay.txt` |
| B9 | if 246 `Kill` 오류 | 없음 | `t.Fatal` | 환경 결함 경로 |
| B10 | if 251 lock 파일 열기 오류 | 없음 | `t.Fatal` | 이 시험 |
| B11 | if 254 kill 뒤 flock 실패 | 없음 | `t.Fatalf("lock not released")` | 이 시험 |
| B12 | if 257 unlock 오류 | 없음 | `t.Fatal` | 환경 결함 경로 |
| B13 | if 260 close 오류 | 없음 | `t.Fatal` | 환경 결함 경로 |
| B14 | if 264 state 파일 열기 오류 | 없음 | `t.Fatal` | 이 시험 |
| B15 | if 269 state 가 `active` 아님 | 없음 | `t.Fatalf("state=…")` | 이 시험 |
| B16 | if 272 죽은 active 재개가 `ErrInvalid` 아님 | 없음 | `t.Fatalf("crashed active")` | 이 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Begin`(생산) | 자식 · 부모 양쪽에서 잠금 시도 | `ErrBusy` / `ErrInvalid` / 오류 — 바꾸지 않음 | AST + `session_linux.go:257` |
| `exec.Command` · `Start` · `Kill` · `Wait` | 자식 프로세스 수명 관리 | 부모가 kill 로 끝냄 | AST |
| `unix.Flock` | 자식 사망 뒤 잠금 해제 확인 | `LOCK_NB` | AST |
| `readState` | state `active` 확인 | 오류면 FAIL | AST |

## State mutations and fallbacks

- 파일 시스템: 시험 전용 `fixture(t)` 임시 디렉터리 안의 `2026-11-06.lock` · `.state`. 운영 설정·저널 접근 없음. 폴백 없음.

## a202 편집 계획

- B3 `select {}` 를 시한이 걸린 대기(`time.Sleep`)로 바꾸고, 그 시한이 지나면 자식이 스스로 끝난다(`os.Exit`). 다른 갈래의 조건·순서·단언 불변.
- 편집 뒤 예상: `select` 갈래 B3 이 사라져 분기 15(B1~B15, 옛 B4~B16 이 하나씩 당겨짐).

## Safety conclusion

- Safe edit boundary: 자식 갈래의 마지막 대기 한 문장(B3)만. 부모 갈래 불변.
- High-risk impact: no — 시험 전용, 생산 diff 0.
