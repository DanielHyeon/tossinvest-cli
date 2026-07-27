package console

// templates_overview.go is the markup for the operator overview (change
// console-operator-overview).
//
// Same rules as templates.go and templates_portfolio.go: no script, no external
// asset, no build step, no form. The last one is not a coincidence of this
// screen's contents — it is the screen's contract. There is nothing here to
// submit, so there is nothing to confirm, so there is no confirmation friction of
// any kind (D11; 사용자 지시 2026-07-27).
//
// Every panel renders in every state. A value that could not be read prints
// "미측정" with the sentence that says what to do about it, and it prints on the
// row it belongs to rather than making the row disappear: a row that vanishes
// without a label is not an unmeasured marker, it is concealment.

const overviewTemplates = `
{{/*
  "unmeasured" is the one place a missing value is worded. It always carries the
  reason — a bare em dash does not tell an operator whether to wait, to fix the
  credentials, to start the engine, to wire a seam, to open another screen, to
  edit a config file or to go and look at a ledger row, and those are seven
  different afternoons.
*/}}
{{define "unmeasured"}}<span class="muted">미측정</span>{{if .Code}} <code>{{.Code}}</code>{{end}}
<br><span class="muted">{{.Why}}</span>{{end}}

{{define "cell"}}{{if .Known}}{{.Value}}{{else}}{{template "unmeasured" .}}{{end}}{{end}}

{{define "overview"}}
{{template "head" .}}
<h1>개요</h1>
<p class="muted">지금 계좌가 어떤 상태인지를 한 화면에 모은 <strong>읽기 전용</strong> 화면이다.
누를 것이 없다 — 주문·정정·취소는 물론 어떤 설정 변경 폼도 이 화면에는 없다.
<strong>브로커를 호출하지 않는다</strong>: 보유 수치는 <a href="/positions">포지션</a> 화면이 채운 캐시를
갱신 없이 읽기만 한다. 검증 승인은 <a href="/">검증 콘솔</a>에 그대로 있다.
<span class="muted">기준 시각 {{.Snap.NowText}} · {{.RefreshSeconds}}초마다 자동 재로드</span></p>

{{/* 원장 상태는 화면 전체에 한 번. 계좌·오늘·최근 세 패널이 같은 핸들 하나로 읽으므로
     같은 사실을 세 번 적을 이유가 없고, 세 번 적으면 서로 다른 답이 나올 자리가 생긴다. */}}
{{template "journalstate" .Snap.Journal}}

{{/* --- 엔진 --- */}}
<section>
  <h2>엔진</h2>
  {{with .Snap.Engine}}
  {{if .Wired}}
  <dl>
    <dt>상태</dt>
    <dd>{{if .Running}}<span class="ok">실행 중</span>{{if .PID}} (pid {{.PID}}){{end}}{{else}}<span class="muted">정지</span>{{end}}</dd>
    <dt>기동 시각</dt><dd>{{.StartedAtText}}</dd>
    <dt>마지막 갱신</dt><dd>{{.RefreshedAtText}}</dd>
  </dl>
  {{else}}
  <p class="notice"><strong>엔진 상태 배선이 없다</strong> (활성 마커 경로 미지정) — 정지로 읽지 않는다.</p>
  {{end}}
  {{if .Note}}
  <p class="notice">기동 거부 사유를 포함한 마지막 응답 ({{.NoteAtText}}):</p>
  <pre>{{.Note}}</pre>
  {{else}}
  <p class="muted"><strong>기동 거부 사유 없음.</strong> 사유는 이 콘솔 프로세스에서 [엔진 시작]·[엔진 정지]를
  누른 경우에만 남는다 — 방금 연 화면에 사유가 없는 것이 정상이고, 어딘가에 저장된 사유를 찾을 필요는 없다.
  엔진 기동/정지는 <a href="/">검증 콘솔</a>에 있다.</p>
  {{end}}
  {{end}}
</section>

{{/* --- 계좌 --- */}}
<section>
  <h2>계좌 보유 <span class="muted">시장별</span></h2>
  <p class="muted">시장을 가로지르는 합계는 만들지 않는다 — KR 행은 KRW, US 행은 USD이고 둘을 더한 숫자는
  아무 뜻도 없다.</p>
  <table>
    <tr><th>시장</th><th>평가금액</th><th>평가손익</th><th>보유 종목</th><th>엔진 관리</th><th>관리 외</th></tr>
    {{range .Snap.Account.Markets}}
    <tr>
      <td><strong>{{.Market}}</strong><br><span class="muted">{{.Currency}}</span></td>
      {{if .Blocked}}
      <td colspan="5">{{template "unmeasured" .Unreadable}}
        {{if eq .Unreadable.Code "never_fetched"}}<br><a href="/positions">포지션 화면을 열면 이 값이 채워진다</a>{{end}}
      </td>
      {{else if .Empty}}
      <td colspan="5"><span class="muted">보유 없음 — 이 시장에 보유 종목이 하나도 없다(평가액 0이라는 뜻이 아니라
      대상이 없다는 뜻이다).</span></td>
      {{else}}
      <td>{{template "cell" .Value}}</td>
      <td class="{{if .Gain}}ok{{else}}bad{{end}}">{{template "cell" .PnL}}</td>
      <td>{{template "cell" .Symbols}}</td>
      <td>{{template "cell" .Managed}}</td>
      <td>{{template "cell" .Unmanaged}}</td>
      {{end}}
    </tr>
    {{end}}
  </table>
  {{with .Snap.Account.Broker}}
  {{if .Present}}
  <p class="muted">브로커 보유 캐시 <strong>{{.TakenAt}}</strong> ({{.AgeSeconds}}초 전).
  이 화면은 캐시를 <strong>갱신하지 않는다</strong> — 갱신은 <a href="/positions">포지션</a> 화면을 열 때만
  일어난다. 그래서 여기 숫자는 TTL보다 오래됐을 수 있고, 위 시각이 그것을 말한다.</p>
  {{end}}
  {{if .Error}}<p class="danger">마지막 브로커 조회 실패: <code>{{.Error}}</code></p>{{end}}
  {{end}}
  {{if .Snap.VerifyRunning}}
  <p class="notice"><strong>실계좌 검증이 실행 중이다</strong> ({{.Snap.VerifyNote}}). 이 화면은 원래
  브로커를 부르지 않으므로 여기서 보류되는 것은 없다 — 다만 캐시를 채우는 <a href="/positions">포지션</a>
  화면의 갱신은 검증이 끝날 때까지 보류된다.</p>
  {{end}}
</section>

{{/* --- 오늘 --- */}}
<section>
  <h2>오늘 <span class="muted">시장별 현지 자정 기준</span></h2>
  <p class="muted">원장에 <strong>동결된</strong> 실현손익만 쓴다 — 체결에서 재계산하지 않는다.
  경계는 시장마다 다르다: 원장의 청산 시각은 UTC이고 UTC 자정은 KST 09:00, 즉 KR 장 시작 한 시간 뒤에
  떨어진다. 각 행이 어느 자정을 썼는지 아래에 적는다.</p>
  <table>
    <tr><th>시장</th><th>실현손익</th><th>왕복</th><th>승</th><th>패</th></tr>
    {{range .Snap.Today.Markets}}
    <tr>
      <td><strong>{{.Market}}</strong><br><span class="muted">{{.Currency}}</span>
        <br><span class="muted">{{.Boundary}}</span></td>
      {{if .Blocked}}
      <td colspan="4">{{template "unmeasured" .Unreadable}}</td>
      {{else if .NoTrades}}
      <td colspan="4"><span class="muted">거래 없음 — 오늘 이 시장에서 청산된 왕복이 하나도 없다
      (실현손익 0이라는 뜻이 아니다).</span></td>
      {{else}}
      <td class="{{if .Gain}}ok{{else}}bad{{end}}">{{template "cell" .PnL}}</td>
      <td>{{template "cell" .Trips}}</td>
      <td>{{template "cell" .Wins}}</td>
      <td>{{template "cell" .Losses}}</td>
      {{end}}
    </tr>
    {{end}}
  </table>
</section>

{{/* --- 미체결 --- */}}
<section>
  <h2>살아 있는 주문</h2>
  <dl>
    <dt>미체결 건수</dt><dd>{{template "cell" .Snap.Open.Count}}</dd>
  </dl>
  <p class="muted">이 칸은 <strong>0이 아니라 미측정</strong>이다. 주문 조회 seam은 다음 change
  (<code>console-orders-screen</code>)가 배선하며, 그때까지 0을 적어 두면 "미체결 없음"으로 읽힌다 —
  이 화면이 세우는 규칙을 이 화면 자신에게 먼저 적용한 자리다.</p>
</section>

{{/* --- 안전 --- */}}
<section>
  <h2>안전 <span class="muted">검증 상태 · 잔여물 · Guardian 한도</span></h2>
  <table>
    <tr><th>시장</th><th>검증 진행</th><th>기록</th></tr>
    {{range .Snap.Safety.Markets}}
    <tr>
      <td><strong>{{.Market}}</strong></td>
      <td>{{if .Present}}{{.Done}} / {{.Total}} 단계{{else}}<span class="muted">기록 없음</span>{{end}}
        {{if .Error}}<br><span class="bad">{{.Error}}</span>{{end}}</td>
      <td><code>{{if .Record}}{{.Record}}{{else}}(경로 미지정){{end}}</code></td>
    </tr>
    {{end}}
  </table>

  {{$any := false}}
  {{range .Snap.Safety.Markets}}{{if .Leftovers}}{{$any = true}}
  <p class="danger"><strong>{{.Market}} 계좌에 아직 살아 있는 객체:</strong></p>
  <ul>{{range .Outstanding}}<li>{{.Kind}} <code>{{.ID}}</code> ({{.Symbol}}){{if .Deliberate}}
    — 존속 측정을 위해 의도적으로 남긴 것{{else}} — <strong>취소되지 않았다.</strong>{{end}}</li>{{end}}</ul>
  {{end}}{{end}}
  {{if $any}}
  <p class="notice">잔여물은 <strong>다음 검증의 노출 상한을 먼저 채운다</strong> — 그것이 남아 있는 동안
  이어지는 단계들은 아무것도 보낼 수 없다. <a href="/verify">검증 화면</a>에서 이어하기가 먼저 취소한다.</p>
  {{else}}
  <p class="muted">두 시장 모두 잔여물 없음. 이 확인은 <strong>KR과 US 기록을 각각</strong> 읽은 결과다 —
  두 시장의 증거 기록은 별개 파일이고, 한쪽만 읽는 화면은 다른 쪽 잔여물을 감춘다.</p>
  {{end}}

  <h2>Guardian 한도</h2>
  {{with .Snap.Safety.Guardian}}
  <p class="muted">이 화면은 한도를 <strong>읽기만</strong> 한다 — 게이트·한도·kill switch는 콘솔에서
  편집할 수 없다. 게이트가 열려 있는지는 여기서 판정하지 않는다: 그 판정은 엔진 프로세스가 자신의
  기동 인터록으로 내리고, 콘솔은 그 프로세스가 출력한 사유를 위 엔진 패널에 그대로 옮길 뿐이다.</p>
  {{if .Read}}
  <p class="muted">한도 통화 {{if .Currency}}<strong>{{.Currency}}</strong>{{else}}<span class="muted">(미지정)</span>{{end}}.</p>
  {{else if .Wired}}
  <p class="muted">한도 통화 <span class="muted">미측정</span> <code>config_unreadable</code> —
  설정을 읽지 못했으므로 통화도 읽지 못했다. <strong>미지정과는 다르다</strong>: 미지정은 파일이
  말하는 바이고, 이것은 파일을 아무도 읽지 못했다는 뜻이다.</p>
  {{end}}
  {{if .Err}}<p class="danger">한도를 읽지 못했다: <code>{{.Err}}</code></p>{{end}}
  <table>
    <tr><th>한도</th><th>설정값</th></tr>
    {{range .Limits}}<tr><td>{{.Label}}</td><td>{{template "cell" .Value}}</td></tr>{{end}}
  </table>

  <h2>오늘 소진분</h2>
  {{range .Axes}}
  <p><strong>{{.Label}}</strong></p>
  {{if .Structural}}
  <p class="notice">{{.Structural}}</p>
  {{else}}
  <table>
    <tr><th>시장</th><th>소진</th></tr>
    {{range .Rows}}
    <tr><td><strong>{{.Market}}</strong></td>
      <td>{{template "cell" .Consumed}}{{if .Note}}<br><span class="muted">{{.Note}}</span>{{end}}</td></tr>
    {{end}}
  </table>
  {{end}}
  {{end}}
  <p class="muted">산출되는 축은 <strong>하나뿐</strong>이다 — 원장에 동결된 오늘 실현손익 대 일일 손실
  한도. 나머지 축은 기록 자체가 없어서 구조적으로 미측정이며, 보유에서 그럴듯한 숫자를 만들어 여기 적으면
  있지도 않은 보호를 믿게 만든다.</p>
  {{end}}
</section>

{{/* --- 최근 --- */}}
<section>
  <h2>최근 exit 판정</h2>
  {{with .Snap.Recent}}
  {{if .Blocked}}
  <p class="notice">{{template "unmeasured" .Unreadable}}</p>
  {{else if .Events}}
  <table>
    <tr><th>시각</th><th>심볼</th><th>관측가</th><th>워터마크</th><th>기준선</th><th>단계</th><th>발의</th></tr>
    {{range .Events}}
    <tr><td>{{.At}}</td><td><code>{{.Symbol}}</code></td><td>{{.Observed}}</td>
      <td>{{.HighWater}}</td><td>{{.Baseline}}</td><td>{{.Level}}</td><td>{{.Proposal}}</td></tr>
    {{end}}
  </table>
  <p class="muted">전체 스트림은 <a href="/history">거래 이력</a>에 있다.</p>
  {{else}}
  <p class="muted">exit 판정 기록 없음 — 원장은 읽혔고, 판정이 하나도 없다.</p>
  {{end}}
  {{end}}
  <p class="muted"><strong>체결은 이 화면에 없다.</strong> <code>journal.ReadOnly</code>에 체결 접근자가
  없고, 만드는 것은 원장 읽기 표면 변경이라 <code>console-orders-screen</code>의 일이다.</p>
</section>
{{template "foot" .}}
{{end}}
`
