package console

const positionPolicyTemplates = `
{{define "position-policy"}}
{{template "head" .}}
<h1>포지션 정책 <span class="muted">관리 상태</span></h1>
<p class="muted page-intro">StockOS lane console처럼 범위·현재값·실효값·적용 시점을 한 화면에서 확인한다.
직접 입력은 없고 서버가 제공한 현재 행 action과 preset만 선택한다.</p>
{{if .Notice}}<p class="notice" role="status">{{.Notice}}</p>{{end}}
{{if .LoadErr}}<p class="danger" role="alert">불러오기 실패: <code>{{.LoadErr}}</code></p>{{end}}
{{if not .Wired}}<p class="notice">engine-owned PositionPolicyCommander가 배선되지 않아 조회만 가능하다.</p>{{end}}

<section aria-labelledby="auto-adoption-heading">
<h2 id="auto-adoption-heading">{{.Descriptor.AutoAdoptionSection}}</h2>
<p class="muted">기본 정책, 파일에 저장된 desired, 실행 중 엔진이 시작할 때 읽은 effective를 분리해 표시한다.
desired와 effective가 달라도 어느 한쪽으로 덮어쓰지 않는다.</p>
<table class="data-table"><caption>자동편입 설정의 기본값·저장값·실행값</caption><thead><tr>
<th scope="col">구분</th><th scope="col">enabled</th><th scope="col">합성 손절</th><th scope="col">include</th><th scope="col">exclude</th>
</tr></thead><tbody>
<tr><th scope="row" data-label="구분">기본값</th><td data-label="enabled">OFF</td><td data-label="합성 손절">{{.Descriptor.StopDefault}}</td><td data-label="include">없음</td><td data-label="exclude">없음</td></tr>
<tr><th scope="row" data-label="구분">저장값 · desired</th><td data-label="enabled"><strong>{{.Desired.EnabledText}}</strong></td><td data-label="합성 손절"><strong>{{.Desired.StopText}}</strong></td>
<td data-label="include"><code>{{.Desired.IncludesText}}</code></td><td data-label="exclude"><code>{{.Desired.ExcludesText}}</code></td></tr>
<tr><th scope="row" data-label="구분">실행값 · effective</th><td data-label="enabled"><strong>{{.Effective.EnabledText}}</strong></td><td data-label="합성 손절"><strong>{{.Effective.StopText}}</strong></td>
<td data-label="include"><code>{{.Effective.IncludesText}}</code></td><td data-label="exclude"><code>{{.Effective.ExcludesText}}</code></td></tr>
</tbody></table>
{{if .DesiredErr}}<p class="notice" role="status">desired 알 수 없음 · <code>{{.DesiredErr}}</code></p>{{end}}
{{if .DesiredVerdict}}<p class="danger" role="alert">desired validation 거부 · <code>{{.DesiredVerdict}}</code></p>{{end}}
{{if .RuntimeErr}}<p class="notice" role="status">effective 알 수 없음 · <code>{{.RuntimeErr}}</code></p>{{end}}
<p class="muted">손절 option 2~20%, 0.5% step · {{.Descriptor.ExcludePrecedence}} · 1주: {{.Descriptor.OneShareBehavior}} · 적용 시점: {{.Descriptor.ApplyTiming}}</p>

<h3>실행 중 편입 대사 차단</h3>
{{if .Blocks}}
<p class="muted">source <code>{{.BlockSource}}</code>. 아래는 편입을 실제로 보류하는 tracker projection이며 원장 전체 cause 목록이 아니다.</p>
<div class="cards">{{range .Blocks}}<article class="card"><h3>{{.Target}}</h3>
<dl><dt>scope</dt><dd><code>{{.Scope}}</code></dd><dt>reason</dt><dd><code>{{.Reason}}</code></dd>
<dt>detail</dt><dd>{{if .Detail}}{{.Detail}}{{else}}—{{end}}</dd><dt>시작·경과</dt><dd>{{.StartedAt}} · {{.Age}}</dd>
<dt>지속성</dt><dd>{{if .Permanent}}<strong class="danger">영구 차단</strong>{{else}}일시 차단{{end}}</dd></dl></article>{{end}}</div>
{{else}}<p class="notice">현재 runtime projection에 active 대사 차단이 없다.</p>{{end}}
</section>

<section aria-labelledby="position-policy-heading">
<h2 id="position-policy-heading">{{.Descriptor.PositionSection}}</h2>
<p>기본값 <strong>공통 정책 상속</strong> · desired는 generation별 override · effective는 저장된 exit snapshot이다.
적용 시점: <strong>{{.Descriptor.ApplyTiming}}</strong> · provenance: <code>{{.Descriptor.Provenance}}</code>.</p>
{{if .Rows}}
<table class="data-table"><caption>현재 포지션 policy scope</caption><thead><tr>
<th>계좌·종목</th><th>generation</th><th>현재 상태</th><th>desired/effective</th><th>현재 행 action</th>
</tr></thead><tbody>{{range .Rows}}<tr>
<th scope="row" data-label="계좌·종목"><code>{{.State.AccountRef}}</code><span class="submetric">{{.State.Market}} · {{.State.Symbol}}{{if .Name}} · {{.Name}}{{end}}</span></th>
<td data-label="generation">{{.State.AdoptionGeneration}} <span class="submetric">version {{.State.Version}}</span></td>
<td data-label="현재 상태"><span class="status-pill">{{.Management.Label}}</span>
<span class="submetric">status <code>{{.Management.Status}}</code> · reason <code>{{.Management.Reason}}</code></span>
<span class="submetric">policy lifecycle {{.State.Status}} · provenance {{.State.Provenance}} · eligibility {{.State.Eligibility}}</span>
{{if .Block.Present}}<span class="submetric danger">대사 차단 · {{.Block.Target}} · {{.Block.Reason}}{{if .Block.Detail}} · {{.Block.Detail}}{{end}} · {{.Block.Age}}{{if .Block.Permanent}} · 영구 차단{{end}}</span>{{end}}
{{if .State.HasPendingExit}}<span class="submetric danger">active exit 있음</span>{{end}}</td>
<td data-label="desired/effective"><strong>{{.State.DesiredLabel}}</strong><span class="submetric">effective {{.State.EffectiveLabel}}</span></td>
<td data-label="현재 행 action"><div class="actions">{{range .Actions}}<form method="post" action="/position-management/preview">
<input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="action_token" value="{{.Token}}">
<button type="submit" {{if .Danger}}class="secondary"{{end}} aria-label="{{.Label}} preview">{{.Label}}</button></form>{{end}}</div></td>
</tr>{{end}}</tbody></table>
{{else}}<p class="notice">관리할 현재 포지션 행이 없다.</p>{{end}}
</section>
<details class="explain"><summary>서버 제공 손절 option 37개 보기</summary><div class="filter-group">{{range .Descriptor.StopOptions}}<span class="status-pill" data-option-id="{{.ID}}">{{.Label}}</span>{{end}}</div></details>
<p class="muted">이 change는 LIVE/automation toggle을 켜지 않는다. 외부 편입 ON/OFF는 canonical control plane 승인과 함께 별도 적용된다.</p>
{{template "foot" .}}
{{end}}

{{define "position-policy-preview"}}
{{template "head" .}}
<h1>정책 변경 preview</h1>
<section><dl><dt>범위</dt><dd><code>{{.Preview.Before.AccountRef}}</code> · {{.Preview.Before.Market}} · {{.Preview.Before.Symbol}}</dd>
<dt>generation/version</dt><dd>{{.Preview.Before.AdoptionGeneration}} / {{.Preview.Before.Version}}</dd>
<dt>before</dt><dd>{{.Preview.Before.Status}} · desired {{.Preview.Before.DesiredLabel}} · effective {{.Preview.Before.EffectiveLabel}}</dd>
<dt>after</dt><dd>{{.Preview.After.Status}} · desired {{.Preview.After.DesiredLabel}} · effective {{.Preview.After.EffectiveLabel}}</dd>
<dt>reason</dt><dd><code>{{.Preview.Reason}}</code> (server-defined)</dd>
<dt>적용 효과</dt><dd>{{if .Release}}현재 generation을 exit observer 대상에서 제거한다. 보존된 과거 snapshot은 다시 실행되지 않는다.{{else if .Readopt}}엔진 관측가로 fresh t0와 합성 손절을 만들고 high-water·rung·pending을 초기화한 새 generation을 시작한다.{{else}}기존 exit snapshot과 LIVE toggle은 변경하지 않는다.{{end}}</dd>
<dt>restart</dt><dd>불필요 · LIVE toggle은 변경하지 않음</dd></dl></section>
{{if .Danger}}<p class="danger" role="alert"><strong>보호 공백 위험.</strong> active exit가 있으면 409로 거부된다.
이 화면을 읽고 최소 {{.WaitSecs}}초 후 명시적으로 체크한 뒤 진행하라. 문구 입력은 요구하지 않는다.</p>{{end}}
{{if not .Danger}}<p class="notice">일반 정책 변경은 추가 대기나 입력 없이 이 preview 그대로 적용된다.</p>{{end}}
<form method="post" action="/position-management/apply">
<input type="hidden" name="csrf" value="{{.CSRF}}"><input type="hidden" name="capability" value="{{.Token}}">
{{if .Danger}}<label><input type="checkbox" name="confirm" value="yes" required> 보호 공백 설명을 확인했다</label>{{end}}
<button type="submit">{{if .Danger}}3초 확인 후 danger action 적용{{else}}preview 그대로 적용{{end}}</button></form>
<p><a href="/position-management">취소하고 현재값 다시 보기</a></p>
{{template "foot" .}}
{{end}}
`
