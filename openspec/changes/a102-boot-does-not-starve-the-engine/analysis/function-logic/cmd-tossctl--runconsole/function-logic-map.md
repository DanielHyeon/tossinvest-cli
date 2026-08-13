# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go` (211-539)
- AST evidence: `ast.json` — AST 기준 branches **43** / returns 21 / calls 113 / defers 3 / go 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `ee2c242c2804def45784f4ef705cd3b7bbeeff2e2555d1233b90893d8929b0f2`
- 작성 사유: a102 §3(D7) — soak 자동 시작 블록을 goroutine으로 바꾸고 start seam을
  `awaitEngineReady`로 감쌌다. **기존 함수의 내부를 편집하므로 편집 전에 만들었다.**
  산출물이 automation gate가 읽는 attestation을 갱신하는 프로세스이므로 High-risk다.

> **세 판을 병기한다.** 1판(편집 전) `:211-529` · 분기 44 · 이탈 21 · 호출 114 · `go` 0 ·
> SHA `08562e7541dd…`. 2판(`6cd643ca`) `:212-545` · 분기 44 · 이탈 22 · 호출 119 · `go` 1 ·
> SHA `949b8a26cbc6…`. **3판(§3.9, 이 문서)** `:211-539` · 분기 **43** · 이탈 **21** ·
> 호출 **113** · `go` **0** · SHA `ee2c242c2804…`.
>
> **3판이 분기를 하나 줄였다.** 2판은 goroutine 안에 `if note != ""`를 남겨 뒀고, §3.9의
> D7c가 부팅 블록 전체를 `startSoakAutostartAsync`로 옮기면서 그 판정과 `go` 문이 함께
> 나갔다. 편집 전과 비교하면 이 change가 이 함수에 남긴 순변화는 **분기 −1**이다.
> 커버리지는 세 번 다 **0.0%**다.
>
> a101이 남긴 같은 함수의 FLM은 1판과 **같은 revision**(`08562e75…`)을 본다 — a102는 그
> 위에서 시작했다.

## 이 함수가 하는 일

콘솔 프로세스의 **조립 전체**다. 경로를 풀고, 자격증명·기록·attestation의 위치를 정하고,
브로커 seam·엔진 control plane·업데이트 seam을 만들고, 마지막에
`console.ListenAndServe(ctx, console.Options{...})`에 전부 넘긴다. 44개 분기 중 대부분은
**「선택 seam이 nil이면 그 화면만 조회 전용으로 뜬다」**는 열화 판정이고, 나머지는 초기 오류
반환이다.

a102가 닿는 것은 **한 블록**이다 — 1판의 `:378`(`if note := runConfiguredSoakAutostart(...)`),
2판의 `:391-396`(그 블록을 감싼 goroutine), **3판의 `:387-388`(그 블록 전체를 대신하는 호출 하나)**.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 가능 | cobra | B1이 `context.Background()`로 대체 |
| `ctx` (signal-derived) | `:218` `signal.NotifyContext` | SIGINT/SIGTERM | 취소되면 서빙이 끝난다 |
| `engineMarkerPath` | **빈 문자열 가능** | `engineJournalDir` 실패 시 (`:282` else) | 빈 경로의 `enginelock.Read`는 `Running=false` — 관대한 reader |
| `soakBoot` seam | nil 가능 | `newConsoleSoakBoot(root)` | B31 — nil이면 load도 nil |
| `bootSurveyIfAbsent` | non-nil | `:372-377` closure | a101의 부팅 seam |
| `soakRecord` | 해석됨 | `resolveSoakRecord` | 실패는 초기 return (B4) |

> **관통 불변식**: 선택 기능의 부재는 **출력 한 줄**이고 기동 실패가 아니다. 엔진 control
> plane에 못 붙어도(`:401`·`:417`), 성과 DB가 없어도, 콘솔은 뜬다.
> **a102가 더한 대기는 이 불변식의 시험이었다** — 대기는 최대 120초이고, 그것이 화면 앞에
> 서면 spec의 SHALL NOT("이 대기는 운영자 콘솔 화면을 지연시켜서는 안 된다")을 바로 깬다.
> 그래서 대기는 **goroutine**이고, 3판에서는 그 goroutine이 `startSoakAutostartAsync` 안에
> 있다 — 여기서는 보이지 않고, **실행으로 단언된다**.

## Branches and early returns

44개 분기의 개별 열거는 `branch-test-map.md`에 측정값과 함께 있다. 여기서는 **a102가 닿는
구간**만 적는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B29 (`:338`) | `engineBoot != nil` | `engineBootLoad` 대입 | — | (a101) |
| B30 (`:345`) | `engineBootNote != ""` | stderr 한 줄 — **엔진 자동 시작** | — | (a101) |
| B31 (`:366`) | `soakBoot != nil` | `soakBootLoad` 대입 | — | (a101) |

**a102의 편집 지점이었던 B32는 3판에 없다.** 1판의 `if note := runConfiguredSoakAutostart(…)`
블록이 `startSoakAutostartAsync` 호출 하나로 바뀌면서 그 판정이 이 함수를 떠났다.

**분기는 줄었다.** 새 판정 일곱(준비 확인 / 엔진 없음 / 상한 초과 / 콘솔 종료 / 시체 마커 /
spawn 직렬화 / 비동기)은 **전부 새 파일 `cmd/tossctl/engineready.go`로 나갔다** — 여기 썼으면
0.0%짜리 판정이 일곱 생겼을 것이다(a101이 같은 이유로 `bootSurvey`를 만들었다).
§3.9(D7c)는 2판이 여기 남겨 뒀던 출력 판정까지 그쪽으로 보냈다.

**새 함수를 `soakautostart.go`가 아니라 새 파일에 둔 이유** (design D6은 파일을 그렇게
적었다): `soakautostart.go`의 마지막 줄 뒤에 함수를 붙이면 `git diff --unified=0`이
`@@ -174,0 +175,N @@`를 내고, `check_analysis.py`의 `intersects`가 `start <= line <= end+1`
이므로 **무편집인 `rememberSoakApproval`(`:162-174`)이 "수정된 기존 함수"로 잡힌다** (실측).
파일 중간 어디에 넣어도 같은 일이 일어난다. design의 못은 「`runConfiguredSoakAutostart`·
`bootSurvey` 무편집」이고 배정표의 못은 「soakautostart.go(추가만)」인데, **같은 패키지의 새
파일은 그 둘을 더 엄격하게 만족한다** — 그래서 그렇게 했고, 근거를 여기 남긴다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir(root)` `:279` | 엔진 마커 경로 | 실패는 stderr 한 줄, `engineMarkerPath = ""` | ast.json |
| `enginelock.MarkerPath(dir)` `:281` | 마커 파일 위치 | — | ast.json |
| `runConfiguredEngineAutostart` `:341` | 엔진 자동 시작 판정 | 에러 없음. 결과는 사람이 읽는 한 줄 | `engineautostart.go` |
| `bootSurvey(...)` `:373-378` closure | a101의 부팅 seam | `(string, error)` | ast.json |
| **`startSoakAutostartAsync(ctx, engineDir, engineMarkerPath, …)` `:387`** | **부팅 서베이 블록 전체** — 관측 seam·시계·상한·간격·직렬화·비동기가 전부 그 안이다 | **즉시 반환한다** (실행으로 단언) | ast.json · **a102가 더한 것** |
| **`guardedSoakRestart(startSurvey, save)` `:512`** | 버튼의 spawn — 부팅 경로와 같은 게이트를 지난다 | `(string, error)` | ast.json · **§3.9 D7b** |
| `console.ListenAndServe` `:427` | 조립된 Options로 서빙 | 이 함수의 최종 경로 | ast.json |

live binding — 콘솔이 서베이를 세우는 경로는 **둘**이고 둘 다 `restartSoak`을 지난다:
부팅(`bootSurveyIfAbsent`)과 버튼(`startSurvey` → `RestartSoak` closure `:504`).
**§3.9부터 그 둘은 프로세스 안의 뮤텍스 하나(`soakSpawnGate`)를 지난다** — 대기가 비동기가
되면서 둘이 동시에 살아 있을 수 있게 됐고, 둘 다 같은 record 위의 check-then-act이기 때문이다.
**a102는 부팅 쪽만 감쌌다** — 버튼은 사람이 지금 누른 것이므로 기다리게 하지 않는다.
`enginelock.Read`는 이미 이 파일이 `:282`에서 쓰는 것과 같은 관대한 reader다
(새 배관을 만들지 않는다).

## State mutations and fallbacks

- 프로세스 밖 상태를 바꾸는 자리는 **둘**이다: `runConfiguredEngineAutostart`가 엔진
  프로세스를 띄울 수 있고, soak 자동 시작이 서베이 프로세스를 띄운다.
- a102는 **세 번째를 만들지 않았다.** 대기는 마커 파일을 **읽기만** 한다.
- fallback: 상한 초과에도 서베이는 **시작한다**(서베이는 선택 기계장치이고 attestation
  시계를 계속 세워야 한다 — spec의 SHALL). 콘솔 종료만이 "시작하지 않는다"이다.

## 편집이 건드려선 안 되는 것

1. **엔진 autostart 블록(B30)보다 뒤.** 둘은 같은 계좌의 rate budget을 쓴다. a102는 그
   순서를 시간이 아니라 **신호**에 묶을 뿐, 순서를 뒤집지 않는다.
2. **`console.ListenAndServe`가 대기 뒤에 오면 안 된다.** goroutine이 아니면 최대 120초
   동안 운영자 화면이 없다 — a101이 이미 같은 실패를 겪었다(soakStopTimeout 30초 대기).
3. **`runConfiguredSoakAutostart`·`bootSurvey`의 본문은 무편집.** 둘 다 100.0%로 측정되는
   함수이고, a102의 판정은 그 앞에 붙는다. (AST 해시로 확인: 두 함수가 사는
   `soakautostart.go`는 이 change에서 한 글자도 바뀌지 않았다.)
4. **버튼 경로(`RestartSoak` closure)를 대기 뒤에 두지 않는다.** 운영자가 방금 누른 것을 120초
   기다리게 하는 것은 버튼이 고장 난 것과 구별되지 않는다. **다만 §3.9부터 spawn 게이트는
   지난다** — 그것은 대기가 아니라 배타이고, 부팅 경로가 spawn 중인 그 몇 초뿐이다.
5. **노트는 어느 쪽이었는지 말해야 한다.** 조용한 상한 초과는 spec이 금지한다(SHALL NOT).

## Safety conclusion

- Safe edit boundary: **B32 블록 하나**를 호출 하나로 바꾸고, 버튼 closure의 spawn을 게이트
  뒤로 보냈다. 나머지 43개 분기의 조건·순서·문구는 불변이다. **실제로 그것이었다** —
  분기 44 → 43(그 블록의 출력 판정이 나갔다), `go` 문 0 → 0.
- High-risk impact: **yes** — 이 배선의 산출물이 automation gate가 읽는 attestation을
  갱신한다. 방향은 보수적이다: 편집이 잘못돼도 최악은 **서베이가 늦게 서는 것**이고,
  가장 나쁜 경우(상한 초과)에도 서베이는 **선다**.
- 물려받은 공백: **44개 분기 전부 count=0**이고 편집 후에도 그대로다. a102는 그것을 바꾸지
  못한다 — 대신 판정을 전부 측정 가능한 함수로 뺐고(넷 다 100.0%), 남는 배선은 **소스 형태
  회귀 테스트**로 고정했다(`branch-test-map.md`의 "측정으로 보장되지 않는 것").
  소스 형태 테스트는 실행 증거가 아니다. **침묵하지 않고 이름을 붙여 남긴다.**
