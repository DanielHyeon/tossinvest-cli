package console

const optimizationTemplates = `
{{define "optimization"}}
{{template "head" .}}
<div class="optimization-title">
  <div>
    <p class="eyebrow">Optimization lifecycle</p>
    <h1>최적화 <span class="muted">전략 파라미터</span></h1>
    <p class="muted">서버가 정한 preset 하나를 골라 before/after를 검토한 뒤 적용한다. 설정과 LIVE 권한은 항상 분리된다.</p>
  </div>
  <dl class="status-strip" aria-label="최적화 최상위 상태" {{if .LifecycleErr}}data-lifecycle-state="error"{{else if not .LifecycleReady}}data-lifecycle-state="unavailable"{{else}}data-lifecycle-state="ready"{{end}}>
    <div><dt>Desired</dt><dd><strong>{{if .LifecycleReady}}v{{.Snapshot.Version}}{{else}}확인 불가{{end}}</strong></dd></div>
    <div><dt>Effective</dt><dd><strong>{{if .LifecycleReady}}v{{.Snapshot.EffectiveVersion}}{{else}}확인 불가{{end}}</strong>{{if .LifecycleReady}}<small>{{if eq .Snapshot.Version .Snapshot.EffectiveVersion}}desired와 일치{{else}}desired와 불일치{{end}}</small>{{end}}</dd></div>
    <div data-evidence-state="{{.Evidence.Status}}"><dt>성과 근거</dt><dd><strong>{{if not .LifecycleReady}}확인 불가{{else if eq .Evidence.Status "complete"}}완료{{else if eq .Evidence.Status "insufficient"}}부족 · 추천 불가{{else if eq .Evidence.Status "stale"}}오래됨 · 근거 추천 중지{{else}}사용 불가 · 추천 없음{{end}}</strong></dd></div>
    <div><dt>Effective entry</dt><dd><strong>{{if and .LifecycleReady .Snapshot.EffectiveEntry}}ON{{else}}OFF{{end}}</strong><small>{{if and .LifecycleReady .Snapshot.EffectiveEntry}}현재 manifest 승인됨{{else}}manifest 재승인 전 진입 없음{{end}}</small></dd></div>
    <div><dt>LIVE 권한</dt><dd><strong>별도 · 변경 안 함</strong></dd></div>
    <div><dt>재시작</dt><dd>{{if not .LifecycleReady}}확인 불가{{else if .Snapshot.RestartRequired}}필요{{else}}불필요{{end}}</dd></div>
  </dl>
</div>
{{if .Notice}}<p class="notice" role="status">{{.Notice}}</p>{{end}}
{{if .Warning}}<p class="notice" role="alert">{{.Warning}}</p>{{end}}
{{if .LifecycleErr}}<p class="danger" role="alert">lifecycle을 읽지 못했다: <code>{{.LifecycleErr}}</code>. 마지막 값을 현재값처럼 사용하지 않고 모든 변경을 닫았다.</p>{{end}}
{{if and .LifecycleReady (ne .Evidence.Status "complete")}}<p class="notice evidence-banner" role="status"><strong>성과 근거 {{.Evidence.Status}}</strong> · 근거 기반 추천 candidate를 만들지 않는다.{{range .Evidence.Missing}} <code>{{.}}</code>{{end}}</p>{{end}}
<nav class="filter-bar" aria-label="최적화 관련 읽기 전용 화면">
  <a href="/strategy-runtime">전략 lane</a>
  <a href="/strategy-runtime/market-schedule">시장·일정</a>
  <a href="/performance-history">레인 성과</a>
</nav>

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
      <p>무엇을 바꿀 수 있는지와 실제 적용 여부를 먼저 확인한다. 설정 저장은 lane, LIVE, automation gate, kill switch 또는 현재 포지션을 바꾸지 않는다.</p>
      <div class="category-summary">
      {{range .Categories}}
        <article>
          <h3><a href="/optimization?category={{.ID}}">{{.Label}}</a></h3>
          <p>{{.Purpose}}</p>
          <span class="status-pill">{{.Status}}</span>
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
    <section id="exit-protection" aria-labelledby="exit-protection-title">
      <p class="section-kicker">현재 변경 가능 항목 1개 · <code>exit.common-policy</code></p>
      <h2 id="exit-protection-title">청산/보호 · 익절 정책</h2>
      <p>신규 관리 포지션에만 적용할 공통 정책이다. 기존 포지션 snapshot과 LIVE/lane 상태는 자동으로 재결합하지 않는다.</p>
      <ol class="optimization-steps" aria-label="익절 보호 설정 적용 순서">
        <li><strong>1 · preset 선택</strong><span>서버가 제공한 세 정책 중 하나만 선택</span></li>
        <li><strong>2 · before/after 확인</strong><span>별도 읽기 전용 preview에서 영향 범위 검토</span></li>
        <li><strong>3 · 3초 확인 후 적용</strong><span>체크박스와 승인 버튼으로 명시적 적용</span></li>
      </ol>
      {{if .LegacyLoadErr}}<p class="danger">legacy effective config를 읽지 못했다: <code>{{.LegacyLoadErr}}</code></p>{{end}}
      {{if eq .Snapshot.Version 0}}<p class="notice">lifecycle command seam 미배선 · 조회만 가능하다.</p>{{end}}
      {{range .Fields}}
      <fieldset class="setting-row{{if .ConfigurationError}} setting-error{{end}}" {{if .ConfigurationError}}aria-readonly="true"{{end}}>
        <legend>{{.Label}} <code>{{.Key}}</code></legend>
        <p>{{.Description}}</p>
        {{if .ConfigurationError}}<p class="danger configuration-error" role="alert"><strong>설정 오류 · 읽기 전용</strong><br>{{.ConfigurationError}} owner descriptor를 바로잡기 전에는 preset preview와 apply를 제공하지 않는다.</p>{{end}}
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
          <header><h3>{{.Label}}</h3>{{if .Recommended}}<span class="status-pill">추천 · 자동 저장 안 함</span>{{end}}</header>
          <p><code>{{.ID}}</code>{{if .Selected}} · <strong>현재 desired</strong>{{end}}</p>
          <div class="table-scroll" role="region" aria-label="{{.Label}} 보호 단계" tabindex="0">
          <table>
            <tr><th>단계</th><th>목표 수익률</th><th>진입가 대비 보호</th><th>잔량 기준 익절</th></tr>
            {{range $i, $r := .Ladder.Rungs}}
            <tr><td>T{{add1 $i}}</td><td>{{$r.TargetPct}}%</td><td>{{$r.StopPct}}%</td><td>{{$r.PartialRatio}}</td></tr>
            {{end}}
          </table>
          </div>
          {{if eq .ID "COMMON_LADDER_HYBRID_50"}}<p>약 50%를 남기고 T4부터 high-water -6.5% 보호를 사용한다.</p>
          {{else if eq .ID "COMMON_LADDER_RUNNER"}}<p><strong>고정 목표 없음</strong>으로 표시하며 999% sentinel을 입력값으로 노출하지 않는다.</p>
          {{else}}<p>T4에서 잔량을 전량익절하는 균형형이다.</p>{{end}}
          {{if $.ExitPolicyWritable}}
          <form method="post" action="/optimization/exit-policy" class="choice-action">
            <input type="hidden" name="csrf" value="{{$.CSRF}}">
            <input type="hidden" name="base_version" value="{{$.Snapshot.Version}}">
            <input type="hidden" name="category" value="exit-protection">
            <input type="hidden" name="setting_key" value="exit.common-policy">
            <input type="hidden" name="option_id" value="{{.ID}}">
            <button type="submit" {{if .Selected}}disabled aria-disabled="true"{{end}}>{{if .Selected}}현재 desired · 선택됨{{else}}이 preset 선택 · 미리보기{{end}}</button>
          </form>
          {{else}}<p class="muted preset-readonly" role="status">읽기 전용 · owner descriptor 확인 필요</p>
          {{end}}
        </article>
      {{end}}
      </div>

      <section aria-labelledby="broker-protection-title">
        <h3 id="broker-protection-title">브로커 상주 보호</h3>
        <p class="muted">현재 공통 보호선과 브로커에 남아 있는 보호주문을 함께 읽는다. 종목·가격·수량·사유 입력과 LIVE 전체 활성화는 없다.</p>
        {{if .ProtectionLoadErr}}<p class="danger" role="alert">보호 상태를 읽지 못했다: {{.ProtectionLoadErr}}</p>{{end}}
        {{if not .ProtectionWired}}
        <div class="detail-grid" aria-readonly="true">
          <div><strong>a045 브로커 보호: 미검증/사용 불가</strong><p class="muted">엔진이 꺼져도 손절이 남는 기능이다.</p></div>
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
      {{if .History}}
      <details class="history-panel">
        <summary>이 카테고리 rollback 후보 보기</summary>
        <p>Rollback도 과거 row를 수정하지 않고 현재 registry로 새 candidate를 만든다.</p>
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
      <p>설정 저장과 position lifecycle command를 섞지 않는다. 현재 row의 stable identity로만 동작하는 a044 화면을 쓴다.</p>
      <p><a class="button-link" href="/position-management">현재 포지션 상속·override·release 상태 보기</a></p>
      <p class="muted">a044 owner descriptor adapter가 아직 settingmeta registry에 연결되지 않아 이 category는 현재 read-only다. label/default를 임의로 만들지 않았다.</p>
    </section>
  {{end}}

  {{if .CandidateFilters}}
    <section id="candidate-filters" aria-labelledby="candidate-filters-title">
      <h2 id="candidate-filters-title">후보 필터</h2>
      <p class="notice"><strong>미승인 · passed 구조적 0 · verdict 비활성</strong></p>
      <p class="muted">a046 owner가 승인한 registry option 전에는 숫자 0을 기본 threshold처럼 표시하지 않고 모든 행을 읽기 전용으로 유지한다. 주문·RiskIntent·LIVE 상태는 변경하지 않는다.</p>
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
      <p>lane ON, autostart, automation gate, LIVE trading은 서로 다른 권한이다. 이 카테고리는 서버가 확정한 desired/effective와 blocker를 읽기만 하며 어느 권한도 켜지 않는다.</p>
      <div class="actions">
        <a class="button-link" href="/strategy-runtime">전략 lane 상태 보기</a>
        <a class="button-link secondary-link" href="/strategy-runtime/market-schedule">시장·일정 상태 보기</a>
      </div>
    </section>
  {{end}}

  {{if .PerformanceHistory}}
    <section aria-labelledby="performance-history-title">
      <h2 id="performance-history-title">성과·이력</h2>
      <p class="notice"><strong>최근 30일 · 전체 시장 · 전체 lane · complete lineage · 읽기 전용</strong></p>
      <p>서버가 고정한 조회 범위의 성과를 보여주며 이 카테고리에서 값을 입력하거나 저장하지 않는다. 측정 불가를 0으로 꾸며내지 않고 누락 사유를 함께 표시한다.</p>
      <p><a class="button-link" href="/performance-history">레인 성과 상세 보기</a></p>
      {{if .Audit}}<div class="table-scroll" role="region" aria-label="설정 변경 감사 이력" tabindex="0"><table>
        <tr><th>Version</th><th>Key</th><th>Before → After</th><th>Actor / reason</th><th>Audit</th></tr>
        {{range .Audit}}<tr><td>v{{.Version}}</td><td><code>{{.Key}}</code></td><td>{{if .BeforeOptionID}}{{.BeforeOptionID}}{{else}}미승인{{end}} → {{if .AfterOptionID}}{{.AfterOptionID}}{{else}}미승인{{end}}</td>
        <td><code>{{.Actor}}</code><br><code>{{.Reason}}</code></td><td><code>{{.AuditID}}</code></td></tr>{{end}}
      </table></div>{{else}}<p class="muted">아직 적용 이력이 없다.</p>{{end}}
    </section>
  {{end}}
  </div>
</div>
{{template "foot" .}}
{{end}}

{{define "optimization-preview"}}
{{template "head" .}}
<p><a href="/optimization?category={{.Preview.Category}}">최적화 / {{.Preview.Category}}</a></p>
<p class="eyebrow">2단계 · 읽기 전용 검토</p>
<h1>설정 변경 미리보기</h1>
<p class="notice"><strong>아직 적용되지 않았다.</strong> before/after, 적용 시점과 권한 불변을 확인한다.</p>
<section class="preview-review" aria-labelledby="preview-diff-title" aria-readonly="true">
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
<section class="sticky-save" aria-labelledby="preview-approval-title" {{if .Preview.RiskConfirmationRequired}}data-risk-preview data-not-before-ms="{{.NotBeforeUnixMilli}}"{{end}}>
  <p class="section-kicker">3단계 · 최종 확인</p>
  <h2 id="preview-approval-title">이 변경만 승인·적용</h2>
  {{if .Preview.RiskConfirmationRequired}}
  <p class="danger">보호 또는 위험 계약에 영향을 줄 수 있어 3초 확인 대기가 적용된다. 문구나 사유를 직접 입력하지 않는다.</p>
  <p class="notice" role="status" aria-live="polite" data-risk-countdown>{{if .Waiting}}{{.WaitSecs}}초 남음{{else}}승인 가능{{end}}</p>
  {{end}}
  <form method="post" action="/optimization/exit-policy">
    <input type="hidden" name="csrf" value="{{.CSRF}}"><input type="hidden" name="action" value="apply">
    <input type="hidden" name="capability" value="{{.Preview.Capability}}">
    {{if .Preview.RiskConfirmationRequired}}<label class="confirm-check"><input type="checkbox" name="confirm" value="yes" required data-risk-confirm> before/after, 기존 포지션 불변, LIVE 권한 분리를 확인했다.</label>
    {{else}}<input type="hidden" name="confirm" value="yes">{{end}}
    <div class="actions approval-actions"><a class="button-link secondary-link" href="/optimization?category={{.Preview.Category}}">취소</a><button type="submit" {{if .Preview.RiskConfirmationRequired}}data-risk-submit disabled{{end}}>이 changed subset 승인·적용</button></div>
  </form>
</section>
{{if .Preview.RiskConfirmationRequired}}<script>{{optimizationPreviewScript}}</script>{{end}}
{{template "foot" .}}
{{end}}

{{define "optimization-conflict"}}
{{template "head" .}}
<p><a href="/optimization?category={{.Category}}">최적화 / {{.Category}}</a></p>
<h1>설정 version 충돌</h1>
<p class="danger" role="alert">base v{{.BaseVersion}} 승인 중 latest v{{.LatestVersion}}을 확인했다. stale draft를 자동 retry하지 않았다.</p>
<section aria-labelledby="optimization-conflict-diff">
  <h2 id="optimization-conflict-diff">Attempted → latest field diff</h2>
  <p>시도한 draft는 비교용으로만 읽기 전용으로 보존했다. 최신 desired/effective를 확인한 뒤 명시적으로 돌아가거나 새 preview를 만든다.</p>
  <div class="table-scroll" role="region" aria-label="충돌한 설정 비교" tabindex="0"><table>
    <tr><th>Key</th><th>Attempted</th><th>Latest desired</th><th>Latest effective</th></tr>
    {{range .Rows}}<tr><td><code>{{.Key}}</code></td><td>{{.Attempted}}</td><td>{{.LatestDesired}}</td><td>{{.LatestEffective}}</td></tr>{{end}}
  </table></div>
  <div class="actions conflict-actions">
    <a class="button-link secondary-link" href="/optimization?category={{.Category}}">최신값으로 돌아가기</a>
    {{if .CanRepreview}}<form method="post" action="/optimization/exit-policy">
      <input type="hidden" name="csrf" value="{{.CSRF}}"><input type="hidden" name="base_version" value="{{.LatestVersion}}">
      <input type="hidden" name="category" value="{{.Category}}"><input type="hidden" name="setting_key" value="{{.RepreviewKey}}">
      <input type="hidden" name="option_id" value="{{.RepreviewOption}}"><button type="submit">보존한 draft로 새 preview</button>
    </form>{{end}}
  </div>
</section>
{{template "foot" .}}
{{end}}

{{define "performance-history"}}
{{template "head" .}}
<p><a href="/optimization">최적화</a> / <code>performance-history</code></p>
<h1>성과 이력 <span class="muted">레인별 · 읽기 전용</span></h1>
<p class="page-intro muted">StockOS lane-console 정보 구조처럼 필터 상태, 결과, 근거와 누락 이유를 한 화면에 둡니다.
이 화면은 기존 journal·position 관측에서 만든 파생 결과만 읽으며 주문, 설정 저장, lane 토글 또는 LIVE 승인을 제공하지 않는다.</p>

<section aria-labelledby="performance-filter-title">
  <h2 id="performance-filter-title">서버 조회 기본값</h2>
  <div class="filter-bar" aria-label="고정된 성과 조회 범위" aria-readonly="true">
    <span class="status-pill">최근 {{.View.Query.PeriodDays}}일</span>
    <span class="status-pill">전체 시장</span>
    <span class="status-pill">전체 lane</span>
    <span class="status-pill">complete lineage only</span>
  </div>
  <p class="muted">화면 필터일 뿐 거래 정책 기본값이 아니다. 입력하거나 저장할 값이 없다.</p>
</section>

{{if .Unwired}}
<p class="notice"><strong>performance.db 조회 seam 미배선</strong> — 성과를 0으로 꾸미지 않는다.</p>
{{else if .LoadErr}}
<p class="danger">성과를 읽지 못했다. 값은 변경되지 않았다: <code>{{.LoadErr}}</code></p>
{{else}}
<section aria-labelledby="performance-state-title">
  <h2 id="performance-state-title">측정 상태</h2>
  <dl>
    <dt><code>complete</code></dt><dd>{{.View.States.Complete}}건 · 전체 식별자 chain 확인</dd>
    <dt><code>link_missing</code></dt><dd>{{.View.States.LinkMissing}}건 · {{.View.Explanation "link_missing"}}</dd>
    <dt><code>not_measured</code></dt><dd>{{.View.States.NotMeasured}}개 거래 · {{.View.Explanation "not_measured"}}</dd>
    <dt><code>insufficient_sample</code></dt><dd>{{.View.States.InsufficientSample}}개 묶음 · {{.View.Explanation "insufficient_sample"}}</dd>
  </dl>
</section>

{{if .View.Aggregates}}
{{range .View.Aggregates}}
<section aria-labelledby="lane-{{.LaneID}}-{{.PolicyID}}">
  <h2 id="lane-{{.LaneID}}-{{.PolicyID}}"><code>{{.LaneID}}</code> · <code>{{.PolicyID}}</code></h2>
  {{if eq .Status "insufficient_sample"}}
  <p class="notice"><strong>insufficient_sample · 추천 근거로 사용 불가</strong></p>
  {{else}}<p class="ok"><strong>complete</strong></p>{{end}}
  <dl>
    <dt>시장 / 표본 / 기간</dt><dd>{{.Market}} / {{.Samples}}건 / 최근 {{$.View.Query.PeriodDays}}일</dd>
    <dt>lane provenance</dt><dd><code>{{.LaneID}}@{{.LaneVersion}}</code></dd>
    <dt>policy provenance</dt><dd><code>{{.PolicyID}}@{{.PolicyVersion}}</code></dd>
    <dt>query semantics</dt><dd><code>{{.SemanticsVersion}}</code></dd>
    <dt>observation source</dt><dd>{{if .ObservationProvenance}}<code>{{.ObservationProvenance}}</code>{{else}}<code>not_measured</code>{{end}}</dd>
  </dl>
  <table class="data-table">
    <caption>지표 정의와 표본</caption>
    <thead><tr><th scope="col">지표</th><th scope="col">값</th><th scope="col">단위</th><th scope="col">표본</th><th scope="col">기간</th><th scope="col">provenance</th><th scope="col">쉬운 정의</th></tr></thead>
    <tbody>{{range .Metrics}}
      <tr>
        <th scope="row" data-label="지표">{{.Label}}</th>
        <td data-label="값">{{if eq .Status "not_measured"}}<code>not_measured</code>{{else}}<strong>{{.Value}}</strong>{{end}}</td>
        <td data-label="단위">{{.Unit}}</td>
        <td data-label="표본">{{.Samples}}건</td>
        <td data-label="기간">최근 {{$.View.Query.PeriodDays}}일</td>
        <td data-label="provenance"><code>{{.Provenance}}</code></td>
        <td data-label="쉬운 정의">{{.Help}}</td>
      </tr>
    {{end}}</tbody>
  </table>
</section>
{{end}}
{{else}}
<p class="notice">고정 조회 범위에 complete lineage 거래가 없다. 빈 결과는 성과 0이 아닙니다.</p>
{{end}}
{{end}}
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
<p><a href="/optimization?category=strategy-runtime">최적화</a> / <code>strategy-runtime</code></p>
<h1>시장·일정</h1>
<p class="muted">서버가 정의한 시장·세션 선택지와 official exchange calendar 근거를 읽는 화면이다. 임의 종목, 운영 사유, 시각 또는 휴장일을 입력하지 않으며 주문·운영 토글·설정을 변경하지 않는다.</p>
{{if .LoadErr}}<p class="danger">상태를 읽지 못해 닫힌 기본값을 표시한다. 신규 진입은 fail-closed다.</p>{{end}}
{{if .Unwired}}<p class="notice">scheduler seam 미배선 — 닫힌 기본값만 표시한다.</p>{{end}}
<section aria-labelledby="scheduler-state-heading"><h2 id="scheduler-state-heading">Scheduler 상태</h2><div class="table-scroll" role="region" aria-label="Scheduler 상태" tabindex="0"><table>
<tr><th>항목</th><th>기본값</th><th>Desired</th><th>Effective</th></tr>
<tr><th>Scheduler</th><td>OFF</td><td>{{.SchedulerDesired}}</td><td><strong>{{.SchedulerEffective}}</strong></td></tr>
<tr><th>자동 시작</th><td>OFF</td><td>{{.AutoStartDesired}}</td><td>{{.AutoStartEffective}}</td></tr>
<tr><th>시장</th><td>선택 시장 없음</td><td colspan="2">{{.Market}}</td></tr><tr><th>세션</th><td>정규장</td><td colspan="2">{{.Session}}</td></tr>
<tr><th>적용 시점</th><td>다음 엔진 기동</td><td colspan="2">{{.ApplyTiming}}</td></tr></table></div>
<p class="muted">서버 정의 시장 범위: 선택 시장 없음, 한국, 미국. 서버 정의 세션: 정규장.</p></section>
<section aria-labelledby="calendar-heading"><h2 id="calendar-heading">Exchange calendar · 읽기 전용</h2><div class="table-scroll" role="region" aria-label="Exchange calendar · 읽기 전용" tabindex="0"><table>
<tr><th>Source</th><td>{{.CalendarSource}}</td></tr><tr><th>Version</th><td><code>{{.CalendarVersion}}</code></td></tr>
<tr><th>Updated at</th><td>{{.CalendarUpdatedAt}}</td></tr></table></div>
<p>6시간 freshness 또는 장 시작 전 refresh 조건이 실패하면 신규 entry는 대기한다. calendar는 이 화면에서 수정할 수 없다.</p></section>
<section aria-labelledby="decision-heading"><h2 id="decision-heading">현재 결정</h2><div class="table-scroll" role="region" aria-label="현재 결정" tabindex="0"><table>
<tr><th>Typed reason</th><td><code>{{.DecisionReason}}</code></td></tr><tr><th>다음 전환</th><td>{{.NextTransition}}</td></tr></table></div>
<p>{{.DecisionHelp}}</p><p><strong>신규 entry가 대기하거나 꺼져 있어도 exit·reconcile·fill detection은 계속된다.</strong></p></section>
{{template "foot" .}}
{{end}}

{{define "strategy-runtime"}}
{{template "head" .}}
<p><a href="/optimization">최적화</a> / <code>strategy-runtime</code> / <a href="/strategy-runtime/market-schedule">시장·일정</a></p>
<h1>한국 주식 전략 lane</h1>
<p class="muted page-intro">StockOS Parker VWAP conservative v1을 옮기는 읽기 전용 상태 카드다. 서버가 관측한 저장값과 적용값만 보여주며 이 화면에는 입력·저장·활성화·LIVE action이 없다.</p>
{{if .LoadErr}}<p class="danger" role="alert">상태를 읽지 못해 신규 진입 OFF를 표시한다.</p>{{end}}
{{if .Unwired}}<p class="notice">strategy-runtime seam 미배선 — dormant 기본값만 표시한다.</p>{{end}}
<nav class="filter-bar" aria-label="전략 런타임 화면">
  <a href="#strategy-observation">관측 상태</a><a href="#strategy-parameters">파라미터</a><a href="#lane-state">lane</a>
  <a href="#autostart-state">자동 기동</a><a href="#live-state">승인·진입</a><a href="#strategy-blockers">blockers</a>
</nav>

<section id="strategy-observation" aria-labelledby="strategy-observation-title">
  <h2 id="strategy-observation-title">관측 상태</h2>
  <dl>
    <dt>Generated at</dt><dd>{{.GeneratedAt}}</dd>
    <dt>Observed at</dt><dd>{{.ObservedAt}}</dd>
    <dt>Freshness</dt><dd><span class="status-pill {{.Freshness.Class}}" data-testid="runtime-freshness">{{.Freshness.Value}}</span></dd>
  </dl>
  <p><a href="/strategy-runtime">상태 새로고침</a> · GET으로 다시 읽기만 하며 설정이나 주문을 보내지 않는다.</p>
</section>

<section id="strategy-parameters" aria-labelledby="strategy-parameters-title">
  <h2 id="strategy-parameters-title">{{.ParameterSection}} · 읽기 전용</h2>
  <p class="muted">각 카드는 서버 소유 descriptor의 help·기본값·저장값·적용값·provenance를 그대로 표시한다.</p>
  {{range .Fields}}
  <article class="detail-grid" aria-labelledby="strategy-field-{{.Key}}" aria-readonly="true">
    <div>
      <h3 id="strategy-field-{{.Key}}">{{.Label}}</h3>
      <p class="muted">{{.Help}}</p>
      <code>{{.Key}}</code>
    </div>
    <dl>
      <dt>기본값</dt><dd>{{.Default}}</dd>
      <dt>Desired · 저장값</dt><dd>{{.Desired}}</dd>
      <dt>Effective · 적용값</dt><dd><span class="status-pill muted">{{.Effective}}</span></dd>
      <dt>단위 / 범위</dt><dd>{{.Unit}} / {{.Range}}</dd>
      <dt>적용 시점</dt><dd>{{.ApplyTiming}}</dd>
      <dt>Provenance</dt><dd><code>{{.Provenance}}</code></dd>
    </dl>
  </article>
  {{end}}
</section>

<section id="lane-state" aria-labelledby="lane-state-title">
  <h2 id="lane-state-title">{{.LaneSection}}</h2>
  <article class="detail-grid" aria-readonly="true">
    <div><h3>krx_parker_vwap_conservative_v1</h3><p class="muted">lane 권한은 자동 기동·gate·LIVE 승인과 별도다.</p></div>
    <dl>
      <dt>기본값</dt><dd><span class="status-pill {{.Lane.Default.Class}}">{{.Lane.Default.Value}}</span></dd>
      <dt>Desired · 저장값</dt><dd><span class="status-pill {{.Lane.Desired.Class}}">{{.Lane.Desired.Value}}</span></dd>
      <dt>Effective · 적용값</dt><dd><span class="status-pill {{.Lane.Effective.Class}}" data-testid="lane-effective">{{.Lane.Effective.Value}}</span></dd>
      <dt>Reason</dt><dd><code>{{.Lane.Reason}}</code></dd>
    </dl>
  </article>
</section>

<section id="autostart-state" aria-labelledby="autostart-state-title">
  <h2 id="autostart-state-title">{{.AutoStartSection}}</h2>
  <article class="detail-grid" aria-readonly="true">
    <div><h3>엔진 기동 시 자동 시작</h3><p class="muted">다음 엔진 기동 시 별도 activation manifest를 다시 검증한다.</p></div>
    <dl>
      <dt>기본값</dt><dd><span class="status-pill {{.AutoStart.Default.Class}}">{{.AutoStart.Default.Value}}</span></dd>
      <dt>Desired · 저장값</dt><dd><span class="status-pill {{.AutoStart.Desired.Class}}">{{.AutoStart.Desired.Value}}</span></dd>
      <dt>Effective · 적용값</dt><dd><span class="status-pill {{.AutoStart.Effective.Class}}" data-testid="autostart-effective">{{.AutoStart.Effective.Value}}</span></dd>
      <dt>Reason</dt><dd><code>{{.AutoStart.Reason}}</code></dd>
    </dl>
  </article>
</section>

<section id="live-state" aria-labelledby="live-state-title">
  <h2 id="live-state-title">{{.LiveSection}}</h2>
  <p class="muted">Automation gate와 LIVE 승인은 독립 권한이다. 일괄 활성화는 없다.</p>
  <article class="detail-grid" aria-labelledby="gate-approval-title" aria-readonly="true">
    <div><h3 id="gate-approval-title">Automation gate</h3><p class="muted">프로그램 주문 gate의 서버 권위 상태다.</p></div>
    <dl>
      <dt>기본값</dt><dd><span class="status-pill {{.GateApproval.Default.Class}}">{{.GateApproval.Default.Value}}</span></dd>
      <dt>Desired · 저장값</dt><dd><span class="status-pill {{.GateApproval.Desired.Class}}">{{.GateApproval.Desired.Value}}</span></dd>
      <dt>Effective · 적용값</dt><dd><span class="status-pill {{.GateApproval.Effective.Class}}" data-testid="gate-effective">{{.GateApproval.Effective.Value}}</span></dd>
      <dt>Reason</dt><dd><code>{{.GateApproval.Reason}}</code></dd>
    </dl>
  </article>
  <article class="detail-grid" aria-labelledby="live-approval-title" aria-readonly="true">
    <div><h3 id="live-approval-title">LIVE approval</h3><p class="muted">사람이 승인한 LIVE 주문 권한의 서버 권위 상태다.</p></div>
    <dl>
      <dt>기본값</dt><dd><span class="status-pill {{.LiveApproval.Default.Class}}">{{.LiveApproval.Default.Value}}</span></dd>
      <dt>Desired · 저장값</dt><dd><span class="status-pill {{.LiveApproval.Desired.Class}}">{{.LiveApproval.Desired.Value}}</span></dd>
      <dt>Effective · 적용값</dt><dd><span class="status-pill {{.LiveApproval.Effective.Class}}" data-testid="live-effective">{{.LiveApproval.Effective.Value}}</span></dd>
      <dt>Reason</dt><dd><code>{{.LiveApproval.Reason}}</code></dd>
    </dl>
  </article>
  <article class="detail-grid" aria-labelledby="entry-capability-title" aria-readonly="true">
    <div><h3 id="entry-capability-title">Entry capability</h3><p class="muted">authority가 확정한 신규 진입 결과다. 이 화면은 blocker를 다시 계산하지 않는다.</p></div>
    <dl>
      <dt>기본값</dt><dd><span class="status-pill {{.EntryDefault.Class}}">{{.EntryDefault.Value}}</span></dd>
      <dt>Desired · 저장값</dt><dd><span class="status-pill {{.EntryDesired.Class}}">{{.EntryDesired.Value}}</span></dd>
      <dt>Effective</dt><dd><span class="status-pill {{.EntryEffective.Class}}" data-testid="entry-effective">{{.EntryEffective.Value}}</span></dd>
      <dt>첫 refusal</dt><dd><code>{{.FirstRefusal}}</code></dd>
    </dl>
  </article>
</section>

<section id="strategy-blockers" aria-labelledby="strategy-blockers-title">
  <h2 id="strategy-blockers-title">Activation blockers</h2>
  <p class="muted">authority가 제공한 순서대로 첫 거부 근거와 freshness를 확인한다.</p>
  {{range .Blockers}}
  <article class="detail-grid" aria-labelledby="strategy-blocker-{{.Key}}" aria-readonly="true">
    <div><h3 id="strategy-blocker-{{.Key}}">{{.Label}}</h3><code>{{.Key}}</code></div>
    <dl>
      <dt>Desired</dt><dd><span class="status-pill {{.Desired.Class}}">{{.Desired.Value}}</span></dd>
      <dt>Effective</dt><dd><span class="status-pill {{.Effective.Class}}">{{.Effective.Value}}</span></dd>
      <dt>Freshness</dt><dd><span class="status-pill {{.Freshness.Class}}">{{.Freshness.Value}}</span></dd>
      <dt>Reason</dt><dd><code>{{.Reason}}</code></dd>
    </dl>
  </article>
  {{end}}
  <p><strong>신규 entry가 OFF여도 exit·reconcile·보호 감독은 계속된다.</strong></p>
</section>
{{template "foot" .}}
{{end}}
`
