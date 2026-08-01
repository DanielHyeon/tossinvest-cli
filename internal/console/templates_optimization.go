package console

const optimizationTemplates = `
{{define "optimization"}}
{{template "head" .}}
<h1>최적화 · 공통 익절/보호선</h1>
<p class="muted">StockOS의 공통 정책 세 가지를 TossOS의 decimal exit evaluator로 적용한다.
이 화면은 <code>engine.exit_policy.common_policy</code> ID 하나만 저장하며 주문, automation gate,
trading toggle, 편입 목록은 변경하지 않는다.</p>
<p><a href="/strategy-runtime/market-schedule">strategy-runtime &gt; 시장·일정 보기</a></p>
<nav class="filter-bar" aria-label="최적화 화면">
  <a href="#candidate-filters">상태·안전</a><a href="#exit-protection">청산/보호</a><a href="#application-scope">성과·근거</a>
</nav>
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

<section id="exit-protection" aria-labelledby="exit-protection-title">
  <h2 id="exit-protection-title">청산/보호 · 브로커 상주 보호</h2>
  <p class="muted">현재 공통 보호선과 브로커에 남아 있는 보호주문을 함께 읽는다. 이 화면에는 종목·가격·수량·사유 입력란과 LIVE 전체 활성화가 없다.</p>
  {{if .ProtectionLoadErr}}<p class="danger" role="alert">보호 상태를 읽지 못했다: {{.ProtectionLoadErr}}</p>{{end}}
  {{if not .ProtectionWired}}
  <div class="detail-grid" aria-readonly="true">
    <div><strong>브로커 보호주문</strong><p class="muted">엔진이 꺼져도 손절이 남는 기능이다.</p></div>
    <dl><dt>Capability</dt><dd>지원 확인 전 사용 불가</dd><dt>Activation</dt><dd>OFF</dd>
      <dt>Desired</dt><dd>OFF</dd><dt>Effective</dt><dd>UNWIRED</dd><dt>적용 시점</dt><dd>운영자 별도 승인 후 다음 엔진 기동</dd>
      <dt>Provenance</dt><dd>signed attestation 및 engine command seam 없음</dd></dl>
  </div>
  {{end}}
  {{range .Protections}}
  <article class="detail-grid" aria-label="{{.Symbol}} 브로커 보호 상태">
    <div><strong>{{.Symbol}}</strong><p class="muted">{{.Explanation}}</p></div>
    <dl><dt>Capability</dt><dd>{{.Capability}}</dd><dt>Activation</dt><dd>{{.Activation}}</dd>
      <dt>Desired / Effective</dt><dd>{{.Desired}} / <strong>{{.Effective}}</strong></dd>
      <dt>Effective trigger</dt><dd>{{.EffectiveTrigger}}</dd><dt>보호 수량</dt><dd>{{.ProtectedQuantity}}</dd>
      <dt>Broker ID</dt><dd><code>{{.BrokerID}}</code></dd><dt>Updated at</dt><dd>{{.UpdatedAt}}</dd>
      <dt>Reconcile</dt><dd>{{.ReconcileReason}}</dd><dt>적용 시점</dt><dd>{{.ApplyTiming}}</dd>
      <dt>Provenance</dt><dd>{{.Provenance}}</dd></dl>
    {{if .WeakeningActionToken}}
    <form method="post" action="/optimization/exit-protection/preview">
      <input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="action_token" value="{{.WeakeningActionToken}}">
      <button type="submit" class="secondary">{{.WeakeningAction}}</button>
    </form>
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

<section id="application-scope">
  <h2>적용 범위</h2>
  <p>저장은 <strong>다음 엔진 기동</strong>부터 새로 관리되는 자체 진입과 외부 매수 편입분에 적용된다.
  <strong>기존 포지션</strong>은 exit state에 저장된 정책을 계속 사용하며 자동 변경되지 않는다.
  외부 매수는 편입 관측가를 entry/high-water t0로 사용하므로 과거 수익 때문에 즉시 익절하지 않는다.</p>
  {{if .EngineRunning}}<p class="notice">현재 엔진이 실행 중이다. 반영하려면 사람이 엔진을 재기동해야 한다.</p>{{end}}
</section>
{{template "foot" .}}
{{end}}

{{define "protection-preview"}}
{{template "head" .}}
<p><a href="/optimization?category=exit-protection">최적화 · 청산/보호</a></p>
<h1>보호 약화 확인</h1>
<p class="danger" role="alert"><strong>브로커 보호가 약해질 수 있다.</strong> preview 뒤 최소 3초 동안 영향 범위를 확인한다.</p>
<dl><dt>종목</dt><dd>{{.Preview.Symbol}}</dd><dt>Before</dt><dd>{{.Preview.Before}}</dd><dt>After</dt><dd>{{.Preview.After}}</dd>
  <dt>영향 포지션</dt><dd>{{.Preview.AffectedPositions}}</dd><dt>영향 수량</dt><dd>{{.Preview.AffectedQuantity}}</dd>
  <dt>보호 공백 가능성</dt><dd>{{.Preview.CoverageGap}}</dd><dt>적용 시점</dt><dd>{{.Preview.ApplyTiming}}</dd></dl>
<form method="post" action="/optimization/exit-protection/apply">
  <input type="hidden" name="csrf" value="{{.CSRF}}"><input type="hidden" name="capability" value="{{.Preview.Capability}}">
  <label><input type="checkbox" name="confirm" value="yes" required> 위 before/after와 보호 공백 가능성을 확인했다.</label>
  <p><button type="submit">3초 후 현재 행에만 적용</button></p>
</form>
<p><a href="/optimization?category=exit-protection">취소하고 현재 상태 다시 보기</a></p>
{{template "foot" .}}
{{end}}

{{define "market-schedule"}}
{{template "head" .}}
<p><a href="/optimization">최적화</a> / <code>strategy-runtime</code></p>
<h1>시장·일정</h1>
<p class="muted">서버가 정의한 시장·세션 선택지와 official exchange calendar 근거를 읽는 화면이다.
임의 종목, 운영 사유, 시각 또는 휴장일을 입력하지 않는다. 이 화면은 주문·운영 토글·설정을 변경하지 않는다.</p>
{{if .LoadErr}}<p class="danger">상태를 읽지 못해 닫힌 기본값을 표시한다. 신규 진입은 fail-closed다.</p>{{end}}
{{if .Unwired}}<p class="notice">scheduler seam 미배선 — 닫힌 기본값만 표시한다.</p>{{end}}

<section aria-labelledby="scheduler-state-heading">
  <h2 id="scheduler-state-heading">Scheduler 상태</h2>
  <table>
    <tr><th>항목</th><th>기본값</th><th>Desired</th><th>Effective</th></tr>
    <tr><th>Scheduler</th><td>OFF</td><td>{{.SchedulerDesired}}</td><td><strong>{{.SchedulerEffective}}</strong></td></tr>
    <tr><th>자동 시작</th><td>OFF</td><td>{{.AutoStartDesired}}</td><td>{{.AutoStartEffective}}</td></tr>
    <tr><th>시장</th><td>선택 시장 없음</td><td colspan="2">{{.Market}}</td></tr>
    <tr><th>세션</th><td>정규장</td><td colspan="2">{{.Session}}</td></tr>
    <tr><th>적용 시점</th><td>다음 엔진 기동</td><td colspan="2">{{.ApplyTiming}}</td></tr>
  </table>
  <p class="muted">서버 정의 시장 범위: 선택 시장 없음, 한국, 미국. 서버 정의 세션: 정규장.</p>
</section>

<section aria-labelledby="calendar-heading">
  <h2 id="calendar-heading">Exchange calendar · 읽기 전용</h2>
  <table>
    <tr><th>Source</th><td>{{.CalendarSource}}</td></tr>
    <tr><th>Version</th><td><code>{{.CalendarVersion}}</code></td></tr>
    <tr><th>Updated at</th><td>{{.CalendarUpdatedAt}}</td></tr>
  </table>
  <p>6시간 freshness 또는 장 시작 전 refresh 조건이 실패하면 신규 entry는 대기한다. calendar는 이 화면에서 수정할 수 없다.</p>
</section>

<section aria-labelledby="decision-heading">
  <h2 id="decision-heading">현재 결정</h2>
  <table>
    <tr><th>Typed reason</th><td><code>{{.DecisionReason}}</code></td></tr>
    <tr><th>다음 전환</th><td>{{.NextTransition}}</td></tr>
  </table>
  <p>{{.DecisionHelp}}</p>
  <p><strong>신규 entry가 대기하거나 꺼져 있어도 exit·reconcile·fill detection은 계속된다.</strong></p>
</section>
{{template "foot" .}}
{{end}}

{{define "strategy-runtime"}}
{{template "head" .}}
<p><a href="/optimization">최적화</a> / <code>strategy-runtime</code></p>
<h1>한국 주식 전략 lane</h1>
<p class="muted">StockOS Parker VWAP conservative v1을 옮기는 읽기 전용 상태 카드다. 서버 고정 preset만 보여주며 이 화면에는 입력·저장·활성화·LIVE action이 없다.</p>
{{if .LoadErr}}<p class="danger" role="alert">상태를 읽지 못해 신규 진입 OFF를 표시한다.</p>{{end}}
{{if .Unwired}}<p class="notice">strategy-runtime seam 미배선 — dormant 기본값만 표시한다.</p>{{end}}
<section aria-labelledby="strategy-parameters"><h2 id="strategy-parameters">전략 파라미터 · 읽기 전용</h2>
<table><tr><th>항목</th><th>기본값</th><th>Desired</th><th>Effective</th><th>단위·범위</th><th>출처</th><th>적용 시점</th></tr>
<tr><th>최소 VWAP 기울기</th><td>0.08</td><td>0.08</td><td>미구성</td><td>% · 0 이상</td><td rowspan="11"><code>d75113d3</code><br><code>09260ac…</code><br>manifest 미재현</td><td>bar 평가 시</td></tr>
<tr><th>EMA9 접촉 허용</th><td>0.25</td><td>0.25</td><td>미구성</td><td>% · 0 이상</td><td>bar 평가 시</td></tr>
<tr><th>최소 LVN 전방 공간</th><td>1.2</td><td>1.2</td><td>미구성</td><td>% · 0 이상</td><td>bar 평가 시</td></tr>
<tr><th>최소 기대 RR</th><td>1.5</td><td>1.5</td><td>미구성</td><td>R · 0 이상</td><td>bar 평가 시</td></tr>
<tr><th>얽힘 band</th><td>0.35</td><td>0.35</td><td>미구성</td><td>% · 0 이상</td><td>bar 평가 시</td></tr>
<tr><th>최대 band 확장률</th><td>1.8</td><td>1.8</td><td>미구성</td><td>배 · 0 이상</td><td>bar 평가 시</td></tr>
<tr><th>하드 스톱</th><td>0.7</td><td>0.7</td><td>미구성</td><td>% · 고정</td><td>결정 생성 시</td></tr>
<tr><th>목표</th><td>3.0</td><td>3.0</td><td>미구성</td><td>R · 고정</td><td>결정 생성 시</td></tr>
<tr><th>시초 제외</th><td>10</td><td>10</td><td>미구성</td><td>분 · 고정</td><td>세션 평가 시</td></tr>
<tr><th>최대 신호 나이</th><td>15</td><td>15</td><td>미구성</td><td>초 · 0~15</td><td>결정 생성 시</td></tr>
<tr><th>최대 진입 괴리</th><td>0.20</td><td>0.20</td><td>미구성</td><td>% · 0~0.20</td><td>dispatch 전</td></tr></table>
<p>각 값은 서버 소유 fixed preset이다. source manifest가 <code>09260ac…</code>를 재현하기 전에는 effective 값이 아니다.</p></section>
<section aria-labelledby="lane-state"><h2 id="lane-state">lane 상태</h2><table><tr><th>항목</th><th>기본값</th><th>Desired</th><th>Effective</th></tr><tr><th>krx_parker_vwap_conservative_v1</th><td>OFF</td><td>{{.LaneDesired}}</td><td><strong>{{.LaneEffective}}</strong></td></tr></table></section>
<section aria-labelledby="autostart-state"><h2 id="autostart-state">자동 기동</h2><table><tr><th>기본값</th><th>Desired</th><th>Effective</th></tr><tr><td>OFF</td><td>{{.AutoStartDesired}}</td><td>{{.AutoStartEffective}}</td></tr></table><p>다음 엔진 기동 시 별도 activation을 다시 검증한다.</p></section>
<section aria-labelledby="live-state"><h2 id="live-state">LIVE 주문 승인</h2><table><tr><th>권한</th><th>Desired</th><th>Effective</th></tr><tr><th>Automation gate</th><td>{{.GateDesired}}</td><td>{{.GateEffective}}</td></tr><tr><th>LIVE approval</th><td>{{.LiveDesired}}</td><td>{{.LiveEffective}}</td></tr></table><p>lane·자동 기동·gate·LIVE는 서로 다른 권한이다. 일괄 활성화는 없다.</p></section>
<section aria-labelledby="strategy-blockers"><h2 id="strategy-blockers">Activation blockers</h2><table><tr><th>a045 브로커 보호 상태</th><td>{{.Protection}}</td></tr><tr><th>a046 후보 provenance</th><td>{{.Candidate}}</td></tr><tr><th>a048 scheduler/calendar claim</th><td>{{.Scheduler}}</td></tr><tr><th>StockOS source manifest</th><td>{{.SourceManifest}}</td></tr><tr><th>첫 refusal</th><td><code>{{.Reason}}</code></td></tr></table><p><strong>신규 entry가 OFF여도 exit·reconcile·보호 감독은 계속된다.</strong></p></section>
{{template "foot" .}}
{{end}}
`
