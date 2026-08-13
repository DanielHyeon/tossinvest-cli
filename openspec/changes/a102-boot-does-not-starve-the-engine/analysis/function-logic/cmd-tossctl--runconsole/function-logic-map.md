# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go` (212-545)
- AST evidence: `ast.json` — AST 기준 branches **44** / returns 22 / calls 119 / defers 3 / go 1
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `949b8a26cbc612fa4b64f12e0dd5480b6dedd871a451852f1dff24bd20823b31`
- 작성 사유: a102 §3(D7) — soak 자동 시작 블록을 goroutine으로 바꾸고 start seam을
  `awaitEngineReady`로 감쌌다. **기존 함수의 내부를 편집하므로 편집 전에 만들었다.**
  산출물이 automation gate가 읽는 attestation을 갱신하는 프로세스이므로 High-risk다.

> **두 판을 병기한다.** 1판(편집 전)은 `:211-529` · 분기 44 · 이탈 21 · 호출 114 · `go` 0 ·
> SHA `08562e7541dd…`였고, 2판(GREEN 후, 이 문서)은 `:212-545` · 분기 **44**(그대로) ·
> 이탈 **22** · 호출 **119** · `go` **1** · SHA `949b8a26cbc6…`다.
> **분기가 하나도 늘지 않았다** — a101이 세운 규율(판정은 seam 함수에, runConsole은 호출과
> 출력만)을 그대로 지킨 결과다. 커버리지는 두 번 다 **0.0%**다.
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
2판의 `:391-396`(그 블록을 감싼 goroutine).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 가능 | cobra | B1이 `context.Background()`로 대체 |
| `ctx` (signal-derived) | `:219` `signal.NotifyContext` | SIGINT/SIGTERM | 취소되면 서빙이 끝난다 |
| `engineMarkerPath` | **빈 문자열 가능** | `engineJournalDir` 실패 시 (`:283` else) | 빈 경로의 `enginelock.Read`는 `Running=false` — 관대한 reader |
| `soakBoot` seam | nil 가능 | `newConsoleSoakBoot(root)` | B31 — nil이면 load도 nil |
| `bootSurveyIfAbsent` | non-nil | `:373-378` closure | a101의 부팅 seam |
| `soakRecord` | 해석됨 | `resolveSoakRecord` | 실패는 초기 return (B4) |

> **관통 불변식**: 선택 기능의 부재는 **출력 한 줄**이고 기동 실패가 아니다. 엔진 control
> plane에 못 붙어도(`:407`·`:423`), 성과 DB가 없어도, 콘솔은 뜬다.
> **a102가 더한 대기는 이 불변식의 시험이었다** — 대기는 최대 120초이고, 그것이 화면 앞에
> 서면 spec의 SHALL NOT("이 대기는 운영자 콘솔 화면을 지연시켜서는 안 된다")을 바로 깬다.
> 그래서 대기는 **goroutine**이다(`:391`).

## Branches and early returns

44개 분기의 개별 열거는 `branch-test-map.md`에 측정값과 함께 있다. 여기서는 **a102가 닿는
구간**만 적는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B29 (`:339`) | `engineBoot != nil` | `engineBootLoad` 대입 | — | (a101) |
| B30 (`:346`) | `engineBootNote != ""` | stderr 한 줄 — **엔진 자동 시작** | — | (a101) |
| B31 (`:367`) | `soakBoot != nil` | `soakBootLoad` 대입 | — | (a101) |
| **B32 (`:393`)** | **`note := runConfiguredSoakAutostart(...) != ""`** | **stderr 한 줄 — soak 자동 시작. 2판에서는 goroutine 안이다** | — | **a102의 편집 지점** |

**분기는 늘지 않았다.** a102는 B32의 블록을 `go func(){ … }()`(`:391-396`) 안으로 옮기고,
그 안에서 `runConfiguredSoakAutostart`가 받는 start seam을 한 겹 감싼 것으로 바꿨다.
새 판정 넷(준비 확인 / 엔진 없음 / 상한 초과 / 콘솔 종료)은 **전부 새 파일
`cmd/tossctl/engineready.go`로 나갔다** — 여기 썼으면 0.0%짜리 판정이 넷 생겼을 것이다
(a101이 같은 이유로 `bootSurvey`를 만들었다).

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
| `engineJournalDir(root)` `:280` | 엔진 마커 경로 | 실패는 stderr 한 줄, `engineMarkerPath = ""` | ast.json |
| `enginelock.MarkerPath(dir)` `:282` | 마커 파일 위치 | — | ast.json |
| `runConfiguredEngineAutostart` `:342` | 엔진 자동 시작 판정 | 에러 없음. 결과는 사람이 읽는 한 줄 | `engineautostart.go` |
| `bootSurvey(...)` `:374-379` closure | a101의 부팅 seam | `(string, error)` | ast.json |
| **`soakStartAfterEngineReady(ctx, engineReadiness, clock.System(), bootSurveyIfAbsent)` `:392`** | **대기를 start seam 앞에 붙인다** | 에러는 콘솔 종료·배선 부재 둘뿐 | ast.json · **a102가 더한 것** |
| **`runConfiguredSoakAutostart(soakBootLoad, start)` `:393`** | **soak 자동 시작 판정** (a101, 무편집) | 에러 없음 | ast.json |
| `console.ListenAndServe` `:433` | 조립된 Options로 서빙 | 이 함수의 최종 경로 | ast.json |

live binding — 콘솔이 서베이를 세우는 경로는 **둘**이고 둘 다 `restartSoak`을 지난다:
부팅(`bootSurveyIfAbsent`)과 버튼(`startSurvey` → `RestartSoak` closure `:519`).
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
   함수이고, a102의 판정은 그 앞에 붙는다.
4. **버튼 경로(`RestartSoak` closure)는 감싸지 않는다.** 운영자가 방금 누른 것을 120초
   기다리게 하는 것은 버튼이 고장 난 것과 구별되지 않는다.
5. **노트는 어느 쪽이었는지 말해야 한다.** 조용한 상한 초과는 spec이 금지한다(SHALL NOT).

## Safety conclusion

- Safe edit boundary: **B32 블록 하나**를 goroutine으로 감싸고, 그 안에서 start seam을
  한 겹 감싼 것으로 바꿨다. 나머지 43개 분기의 조건·순서·문구는 불변이다. **실제로 그것이었다**
  — 분기 44 그대로, `go` 문 0 → 1.
- High-risk impact: **yes** — 이 배선의 산출물이 automation gate가 읽는 attestation을
  갱신한다. 방향은 보수적이다: 편집이 잘못돼도 최악은 **서베이가 늦게 서는 것**이고,
  가장 나쁜 경우(상한 초과)에도 서베이는 **선다**.
- 물려받은 공백: **44개 분기 전부 count=0**이고 편집 후에도 그대로다. a102는 그것을 바꾸지
  못한다 — 대신 판정을 전부 측정 가능한 함수로 뺐고(넷 다 100.0%), 남는 배선은 **소스 형태
  회귀 테스트**로 고정했다(`branch-test-map.md`의 "측정으로 보장되지 않는 것").
  소스 형태 테스트는 실행 증거가 아니다. **침묵하지 않고 이름을 붙여 남긴다.**
