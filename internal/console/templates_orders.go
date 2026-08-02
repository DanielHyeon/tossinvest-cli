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
<p class="muted page-intro">활성 상태와 최근 주문을 한 화면에서 확인한다. 읽기 전용 — 이 화면에서 주문을 내거나 정정·취소할 수 없다.
{{.Snap.NowText}} 기준. 보호 정책은 <a href="/optimization?category=exit-protection">exit 보호 최적화</a>에서 확인한다.</p>

{{template "journalstate" .Snap.Journal}}
{{template "ordercounts" .Screen}}
{{template "orderfilters" .Snap}}
{{template "ordertable" .Snap}}

{{template "foot" .}}
{{end}}

{{/*
  "ordercounts" is the answer to the question the screen exists for.

  합계는 두 목록을 **둘 다 읽었을 때만** 숫자다: 조건주문도 노출 상한을 채우는
  잔여물이고, 한쪽만 세고 낸 합계는 잔여물을 화면에서 지운 확신에 찬 숫자다.

  그리고 **건수와 그 출처는 한 문장이다**(D9). 굵은 숫자를 내고 실패·나이·보류를
  다른 문단에 두면 읽는 사람이 둘을 잇지 않는다 — 세 시간 전 캐시의 "0건"을 지금
  계좌의 사실로 읽는다. 그래서 조회 시각·경과·보류·실패가 같은 섹션의 dl 바로 아래,
  합계와 붙어 있다.
*/}}
{{define "ordercounts"}}
<section class="order-summary">
  <h2>미체결 주문 현황</h2>
  <dl>
    <dt>일반 주문 <span class="muted">OPEN 그룹</span></dt>
    <dd>{{if .OpenLive.Known}}{{.OpenLive.Value}}{{else}}<span class="muted">미측정
      ({{.OpenLive.Code}})</span> — {{.OpenLive.Why}}{{end}}</dd>
    <dt>조건주문</dt>
    <dd>{{if .ConditionalLive.Known}}{{.ConditionalLive.Value}}{{else}}<span class="muted">미측정
      ({{.ConditionalLive.Code}})</span> — {{.ConditionalLive.Why}}{{end}}</dd>
    <dt>합계</dt>
    <dd>{{if .Live.Known}}<strong>{{.Live.Value}}</strong>{{if .Broker.Present}}
      <span class="muted">— {{.Broker.TakenAt}}에 읽은 값 ({{.Broker.AgeSeconds}}초 전)</span>{{end}}
      {{else}}<span class="muted">미측정
      {{if .Live.Code}}({{.Live.Code}}){{end}}</span> — {{.Live.Why}}{{end}}</dd>
  </dl>
  {{template "orderbroker" .}}
  {{/*
    A link and not a <details>: this screen reloads itself, and a native fold
    would open on click and close on the next tick. The one that persists is the
    one the address bar carries (change a055 §6).
  */}}
  <div class="explain-link" data-explain="open-basis">
    <a href="{{$.Explain.Href "open-basis"}}">집계 기준 {{$.Explain.Toggle "open-basis"}}</a>
    {{if $.Explain.Is "open-basis"}}<div><p class="muted">미체결 조회는 브로커의 <code>status=OPEN</code> 그룹이다 — 대기 주문을
    <strong>전량</strong> 돌려주는 호출이므로 이 건수는 하한이 아니라 숫자다. 페이지 경계에 걸려
    잘릴 수 있는 조회로 미체결을 세면 101번째의 살아 있는 주문이 표에서도 집계에서도 사라진다.</p></div>{{end}}
  </div>

  <dl>
    <dt>종결 주문 <span class="muted">CLOSED 그룹 한 페이지</span></dt>
    <dd>{{if .ClosedCount.Known}}{{.ClosedCount.Value}}{{else}}<span class="muted">미측정
      ({{.ClosedCount.Code}})</span> — {{.ClosedCount.Why}}{{end}}</dd>
  </dl>
  <div class="explain-link" data-explain="closed-basis">
    <a href="{{$.Explain.Href "closed-basis"}}">종결 목록 기준 {{$.Explain.Toggle "closed-basis"}}</a>
    {{if $.Explain.Is "closed-basis"}}<div><p class="muted">종결 목록은 따로 부른다. <strong>취소·거부된 주문은 체결이 되지 않으므로
    왕복 기록에 영영 나타나지 않는다</strong> — 거래 이력 화면이 대신하지 못하는 정보다.
    이 목록은 페이지네이션되므로 건수가 "N건 이상"일 수 있다.</p></div>{{end}}
  </div>

  {{if .Truncated}}
  <p class="notice"><strong>페이지가 잘렸다.</strong> 브로커가 다음 페이지가 있다고 답했으므로
  그 건수는 확정된 숫자가 아니라 하한이다.</p>
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
위 값은 {{.Broker.TakenAt}}에 읽은 것이다({{.Broker.AgeSeconds}}초 전).</p>
{{else if .Broker.Present}}
<p class="muted">브로커 조회 {{.Broker.TakenAt}} ({{.Broker.AgeSeconds}}초 전) ·
갱신 1회 = 미체결 1콜 + 종결 1콜 + 조건주문 1콜{{if .Broker.Stale}} ·
<strong class="bad">캐시 TTL({{.Broker.TTLSeconds}}초)을 넘었다 — 지금 계좌의 값이 아니다</strong>{{end}}{{if .Broker.Error}} ·
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
  <nav class="filter-bar" aria-label="주문 필터">
  {{range .Filters}}
  <span class="filter-group"><span class="muted">{{.Label}}</span>
  {{range .Options}}{{if .On}}<strong>{{.Label}}</strong>{{else}}<a href="{{.Href}}">{{.Label}}</a>{{end}}
  {{end}}</span>
  {{end}}
  </nav>
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
  {{if .Rows}}
  <table class="data-table orders-table">
    <caption>주문 목록</caption>
    <thead><tr>
      <th scope="col">시각(UTC)</th><th scope="col">심볼·시장</th><th scope="col">방향·구분</th>
      <th scope="col">상태·체결수량</th><th scope="col">주문수량</th><th scope="col">주문가</th><th scope="col">발주 주체</th>
    </tr></thead>
    <tbody>
    {{range .Rows}}
      <tr class="order-row">
        <td data-label="시각">{{.At}}</td>
        <th scope="row" data-label="종목"><code>{{.Symbol}}</code><span class="submetric">{{.Market}}</span></th>
        <td data-label="방향·구분"><strong>{{.Side}}</strong><span class="submetric">{{if .Conditional}}조건주문{{else}}{{.StateText}}{{end}}</span></td>
        <td data-label="상태·체결"><strong>{{.Status}}</strong><span class="submetric">체결 {{.Filled}}</span>
          {{if .Detail}}<span class="submetric">{{.Detail}}</span>{{end}}
          <details class="row-details"><summary>주문 추적 정보</summary><div class="detail-grid">
            <span>주문번호 <code>{{.ID}}</code></span><span>평균체결가 <strong>{{.AvgPrice}}</strong></span>
            {{if .ExitEvidence}}
            <span>exit 근거 <strong class="status-pill {{if .ExitLine.Fresh}}ok{{else}}bad{{end}}">{{.ExitLine.StatusText}}</strong></span>
            {{if .ExitLine.Reason}}<span class="bad">{{.ExitLine.Reason}}</span>{{end}}
            <span>판단 <strong>{{.ExitLine.ActionText}}</strong> · 단계 <strong>{{.ExitLine.Stage}}</strong></span>
            <span>관측가 <strong>{{.ExitLine.ObservedPrice}}</strong> · 현재 보호선 <strong>{{.ExitLine.CurrentProtection}}</strong></span>
            <span>다음 익절 <strong>{{.ExitLine.NextTarget}}</strong> · 다음 보호선 <strong>{{.ExitLine.NextProtection}}</strong></span>
            <span>예상 수량 <strong>{{.ExitLine.ProjectedQuantity}}</strong> · 평가 시각 {{.ExitLine.EvaluatedAt}}</span>
            <span>decision <code>{{.ExitLine.DecisionID}}</code> · snapshot <code>{{.ExitLine.SnapshotID}}</code></span>
            <span>attempt <code>{{.ExitAttemptID}}</code> · intent <code>{{.ExitIntentID}}</code></span>
            <span>observation <code>{{.ExitLine.ObservationID}}</code> · {{.ExitLine.ObservationSource}}</span>
            <span>정책 {{.ExitLine.Policy}} · 유효 근거 {{.ExitLine.EffectiveSource}}</span>
            {{else}}
            <span class="muted"><strong>근거 미연결</strong> — 명시적 주문 시도 lineage가 없어 심볼·가격·시각으로 추측하지 않는다.</span>
            {{end}}
          </div></details>
        </td>
        <td data-label="수량"><strong>{{.Quantity}}</strong></td>
        <td data-label="가격"><strong>{{.Price}}</strong></td>
        <td data-label="발주 주체">{{.OriginText}}</td>
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
  {{if .AnyConditionalOrigin}}
  <p class="muted"><strong>조건주문의 발주 주체</strong>는 원장이 읽혀도 대개 "불명"이다. 원장의
  주문 시도 기록(<code>mutation_attempts</code>)은 <strong>일반 주문</strong>의 발주·취소·정정만
  남기고, 조건주문 등록은 그 기록을 만들지 않는다 — 물어본 적 없는 질문에 "그 밖"이라고 답하면
  엔진이 낸 조건주문이 수동 주문으로 읽힌다. 기록에 id가 있으면 그때는 "엔진 발주"로 적는다.</p>
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
