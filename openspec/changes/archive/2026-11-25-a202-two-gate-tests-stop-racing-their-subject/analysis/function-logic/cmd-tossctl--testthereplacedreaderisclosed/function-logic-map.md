# Function Logic Map: `TestTheReplacedReaderIsClosed`

- Source: `cmd/tossctl/a109_the_reattachment_does_not_strand_itself_test.go` (191-212, base `7bc35e5d1`)
- AST evidence: `ast.json`(**편집 뒤** 재추출, revision current, 191-214 — 편집 전 갈래는 아래 「편집 전」 표와 커밋 `bbf0d9dcd` 의 같은 파일)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: Low — 시험 함수. 생산 `strategyRuntimeAttachment.attempt`(`httpapi_strategy_attach.go:180-215`) 는 부르기만 하고 바꾸지 않는다.
- 생산 호출자: 없음(시험).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `attachment` | `a109Attachment` 로 만든 자리, resolve 는 `fresh, true` | 시험 픽스처 | — |
| `evicted` | 자리에 먼저 앉힌 `a109ClosingReader` | 시험 | — |
| 닫힘 시점 | 생산 `attempt` 는 `a.trying = false`(183) → `a.mu.Unlock()`(210) → **그 다음 줄** `closeEvictedStrategyReader(evicted)`(211) — 잠금 밖에서 닫음(의도) | 생산 코드 | 편집 전 시험은 `!inFlight()` 까지만 기다려 210~211 사이 창에서 0 을 봄(`analysis/red/tossctl-old-test-prod-sleep.txt`) |

## Branches and early returns — 편집 전(base `7bc35e5d1`)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (반환) | 198 resolve 클로저 `return fresh, true` | `attempts` +1 | 클로저 반환 | 이 시험 |
| (반환) | 205 대기 술어 클로저 `return attempts == 1 && !inFlight()` — **편집 대상(대기 조건)** | 없음 | bool | 이 시험 |
| B1 | if 208 `closes != 1` | 없음 | `t.Fatalf("자리에서 밀려난 reader 를 %d번 닫았다 …")` — **게이트 2회차 실패 지점** | 이 시험 · `analysis/red/tossctl-old-test-prod-sleep.txt` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a109Attachment` | 자리 픽스처 | — | AST |
| `attachment.attach` · `wake` | 재부착 시도 유발 | wake 는 막히지 않음, 시도는 goroutine | 생산 `:158` · `:166` |
| `a109WaitFor` | 술어가 참이 될 때까지 2 s 기한 폴링 | 기한 넘으면 `t.Fatalf("%s: 2s 안에 일어나지 않았다")` — 바꾸지 않음 | `a109_the_request_path_never_dials_test.go:106` |
| `attachment.inFlight` | single-flight 상태 | 잠금 안에서 `trying` | 생산 `:151` |

## State mutations and fallbacks

- 메모리 안 픽스처만. 네트워크 · 파일 · 운영 설정 없음. 폴백 없음.

## Branches and early returns — 편집 뒤(현재, 191-214)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (반환) | 198 resolve 클로저 `return fresh, true` | `attempts` +1 | 클로저 반환 | 이 시험 |
| (반환) | 205 대기 술어 `attempts == 1 && !inFlight()` | 없음 | bool | 이 시험 |
| (반환) | 208 **새** 대기 술어 `evicted.closes >= 1` — a202 | 없음 | bool, 2 s 안에 거짓이면 `a109WaitFor` 가 「밀려난 reader 닫기: 2s 안에 일어나지 않았다」 FAIL | 사본 양성 대조(닫기 삭제) 5/5 FAIL `../../green/tossctl-new-test-prod-noclose.txt` |
| B1 | if 210 `closes != 1` | 없음 | `t.Fatalf("… %d번 닫았다 …")` — 두 번 닫기를 잡음 | 이 시험 |

반환 좌표(AST, 편집 뒤): 198 · 205 · 208.

## a202 편집 — 적용됨

- 「재부착」 대기 뒤에 `a109WaitFor(t, "밀려난 reader 닫기", …closes >= 1)` 한 줄 + 「왜」 주석 한 줄(design D2). B1 단언 · 헬퍼 · 기한 불변.
- 편집 뒤 실측: 분기 1, 반환 3(예상과 같음).

## Safety conclusion

- Safe edit boundary: 시험의 대기 조건만. 단언 불변, 생산 불변.
- High-risk impact: no — 시험 전용, 생산 diff 0.
