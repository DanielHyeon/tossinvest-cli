# a202 design — 기존 시험 2건의 경합 수리

base `7bc35e5d1`(origin/main). 증거: `analysis/red/` · `analysis/green/` · `analysis/verify/` · `analysis/function-logic/` · `analysis/code-context/`.

## §1 원인(측정)

### D1 `TestChildLockAndCrash` — 자식이 교착 검출로 일찍 죽는다

- 자식 갈래(`session_linux_test.go:211-222`)는 `Begin` 으로 잠금을 쥐고 `READY` 를 찍은 뒤 `select {}`(221) 로 막는다.
- 자식은 `exec.Command(os.Args[0], "-test.run=^TestChildLockAndCrash$")` 로 뜬다 — `go test` 가 부모에게 주는 `-test.timeout` 이 자식에는 없어
  살아 있는 타이머가 0 이다. 그러면 Go 런타임은 모든 goroutine 이 막힌 것을 교착으로 보고 `fatal error: all goroutines are asleep - deadlock!` 로 끝낸다.
- 측정(편집 전, `analysis/red/shadowsession-child-before.txt`): 자식 직접 실행 → `READY` 뒤 **0.10 s 만에 exit 2**.
- 결과: 부모가 `READY` 를 읽고(240) `Begin`(243) 에 닿는 사이에 자식이 죽으면 flock 이 풀려 부모가 잠금을 얻고, state `active` 때문에 `ErrInvalid` 를 받는다.
  사본에서 부모를 300 ms 늦추면 3/3 FAIL(`analysis/red/shadowsession-old-test-parent-delay.txt`) — 게이트 1회차와 같은 메시지.
- 평소에 통과하던 이유: 부모가 `READY` 직후 수 ms 안에 `Begin` 에 닿아 0.1 s 창보다 빨랐다. 부하가 그 차이를 뒤집는다.

### D2 `TestTheReplacedReaderIsClosed` — 시험이 닫힘보다 먼저 단언한다

- 생산 `strategyRuntimeAttachment.attempt`: `:183` `a.trying = false` → `:210` `a.mu.Unlock()` → `:211` `closeEvictedStrategyReader(evicted)`.
  닫기를 잠금 밖에 두는 것은 의도다(`:204` 주석 · `:222-224` — 진행 중 읽기와 나란히 불러도 안전).
- 시험은 `attempts == 1 && !inFlight()`(205) 까지만 기다리고 곧바로 `closes == 1` 을 단언(208)한다. `!inFlight()` 는 `:183` 뒤 `:210` 에서
  잠금이 풀리면 참이 되므로 `:211` 이 돌기 전 창이 열려 있다.
- 측정(편집 전): 사본에서 `:210` 과 `:211` 사이에 50 ms sleep → 5/5 FAIL 「0번 닫았다」(`analysis/red/tossctl-old-test-prod-sleep.txt`),
  무변이 사본 5/5 PASS(`analysis/red/tossctl-old-test-control.txt`).

## §2 수리

### D1 수리 — 타이머가 걸린 대기

```go
// 타이머 없는 select {} 는 교착 검출로 자식을 죽여 잠금을 풂 — 부모가 죽일 때까지 시한 있는 잠으로 버팀
time.Sleep(time.Minute)
os.Exit(4) // 부모가 끝내 죽이지 않으면 스스로 끝냄 — 고아 자식이 잠금을 오래 쥐지 않게 함
```

- `time.Sleep` 은 런타임 타이머를 걸어 교착 검출 대상이 아니다. 부모는 지금처럼 `cmd.Process.Kill()`(246) 로 끝낸다 — 정상 실행에서 잠은 수 ms 만 돈다.
- 1분은 시험 시한이 아니라 **고아 상한**이다: 부모가 kill 전에 죽으면(시험 프로세스 강제 종료 등) 자식은 1분 뒤 exit 4 로 스스로 끝난다.
  편집 전에는 이 경우에도 교착 검출이 즉시 끝냈으므로, 이 상한이 없으면 고아가 영원히 남는 새 결함이 생긴다.
- 부모 쪽(240-274) 단언 · 순서 · 기한은 그대로다. `time` import 한 줄이 늘어난다.
- 기각: 자식이 stdin EOF 까지 막기(`StdinPipe` 배관이 늘어남 — 같은 효과에 코드가 더 큼), 부모의 `Begin` 재시도(시험이 재는 「잠금 보유 중 Busy」를 흐림).

### D2 수리 — 닫힘 자체를 기다림

```go
a109WaitFor(t, "재부착", func() bool {
	return attempts.Load() == 1 && !attachment.inFlight()
})
// 닫기는 생산 attempt 가 잠금을 푼 뒤 잠금 밖에서 함 — inFlight 가 꺼져도 아직 안 닫혔을 수 있어 닫힘 자체를 기다림
a109WaitFor(t, "밀려난 reader 닫기", func() bool { return evicted.closes.Load() >= 1 })
```

- 기한은 기존 헬퍼의 2 s 그대로(헬퍼 · 기한 불변). 기존 단언 `closes != 1`(208) 은 그대로 둔다.
- 닫기가 아예 없으면 둘째 대기가 2 s 뒤 「밀려난 reader 닫기: 2s 안에 일어나지 않았다」로 FAIL — 사본 양성 대조로 확인한다(§3).
- 기각: 첫 대기 술어에 `closes` 조건을 합치기(실패 문구가 「재부착」 하나로 뭉개져 무엇이 안 일어났는지 잃음).

## §3 검증 계획(원문은 `analysis/`)

| 대상 | 실험 | 기대 |
|---|---|---|
| D1 | 수리 뒤 자식 직접 실행 5 s 관찰 | 살아 있음, deadlock fatal 없음 |
| D1 | 사본 부모 300 ms 지연 + 새 시험 | PASS |
| D1 | `-run TestChildLockAndCrash -count=200` | 200/200 PASS |
| D2 | 무변이 사본 + 새 시험 | PASS(대조군) |
| D2 | 사본 생산 unlock↔close 50 ms + 새 시험 | PASS |
| D2 | 사본 생산 닫기 호출 삭제 + 새 시험 | FAIL(양성 대조 — 시험이 여전히 닫힘을 잰다) |
| D2 | `-run TestTheReplacedReaderIsClosed -count=200` | 200/200 PASS |
| 둘 | 두 패키지 전체 무태그 · `-tags tossos_testseams` 각 1회, `make lint` | 통과 |

## §4 census(같은 모양의 다른 자리)

### 자식 프로세스 갈래의 타이머 없는 `select {}` — 저장소 `*_test.go` 전수

명령: `rg -n --glob '*_test.go' 'select\s*\{\s*\}|<-make\(chan|for\s*\{\s*\}' .`(원문 `analysis/census/blocking-shapes.txt`)

| 자리 | 모양 | 처분 |
|---|---|---|
| `internal/shadowsession/session_linux_test.go:221` | 자식 갈래 · flock 보유 · `select {}` | **이 change 에서 수리(D1)** |
| `openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype/process_test.go:110` `TestKilledLockOwnerReleases` | 자식 갈래 · flock 보유 · `select {}` — 같은 결함 | **고치지 않음 — 다른 활성 change(a166, 병행 팀메이트) 디렉터리.** 그 시험은 kill 뒤 잠금이 풀렸는지만 보므로 자식이 일찍 죽어도 FAIL 하지 않고 「kill 이 잠금을 풀었다」를 증명하지 못한 채 통과한다(거짓 PASS 쪽). Manager 결정 항목 |
| `internal/strategyworker/bounded_test.go:279` `TestCancellationIsNotAFailure` | 같은 프로세스 안 goroutine 의 `<-make(chan struct{})` | 해당 없음 — 자식 프로세스가 아니고 시험 goroutine 은 계속 돈다 |

### `!inFlight()` 류 대기 뒤 곧바로 사후 상태를 단언하는 다른 시험 — 목록만(고치지 않음)

명령: `rg -n -A4 --glob '*_test.go' '!\w+\.inFlight\(\)' cmd/tossctl`(원문 `analysis/census/inflight-waits.txt`). 생산
`strategyRuntimeAttachment.attempt` 가 잠금을 푼 뒤 하는 일은 닫기(`:211`)와 부착 보고(`:212-214` `reportAttached` 로그)다. 아래는 대기 직후 단언이 있는
자리이고, **그 단언이 잠금 밖 효과에 기대는지는 이 change 에서 분석하지 않았다**(a114 · a115 는 다른 attachment 타입의 `attempt`).

- `a109_the_reattachment_does_not_strand_itself_test.go:73` → `StrategyRuntimeAbsent`(잠금 안 상태)
- `a109_the_request_path_never_dials_test.go:254` → `StrategyRuntimeAbsent`(잠금 안 상태)
- `a109_the_request_path_never_dials_test.go:288` → 로그 「읽기가 실패했다」 개수
- `a109_a_cancelled_read_is_not_a_detachment_test.go:177` → 로그 「다시 붙었다」 0 개(부재 단언 — 경합이면 거짓 PASS 쪽)
- `a114_the_console_reattaches_its_lifecycle_test.go:502` · `:511` → 로그 「실패했다」 개수
- `a114_the_console_reattaches_its_lifecycle_test.go:827-833` → 시도 수
- `a115_the_console_strategy_screen_reattaches_test.go:468` → 경고 「연결할 수 없다」 개수
- `a115_the_console_strategy_screen_reattaches_test.go:523` → `a115LastTry`

## §5 spec delta

없음 — `.openspec.yaml` 에 `skip_specs: true`. 근거: (1) 이 change 는 생산 동작 · 요구사항을 바꾸지 않는다. (2) openspec 1.16.0 은 delta 0 인 change 를
strict 에서 거절하고(원문 `analysis/verify/validate-without-marker.txt`) 그 해법으로 `skip_specs: true` 를 지정하며, 스키마 지침은
「Do not invent a requirement just to satisfy validation」이다(`@fission-ai/openspec/schemas/spec-driven/schema.yaml:36-42`).
(3) 저장소 선례(a190 등)는 모두 spec delta 를 가졌지만 전부 요구사항이 실제로 바뀐 change 였다. 저장소 안 `skip_specs` 사용은 이것이 처음이다 — Manager 확인 항목.

## §6 위험 · 되돌리기

- 위험 Low. 시험 파일 두 개만. 되돌리기 = 이 change 커밋 revert.
- 남는 것: D1 의 1분 고아 상한 동안 부모가 사라진 자식은 임시 디렉터리 잠금을 쥔다(그 디렉터리는 시험 전용 `fixture` 의 것).
