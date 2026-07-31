package console

const optimizationTemplates = `
{{define "optimization"}}
{{template "head" .}}
<h1>최적화 · 공통 익절/보호선</h1>
<p class="muted">StockOS의 공통 정책 세 가지를 TossOS의 decimal exit evaluator로 적용한다.
이 화면은 <code>engine.exit_policy.common_policy</code> ID 하나만 저장하며 주문, automation gate,
trading toggle, 편입 목록은 변경하지 않는다.</p>
{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}
{{if .LoadErr}}<p class="danger">설정을 읽을 수 없다: <code>{{.LoadErr}}</code></p>{{end}}

<section id="candidate-filters" aria-labelledby="candidate-filters-title">
  <h2 id="candidate-filters-title">후보 필터</h2>
  <p class="notice"><strong>미승인 · passed 구조적 0 · verdict 비활성</strong></p>
  <p class="muted">수치 threshold의 evidence activation record가 아직 없다. 숫자 0을 기본값으로
  대신하지 않으며 모든 행은 읽기 전용이다. 승인 가능한 registry option이 생기기 전에는 preview와
  apply를 만들지 않는다. 향후 activation도 후보 판정만 바꾸며 <strong>주문·RiskIntent·LIVE 상태는 변경하지 않는다</strong>.</p>
  <nav class="filter-bar" aria-label="후보 필터 시장 전환">
    <a href="#candidate-filters-KR">KR regular</a>
    <a href="#candidate-filters-US">US regular</a>
  </nav>
  {{range .CandidateFilterMarkets}}
  <article id="candidate-filters-{{.Market}}" aria-label="{{.Market}} {{.Session}} 후보 필터">
    <h3>{{.Market}} · {{.Session}}</h3>
    {{range .Filters}}
    <div class="detail-grid" aria-readonly="true">
      <div>
        <strong><code>{{.Key}}</code> · {{.Label}}</strong>
        <p class="muted">{{.Help}}</p>
      </div>
      <dl>
        <dt>default state</dt><dd>미승인 (<code>{{.DefaultState}}</code>)</dd>
        <dt>desired</dt><dd>미승인 — 숫자 값 없음</dd>
        <dt>effective</dt><dd>미승인 — verdict 비활성</dd>
        <dt>단위 / 유효 범위</dt><dd>{{.Unit}} / {{.ValidRange}}</dd>
        <dt>판정 방향</dt><dd>{{.Direction}}</dd>
        <dt>표본 / evidence</dt><dd><code>{{.SampleState}}</code> / <code>{{.EvidenceState}}</code></dd>
        <dt>적용 시점</dt><dd>{{.ApplyTiming}}</dd>
        <dt>누락</dt><dd>{{range .MissingEvidence}}<code>{{.}}</code> {{end}}</dd>
        <dt>preview / CAS</dt><dd>{{.PreviewContract}} — CAS 필수</dd>
        {{if .LegacyValue}}<dt>legacy provenance</dt><dd><code>{{.LegacyValue}}</code> ·
          <code>{{.Provenance}}</code> — desired/effective로 승격하지 않음</dd>{{end}}
      </dl>
    </div>
    {{end}}
  </article>
  {{end}}
</section>

{{if .Current.Rejected}}<p class="danger">현재 설정은 엔진이 거부한다: {{.Current.Rejected}}</p>{{end}}
{{if eq .Current.CommonPolicy ""}}
<p class="notice"><strong>아직 공통 정책을 승인하지 않았다.</strong> 기존 RATCHET 동작이 유지된다.</p>
{{else}}
<p>현재 선택: <strong><code>{{.Current.CommonPolicy}}</code></strong></p>
{{end}}

<form method="post" action="/optimization/exit-policy">
  <input type="hidden" name="csrf" value="{{.CSRF}}">
  {{range .Policies}}
  <section>
    <h2><label><input type="radio" name="common_policy" value="{{.ID}}" {{if .Selected}}checked{{end}}>
      {{.Label}} {{if .Recommended}}<span class="notice">권장</span>{{end}}</label></h2>
    <p><code>{{.ID}}</code></p>
    <table>
      <tr><th>단계</th><th>목표 수익률</th><th>보호선(진입가 대비)</th><th>잔량 기준 익절</th></tr>
      {{range $i, $r := .Ladder.Rungs}}
      <tr><td>T{{add1 $i}}</td><td>{{$r.TargetPct}}%</td><td>{{$r.StopPct}}%</td><td>{{$r.PartialRatio}}</td></tr>
      {{end}}
    </table>
    {{if eq .ID "COMMON_LADDER_HYBRID_50"}}
    <p>100주 기준 T2에서 잔량 25%, T3에서 잔량 1/3을 익절해 <strong>약 50%</strong>를 남긴다.
    T4부터 고정 전량익절 없이 high-water의 6.5% 아래 보호선을 올린다.</p>
    {{else if eq .ID "COMMON_LADDER_RUNNER"}}
    <p>T4의 999%는 고정 전량익절을 만들지 않는 sentinel이다. <strong>외부 매수</strong> 편입분은
    StockOS A168 계약에 따라 자동 부분익절 없이 보호선 승격과 breach 전량보호만 사용한다.</p>
    {{else}}
    <p>T4에서 잔량 전량익절하는 균형형 정책이다.</p>
    {{end}}
  </section>
  {{end}}
  {{if .Wired}}<p><button type="submit">선택한 정책 승인·저장</button></p>
  {{else}}<p class="notice">정책 저장 seam이 배선되지 않아 조회만 가능하다.</p>{{end}}
</form>

<section>
  <h2>적용 범위</h2>
  <p>저장은 <strong>다음 엔진 기동</strong>부터 새로 관리되는 자체 진입과 외부 매수 편입분에 적용된다.
  <strong>기존 포지션</strong>은 exit state에 저장된 정책을 계속 사용하며 자동 변경되지 않는다.
  외부 매수는 편입 관측가를 entry/high-water t0로 사용하므로 과거 수익 때문에 즉시 익절하지 않는다.</p>
  {{if .EngineRunning}}<p class="notice">현재 엔진이 실행 중이다. 반영하려면 사람이 엔진을 재기동해야 한다.</p>{{end}}
</section>
{{template "foot" .}}
{{end}}
`
