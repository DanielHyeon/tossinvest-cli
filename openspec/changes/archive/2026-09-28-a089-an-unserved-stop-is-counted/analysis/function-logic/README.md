# analysis/function-logic — 범위와 생략 사유

작성 시점: 2026-08-06, **proposal 단계(구현 전)**.
`.claude/CLAUDE.md`「단계 건너뛰기 금지」에 따라, 함수 내부의 분기·early return을
근거로 삼는 문서보다 **먼저** 만든다.

## 포함

| 함수 | 이유 |
|---|---|
| `ExitObserver.ObserveOnce` | 보유 포지션을 주기마다 1회 보는 유일한 지점. B6(`:455`)가 무흔적 |
| `ExitObserver.record` | 지연 시계의 시작(B7)·해제(B8)가 둘 다 여기 |
| `ExitObserver.submit` | 손절이 브로커로 나가는 유일한 함수. 이탈 9개 |
| `ExitObserver.noteDelay` | 시계 본체 |
| `ExitObserver.clearDelay` | 시계 해제 — 유일한 호출자가 `record:1150` |
| `Journal.EnqueueAlert` | outbox 재발 장부 결함의 위치 |

## 생략 (`not-applicable`)

| 함수 | 사유 |
|---|---|
| `ExitObserver.judge` (`:807-840`) | **읽기 전용 문맥.** 내부를 편집하지 않고, 그 분기를 근거로 삼는 주장도 문서에 없다. 범위가 확정되면 필요 시 재생성 |
| `ExitObserver.judgeLadder` (`:916-980`) | 동일. 이탈 6개는 `ObserveOnce` FLM의 하류 열거로만 인용하고 개별 분기를 근거로 쓰지 않는다 |

두 함수를 편집하거나 그 내부 분기를 근거로 삼게 되면 이 생략은 무효이고 산출물을
먼저 만들어야 한다.

## 도구 상태 (5단계)

- CodeGraph hard-evidence 인덱스: **worktree와 일치** (`make sdd-sync` 2026-08-06)
- CodeGraphContext: **300초 타임아웃** — advisory이므로 degraded 진행
- GBrain: `gbrain serve`가 데이터 홈을 점유 중 — advisory이므로 degraded 진행
- 기억 회고: 3키워드 **매칭 0건**

## 확인된 도구 한계

`codegraph callees ObserveOnce`는 8개를 돌려주며 **`o.observe`(`:443`)를 포함하지 않는다.**
4단계(hard evidence)만으로는 함수 내부 호출을 전수로 얻을 수 없다 — 6단계 AST가 필요하다.
