package console

// templates.go is the markup. It is one string on purpose: the console has three
// pages and a refusal, and a template directory with four files in it would be
// four more things to keep in step with an embed directive.
//
// The style sheet is inline and deliberately tiny. Anything that had to be
// fetched — a font, a framework, an icon set — would be a request this page
// cannot make (there is nothing to make it to) and a dependency on somebody
// else's server for a screen that authorises live orders.

const pageTemplates = `
{{define "style"}}
:root { color-scheme: light dark; }
* { box-sizing: border-box; }
body {
  margin: 0; padding: 0 1rem 4rem;
  font: 15px/1.6 ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Noto Sans KR", sans-serif;
  background: #fbfbfa; color: #1c1c1a;
}
main, header > div { max-width: 62rem; margin: 0 auto; }
header { border-bottom: 1px solid #dcdcd6; margin-bottom: 1.5rem; }
header > div { display: flex; gap: 1rem; align-items: baseline; padding: 0.9rem 0; }
header strong { font-size: 0.95rem; letter-spacing: 0.02em; }
nav a { margin-right: 0.9rem; color: #3a3a35; text-decoration: none; }
nav a.on { font-weight: 600; text-decoration: underline; }
h1 { font-size: 1.35rem; margin: 1.4rem 0 0.4rem; }
h2 { font-size: 1.05rem; margin: 1.6rem 0 0.4rem; }
section { border: 1px solid #dcdcd6; border-radius: 6px; padding: 0.9rem 1.1rem; margin: 1rem 0; background: #fff; }
pre { overflow-x: auto; padding: 0.8rem; background: #f4f4f0; border-radius: 4px;
      font: 12.5px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
table { border-collapse: collapse; width: 100%; font-size: 0.9rem; }
th, td { text-align: left; padding: 0.3rem 0.6rem 0.3rem 0; border-bottom: 1px solid #ececE6; vertical-align: top; }
dl { display: grid; grid-template-columns: 12rem 1fr; gap: 0.25rem 1rem; margin: 0.4rem 0; }
dt { color: #6a6a62; } dd { margin: 0; }
.notice { border-left: 3px solid #b07000; background: #fff6e5; padding: 0.7rem 0.9rem; border-radius: 4px; }
.danger { border-left: 3px solid #a01818; background: #fdeeee; padding: 0.7rem 0.9rem; border-radius: 4px; }
.ok { color: #1a6b2a; } .bad { color: #a01818; } .muted { color: #6a6a62; }
input[type=text] { font: 15px/1.4 ui-monospace, Menlo, Consolas, monospace; padding: 0.45rem 0.6rem;
                   width: 22rem; max-width: 100%; border: 1px solid #9a9a92; border-radius: 4px; }
button { font: inherit; padding: 0.45rem 1rem; border-radius: 4px; border: 1px solid #6a6a62;
         background: #1c1c1a; color: #fbfbfa; cursor: pointer; }
button.secondary { background: transparent; color: #1c1c1a; }
form { display: inline; }
@media (prefers-color-scheme: dark) {
  body { background: #16161a; color: #e6e6e0; }
  header { border-color: #33333a; } nav a { color: #c8c8c0; }
  section { background: #1d1d22; border-color: #33333a; }
  pre { background: #111116; } th, td { border-color: #2a2a30; }
  .notice { background: #2e2510; } .danger { background: #2e1414; }
  button.secondary { color: #e6e6e0; }
  input[type=text] { background: #111116; color: #e6e6e0; border-color: #44444c; }
}
{{end}}

{{define "head"}}<!doctype html>
<html lang="ko"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="referrer" content="same-origin">
{{if .Refresh}}<meta http-equiv="refresh" content="{{.RefreshSeconds}}">{{end}}
<title>tossctl console</title>
<style>{{template "style"}}</style>
</head><body>
<header><div>
  <strong>tossctl console</strong>
  <nav>
    {{/*
      Two screens, two names. The root is the verification console — soak,
      attestation, binary provenance, the approval window — and /dashboard is the
      overview of the account. Both were labelled 대시보드 until change
      console-operator-overview, which is one name for two different questions.
    */}}
    <a href="/dashboard" {{if eq .Nav "overview"}}class="on"{{end}}>개요</a>
    <a href="/" {{if eq .Nav "verify-console"}}class="on"{{end}}>검증 콘솔</a>
    <a href="/positions" {{if eq .Nav "positions"}}class="on"{{end}}>포지션</a>
    <a href="/orders" {{if eq .Nav "orders"}}class="on"{{end}}>주문</a>
    <a href="/signals" {{if eq .Nav "signals"}}class="on"{{end}}>발굴 신호</a>
    <a href="/history" {{if eq .Nav "history"}}class="on"{{end}}>거래 이력</a>
    <a href="/settings#adoption" {{if eq .Nav "settings"}}class="on"{{end}}>외부 종목 자동관리</a>
    <a href="/optimization" {{if eq .Nav "optimization"}}class="on"{{end}}>최적화</a>
    <a href="/verify" {{if eq .Nav "verify"}}class="on"{{end}}>검증</a>
    <a href="/report" {{if eq .Nav "report"}}class="on"{{end}}>리포트</a>
  </nav>
  <span class="muted" style="margin-left:auto">127.0.0.1 전용 · 승인 외 읽기 전용</span>
</div></header>
<main>
{{end}}

{{define "foot"}}
</main></body></html>
{{end}}

{{/*
  "hours" is the market-hours advisory (task 1.7 ②). It warns and it never
  blocks: the only thing measured is that a live order sent on a closed day comes
  back 422 order-hours-closed, and the window in which the broker DOES accept one
  is unmeasured. A tool that refused to start outside a hard-wired calendar would
  be asserting the half nobody has observed.
*/}}
{{define "hours"}}
{{if .Outside}}
<p class="notice"><strong>지금은 KR 정규장(평일 09:00–15:30 KST) 밖이다.</strong>
{{.Detail}}
<br>그래도 <strong>시작은 막지 않는다</strong> — 주문 접수 창의 실제 경계는 미측정이고, 휴장 달력을
도구에 박아 넣으면 관측하지 않은 것을 단언하게 된다.</p>
{{else}}
<p class="muted">{{.Detail}}</p>
{{end}}
{{end}}

{{define "refuse"}}<!doctype html>
<html lang="ko"><head><meta charset="utf-8"><title>거부됨 — tossctl console</title>
<style>{{template "style"}}</style></head><body><main>
<h1>{{.Title}}</h1>
<p class="danger">{{.Detail}}</p>
</main></body></html>
{{end}}

{{/*
  "restart" is the interstitial the browser holds while this process is replaced.
  It refreshes to a different URL than itself — the same origin, carrying the
  one-time handoff token — so it writes its own meta tag instead of the shared
  two-second one.
*/}}
{{define "restart"}}<!doctype html>
<html lang="ko"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="referrer" content="same-origin">
<meta http-equiv="refresh" content="{{.RefreshContent}}">
<title>재시작 중 — tossctl console</title>
<style>{{template "style"}}</style></head><body><main>
<h1>콘솔을 다시 시작하는 중</h1>
<p class="notice">같은 바이너리를 같은 포트로 다시 실행한다. <strong>{{.Delay}}초 뒤</strong> 이 페이지가
자동으로 새 프로세스로 돌아간다 — 돌아오지 않으면 아래 링크를 누르거나 터미널에 출력된 새 세션 링크를 열어라.</p>
<p><a href="{{.Target}}">지금 새 프로세스로 이동</a></p>
{{if .Note}}<p class="notice">{{.Note}}</p>{{end}}
<p class="muted">재시작이 하는 일은 하나뿐이다: <strong>새 프로세스 인스턴스</strong>를 만든다.
조건주문 존속 판정은 증거 기록의 <code>process.instance_id</code>(기동마다 새로 발급)를 보므로,
이것이 그 측정에 필요한 프로세스 경계다. 게이트·설정·계좌는 아무것도 바뀌지 않고, 배치 승인은
새 계획으로 처음부터 다시 받는다.</p>
</main></body></html>
{{end}}

{{/*
  "staleconsole" / "stalesoak" are the two halves of the stale-binary warning.
  They are advisory and they are quiet when there is nothing to say: a banner
  that is always on screen is a banner nobody reads.
*/}}
{{define "stalebinary"}}
{{if .ConsoleStale}}
<p class="notice"><strong>이 콘솔은 설치된 바이너리보다 오래되었다.</strong>
실행 중 {{.ConsoleAt}} · 설치됨 {{.InstalledAt}} — [콘솔 재시작]을 누르면 새 바이너리로 올라온다.</p>
{{end}}
{{if .SoakStale}}
<p class="notice"><strong>실행 중인 soak은 설치된 바이너리보다 오래되었다.</strong>
마지막 사이클이 보고한 빌드 {{.SoakBuildAt}} · 설치됨 {{.InstalledAt}} —
soak은 다음 사이클 경계에서 스스로 새 바이너리로 재실행한다. 기다릴 수 없으면 [soak 재시작].</p>
{{end}}
{{end}}

{{define "dashboard"}}
{{template "head" .}}
<h1>대시보드</h1>
<p class="muted">이 화면은 로컬 파일만 읽는다. 네트워크 호출도, 설정 변경도 하지 않는다.
계좌 보유·exit 라인은 <a href="/positions">포지션</a>, 완결 거래는 <a href="/history">거래 이력</a>에 있다.</p>
{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}
{{template "stalebinary" .Snap.Binary}}

<section>
  <h2>도구 프로세스</h2>
  <p class="muted">아래 두 버튼은 <strong>도구를 다시 띄우는 것</strong>이다 — 게이트·주문 능력·위험 한도와는
  무관하고, 승인 절차(새 계획을 보고 다시 승인)도 그대로다.</p>
  {{if .CanRestart}}
  <form method="post" action="/restart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <button type="submit" class="secondary">콘솔 재시작</button>
  </form>
  {{else}}
  <p class="muted">콘솔 재시작 배선이 없다 — 종료(Ctrl-C) 후 직접 다시 시작하라.</p>
  {{end}}
  {{if .CanRestartSoak}}
  <form method="post" action="/soak/restart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <button type="submit" class="secondary">soak 재시작</button>
  </form>
  {{else}}
  <p class="muted">soak 재기동 배선이 없다.</p>
  {{end}}
  <p class="muted">콘솔 재시작은 <strong>새 프로세스 인스턴스</strong>를 만든다 — 프로세스당 1회 검증 상한이
  초기화되고, 조건주문 존속 측정에 필요한 프로세스 경계가 생긴다. soak 재시작은 조회 전용 서베이를
  SIGINT 후 다시 띄우는 것이고 자격증명은 건드리지 않는다.</p>
</section>

<section>
  <h2>엔진 런타임</h2>
  {{with .Snap.Engine}}
  {{if .Wired}}
  <dl>
    <dt>상태</dt>
    <dd>{{if .Running}}<span class="ok">실행 중</span>{{if .PID}} (pid {{.PID}}){{end}}{{else}}<span class="muted">정지</span>{{end}}</dd>
    <dt>기동 시각</dt><dd>{{.StartedAtText}}</dd>
    <dt>마지막 갱신</dt><dd>{{.RefreshedAtText}}</dd>
    <dt>실행 중 빌드</dt><dd>{{.BuildAt}}</dd>
    <dt>활성 마커</dt><dd><code>{{.Marker}}</code></dd>
  </dl>
  {{if .Stale}}
  <p class="notice"><strong>실행 중인 엔진은 설치된 바이너리보다 오래되었다.</strong>
  엔진이 보고한 빌드 {{.BuildAt}} — 새 바이너리로 올리려면 [엔진 정지] 뒤 [엔진 시작].</p>
  {{end}}
  {{else}}
  <p class="muted">엔진 상태 배선이 없다 (활성 마커 경로 미지정).</p>
  {{end}}

  {{if .CanStart}}
  <form method="post" action="/engine/start">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit" class="secondary">엔진 시작</button>
  </form>
  {{end}}
  {{if .CanStop}}
  <form method="post" action="/engine/stop">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit" class="secondary">엔진 정지</button>
  </form>
  {{end}}
  {{if not (or .CanStart .CanStop)}}
  <p class="muted">엔진 기동/정지 배선이 없다 — 터미널에서 <code>tossctl engine run</code>으로 직접 띄워라.</p>
  {{end}}

  {{if .Note}}
  <p class="notice">마지막 응답 ({{.NoteAtText}}):</p>
  <pre>{{.Note}}</pre>
  {{end}}

  <p class="muted">이 두 버튼은 <strong>프로세스를 켜고 끄는 것</strong>이다. 엔진이 주문 능력을 갖는지는
  §0.7로 승인된 게이트 설정과 엔진 자신의 기동 인터록이 결정한다 — 콘솔은 게이트를 켜지 않고,
  인터록을 우회하지도 못한다. 인터록이 미충족이면 엔진 프로세스가 거부하고, 그 사유가 위에 그대로 뜬다.
  정지는 SIGTERM(루프 완주·journal 정합 close)이고, 활성 마커는 자문 신호다 —
  배타는 엔진이 쥔 journal 디렉터리 flock이 담당하며 크래시한 엔진은 최대 {{.StaleAfter}} 동안
  실행 중으로 보인다.</p>
  {{end}}
</section>

<section>
  <h2>soak (조회 전용 서베이)</h2>
  {{with .Snap.Soak}}
  {{if .Error}}<p class="danger">{{.Error}}</p>{{end}}
  {{if .Present}}
  <dl>
    <dt>계좌</dt><dd>{{.AccountRef}}</dd>
    <dt>사이클</dt><dd>{{.Cycles}}</dd>
    <dt>연속일</dt><dd>{{.StreakDays}} / {{.MinDays}}일</dd>
    <dt>기간</dt><dd>{{.FirstAt.Format "2006-01-02 15:04Z"}} → {{.LastAt.Format "2006-01-02 15:04Z"}}</dd>
    <dt>attest 가능</dt><dd>{{if .Ready}}<span class="ok">예</span>{{else}}<span class="bad">아니오</span>{{end}}</dd>
  </dl>
  {{if .Reasons}}<ul>{{range .Reasons}}<li>{{.}}</li>{{end}}</ul>{{end}}
  <table>
    <tr><th>날짜</th><th>사이클</th><th>인증 성공</th><th>인증 실패</th><th>판정</th></tr>
    {{range .Days}}
    <tr><td>{{.Date}}</td><td>{{.Cycles}}</td><td>{{.CredentialOK}}</td><td>{{.AuthFailures}}</td>
        <td>{{if .Pass}}<span class="ok">pass</span>{{else}}<span class="bad">fail</span>{{end}}</td></tr>
    {{end}}
  </table>
  {{else}}
  <p class="muted">아직 기록이 없다. <code>tossctl soak run</code>으로 시작하라. ({{.Record}})</p>
  {{end}}
  {{end}}
</section>

<section>
  <h2>capability attestation</h2>
  {{with .Snap.Attestation}}
  {{if .Present}}
  <dl>
    <dt>계좌</dt><dd>{{.AccountRef}}</dd>
    <dt>soak 일수</dt><dd>{{.SoakDays}}</dd>
    <dt>발급 / 만료</dt><dd>{{.IssuedAt.Format "2006-01-02"}} → {{.ExpiresAt.Format "2006-01-02"}}</dd>
    <dt>상태</dt><dd>{{if .Usable}}<span class="ok">유효</span>{{else}}<span class="bad">사용 불가</span>{{end}}</dd>
    <dt>검증된 endpoint</dt><dd>{{range .Endpoints}}{{.}}<br>{{end}}</dd>
  </dl>
  {{else}}
  <p class="muted">attestation 파일이 없다 ({{.Path}}). <code>tossctl soak attest</code>가 조건을 만족할 때 쓴다.</p>
  {{end}}
  {{if .Reasons}}
  <p class="notice">게이트가 열리지 않는 이유:</p>
  <ul>{{range .Reasons}}<li>{{.}}</li>{{end}}</ul>
  {{end}}
  <p class="muted">이 콘솔은 게이트를 켜지 않는다. 자동매매 토글은 사람이 직접 승인한다.</p>
  {{end}}
</section>

<section>
  <h2>실계좌 검증 진행</h2>
  {{with .Snap.Verify}}
  {{if .Error}}<p class="danger">{{.Error}}</p>{{end}}
  <dl>
    <dt>기록</dt><dd>{{.Record}}</dd>
    <dt>계좌</dt><dd>{{if .AccountRef}}{{.AccountRef}}{{else}}(없음){{end}}</dd>
    <dt>완료 / 전체</dt><dd>{{.Done}} / {{.Total}} 단계</dd>
  </dl>
  {{if .Steps}}
  <table>
    <tr><th>단계</th><th>이름</th><th>판정</th><th>사유</th></tr>
    {{range .Steps}}<tr><td><code>{{.Step}}</code></td><td>{{stepLabel .Step}}</td><td>{{verdict .Verdict}}</td><td>{{.Reason}}</td></tr>{{end}}
  </table>
  {{end}}
  {{if .AwaitingRestart}}
  <p class="notice"><code>{{.AwaitingRestart}}</code> ({{stepLabel .AwaitingRestart}}) 단계는 <strong>새 프로세스</strong>를 기다린다 — 위의 [콘솔 재시작] 뒤 이어하기.</p>
  {{end}}
  {{if .Outstanding}}
  <p class="danger">계좌에 아직 살아 있는 객체:</p>
  <ul>{{range .Outstanding}}<li>{{.Kind}} {{.ID}} ({{.Symbol}}){{if .Deliberate}} — 존속 측정을 위해 의도적으로 남긴 것{{end}}</li>{{end}}</ul>
  {{end}}
  {{end}}
  <p><a href="/verify">검증 화면으로</a></p>
</section>
{{template "foot" .}}
{{end}}

{{define "verify"}}
{{template "head" .}}
<h1>실계좌 검증 <span class="muted">{{.Market}} 시장</span></h1>
<p class="muted">검증 능력은 계좌와 <strong>시장</strong>의 속성이다 — 아래 모든 것은 {{.Market}} 기록만 읽는다.
한 시장의 판정은 다른 시장의 단계를 완료로 만들지 않는다.
{{if .Snap.IsUS}}<a href="/verify?market=KR">KR 검증으로</a>{{else}}<a href="/verify?market=US">US 검증으로</a>{{end}}
· <a href="/report?market={{.Market}}">이 시장 리포트</a></p>
{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}

{{if .Spent}}
<section>
  <p class="notice"><strong>이 프로세스는 이미 검증을 수행했다.</strong> 조건주문 존속은 등록한 프로세스가
  죽은 뒤에만 관측할 수 있으므로 이 경계는 우회하지 않는다. 다른 시장을 이어서 측정할 때도 마찬가지다 —
  재시작한 뒤 그 시장 화면에서 시작하라.</p>
  {{if .CanRestart}}
  <p>아래 버튼이 콘솔을 같은 포트로 다시 띄운다 — 돌아오면 [이어하기]를 누르면 된다.</p>
  <form method="post" action="/restart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <button type="submit">콘솔 재시작</button>
  </form>
  {{else}}
  <p>콘솔을 종료(Ctrl-C)하고 다시 시작한 뒤 이어하기를 눌러라.</p>
  {{end}}
</section>
{{end}}

{{if .Run}}{{with .Run}}
<section>
  <h2>실행 {{.ID}} <span class="muted">{{.Market}} 시장 · 시작 {{.StartedAt.Format "15:04:05Z"}}</span></h2>
  {{if .Remeasuring}}
  <p class="notice"><strong>재측정 {{len .Redo}}단계</strong> — {{.RedoList}}. 판정이 pass인 단계는
  건드리지 않는다. 아래 목록을 승인해야만 요청이 나간다.</p>
  {{end}}

  {{if .Awaiting}}
  <p class="danger"><strong>여기서부터 실제 계좌에 요청이 나간다.</strong> 아래 목록이 이 실행이 보낼 수 있는
  전부다. 목록에 없는 요청은 전송되지 않으며, 보내야 하는 상황이 되면 실행이 멈춘다.</p>
  <pre>{{$.Prompt}}</pre>
  <form method="post" action="/verify/approve">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit">위 목록을 승인하고 실행</button>
  </form>
  <form method="post" action="/verify/abort">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit" class="secondary">거부</button>
  </form>
  <p class="muted">승인은 이 화면의 클릭 한 번이다 — 세션 토큰과 CSRF 토큰을 갖춘 이 폼의 제출만 승인이 된다.
  승인 창(5분)이 지나면 아무것도 전송되지 않고 이 실행은 중단된다. 단계별 확인(--confirm-each)은 웹에서는
  제공하지 않는다 — 콘솔은 배치 승인 전용이다.</p>
  {{end}}

  {{if .Approved}}<p class="ok">승인됨 {{.ApprovedAt.Format "15:04:05Z"}} — 목록에 있는 요청만 나간다.</p>{{end}}
  {{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}

  {{if .Lines}}
  <h2>진행</h2>
  <pre>{{range .Lines}}{{.}}
{{end}}</pre>
  {{end}}

  {{if .Done}}
  <h2>결과</h2>
  {{if .Err}}<p class="danger">{{.Err}}</p>{{end}}
  {{if .Summary.Outcomes}}
  <table>
    <tr><th>단계</th><th>이름</th><th>판정</th><th>사유</th></tr>
    {{range .Summary.Outcomes}}<tr><td><code>{{.Step}}</code></td><td>{{stepLabel .Step}}</td><td>{{verdict .Verdict}}</td><td>{{.Reason}}</td></tr>{{end}}
  </table>
  {{end}}
  {{if .Summary.Halt}}<p class="notice">중단: {{.Summary.Halt}}</p>{{end}}
  {{if .Summary.Outstanding}}
  <p class="danger">계좌에 아직 살아 있는 객체:</p>
  <ul>{{range .Summary.Outstanding}}<li>{{.Kind}} {{.ID}} ({{.Symbol}}){{if .Deliberate}} — 의도적으로 남긴 것, 이어하기가 취소한다{{else}} — <strong>취소되지 않았다. 직접 처리하라.</strong>{{end}}</li>{{end}}</ul>
  {{end}}
  {{if .AwaitingRestart}}
  <p class="notice"><strong>여기서 프로세스 경계다.</strong> 조건주문이 이 프로세스보다 오래 사는지가
  다음 측정이므로, 같은 프로세스에서 이어갈 수는 없다.</p>
  {{if $.CanRestart}}
  <form method="post" action="/restart">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <button type="submit">콘솔 재시작 후 이어하기</button>
  </form>
  {{else}}
  <p>콘솔을 종료(Ctrl-C)하고 다시 시작한 뒤 이어하기를 눌러라.</p>
  {{end}}
  {{end}}
  <p><a href="/report">리포트 보기</a></p>
  {{end}}
</section>
{{end}}{{end}}

{{if .ShowStart}}
{{template "hours" .Snap.Session}}

<section>
  {{if .Resuming}}
  <p class="notice">이 기록에는 이미 검증이 있다. 시작하면 <strong>이어서</strong> 진행한다 —
  판정이 끝난 단계는 다시 주문을 내지 않는다.</p>
  {{end}}
  {{if .Snap.Verify.AwaitingRestart}}
  <p class="notice"><code>{{.Snap.Verify.AwaitingRestart}}</code> ({{stepLabel .Snap.Verify.AwaitingRestart}}) 단계가 새 프로세스를 기다리고 있었다. 이 콘솔이 그 새 프로세스다.</p>
  {{end}}
  {{if .Snap.Verify.Cleanup}}
  <p class="notice"><strong>이전 실행이 남긴 객체 {{.Snap.Verify.CleanupCount}}건</strong>을 이 실행이 먼저 취소한다 —
  승인 목록의 첫 줄에 표시된다. 그것이 노출 상한을 채우고 있는 동안에는 아래 단계들이 아무것도 보낼 수 없다.</p>
  {{end}}
  <p>시작하면 이 실행이 보낼 수 있는 <strong>모든 라이브 요청의 목록</strong>이 표시되고,
  그 목록을 승인하는 버튼을 누르기 전에는 아무것도 전송되지 않는다.</p>
  {{$nothingToResume := and .Resuming (not .Snap.Verify.Pending) (not .Snap.Verify.Cleanup)}}
  {{if $nothingToResume}}
  <p class="notice"><strong>이어할 단계가 없다</strong> — 모든 단계에 판정이 있어 이어하기는 아무것도
  측정하지 않는다(2026-07-27에 이 버튼이 두 번 눌려 장중 창이 측정 0건으로 끝났다).
  {{if .Snap.Verify.Redo}}다시 측정하려면 아래 <strong>재측정</strong>을 사용하라.{{end}}</p>
  {{end}}
  <form method="post" action="/verify/start">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <input type="hidden" name="market" value="{{.Market}}">
    <input type="hidden" name="mode" value="resume">
    <button type="submit" {{if or .Spent $nothingToResume}}disabled{{end}}>{{if .Resuming}}이어하기{{else}}검증 시작{{end}}</button>
  </form>
</section>

{{with .Snap.Verify}}
{{if .Redo}}
<section>
  <h2>재측정</h2>
  <p>재측정 대상은 <strong>{{.RedoCount}}개</strong>다 — 마지막 판정이 <code>fail</code>·<code>skipped</code>인
  단계, 그리고 <strong>통과했지만 그 통과가 만든 조건주문이 사라져 아래 단계들이 측정할 수 없게 된 단계</strong>다.
  이어하기는 이 단계들을 건너뛴다(판정이 이미 terminal이므로). 재측정은 이 단계들만 다시 시도한다.</p>
  <table>
    <tr><th>단계</th><th>이름</th><th>마지막 판정</th><th>사유</th></tr>
    {{$redo := .Redo}}
    {{range .Steps}}{{if inRedo $redo .Step}}<tr><td><code>{{.Step}}</code></td><td>{{stepLabel .Step}}</td><td>{{verdict .Verdict}}</td><td>{{.Reason}}</td></tr>{{end}}{{end}}
  </table>
  <p class="muted">대상: {{.RedoList}}</p>
  <p class="notice"><code>deferred</code>·<code>refused</code> 단계는 대상이 아니다. <code>pass</code>도
  대상이 아니다 — 이미 측정된 속성을 위해 실주문을 다시 내지 않는다. 단 하나의 예외는
  <strong>그 통과가 확립한 성질이 더 이상 참이 아닌 경우</strong>다: 조건주문 등록이 남긴 조건주문이
  사라지면 존속·정정·취소는 영원히 <code>skipped</code>가 되므로, 그 전제를 다시 세우는 것은
  아는 것을 다시 재는 것이 아니다. 대상 목록은 폼이 아니라 <strong>증거 기록</strong>에서 계산한다.
  재측정도 계획을 처음부터 다시 만들고 <strong>그 계획을 다시 승인</strong>해야 요청이 나간다.
  설계상 생략되는 단계(보유 0, <code>--include-ttl-edge</code> 옵트인)는 preflight가 다시 걸러 무해하다.</p>
  <form method="post" action="/verify/start">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <input type="hidden" name="market" value="{{$.Market}}">
    <input type="hidden" name="mode" value="redo">
    <button type="submit" {{if $.Spent}}disabled{{end}}>재측정 {{.RedoCount}}단계</button>
  </form>
</section>
{{end}}
{{end}}

<section>
  <h2>단계 목록</h2>
  <p class="muted"><code>tossctl verify run --list</code>와 같은 내용이다. 계좌를 건드리지 않는다.</p>
  <pre>{{.Steps}}</pre>
</section>
{{end}}
{{template "foot" .}}
{{end}}

{{define "report"}}
{{template "head" .}}
<h1>리포트</h1>
{{if .Error}}<p class="danger">{{.Error}}</p>{{end}}
<p><a href="/report.json">JSON 내려받기</a> — 측정된 속성과 미검증 목록 전체.</p>
<section><pre>{{.Text}}</pre></section>
{{template "foot" .}}
{{end}}
`
