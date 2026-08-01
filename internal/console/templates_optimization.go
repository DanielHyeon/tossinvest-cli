package console

const optimizationTemplates = `
{{define "optimization"}}
{{template "head" .}}
<div class="optimization-title">
  <div>
    <h1>전략 최적화</h1>
    <p class="muted">설정과 LIVE 권한을 분리한 versioned lifecycle입니다. 임의 숫자·문자·종목을 입력하지 않고 owner가 제공한 선택지만 검토합니다.</p>
  </div>
  <dl class="status-strip" aria-label="최적화 최상위 상태">
    <div><dt>Desired</dt><dd><strong>v{{.Snapshot.Version}}</strong></dd></div>
    <div><dt>Effective</dt><dd><strong>v{{.Snapshot.EffectiveVersion}}</strong></dd></div>
    <div><dt>LIVE 권한</dt><dd><strong class="bad">별도 · 변경 안 함</strong></dd></div>
    <div><dt>재시작</dt><dd>{{if .Snapshot.RestartRequired}}필요{{else}}불필요{{end}}</dd></div>
  </dl>
</div>
{{if .Notice}}<p class="notice" role="status">{{.Notice}}</p>{{end}}
{{if .Warning}}<p class="notice" role="alert">{{.Warning}}</p>{{end}}
{{if .LifecycleErr}}<p class="danger" role="alert">lifecycle을 읽지 못했습니다: <code>{{.LifecycleErr}}</code>. 마지막 값을 현재값처럼 사용하지 않고 모든 변경을 닫았습니다.</p>{{end}}

<div class="optimization-shell">
  <nav class="optimization-nav" aria-label="최적화 카테고리">
    {{range .Categories}}
    <a href="/optimization?category={{.ID}}" {{if .Active}}class="on" aria-current="page"{{end}}>
      <span>{{.Label}}</span><small>{{.Status}}</small>
    </a>
    {{end}}
  </nav>

  <div class="optimization-content">
  {{if .Overview}}
    <section aria-labelledby="optimization-overview-title">
      <h2 id="optimization-overview-title">개요</h2>
      <p>무엇을 바꿀 수 있는지와 실제 적용 여부를 먼저 확인합니다. 설정 저장은 lane, LIVE, automation gate, kill switch 또는 현재 포지션을 바꾸지 않습니다.</p>
      <div class="category-summary">
      {{range .Categories}}
        <article>
          <h3><a href="/optimization?category={{.ID}}">{{.Label}}</a></h3>
          <p>{{.Purpose}}</p>
          <span class="state-badge">{{.Status}}</span>
        </article>
      {{end}}
      </div>
      <dl>
        <dt>마지막 actor</dt><dd>{{if .Snapshot.Actor}}<code>{{.Snapshot.Actor}}</code>{{else}}없음{{end}}</dd>
        <dt>마지막 reason</dt><dd>{{if .Snapshot.Reason}}<code>{{.Snapshot.Reason}}</code>{{else}}없음{{end}}</dd>
        <dt>settings digest</dt><dd>{{if .Snapshot.SettingsDigest}}<code>{{.Snapshot.SettingsDigest}}</code>{{else}}미생성{{end}}</dd>
        <dt>evidence</dt><dd><code>{{.Evidence.Status}}</code>{{range .Evidence.Missing}} · {{.}}{{end}}</dd>
        <dt>effective entry</dt><dd>{{if .Snapshot.EffectiveEntry}}ON{{else}}OFF · manifest 재승인 필요{{end}}</dd>
      </dl>
    </section>
  {{end}}

  {{if .ExitProtection}}
    <section aria-labelledby="exit-protection-title">
      <h2 id="exit-protection-title">익절·보호</h2>
      <p>신규 관리 포지션에만 적용할 공통 정책입니다. 기존 포지션 snapshot과 LIVE/lane 상태는 자동으로 재결합하지 않습니다.</p>
      {{if .LegacyLoadErr}}<p class="danger">legacy effective config를 읽지 못했습니다: <code>{{.LegacyLoadErr}}</code></p>{{end}}
      {{if eq .Snapshot.Version 0}}<p class="notice">lifecycle command seam 미배선 · 조회만 가능합니다.</p>{{end}}
      {{range .Fields}}
      <fieldset class="setting-row">
        <legend>{{.Label}} <code>{{.Key}}</code></legend>
        <p>{{.Description}}</p>
        <dl class="setting-values">
          <div><dt>기본값</dt><dd>{{.Default}}</dd></div>
          <div><dt>Desired</dt><dd><strong>{{.Desired}}</strong></dd></div>
          <div><dt>Effective</dt><dd><strong>{{.Effective}}</strong></dd></div>
          <div><dt>단위</dt><dd>{{.Unit}}</dd></div>
          <div><dt>적용</dt><dd>{{.ApplyTiming}}</dd></div>
          <div><dt>안전 방향</dt><dd>{{.Safety}}</dd></div>
          <div><dt>Owner</dt><dd><code>{{.Owner}}</code></dd></div>
          <div><dt>출처</dt><dd><code>{{.PolicyID}}</code> {{.PolicyVersion}} {{if .Evidence}}· evidence <code>{{.Evidence}}</code>{{end}}</dd></div>
        </dl>
      </fieldset>
      {{end}}

      <div class="choice-grid" aria-label="공통 익절 보호 정책 preset">
      {{range .Policies}}
        <article class="choice-tile" {{if .Selected}}data-selected="true"{{end}}>
          <header><h3>{{.Label}}</h3>{{if .Recommended}}<span class="state-badge">추천 · 자동 저장 안 함</span>{{end}}</header>
          <p><code>{{.ID}}</code>{{if .Selected}} · <strong>현재 desired</strong>{{end}}</p>
          <div class="table-scroll" role="region" aria-label="{{.Label}} 보호 단계" tabindex="0">
          <table>
            <tr><th>단계</th><th>목표 수익률</th><th>진입가 대비 보호</th><th>잔량 기준 익절</th></tr>
            {{range $i, $r := .Ladder.Rungs}}
            <tr><td>T{{add1 $i}}</td><td>{{$r.TargetPct}}%</td><td>{{$r.StopPct}}%</td><td>{{$r.PartialRatio}}</td></tr>
            {{end}}
          </table>
          </div>
          {{if eq .ID "COMMON_LADDER_HYBRID_50"}}<p>약 50%를 남기고 T4부터 high-water -6.5% 보호를 사용합니다.</p>
          {{else if eq .ID "COMMON_LADDER_RUNNER"}}<p><strong>고정 목표 없음</strong>으로 표시하며 999% sentinel을 입력값으로 노출하지 않습니다.</p>
          {{else}}<p>T4에서 잔량을 전량익절하는 균형형입니다.</p>{{end}}
          {{if $.LifecycleReady}}
          <form method="post" action="/optimization/exit-policy" class="choice-action">
            <input type="hidden" name="csrf" value="{{$.CSRF}}">
            <input type="hidden" name="base_version" value="{{$.Snapshot.Version}}">
            <input type="hidden" name="category" value="exit-protection">
            <input type="hidden" name="setting_key" value="exit.common-policy">
            <input type="hidden" name="option_id" value="{{.ID}}">
            <button type="submit" {{if .Selected}}disabled{{end}}>{{if .Selected}}선택됨{{else}}이 preset 미리보기{{end}}</button>
          </form>
          {{end}}
        </article>
      {{end}}
      </div>

      <aside class="notice" aria-label="브로커 보호 capability">
        <strong>a045 브로커 보호: 미검증/사용 불가</strong>
        <p>보호 capability owner provider가 통합되기 전에는 주문 유형·기본값·활성화 버튼을 만들지 않습니다. 준비 이후에도 기본은 OFF입니다.</p>
      </aside>
      {{if .History}}
      <details class="history-panel">
        <summary>이 카테고리 rollback 후보 보기</summary>
        <p>Rollback도 과거 row를 수정하지 않고 현재 registry로 새 candidate를 만듭니다.</p>
        <div class="table-scroll" role="region" aria-label="익절 보호 설정 이력" tabindex="0">
        <table><tr><th>Version</th><th>Actor / reason</th><th>Desired</th><th>동작</th></tr>
        {{range .History}}<tr><td>v{{.Version}}</td><td><code>{{.Actor}}</code><br><code>{{.Reason}}</code></td>
          <td>{{index .Desired "exit.common-policy"}}</td><td>
          {{if ne .Version $.Snapshot.Version}}<form method="post" action="/optimization/exit-policy">
            <input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="action" value="rollback-preview">
            <input type="hidden" name="base_version" value="{{$.Snapshot.Version}}"><input type="hidden" name="target_version" value="{{.Version}}">
            <input type="hidden" name="category" value="exit-protection"><button type="submit">이 version rollback 미리보기</button>
          </form>{{else}}현재{{end}}</td></tr>{{end}}
        </table></div>
      </details>
      {{end}}
    </section>
  {{end}}

  {{if .PositionManagement}}
    <section aria-labelledby="position-management-title">
      <h2 id="position-management-title">종목별 관리</h2>
      <p>설정 저장과 position lifecycle command를 섞지 않습니다. 현재 row의 stable identity로만 동작하는 a044 화면을 사용합니다.</p>
      <p><a class="button-link" href="/position-management">현재 포지션 상속·override·release 상태 보기</a></p>
      <p class="muted">a044 owner descriptor adapter가 아직 settingmeta registry에 연결되지 않아 이 category는 현재 read-only입니다. label/default를 임의로 만들지 않았습니다.</p>
    </section>
  {{end}}

  {{if .CandidateFilters}}
    <section id="candidate-filters" aria-labelledby="candidate-filters-title">
      <h2 id="candidate-filters-title">후보 필터</h2>
      <p class="notice"><strong>미승인 · passed 구조적 0 · verdict 비활성</strong></p>
      <p class="muted">a046 owner가 승인한 registry option 전에는 숫자 0을 기본 threshold처럼 표시하지 않고 모든 행을 읽기 전용으로 유지합니다. 주문·RiskIntent·LIVE 상태는 변경하지 않는다.</p>
      <nav class="filter-bar" aria-label="후보 필터 시장 전환"><a href="#candidate-filters-KR">KR regular</a><a href="#candidate-filters-US">US regular</a></nav>
      {{range .CandidateFilterMarkets}}
      <article id="candidate-filters-{{.Market}}" aria-label="{{.Market}} {{.Session}} 후보 필터">
        <h3>{{.Market}} · {{.Session}}</h3>
        {{range .Filters}}<div class="detail-grid" aria-readonly="true"><div><strong><code>{{.Key}}</code> · {{.Label}}</strong><p class="muted">{{.Help}}</p></div>
        <dl><dt>default state</dt><dd>미승인 (<code>{{.DefaultState}}</code>)</dd><dt>desired</dt><dd>미승인 — 숫자 값 없음</dd>
        <dt>effective</dt><dd>미승인 — verdict 비활성</dd><dt>단위 / 유효 범위</dt><dd>{{.Unit}} / {{.ValidRange}}</dd>
        <dt>판정 방향</dt><dd>{{.Direction}}</dd><dt>표본 / evidence</dt><dd><code>{{.SampleState}}</code> / <code>{{.EvidenceState}}</code></dd>
        <dt>적용 시점</dt><dd>{{.ApplyTiming}}</dd><dt>누락</dt><dd>{{range .MissingEvidence}}<code>{{.}}</code> {{end}}</dd>
        <dt>preview / CAS</dt><dd>{{.PreviewContract}} — CAS 필수</dd>{{if .LegacyValue}}<dt>legacy provenance</dt><dd><code>{{.LegacyValue}}</code> · <code>{{.Provenance}}</code> — desired/effective로 승격하지 않음</dd>{{end}}</dl></div>{{end}}
      </article>{{end}}
    </section>
  {{end}}

  {{if .StrategyRuntime}}
    <section aria-labelledby="strategy-runtime-title">
      <h2 id="strategy-runtime-title">전략·실행</h2>
      <p class="notice"><strong>lane desired OFF · engine autostart OFF · scheduler desired OFF</strong></p>
      <p>a047 owner descriptor가 통합되기 전에는 lane·시장 scope·세션 setting을 만들지 않습니다. lane ON, autostart, automation gate, LIVE trading은 서로 다른 권한이며 이 화면은 어느 것도 켜지 않습니다.</p>
      <p><a class="button-link" href="/strategy-runtime/market-schedule">시장·일정 read-only 상태 보기</a></p>
    </section>
  {{end}}

  {{if .PerformanceHistory}}
    <section aria-labelledby="performance-history-title">
      <h2 id="performance-history-title">성과·이력</h2>
      <p class="notice"><strong>a049 provider unavailable · 읽기 전용</strong></p>
      <p>최근 30일 · 전체 시장 · 전체 lane · complete lineage만이라는 필터 계약을 유지합니다. provider가 없으므로 P&amp;L, realized R, PF, MDD, slippage, markout, MFE/MAE를 0으로 꾸며내지 않습니다.</p>
      {{if .Audit}}<div class="table-scroll" role="region" aria-label="설정 변경 감사 이력" tabindex="0"><table>
        <tr><th>Version</th><th>Key</th><th>Before → After</th><th>Actor / reason</th><th>Audit</th></tr>
        {{range .Audit}}<tr><td>v{{.Version}}</td><td><code>{{.Key}}</code></td><td>{{if .BeforeOptionID}}{{.BeforeOptionID}}{{else}}미승인{{end}} → {{if .AfterOptionID}}{{.AfterOptionID}}{{else}}미승인{{end}}</td>
        <td><code>{{.Actor}}</code><br><code>{{.Reason}}</code></td><td><code>{{.AuditID}}</code></td></tr>{{end}}
      </table></div>{{else}}<p class="muted">아직 적용 이력이 없습니다.</p>{{end}}
    </section>
  {{end}}
  </div>
</div>
{{template "foot" .}}
{{end}}

{{define "optimization-preview"}}
{{template "head" .}}
<p><a href="/optimization?category={{.Preview.Category}}">최적화 / {{.Preview.Category}}</a></p>
<h1>설정 변경 미리보기</h1>
<p class="notice"><strong>아직 적용되지 않았습니다.</strong> before/after, 적용 시점과 권한 불변을 확인하세요.</p>
<section aria-labelledby="preview-diff-title">
  <h2 id="preview-diff-title">Changed subset · {{len .Preview.Changes}}개</h2>
  <div class="table-scroll" role="region" aria-label="설정 변경 before after" tabindex="0"><table>
    <tr><th>Key</th><th>Before</th><th>After</th><th>적용</th><th>안전 방향</th></tr>
    {{range .Preview.Changes}}<tr><td><code>{{.Key}}</code></td><td>{{if .BeforeOptionID}}{{.BeforeOptionID}}{{else}}미승인{{end}}</td>
    <td><strong>{{if .AfterOptionID}}{{.AfterOptionID}}{{else}}미승인{{end}}</strong></td><td>{{.ApplyTiming}}</td><td>{{.Safety}}</td></tr>{{end}}
  </table></div>
  <dl>
    <dt>Base version</dt><dd>v{{.Preview.BaseVersion}}</dd><dt>Evidence</dt><dd><code>{{.Preview.Evidence.Status}}</code> {{.Preview.Evidence.Digest}}</dd>
    <dt>기존 포지션</dt><dd>{{if .Preview.ExistingPositionsUnchanged}}변경 안 함{{else}}거부{{end}}</dd>
    <dt>LIVE/lane/gate</dt><dd>{{if .Preview.LiveStateUnchanged}}변경 안 함{{else}}거부{{end}}</dd>
    <dt>재시작</dt><dd>{{if .Preview.RestartRequired}}필요{{else}}불필요{{end}}</dd>
    <dt>effective entry</dt><dd>{{if .Preview.EffectiveEntryAfterApply}}유지{{else}}OFF · manifest 재승인 필요{{end}}</dd>
  </dl>
</section>
<section class="sticky-save" aria-labelledby="preview-approval-title">
  <h2 id="preview-approval-title">명시적 승인</h2>
  {{if .Preview.RiskConfirmationRequired}}
  <p class="danger">보호 또는 위험 계약에 영향을 줄 수 있어 <strong>{{.WaitSecs}}초</strong> 확인 대기가 적용됩니다. typed phrase나 자유 reason 입력은 없습니다.</p>
  {{end}}
  <form method="post" action="/optimization/exit-policy">
    <input type="hidden" name="csrf" value="{{.CSRF}}"><input type="hidden" name="action" value="apply">
    <input type="hidden" name="capability" value="{{.Preview.Capability}}">
    {{if .Preview.RiskConfirmationRequired}}<label class="confirm-check"><input type="checkbox" name="confirm" value="yes" required> before/after, 기존 포지션 불변, LIVE 권한 분리를 확인했습니다.</label>
    {{else}}<input type="hidden" name="confirm" value="yes">{{end}}
    <div class="actions"><a class="button-link secondary-link" href="/optimization?category={{.Preview.Category}}">취소</a><button type="submit">이 changed subset 승인·적용</button></div>
  </form>
</section>
{{template "foot" .}}
{{end}}

{{define "market-schedule"}}
{{template "head" .}}
<p><a href="/optimization?category=strategy-runtime">최적화</a> / <code>strategy-runtime</code></p>
<h1>시장·일정</h1>
<p class="muted">서버가 정의한 시장·세션 선택지와 official exchange calendar 근거를 읽는 화면입니다. 임의 종목, 운영 사유, 시각 또는 휴장일을 입력하지 않으며 주문·운영 토글·설정을 변경하지 않습니다.</p>
{{if .LoadErr}}<p class="danger">상태를 읽지 못해 닫힌 기본값을 표시한다. 신규 진입은 fail-closed입니다.</p>{{end}}
{{if .Unwired}}<p class="notice">scheduler seam 미배선 — 닫힌 기본값만 표시합니다.</p>{{end}}
<section aria-labelledby="scheduler-state-heading"><h2 id="scheduler-state-heading">Scheduler 상태</h2><table>
<tr><th>항목</th><th>기본값</th><th>Desired</th><th>Effective</th></tr>
<tr><th>Scheduler</th><td>OFF</td><td>{{.SchedulerDesired}}</td><td><strong>{{.SchedulerEffective}}</strong></td></tr>
<tr><th>자동 시작</th><td>OFF</td><td>{{.AutoStartDesired}}</td><td>{{.AutoStartEffective}}</td></tr>
<tr><th>시장</th><td>선택 시장 없음</td><td colspan="2">{{.Market}}</td></tr><tr><th>세션</th><td>정규장</td><td colspan="2">{{.Session}}</td></tr>
<tr><th>적용 시점</th><td>다음 엔진 기동</td><td colspan="2">{{.ApplyTiming}}</td></tr></table>
<p class="muted">서버 정의 시장 범위: 선택 시장 없음, 한국, 미국. 서버 정의 세션: 정규장.</p></section>
<section aria-labelledby="calendar-heading"><h2 id="calendar-heading">Exchange calendar · 읽기 전용</h2><table>
<tr><th>Source</th><td>{{.CalendarSource}}</td></tr><tr><th>Version</th><td><code>{{.CalendarVersion}}</code></td></tr>
<tr><th>Updated at</th><td>{{.CalendarUpdatedAt}}</td></tr></table>
<p>6시간 freshness 또는 장 시작 전 refresh 조건이 실패하면 신규 entry는 대기합니다. calendar는 이 화면에서 수정할 수 없습니다.</p></section>
<section aria-labelledby="decision-heading"><h2 id="decision-heading">현재 결정</h2><table>
<tr><th>Typed reason</th><td><code>{{.DecisionReason}}</code></td></tr><tr><th>다음 전환</th><td>{{.NextTransition}}</td></tr></table>
<p>{{.DecisionHelp}}</p><p><strong>신규 entry가 대기하거나 꺼져 있어도 exit·reconcile·fill detection은 계속된다.</strong></p></section>
{{template "foot" .}}
{{end}}
`
