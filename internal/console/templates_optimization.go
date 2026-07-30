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
