# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (183-301)
- AST evidence: `ast.json` — AST 기준 branches **18** / returns 14 / calls 53 / defers 9
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `8ad1cc88b9e0a0181c0a1c50a1fbcbdad4a6f85a6786e316c527ba5565274110`
- 작성 사유: a102 §3(D4·D5) — 6단계의 `enginelock.Hold` 반환이 핸들로 바뀌고, 7단계의
  `engineRuntimeFactory` 호출이 ready seam을 하나 더 넘긴다. **기존 함수의 내부를 편집하므로
  편집 전에 만든다.** 주문·손절 루프의 기동 경로이므로 High-risk다.

> **두 판을 병기한다.** 1판(편집 전)은 `:183-296` · 분기 18 · 이탈 14 · 호출 51 ·
> SHA `f13e36b35e08…` · 블록 37개 중 21개 실행(526건 통과)였고, 2판(GREEN 후, 이 문서)은
> `:183-301` · 분기 **18**(그대로) · 이탈 14(그대로) · 호출 **53** · SHA `8ad1cc88b9e0…` ·
> 블록 **39개 중 22개 실행**(550건 통과)다. **분기와 이탈이 하나도 늘지 않았다** —
> 늘어난 것은 호출 둘(`marker.Release`·`marker.Ready`)뿐이다.
> a098의 FLM은 다른 base다 — 이것이 a102의 재기준이다.

## 이 함수가 하는 일

`tossctl engine run`의 **기동 순서 그 자체**다. 파일 머리 주석이 7단계로 못박아 뒀고,
순서가 곧 요구사항이다(engine-safety 「엔진 런타임 수명주기」).

1. flock (`:196`) → 2. 조립 (`:206`) → 3·4. 게이트 거절 (`:207-221`) →
5. verify runlock (`:229-235`) → **6. 자문 마커 (`:238-248`)** → **7. 루프 (`:250-300`)**

a102가 닿는 것은 **6과 7의 경계**다: 6이 만든 핸들의 `Ready`를 7이 만드는 recovery 클로저에
넘긴다. 1~5의 거절 순서는 건드리지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 가능 | cobra | B1이 `context.Background()`로 대체 |
| `root.configDir` | 비었거나 디렉터리 | CLI 플래그 | `engineJournalDir` 실패는 B2 즉시 return |
| 저널 디렉터리의 flock | 남이 쥐었을 수 있음 | 커널 | B3 — `ErrAlreadyRunning` 즉시 return, **조립 전에** |
| `ectx.Automation.Verified` | bool | 조립된 프로필 | B7 — `errEngineGateOff` |
| verify runlock 신선도 | — | `runlock.Fresh` | B9 — `errVerifyInProgress` |
| 마커 쓰기 | 실패 가능 | `enginelock.Hold` | **B10 — 거절이 아니다.** stderr 한 줄 |
| control-plane 서버 5개 | 실패 가능 | `engine.Start*` | B13~B18 — 각각 즉시 return |

> **관통 불변식 둘.**
> (1) **거절은 루프가 하나라도 서기 전에 끝난다.** B3·B4·B7·B9는 전부 `rt.Run` 앞이다.
> (2) **마커는 거절 사유가 아니다** (B10/B11). 배타는 이미 flock이 쥐었고, 못 쓴 마커의
> 대가는 콘솔이 못 그리는 상태 한 줄이다. a102는 이 비대칭을 **그대로 둔다** — ready 신호를
> 못 쓰는 것이 엔진을 세우지 않을 이유가 되면, 겹2가 겹1보다 위험해진다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:185` | `ctx == nil` | `ctx = context.Background()` | — |
| B2 | `:191` | `engineJournalDir` 오류 | — | `:192` return err |
| B3 | `:197` | `enginelock.Acquire` 오류 | — | `:198` return err (**조립 전**) |
| B4 | `:207` | `engineAssemble` 오류 | — | `:213`/`:215` |
| B5 | `:208` | `UnmetInterlockClauses(err) != nil` | stderr 헤더 | `:213` `errEngineInterlockUnmet` |
| B6 | `:210` | `range clauses` | 절마다 stderr 한 줄 | — |
| B7 | `:219` | `!ectx.Automation.Verified` | — | `:220` `errEngineGateOff` |
| B8 | `:229` | `engineVerifyLockPath` 성공 | — | — |
| B9 | `:230` | verify runlock 신선 | stderr 한 줄 | `:233` `errVerifyInProgress` |
| B10 | `:241` | `merr != nil` (마커 못 씀) | stderr note | **return 없음** |
| B11 | `:245` | else | stdout `active marker …` | — |
| B12 | `:257` | `engineRuntimeFactory` 오류 | — | `:258` |
| B13 | `:261` | `NewPositionPolicyCommandService` 오류 | — | `:262` |
| B14 | `:265` | `StartPositionPolicyCommandServer` 오류 | — | `:266` |
| B15 | `:270` | `StartPositionPolicyRuntimeServer` 오류 | — | `:271` |
| B16 | `:275` | `strategyprojectionrpc.Start` 오류 | — | `:276` |
| B17 | `:283` | `ectx.AlertOperations` 오류 | — | `:284` |
| B18 | `:287` | `StartAlertControlServer` 오류 | — | `:288` |

정상 이탈: `:300` `return rt.Run(runCtx)`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginelock.Acquire(dir)` `:196` | 배타 | `ErrAlreadyRunning`은 대기가 아니라 거절 | ast.json |
| `engineAssemble` `:206` | 프로필 조립 (seam 변수) | 인터록 미충족은 절 열거 후 거절 | `engine.go:303` |
| `runlock.Fresh` `:230` | 검증 진행 확인 | 파일 mtime | ast.json |
| **`enginelock.Hold(ctx, markerPath, clk.Now())` `:239`** | **6단계 마커** | 오류는 거절이 아니다 | ast.json · **a102 D4의 편집 지점** |
| **`engineRuntimeFactory(ctx, ectx, clk, logger, ready)` `:256`** | **7단계 루프 조립** (seam 변수) | 오류는 즉시 return | `engine.go:354` · **a102 D5의 편집 지점** |
| `watchStopSignals(cancel, out)` `:297` | 2회 시그널 규율 | — | ast.json |
| `rt.Run(runCtx)` `:300` | 감독 실행 | 이 함수의 최종 return | ast.json |

live binding — `defer` 9개가 역순 해제를 만든다: lock → ectx.Close → **marker.Release** →
policyControl → policyRuntime → strategyRuntime → alertControl → cancel → stopWatching.
**마커 해제가 control plane보다 먼저**이므로, 정상 종료에서 콘솔은 control plane이 닫히기
전에 "엔진 없음"을 본다 — 편집 전과 동일하게 유지한다.

## State mutations and fallbacks

- 프로세스 밖 상태: flock 파일, 마커 파일, control-plane 소켓 4개, 저널(조립이 연다).
- fallback은 **정확히 하나**다 — B10의 마커 실패. 나머지 모든 실패는 즉시 return이다.
- a102는 fallback을 **추가하지 않는다.** ready 마킹 실패도 같은 자리에서 침묵한다
  (`Ready`는 값을 돌려주지 않는다 — 마커의 어떤 실패도 루프를 세우지 않는다는 규율).

## Safety conclusion

- Safe edit boundary: `:239-240`(핸들 수신 + `defer marker.Release()`)과 `:256`(ready seam
  인자 추가) **두 줄**. **실제로 그 두 줄이었다.**
  1~5단계의 조건·순서·문구는 불변이다. **거절의 개수와 순서가 바뀌면 안 된다.**
- High-risk impact: **yes** — 손절을 두는 루프의 기동 경로다. 다만 a102의 방향은 보수적이다:
  ready 신호는 **추가 정보**일 뿐이고, 그것을 못 써도 오늘과 같은 엔진이 뜬다.
- 물려받은 공백: B1·B2·B10·B13~B18의 본문이 편집 전·후 모두 count=0이다(`branch-test-map.md`).
  **B10이 특히 문제다** — a102가 그 줄 바로 위를 편집하는데 그 분기는 미측정이다.
  a102는 B10을 새로 메우지 않는다(마커 쓰기 실패를 만드는 것은 이 change의 범위가 아니다).
  대신 **핸들의 no-op 계약을 `internal/enginelock` 쪽에서** 고정한다 — 실패한 Hold의
  `Ready()`·`Release()`가 아무 일도 하지 않는다는 테스트가 그것이다.
