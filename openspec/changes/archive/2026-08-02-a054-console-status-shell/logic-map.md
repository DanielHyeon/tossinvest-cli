# Function Logic Map · Branch Test Map — a054

기존 함수 내부 로직을 바꾸기 전에 작성한다(`.claude/CLAUDE.md` 진입 조건 3).
근거는 base commit `b331f664d8584d7ff8de7cbc8cfd4e50a1e5bc7c`의 현재 HEAD다.

## 0. 착수 증거

| 항목 | 값 |
|---|---|
| base commit | `b331f66` = HEAD (병행 커밋 없음) |
| `make sdd-sync` | `all indexes current` (GBrain advisory는 busy — advisory이므로 게이트 아님) |
| 착수 전 테스트 | `go test ./...` → **5840 passed / 78 packages** |

## 1. `"/"` 문자열 전수 분류 (task 1.4)

design §4 표와 대조했다. 세 종류가 섞여 있고 **셋 중 둘은 건드리면 안 된다.**

| 위치 | 종류 | 처리 |
|---|---|---|
| [console.go:715](../../../internal/console/console.go#L715) `mux.HandleFunc("/", …)` | 라우트 | 리다이렉트 핸들러로 교체 |
| [pages.go:98](../../../internal/console/pages.go#L98) `r.URL.Path != "/"` | 경로 가드 | 검증 콘솔의 새 경로로 |
| [restart.go:149](../../../internal/console/restart.go#L149) `restartTarget` | 복귀 경로 | 검증 콘솔 — 재시작 컨트롤이 그 화면에 있다 |
| [restart.go:321](../../../internal/console/restart.go#L321) `redirectDashboard` | 결과 안내 | 검증 콘솔 + `?notice=` 보존 |
| [remote.go:327](../../../internal/console/remote.go#L327) | 로그인 후 착지 | 개요 |
| [templates.go:217](../../../internal/console/templates.go#L217) nav href | 링크 | 검증 콘솔 새 경로 |
| [templates_overview.go:36](../../../internal/console/templates_overview.go#L36), [:63](../../../internal/console/templates_overview.go#L63) | 본문 링크 | 검증 콘솔 새 경로 |
| [remote.go:435](../../../internal/console/remote.go#L435), [:503](../../../internal/console/remote.go#L503), [restart.go:259](../../../internal/console/restart.go#L259) | **쿠키 `Path: "/"`** | **무변경** |
| [console.go:474-476](../../../internal/console/console.go#L474-L476) | public URL 정규화 | **무변경** |

쿠키 스코프를 좁히면 `/positions`에서 세션이 사라진다. Branch Test: T3.5.

## 2. 편집 대상 함수의 Logic Map

### 2.1 `handleDashboard` ([pages.go:97](../../../internal/console/pages.go#L97))

```
현재                                     뒤
─────────────────────────────────────    ─────────────────────────────────────
① path != "/" → 404 refuse               ① path != <검증 콘솔 경로> → 404 refuse
② page := dashboardPage{Nav, CSRF,       ② page := verifyConsolePage{chrome, …}
   Notice, Snap, CanRestart, CanRestartSoak}     Nav = "verify-console" (무변경)
③ run != nil → Run, Refresh =            ③ 무변경
   !Done && !Awaiting
④ render "dashboard"                     ④ render "verify-console"
```

바뀌지 않는 것: `c.snapshot()` 호출 1회, CSRF, 재시작 seam 판정, Refresh 조건식.
**Branch Test**: T2.1②(Refresh 필드 누락 검출), T3.1, T3.6.

| 분기 | 조건 | 테스트 |
|---|---|---|
| B1 | 경로 불일치 | T3.1 — 옛 `/`는 404가 아니라 303 |
| B2 | run 없음 | T3.6 — 컨트롤 존재 |
| B3 | run working | T2.1① — 2초 주기 |
| B4 | run awaiting | T2.1① — 재로드 안 걸림 + T2.6 승인 대기 표시 |

### 2.2 `restartTarget` ([restart.go:145](../../../internal/console/restart.go#L145))

```
현재                                뒤
token == ""  → "/"                  → <검증 콘솔 경로>
token != ""  → "/?handoff=" + esc   → <검증 콘솔 경로> + "?handoff=" + esc
```

토큰 소비는 `session0`이 하고 이 함수는 착지 경로만 정한다. 소비 로직 무변경.
**Branch Test**: T3.2(왕복), T3.3(재사용 거부).

### 2.3 `redirectDashboard` ([restart.go:320](../../../internal/console/restart.go#L320))

```
현재                                          뒤
target := "/"                                 target := <검증 콘솔 경로>
notice != "" → target += "?notice=" + esc     무변경
303                                           무변경
```

이 함수의 호출자는 자기 재시작·soak 재기동 결과다. 그 컨트롤은 검증 콘솔에 있다.
**Branch Test**: T3.3.

### 2.4 `handleRemoteLogin` 계열 ([remote.go:327](../../../internal/console/remote.go#L327))

로그인 성공 후 착지 경로 문자열 하나만 개요로 바꾼다. 세션 발급·감사 기록·쿠키 설정은
무변경. 특히 [remote.go:435](../../../internal/console/remote.go#L435)·[:503](../../../internal/console/remote.go#L503)의 `Path: "/"`는 쿠키 스코프이고 라우트가 아니다.
**Branch Test**: T3.4, T3.5.

### 2.5 각 page struct의 chrome 임베딩 (20개 렌더 지점)

```
현재                              뒤
type xPage struct {               type xPage struct {
    Nav     string                    chrome        // Nav, Refresh, Status
    Refresh bool                      …
    …                             }
}
func (xPage) RefreshSeconds() int   ← 화면이 정의한 것은 그대로 이긴다(승격 가림)
```

**임베딩이 조용히 바꿀 수 있는 것** — 검증했다(scratchpad 실측):

- 승격된 exported 필드는 unexported 임베딩을 거쳐도 `html/template`에서 읽힌다
  (`flagEmbedRO`는 exported 필드로 내려갈 때 전파되지 않는다).
- 바깥 타입의 `RefreshSeconds()`가 승격된 것을 가린다 — depth 0이 이긴다.
- 따라서 **화면이 `RefreshSeconds()`를 실수로 지우면 컴파일도 테스트도 실패하지 않고
  주기가 0이 된다.** 같은 이유로 `Refresh`를 세팅하지 않으면 자동 재로드가 조용히 꺼진다.

**Branch Test**: T2.1을 두 방향으로 — ① 주기 값, ② 걸림/안 걸림. 화면별로 고정한다.

`restartPage`·refuse는 `head`를 쓰지 않는 독립 문서다(templates.go:256, :272). chrome을
넣지 않는다 — 넣으면 쓰지 않는 필드가 늘고, `restartPage`는 `Refresh()`를 **메서드**로
갖고 있어 승격 필드와 이름이 겹친다.

### 2.6 `readEngine` / `SessionAdvisoryFor` / `binstamp.Of` — 호출자 확인 (task 1.3)

| 심볼 | 성격 | 표시줄이 늘리는 비용 |
|---|---|---|
| `enginelock.Read` | 마커 파일 1회 읽기 | 화면당 1회 (이미 `snapshot()`이 하던 것과 동일 경로) |
| `binstamp.Of` | `os.Stat` — **해시 아님** | 화면당 1회 |
| `verifylive.SessionAdvisoryFor` | 순수 계산(요일·시각) | 0 |
| 화면 `RefreshSeconds()` | 상수 또는 TTL 나눗셈 | 0 |

`c.snapshot()` **전체를 부르지 않는다.** soak 파싱·attestation·검증 기록은 표시줄이 쓰지
않고, 2초 주기 화면에 얹을 이유가 없다.
**Branch Test**: T2.8 — TTL당 holdings 1콜 상한 유지.

## 3. 이름 일치의 판정 방식 (task 3.7)

`aria-current`가 붙은 nav 항목의 텍스트와 `<h1>`의 **핵심 텍스트**를 대조한다.
핵심 텍스트 = `<h1>`에서 `<span class="muted">…</span>` 수식어를 제거한 나머지.

현재 HEAD의 불일치 (렌더 실측):

| 경로 | nav 라벨 | 현재 `<h1>` | 판정 |
|---|---|---|---|
| `/` | 검증 콘솔 | 대시보드 | **불일치** |
| `/dashboard` | 개요 | 개요 | 일치 |
| `/positions` | 포지션 | 포지션 | 일치 |
| `/orders` | 주문 | 주문 | 일치 |
| `/signals` | 발굴 신호 | 발굴 신호 | 일치 |
| `/history` | 거래 이력 | 거래 이력 | 일치 |
| `/settings` | 외부 종목 자동관리 | 외부 종목 자동관리 설정 | **불일치**(수식어 미분리) |
| `/optimization` | 최적화 | 전략 최적화 | **불일치** |
| `/performance-history` | 성과 이력 | 레인 성과 · 읽기 전용 | **불일치** |
| `/position-management` | 포지션 정책 | 포지션 관리 | **불일치** |
| `/verify` | 검증 | 실계좌 검증 KR 시장 | **불일치** |
| `/report` | 리포트 | 리포트 | 일치 |

nav 라벨은 `a055`이 소유하므로 이 change는 **`<h1>` 쪽을 고친다**. 수식어는 버리지 않고
`<span class="muted">`로 옮긴다 — "실계좌"는 이 화면이 무엇을 하는지 말하는 단어다.

nav 항목이 없는 화면(`/strategy-runtime`, `/strategy-runtime/market-schedule`, preview 화면)은
`aria-current`가 없으므로 이 검사의 대상이 아니다.
