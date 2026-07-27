package console

// templates_settings.go is the adoption-settings screen's markup (change
// console-adoption-controls task 3.1). Same rules as every other screen: no
// script, no external asset, every state has words.

const settingsTemplates = `
{{define "settings"}}
{{template "head" .}}
<h1>편입 설정</h1>
<p class="muted">수동 매수 보유를 엔진의 exit 관리(손절·익절·래칫)에 편입하는 규칙을 편집한다.
편입 실행은 <strong>엔진 대사 루프</strong>의 몫이고, 이 화면은 config의 <code>engine.adoption</code>
블록만 기록한다. automation gate(운영 게이트)는 이 콘솔에서 편집할 수 없다 — 게이트 ON은 콘솔 밖
승인 절차다.</p>

{{if .Notice}}<p class="notice">{{.Notice}}</p>{{end}}

{{if not .Wired}}
<section><p class="notice"><strong>저장이 배선되지 않았다.</strong> 이 빌드의 콘솔에는 편입 설정
저장 seam이 주입되지 않았다 — 표시는 되지만 저장할 수 없다. config.json을 직접 편집하라.</p></section>
{{else}}
{{if .LoadErr}}<p class="danger">설정 파일을 읽을 수 없다: <code>{{.LoadErr}}</code></p>{{end}}
{{if .Verdict}}
<p class="danger"><strong>현재 블록은 엔진이 거부한다.</strong> {{.Verdict}}
<br>아래 값은 파일에 적힌 원문 그대로다 — 고쳐서 저장하면 목록은 유실되지 않는다.</p>
{{end}}

<section>
  <h2>규칙</h2>
  <form method="post" action="/settings/save">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <p><label><input type="checkbox" name="enabled" {{if .Block.Enabled}}checked{{end}}>
    <strong>자동 편입</strong> — 계좌의 모든 무기록 보유(수동 매수)를 다음 대사 주기에 편입한다</label></p>
    <p><label>합성 손절폭 — 편입가에서 얼마나 아래에 손절선을 둘지 (기본값 5%)<br>
    <input type="range" name="default_stop_pct" min="0.02" max="0.2" step="0.005"
      value="{{.StopPctSlider}}" style="width:20rem;vertical-align:middle"
      oninput="document.getElementById('pctout').textContent=(Math.round(this.value*1000)/10)+'%'">
    <output id="pctout"><strong>{{.StopPctPercent}}</strong></output></label>
    <span class="muted">(마우스로 조절 — 2% ~ 20%)</span></p>
    <details>
      <summary class="muted">고급 — 목록 직접 편집 (보통은 포지션 화면의 버튼으로 관리한다)</summary>
      <p><label>제외 목록 (쉼표 구분 — 어떤 경로로도 편입하지 않는다, 지정보다 우선)<br>
      <input type="text" name="exclude_symbols" value="{{.Excludes}}"></label></p>
      <p><label>지정 목록 (자동 편입이 꺼져 있어도 이 심볼은 편입한다)<br>
      <input type="text" name="include_symbols" value="{{.Includes}}"></label></p>
    </details>
    <p><button type="submit">저장</button></p>
  </form>
  <p class="muted">저장은 즉시 파일에 기록되지만 <strong>다음 엔진 기동부터 반영</strong>된다
  {{if .EngineRunning}}— <strong>지금 엔진이 실행 중이다</strong>: 기동 시점 설정으로 계속 동작하므로
  반영하려면 엔진을 재시작해야 한다{{end}}.
  지정은 <strong>상시 규칙</strong>이다 — 청산 후 재매수해도 다시 편입된다. 지정 해제·제외 추가는
  <strong>이미 편입된 포지션에 아무 효과가 없다</strong>(편입 해제 기능은 존재하지 않는다 — 가용
  수단은 사전 제외·자동 편입 끄기·flatten뿐이다). 편입된 보유는 편입일 관측가를 t0으로 관리되며
  +0.8R부터 편입가 기준 본전이 보호 바닥이 된다.</p>
</section>
{{end}}
{{template "foot" .}}
{{end}}
`
