# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go` (412-467)
- AST evidence: `ast.json` — branches 7, returns 5, calls 16, assignments 15, defers 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **proposal 단계(구현 전)**. 이 함수의 분기를 근거로 삼는 문서보다 먼저 만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.opts.SLO` | nil 허용 | 주입 | nil이면 양보 판정을 건너뛴다 (B1 단락 평가) |
| `SLO.FillDetectionBehind()` | bool | filldetect Detector 어댑터 | true면 **주기 전체 양보**, outage 시계는 계속 흐른다 |
| `states` (`workingSet`) | 0..N 보유 포지션 | journal | 오류면 `cycle.Err`, 0개면 outage 시계 **리셋** |
| `quotes` (`observe`) | symbol→observedQuote | 브로커 price read | **전 종목 실패만 오류**(`:760-762`). 부분 실패는 오류가 아니다 |
| `o.lastObserved` / `o.outageRaised` | 프로세스 메모리 | 이 함수 | 재시작 시 소실 |

**파일 헤더가 선언한 불변식(32-41행)**

> "we chose not to look" and "we could not look" leave a position equally unprotected.

outage ladder는 `15s → 신규 진입 차단`, `60s → EXIT_OBSERVATION_OUTAGE + ENTRY_BLOCKED + critical`.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:418` | `SLO != nil && FillDetectionBehind()` | `cycle.Deferred=true`, `checkOutage` | `return` `:424` | 양보해도 outage 시계가 흐른다 |
| B2 `:428` | `workingSet` 오류 | `cycle.Err=err` | `return` `:430` | 원장 오류가 사이클 오류로 보고된다 |
| B3 `:432` | `len(states)==0` | `lastObserved=now`, `outageRaised=false` | `return` `:440` | 무보유 계정은 outage로 승격되지 않는다 |
| B4 `:444` | `observe` 오류 (= **전 종목** 미응답) | `cycle.Err=err`, `checkOutage` | `return` `:447` | 전 종목 실패가 사다리를 탄다 |
| B5 `:453` | `range states` | — | (순회) | 보유 포지션마다 정확히 1회 방문 |
| **B6 `:455`** | **`quotes[symbol]` 부재** | **없음** | **`continue` `:459`** | **⚠ 테스트 없음 — 이 change의 대상** |
| B7 `:462` | `judge` 오류 && `cycle.Err==nil` | `cycle.Err=err` | (계속) | 첫 오류만 사이클에 실린다 |

### B6가 이 함수의 결함이다

B1은 헤더의 불변식을 지킨다 — 양보해도 `checkOutage`가 돈다. B4도 지킨다.
**B6만 지키지 않는다.** 시세가 안 온 종목 하나는 `continue`로 사라지고

- 로그 없음 — `reportCycle` `:381-384`는 `cycle.Err != nil`일 때만 쓴다
- 알림 없음
- `exit_events` 행 없음 — `record`에 도달하지 못한다
- outage 시계 무변화 — `checkOutage`는 B1·B4에서만 불린다
- `cycle.Observed` `:451` · `cycle.Judged` `:461`은 계산되지만 **어디에도 기록되지 않는다**

사다리는 **계정 단위**이고 `observe`는 **전 종목** 실패에서만 오류를 낸다(`:760-762`).
한 종목이 무기한 조용히 빠져도 사다리가 반응하지 않는다.

**"we could not look"에 해당하면서 어느 rung도 타지 않는 유일한 경로다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SLO.FillDetectionBehind` `:418` | 체결 감지 SLO 양보 | bool, 오류 없음 | AST |
| `o.checkOutage` `:423,:446` | outage 사다리 2번째 rung | 오류 없음 | AST · **B6에서 미호출** |
| `o.workingSet` `:427` | 보유 포지션 집합 | 오류 → `cycle.Err` | AST |
| `symbolsOf` `:443` | 조회할 심볼 목록 | 순수 | AST |
| `o.observe` `:443` | 브로커 price read | **전 종목 실패만 오류** | AST · `:760-762` |
| `o.judge` `:462` | 포지션별 판정 | 오류 → `cycle.Err`(첫 건만) | AST |

**주의**: `codegraph callees ObserveOnce`는 8개를 돌려주며 **`o.observe`를 포함하지 않는다.**
hard evidence(4단계)만으로는 이 함수의 호출을 전수로 얻을 수 없다 — AST(6단계)가 필요하다.

## State mutations and fallbacks

- `o.observationSequence.Add(1)` `:415` — 단조 증가, 이탈 경로와 무관하게 소비된다
- `o.lastObserved` — B3(무보유)와 `:449`(관측 성공)에서만 전진. **B6는 건드리지 않는다**
- `o.outageRaised` — B3와 `:450`에서 false로 되돌아간다
- fallback 없음. B6의 주석은 "the next cycle may answer"라고 쓰지만 **다음 주기가 답한다는
  보장도, 답하지 않은 주기를 세는 장치도 없다**

## Safety conclusion

- **Safe edit boundary**: B6에 계측을 더하는 것은 판정·발의·제출 어디에도 닿지 않는다.
  B5 순회는 `judge` 호출 **전**이므로 §0.3 손절 즉시성 경로 밖이다.
- **High-risk impact**: **yes** — 이 함수는 손절이 평가되는 유일한 입구다. B6를 지나간
  포지션은 그 주기에 손절이 **평가되지 않는다**.
- **계측의 올바른 자리인 이유**: B5는 보유 포지션을 주기마다 정확히 한 번 본다. 하류 사슬
  (`judgeLadder` 이탈 6 · `record` 5 · `submit` 9)의 어느 이탈로 끝나든 그 순회는 이미
  지나갔다. 하류를 개별 계측하는 설계(a089 초안의 P1~P9)는 이 지점 하나로 대체된다.
