package console

// templates_orders.go is the orders screen's markup (change
// console-orders-screen, §5). It joins the same template set as everything else,
// so it shares "head", "foot", "journalstate" and the one stylesheet.
//
// There is no <form> in this file and no button. That is asserted rather than
// promised: orders_test.go renders the page and fails on a form, on an input, on
// a POST target and on any word that would be a confirmation prompt.

const ordersTemplates = `
{{define "orders"}}{{template "head" .}}
<h1>주문</h1>
<p class="muted">계좌의 주문 기록이다. 읽기 전용 — 이 화면에서 주문을 내거나 정정·취소할 수 없다.
{{.Snap.NowText}} 기준.</p>

{{template "journalstate" .Snap.Journal}}
{{template "ordercounts" .Snap}}
{{template "orderbroker" .Snap}}
{{template "orderfilters" .Snap}}
{{template "ordertable" .Snap}}

{{template "foot" .}}
{{end}}

{{/*
  "ordercounts" is the answer to the question the screen exists for. 합계는 두
  목록을 **둘 다 읽었을 때만** 숫자다: 조건주문도 노출 상한을 채우는 잔여물이고,
  한쪽만 세고 낸 합계는 잔여물을 화면에서 지운 확신에 찬 숫자다.
*/}}
{{define "ordercounts"}}
<section>
  <h2>미체결</h2>
  <dl>
    <dt>일반 주문</dt>
    <dd>{{if .PlainLive.Known}}{{.PlainLive.Value}}{{else}}<span class="muted">미측정
      ({{.PlainLive.Code}})</span> — {{.PlainLive.Why}}{{end}}</dd>
    <dt>조건주문</dt>
    <dd>{{if .ConditionalLive.Known}}{{.ConditionalLive.Value}}{{else}}<span class="muted">미측정
      ({{.ConditionalLive.Code}})</span> — {{.ConditionalLive.Why}}{{end}}</dd>
    <dt>합계</dt>
    <dd>{{if .Live.Known}}<strong>{{.Live.Value}}</strong>{{else}}<span class="muted">미측정
      {{if .Live.Code}}({{.Live.Code}}){{end}}</span> — {{.Live.Why}}{{end}}</dd>
  </dl>
  {{if .Truncated}}
  <p class="notice"><strong>페이지가 잘렸다.</strong> 브로커가 다음 페이지가 있다고 답했으므로
  건수는 확정된 숫자가 아니라 하한이다. 잘린 뒤쪽에 잔여물이 있을 수 있다.</p>
  {{end}}
</section>
{{end}}

{{/*
  "orderbroker" is where the numbers came from and how old they are. 갱신 보류와
  콜드 캐시는 다른 상태이고, 화면이 그 둘을 같은 말로 적으면 운영자는 기다릴지
  고칠지 알 수 없다.
*/}}
{{define "orderbroker"}}
{{if not .Broker.Wired}}
<p class="notice"><strong>주문 조회가 배선되지 않았다.</strong> 이 빌드의 콘솔에 주문 조회 seam이
연결되어 있지 않다 — 건수는 0이 아니라 미측정이다.</p>
{{else if and .Broker.Held (not .Broker.Present)}}
<p class="notice"><strong>검증 중 — 데이터 없음.</strong> {{.Broker.HeldReason}}.
검증이 rate limit을 쓰고 있어 이 화면은 브로커를 부르지 않는다.</p>
{{else if .Broker.Held}}
<p class="notice"><strong>검증 중 — 갱신 보류.</strong> {{.Broker.HeldReason}}.
아래 값은 {{.Broker.TakenAt}}에 읽은 것이다({{.Broker.AgeSeconds}}초 전).</p>
{{else if .Broker.Present}}
<p class="muted">브로커 조회 {{.Broker.TakenAt}} ({{.Broker.AgeSeconds}}초 전) ·
갱신 1회 = 주문 조회 1콜 + 조건주문 조회 1콜{{if .Broker.Error}} ·
<span class="bad">마지막 갱신 실패: {{.Broker.Error}}</span>{{end}}</p>
{{else}}
<p class="notice"><strong>아직 읽지 않았다.</strong>
{{if .Broker.Error}}<code>{{.Broker.Error}}</code>{{end}}</p>
{{end}}
{{end}}

{{/*
  "orderfilters" — 링크와 쿼리 파라미터뿐이다(JS 없음). 목록이 미측정이면 필터를
  작동시키지 않는다: "0/—건"은 "0건이 일치"로 읽힌다.
*/}}
{{define "orderfilters"}}
<section>
  <h2>필터</h2>
  {{if .Enabled}}
  {{range .Filters}}
  <p><span class="muted">{{.Label}}</span>
  {{range .Options}}{{if .On}}<strong>{{.Label}}</strong>{{else}}<a href="{{.Href}}">{{.Label}}</a>{{end}}
  {{end}}</p>
  {{end}}
  <p class="muted">{{.Selected.SummaryText}} — <strong>{{.Shown}}</strong>/{{.Total}}건{{if .Truncated}}
  이상{{end}} 표시</p>
  {{else}}
  <p class="notice"><strong>필터를 쓸 수 없다.</strong> 목록을 가져오지 못했으므로 걸러 낼 것이 없다.
  건수를 표시하지 않는 이유도 같다 — 0건은 "일치하는 것이 없다"로 읽히는데, 여기서는 아무것도
  읽지 못한 것이다.</p>
  {{end}}
</section>
{{end}}

{{define "ordertable"}}
<section>
  <h2>주문 목록</h2>
  {{if .Rows}}
  <table>
    <thead><tr>
      <th>시각(UTC)</th><th>심볼</th><th>시장</th><th>방향</th><th>구분</th><th>상태</th>
      <th>주문수량</th><th>체결수량</th><th>주문가</th><th>평균체결가</th><th>주문번호</th><th>발주 주체</th>
    </tr></thead>
    <tbody>
    {{range .Rows}}
      <tr>
        <td>{{.At}}</td>
        <td>{{.Symbol}}</td>
        <td>{{.Market}}</td>
        <td>{{.Side}}</td>
        <td>{{.StateText}}{{if .Conditional}} <span class="muted">조건주문</span>{{end}}</td>
        <td>{{.Status}}{{if .Detail}}<br><span class="muted">{{.Detail}}</span>{{end}}</td>
        <td>{{.Quantity}}</td>
        <td>{{.Filled}}</td>
        <td>{{.Price}}</td>
        <td>{{.AvgPrice}}</td>
        <td><code>{{.ID}}</code></td>
        <td>{{.OriginText}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
  {{else if .Enabled}}
  <p class="muted">{{if .Filtered}}이 필터에 일치하는 주문이 없다 (전체 {{.Total}}건).
  {{else}}주문이 없다 — 브로커가 빈 목록으로 답했다.{{end}}</p>
  {{else}}
  <p class="muted">목록을 가져오지 못했다. 위의 사유를 보라 — "주문 없음"이 아니다.</p>
  {{end}}

  {{if not .Journal.Readable}}
  <p class="notice"><strong>발주 주체는 전부 "불명"이다.</strong> 원장을 읽지 못했으므로 어느 주문이
  엔진이 낸 것인지 판정할 수 없다. "그 밖"으로 적으면 엔진이 아무 일도 안 한 것처럼 보이므로
  그렇게 적지 않는다. 사유는 위의 원장 안내에 있다.</p>
  {{end}}
  {{if .AnyUnresolved}}
  <p class="muted">"상태 불명"은 브로커가 보낸 조합을 이 빌드가 설명할 수 없다는 뜻이다 —
  종결로 읽지 않는다(fail-closed).</p>
  {{end}}
  {{if .AnyUnparsedTime}}
  <p class="muted">시각이 RFC3339로 해석되지 않은 행은 브로커의 문자열을 그대로 적었다.
  버리면 시각이 없는 주문처럼 보인다.</p>
  {{end}}
  <p class="muted">주문을 내거나 정정·취소하는 수단은 이 화면에 없다. 취소가 필요하면
  <code>tossctl</code>에서 한다.</p>
</section>
{{end}}
`
