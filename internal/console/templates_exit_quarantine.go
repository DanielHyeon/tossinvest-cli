package console

// The release preview lives in its own template source (change a069) so the
// position-policy set keeps saying one thing. Both are parsed into the same
// template chain in pages.go and share "head", "foot" and the one stylesheet.

const exitQuarantineTemplates = `
{{define "quarantine-release-preview"}}
{{template "head" .}}
<h1>판정 격리 해제 preview</h1>
<section><dl>
<dt>종목</dt><dd>{{.Row.Market}} · {{.Row.Symbol}}</dd>
<dt>generation / quarantine version</dt><dd>{{.Row.Generation}} / v{{.Row.Version}}</dd>
<dt>격리 사유</dt><dd>{{.Reason}} <span class="submetric"><code>{{.Row.Reason}}</code></span></dd>
<dt>기록된 증거</dt><dd><code>{{if .Row.Evidence}}{{.Row.Evidence}}{{else}}없음{{end}}</code></dd>
<dt>격리 시각</dt><dd>{{.Row.QuarantinedAt}}</dd>
<dt>해제 후 유지되는 보호선</dt><dd>{{if .Row.Protection}}<strong>{{.Row.Protection}}</strong>{{else}}알 수 없음 <span class="submetric"><code>{{.Row.ProtectionUnknown}}</code></span>{{end}}</dd>
<dt>바뀌지 않는 것</dt><dd>진입가 · 초기 손절 · high-water · 활성 rung · adoption generation. 이 해제는 lifecycle을 건드리지 않는다.</dd>
<dt>적용 효과</dt><dd>다음 관측 주기부터 이 포지션이 판정 대상에 돌아온다. 저장된 기준선으로 손절 평가가 재개된다.</dd>
</dl></section>
<p class="danger" role="alert"><strong>해제는 수리가 아니라 재시도다.</strong>
격리 원인이 그대로면 다음 판정이 같은 결론을 내고 새 version으로 <strong>다시 격리된다</strong>.
이 화면을 읽고 최소 {{.WaitSecs}}초 후 체크한 뒤 진행하라. 문구 입력은 요구하지 않는다.</p>
<form method="post" action="/position-management/quarantine/apply">
<input type="hidden" name="csrf" value="{{.CSRF}}"><input type="hidden" name="capability" value="{{.Token}}">
<label><input type="checkbox" name="confirm" value="yes" required> 재격리 가능성과 유지되는 기준선을 확인했다</label>
<button type="submit">3초 확인 후 판정 격리 해제</button></form>
<p><a href="/position-management">취소하고 현재값 다시 보기</a></p>
{{template "foot" .}}
{{end}}
`
