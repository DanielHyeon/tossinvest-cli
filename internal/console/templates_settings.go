package console

// templates_settings.go is the adoption-settings screen's markup (change
// console-adoption-controls task 3.1). Same rules as every other screen: no
// script, no external asset, every state has words.

const settingsTemplates = `
{{define "settings"}}
{{template "head" .}}
<h1>편입 설정</h1>
<p class="muted">수동 매수 보유를 엔진의 exit 관리(손절·익절·래칫)에 편입하는 규칙과, Guardian이
자동 진입에 적용하는 한도를 편집한다. 실행은 <strong>엔진</strong>의 몫이고 이 화면은 config의
<code>engine.adoption</code> 블록과 게이트 블록의 <strong>한도</strong>만 기록한다.
automation gate(운영 게이트)의 ON/OFF와 kill switch는 이 콘솔에서 편집할 수 없다 — 게이트 ON은 콘솔 밖
승인 절차이고, 한도 저장은 그 스위치의 바이트에 손대지 않는다.</p>

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

<section id="guardian-limits">
  <h2>Guardian 한도</h2>
  <p class="muted">자동 진입 하나가 얼마나 커질 수 있는지, 계좌가 얼마나 열릴 수 있는지, 하루에 얼마까지
  잃으면 진입이 멈추는지. 다섯 값은 <strong>진입</strong>에만 적용된다 — 청산은 RISK_REDUCING이라 이
  한도의 대상이 아니므로, 여기를 아무리 조여도 손절·비상 청산은 느려지지 않는다.</p>

  {{if .LimitsLoadErr}}<p class="danger">한도를 읽을 수 없다: <code>{{.LimitsLoadErr}}</code></p>{{end}}

  <table id="guardian-current">
    <tr><th>한도</th><th>설정값</th></tr>
    {{range .LimitRows}}<tr><td>{{.Label}}</td><td>{{.Value}}{{if .Unit}} {{.Unit}}{{end}}</td></tr>
    {{end}}
    <tr><td>한도 통화</td><td>{{.LimitCurrencyText}}</td></tr>
  </table>

  {{if .LimitsUnset}}
  <p class="notice">다섯 값이 <strong>미설정</strong>이다. 게이트가 켜져 있어도 엔진은 이 상태로 기동하지
  않는다. 아래 프리셋 하나를 누르면 다섯 값이 한 번에 기록된다 — 화면은 파일에 없는 숫자를 미리
  그리지 않으므로, 누르기 전까지 엔진이 보는 것도 이 미설정 그대로다.</p>
  {{else if .LimitsPartlyConfigured}}
  <p class="danger"><strong>일부만 설정돼 있다.</strong> 기동 인터록은 다섯 값이 전부 양수·유한하고
  통화가 있어야 통과시키므로 지금 상태로는 <strong>기동을 거부</strong>한다({{.LimitVerdict}}).
  콘솔은 나머지를 임의로 채우지 않는다 — 부분적으로 무제한인 게이트를 조용히 완성하는 것이기
  때문이다. <strong>프리셋</strong> 하나로 다섯 값을 함께 맞추거나, 고급에서 직접 채운다.</p>
  {{else}}
  <p class="muted">현재 값은 {{with .MatchingTier}}<strong>{{.}}</strong>와 일치한다{{else}}
  <strong>사용자 지정값</strong>이다{{end}} — 레지스트리와 대조한 결과이고, 누가 골랐는지에 대한
  주장은 아니다(같은 숫자를 손으로 적었을 수도 있다).</p>
  {{end}}

  <p class="muted">automation gate는 지금 <strong>{{if .Gate.Enabled}}ON{{else}}OFF{{end}}</strong>이다.
  스위치는 <a href="#operating">아래 운영 섹션</a>에 있고, 이 섹션의 저장은 그 키의 바이트를
  건드리지 않는다.</p>

  {{if not .LimitsWired}}
  <p class="notice"><strong>한도 저장이 배선되지 않았다.</strong> 이 빌드의 콘솔에는 Guardian 한도 저장
  seam이 주입되지 않았다 — 표시는 되지만 저장할 수 없다. config.json을 직접 편집하라.</p>
  {{else}}
  <h3>프리셋</h3>
  <p class="muted">StockOS <code>risk_profiles.py</code>의 티어를 옮긴 것이다. 클릭 한 번이 다섯 값과
  통화를 함께 기록한다. <strong>등록된 티어 위로는 올릴 수 없다</strong> — 상한은 그 통화에 등록된
  티어의 필드별 최대이고, 그보다 큰 값은 저장이 거부된다.</p>
  {{range .LimitTierCards}}
  <form method="post" action="/settings/limits/preset" onsubmit="return confirm('{{.Label}} 한도를 기록한다. 다음 엔진 기동부터 반영된다.')">
    <input type="hidden" name="csrf" value="{{$.CSRF}}">
    <input type="hidden" name="tier" value="{{.ID}}">
    <p><button type="submit">{{.Label}}{{if .Recommended}} · 권장{{end}}</button>
    <span class="muted">[{{.Limits.Currency}}] {{.Summary}}</span></p>
  </form>
  {{end}}

  <details>
    <summary class="muted">고급 — 개별 값 직접 기입 (보통은 위의 프리셋으로 관리한다)</summary>
    <p class="muted">낮추는 방향은 양수인 한 자유롭다. 다섯 값 중 하나라도 비거나 0이면 저장이
    거부된다 — 엔진이 기동을 거부할 블록은 기록하지 않는다.</p>
    <form method="post" action="/settings/limits">
      <input type="hidden" name="csrf" value="{{.CSRF}}">
      <p><label>1회 주문 수량 상한 (주)<br>
      <input type="text" name="max_order_quantity" value="{{.FieldQuantity}}"></label></p>
      <p><label>1회 주문 금액 상한<br>
      <input type="text" name="max_order_notional" value="{{.FieldNotional}}"></label></p>
      <p><label>계좌 개방 노출 상한<br>
      <input type="text" name="max_total_exposure" value="{{.FieldExposure}}"></label></p>
      <p><label>일일 손실 한도(금액)<br>
      <input type="text" name="max_daily_loss_amount" value="{{.FieldDailyLoss}}"></label></p>
      <p><label>일일 손실 한도(비율 — 0.01 = 1%)<br>
      <input type="text" name="max_daily_loss_ratio" value="{{.FieldRatio}}"></label></p>
      <p><label>한도 통화 (KRW 또는 USD)<br>
      <input type="text" name="limit_currency" value="{{.FieldCurrency}}"></label></p>
      <p><button type="submit">한도 저장</button></p>
    </form>
  </details>
  {{end}}

  <p class="muted">한도 통화는 하나다. Guardian 체인은 한도 통화와 다른 통화의 진입을 거부하므로
  <strong>USD 한도를 기록하면 국내 자동 진입이, KRW 한도를 기록하면 미국 자동 진입이 닫힌다</strong>.
  저장은 즉시 파일에 기록되지만 <strong>다음 엔진 기동부터 반영</strong>된다
  {{if .EngineRunning}}— <strong>지금 엔진이 실행 중이다</strong>: 기동 시점 설정으로 계속 동작한다{{end}}.</p>
</section>

<section id="operating">
  <h2>운영</h2>
  <p class="muted">엔진이 실제로 돌기 위해 필요한 마지막 두 가지다. 이전에는 둘 다
  <code>config.json</code> 손편집이었다 — 그 경로는 한도가 다 설정됐는지도, JSON이 여전히 유효한지도
  확인하지 않고 audit에도 남기지 않는다.</p>

  <h3>거래 정책</h3>
  <p class="muted">엔진의 청산 경로가 실제로 쓰는 네 개다. <code>amend</code>·<code>conditional</code>·
  <code>fractional</code>은 이 빌드의 어느 루프도 쓰지 않으므로 여기 없고, 저장은 그 세 값을
  파일에 적힌 그대로 둔다.</p>

  {{if .TradingLoadErr}}<p class="danger">거래 정책을 읽을 수 없다: <code>{{.TradingLoadErr}}</code></p>{{end}}

  {{if not .TradingWired}}
  <p class="notice"><strong>거래 정책 저장이 배선되지 않았다.</strong> 표시는 되지만 저장할 수 없다.</p>
  <table><tr><th>토글</th><th>상태</th></tr>
  {{range .TradingRows}}<tr><td>{{.Label}}</td><td>{{if .On}}켜짐{{else}}꺼짐{{end}}</td></tr>{{end}}
  </table>
  {{else}}
  <form method="post" action="/settings/trading">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    {{range .TradingRows}}
    <p><label><input type="checkbox" name="{{.Key}}" {{if .On}}checked{{end}}>
    <strong>{{.Label}}</strong></label><br>
    <span class="muted">{{.Why}}</span></p>
    {{end}}
    <p><button type="submit">거래 정책 저장</button></p>
  </form>
  <p class="muted">네 개가 모두 켜져야 기동 인터록 3절을 통과한다. 이 조합은 "엔진이 내 주식을
  팔아도 된다"는 승인이다 — 켜는 것은 사람이고, 저장은 audit에 남는다.</p>
  {{end}}

  <h3>자동화 게이트</h3>
  <p class="muted">지금 <strong>{{if .Gate.Enabled}}ON{{else}}OFF{{end}}</strong>. 이 스위치는
  <code>engine.automation_gate.enabled</code> 한 키만 쓴다 — 다섯 한도의 바이트는 건드리지 않는다.</p>

  <table>
    <tr><th>기동 인터록 사전 판정</th><th>결과</th></tr>
    {{range .GatePreflight}}
    <tr><td>{{.Name}}<br><span class="muted">{{.Detail}}</span></td>
    <td>{{if not .Judgeable}}판정 불가{{else if .OK}}통과{{else}}<strong>미충족</strong>{{end}}</td></tr>
    {{end}}
  </table>
  <p class="muted">이 판정은 <strong>파일이 답하는 절에 한한다</strong>. Guardian 주입·게이트웨이
  배선·전송의 키 운반은 구성된 프로세스에서만 알 수 있으므로 화면이 판정하지 못한다 —
  <strong>사전 판정을 통과해도 기동이 보장되지 않는다.</strong> 최종 판단은 엔진의 인터록이고,
  거부하면 미충족 항목을 열거한다.</p>

  {{if .GateBlockers}}
  <p class="danger"><strong>지금 켜면 기동이 거부된다.</strong> 미충족: {{range $i, $c := .GateBlockers}}{{if $i}}, {{end}}{{$c}}{{end}}.
  기록은 되지만 엔진은 뜨지 않는다.</p>
  {{end}}

  <p class="notice"><strong>게이트를 켜면 시작되는 것.</strong> 다음 엔진 기동에서 대사·exit 관측·
  체결 감지 세 루프가 돈다. 그 중 대사 루프의 마지막 단계가 <strong>편입</strong>이고,
  <strong>편입은 되돌릴 수 없다</strong>(해제 기능은 존재하지 않는다 — 가용 수단은 사전 제외·자동
  편입 끄기·flatten뿐이다).
  {{if .AdoptionArmed}}지금 설정으로는 <strong>첫 대사 주기에 편입이 일어난다</strong>.
  현재 제외 목록: <strong>{{.AdoptionExcludes}}</strong>.{{else}}지금은 자동 편입이 꺼져 있고 지정
  목록도 비어 있어 편입 대상이 없다.{{end}}
  보호는 <strong>이 프로세스가 살아 있는 동안만</strong> 유효하다 — 브로커에 손절 주문이 남지
  않으므로 프로세스가 죽으면 보호도 사라진다. 노출을 늘리는 주문은 게이트웨이에서 계속 거부된다.</p>

  {{if not .GateWired}}
  <p class="notice"><strong>게이트 저장이 배선되지 않았다.</strong> 표시는 되지만 저장할 수 없다.</p>
  {{else}}
  <form method="post" action="/settings/gate">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <p><label><input type="checkbox" name="enabled" {{if .Gate.Enabled}}checked{{end}}>
    <strong>자동화 게이트 ON</strong></label></p>
    <p><button type="submit">게이트 저장</button></p>
  </form>
  {{end}}
</section>
{{template "foot" .}}
{{end}}
`
