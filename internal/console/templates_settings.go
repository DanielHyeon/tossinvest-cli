package console

// templates_settings.go is the settings screen's markup, in four tabs (change
// a055-console-settings-cadence).
//
// Same rules as every other screen: no script, no external asset, every state
// has words. Two rules are new here.
//
//	one card, one shape    every form says 현재값 → 적용 후 → 사유 → 결과, in that
//	                       order, with the same markup. A card either renders a
//	                       submit control or renders a NAMED reason where one
//	                       would be — never an empty space.
//	one control, one tab   a control appears on exactly one tab. A value another
//	                       tab owns is still displayed here, read-only, with a
//	                       link to the tab that has the switch: hiding the
//	                       relationship between the gate and the limits would be
//	                       a worse screen than the one this replaces.
//
// The inline `onsubmit="return confirm(…)"` handlers are gone. The deployed CSP
// is default-src 'none' with no script-src, so none of them ever ran — they were
// the appearance of a confirmation and not a confirmation. What replaces them is
// the 적용 후 preview, which is server-rendered and therefore actually visible.

const settingsTemplates = `
{{define "settings"}}
{{template "head" .}}
<h1>설정 <span class="muted">{{.CurrentTab.Label}}</span></h1>
<p class="page-intro">{{.CurrentTab.Lead}}</p>

{{/*
  The tab bar carries aria-current="page" because it IS the current page. The
  screen-name check reads the top navigation by its aria-label, not by the first
  aria-current in the document, so the two cannot be confused.
*/}}
<nav class="settings-tabs" aria-label="설정 분류">
  {{range .Tabs}}<a href="{{.Path}}"{{if eq .Key $.Tab}} class="on" aria-current="page"{{end}}>{{.Label}}</a>{{end}}
</nav>

{{/*
  The three facts anyone changing a setting needs in front of them. Each links to
  the tab that owns the switch when this tab only displays the value.
*/}}
<dl class="console-summary settings-state" aria-label="현재 운영 상태">
  {{range .OperatingState}}
  <div><dt>{{.Label}}</dt>
    <dd data-state-tone="{{.Tone}}">{{.Value}}{{if .Href}}<small><a href="{{.Href}}">스위치가 있는 탭으로</a></small>{{end}}</dd></div>
  {{end}}
</dl>

{{with .OrphanNotice}}<p class="notice">{{.}}</p>{{end}}

{{if eq .Tab "standing"}}{{template "settings-standing" .}}
{{else if eq .Tab "daily"}}{{template "settings-daily" .}}
{{else if eq .Tab "strategy"}}{{template "settings-strategy" .}}
{{else}}{{template "settings-tools" .}}{{end}}
{{template "foot" .}}
{{end}}

{{/*
  "cardblocked" is what a card renders in place of its submit control.

  This console does not disable a form whose seam is unwired — it declines to
  render it — so "the save surface is missing" is the normal case and an empty
  space is the failure. Every one of these carries the reason's NAME in an
  attribute, which is what makes the requirement checkable: a check that looked
  for a disabled attribute would find nothing and pass.
*/}}
{{define "cardblocked"}}
{{range .Blocks}}<p class="notice" data-save-block="{{.Name}}"><strong>{{.Name}}</strong> — {{.Detail}}</p>
{{end}}{{end}}

{{/* "cardcautions" is a reason that does NOT stop the save. The engine has the
     last word; the screen says what it already knows before the click. */}}
{{define "cardcautions"}}
{{range .Cautions}}<p class="notice" data-save-caution="{{.Name}}"><strong>{{.Name}}</strong> — {{.Detail}}</p>
{{end}}{{end}}

{{/* --- 상시 — 비가역 · 주 1회 미만 --------------------------------------------- */}}

{{define "settings-standing"}}
<section id="adoption" data-settings-card="adoption">
  <h2>외부 종목 편입 규칙</h2>
  <p>내가 직접 <strong>수동 매수</strong>한 무기록 보유를 엔진 관리에 편입하고, 편입 완료 후
  <strong>기존 공통 익절·보호선·손익 극대화 정책</strong>을 그대로 적용하는 규칙이다.
  <strong>저장 자체는 편입이나 주문을 실행하지 않는다</strong> — 이 화면은 계좌·주문·원장을
  건드리지 않고 <code>engine.adoption</code> 설정만 기록하며, 실제 편입은 다음 엔진 기동의
  <strong>엔진 대사 루프</strong>가 수행한다. 이 편입 카드의 저장으로는 automation gate(운영 게이트)의
  ON/OFF와 kill switch를 <strong>콘솔에서 편집할 수 없다</strong> — 게이트는 아래의 별도 승인 표면이다.</p>

  {{with $.NoticeFor "adoption"}}<p class="notice" data-save-result="adoption">{{.}}</p>{{end}}
  {{if .LoadErr}}<p class="danger">설정 파일을 읽을 수 없다: <code>{{.LoadErr}}</code></p>{{end}}
  {{if .Verdict}}<p class="danger"><strong>현재 블록은 엔진이 거부한다.</strong> {{.Verdict}}
  <br>아래 값은 파일에 적힌 원문 그대로다 — 고쳐서 저장하면 목록은 유실되지 않는다.</p>{{end}}

  <dl class="card-current" aria-label="편입 규칙 현재값">
    <dt>자동 편입</dt><dd>{{if .Block.Enabled}}켜짐{{else}}꺼짐{{end}}</dd>
    <dt>합성 손절폭</dt><dd>{{.StopPctPercent}}</dd>
    <dt>제외 목록</dt><dd>{{.AdoptionExcludes}}</dd>
    <dt>지정 목록</dt><dd>{{if .Includes}}{{.Includes}}{{else}}없음{{end}}</dd>
  </dl>

  {{if .AdoptionGuard.Blocked}}{{template "cardblocked" .AdoptionGuard}}
  {{else}}
  <form method="post" action="/settings/save">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <p><label><input type="checkbox" name="enabled" {{if .Block.Enabled}}checked{{end}}>
    <strong>자동 편입</strong> — 계좌의 모든 무기록 보유(수동 매수)를 다음 대사 주기에 편입한다</label></p>
    <p><label>합성 손절폭 — 편입가에서 얼마나 아래에 손절선을 둘지<br>
    <input type="number" name="default_stop_percent" min="2" max="20" step="0.5"
      value="{{.StopPctSlider}}" style="width:8rem"> %</label>
    <span class="muted">(2% ~ 20%, 0.5% 단위 · 기본값 5%)</span></p>
    <details class="explain">
      <summary>고급 — 목록 직접 편집 (보통은 포지션 화면의 버튼으로 관리한다)</summary>
      <p><label>제외 목록 (쉼표 구분 — 어떤 경로로도 편입하지 않는다, 지정보다 우선)<br>
      <input type="text" name="exclude_symbols" value="{{.Excludes}}"></label></p>
      <p><label>지정 목록 (자동 편입이 꺼져 있어도 이 심볼은 편입한다)<br>
      <input type="text" name="include_symbols" value="{{.Includes}}"></label></p>
    </details>
    <dl class="card-preview" aria-label="편입 규칙 적용 후">
      <dt>적용 후</dt><dd>체크·숫자·목록이 <code>engine.adoption</code>에 그대로 기록되고 audit에 남는다</dd>
      <dt>실행되는 것</dt><dd>없음 — 저장은 편입도 주문도 하지 않는다</dd>
    </dl>
    <p><button type="submit">편입 규칙 저장</button></p>
  </form>
  {{end}}
  {{template "cardcautions" .AdoptionGuard}}
  <p class="notice" data-effect-timing="adoption">반영 시점 — {{.EffectNotice}}</p>

  <details class="explain">
    <summary>편입의 귀결과 이 규칙의 경계</summary>
    <p class="muted">자동 편입 ON은 제외 종목을 뺀 모든 외부 보유를 후보로 만들고, 종목별 지정은
    자동 편입이 꺼져 있어도 선택한 심볼만 후보로 만든다. 제외가 항상 우선한다. 지정은
    <strong>상시 규칙</strong>이다 — 청산 후 재매수해도 다시 편입된다. 지정 해제·제외 추가는
    이미 편입된 포지션에 아무 효과가 없다(편입 해제 기능은 존재하지 않는다 — 가용 수단은
    사전 제외·자동 편입 끄기·flatten뿐이다). 편입된 보유는 편입일 관측가를 t0으로 관리되며
    +0.8R부터 편입가 기준 본전이 보호 바닥이 된다. 이 편입 섹션의 저장으로는 automation
    gate와 kill switch를 편집할 수 없다.</p>
  </details>
</section>

<section id="gate" data-settings-card="gate">
  <h2>자동화 게이트</h2>
  <p>지금 <strong>{{if .Gate.Enabled}}ON{{else}}OFF{{end}}</strong>. 이 스위치는
  <code>engine.automation_gate.enabled</code> 한 키만 쓴다 — 다섯 한도의 바이트는 건드리지 않는다.</p>

  {{with $.NoticeFor "gate"}}<p class="notice" data-save-result="gate">{{.}}</p>{{end}}

  <div class="table-scroll" role="region" aria-label="기동 인터록 사전 판정" tabindex="0"><table>
    <tr><th>기동 인터록 사전 판정</th><th>결과</th></tr>
    {{range .GatePreflight}}
    <tr><td>{{.Name}}<br><span class="muted">{{.Detail}}</span></td>
    <td>{{if not .Judgeable}}판정 불가{{else if .OK}}통과{{else}}<strong>미충족</strong>{{end}}</td></tr>
    {{end}}
  </table></div>
  <p class="notice">이 판정은 <strong>파일이 답하는 절에 한한다</strong>. Guardian 주입·게이트웨이
  배선·전송의 키 운반은 구성된 프로세스에서만 알 수 있으므로 화면이 판정하지 못한다 —
  <strong>사전 판정을 통과해도 기동이 보장되지 않는다.</strong> 최종 판단은 엔진의 인터록이다.</p>

  <p class="notice"><strong>게이트를 켜면 시작되는 것.</strong> 다음 엔진 기동에서 대사·exit 관측·
  체결 감지 세 루프가 돈다. 그 중 대사 루프의 마지막 단계가 <strong>편입</strong>이고,
  <strong>편입은 되돌릴 수 없다</strong>.
  {{if .AdoptionArmed}}지금 설정으로는 <strong>첫 대사 주기에 편입이 일어난다</strong>.
  현재 제외 목록: <strong>{{.AdoptionExcludes}}</strong>.{{else}}지금은 자동 편입이 꺼져 있고 지정
  목록도 비어 있어 편입 대상이 없다.{{end}}
  보호는 <strong>이 프로세스가 살아 있는 동안만</strong> 유효하다 — 브로커에 손절 주문이 남지
  않으므로 프로세스가 죽으면 보호도 사라진다. 노출을 늘리는 주문은 게이트웨이에서 계속 거부된다.</p>

  {{if .GateGuard.Blocked}}{{template "cardblocked" .GateGuard}}
  {{else}}
  <form method="post" action="/settings/gate">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <p><label><input type="checkbox" name="enabled" {{if .Gate.Enabled}}checked{{end}}>
    <strong>자동화 게이트 ON</strong></label></p>
    <dl class="card-preview" aria-label="게이트 적용 후">
      <dt>적용 후</dt><dd>{{if .Gate.Enabled}}ON → 체크 상태대로{{else}}OFF → 체크 상태대로{{end}} 기록된다</dd>
      <dt>실행되는 것</dt><dd>없음 — 저장은 엔진을 기동하지 않는다</dd>
    </dl>
    <p><button type="submit">게이트 저장</button></p>
  </form>
  {{end}}
  {{template "cardcautions" .GateGuard}}
  <p class="notice" data-effect-timing="gate">반영 시점 — {{.EffectNotice}}</p>
</section>

<section id="autostart" data-settings-card="autostart">
  <h2>엔진 자동 시작</h2>
  <p>지금 <strong>{{if .Autostart}}ON{{else}}OFF{{end}}</strong>.
  이 값은 <code>engine.autostart</code> 한 키만 기록한다. ON 저장 직후 엔진 기동을 한 번
  시도하고, 이후 TossOS 콘솔이 부팅·재기동될 때도 같은 시도를 한다.
  OFF로 저장해도 현재 엔진은 정지하지 않는다 — 현재 프로세스를 끄려면 검증 콘솔의
  [엔진 정지]를 쓴다.</p>

  {{with $.NoticeFor "autostart"}}<p class="notice" data-save-result="autostart">{{.}}</p>{{end}}
  {{if .AutostartLoadErr}}
  <p class="danger">엔진 자동 시작 설정을 읽을 수 없다: <code>{{.AutostartLoadErr}}</code></p>{{end}}

  {{if .AutostartGuard.Blocked}}{{template "cardblocked" .AutostartGuard}}
  {{else}}
  <form method="post" action="/settings/autostart">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <p><label><input type="checkbox" name="enabled" {{if .Autostart}}checked{{end}}>
    <strong>엔진 자동 시작 ON</strong></label></p>
    <dl class="card-preview" aria-label="자동 시작 적용 후">
      <dt>적용 후</dt><dd>체크 상태가 기록되고, ON이면 즉시 기동을 <strong>한 번</strong> 시도한다</dd>
      <dt>최종 판단</dt><dd>automation gate · Guardian 한도 · 능력 증명서 · 거래 정책 ·
      journal 단일 writer 인터록이 그대로 결정한다</dd>
    </dl>
    <p><button type="submit">자동 시작 저장·적용</button></p>
  </form>
  {{end}}
  {{template "cardcautions" .AutostartGuard}}
  {{if .AutostartNote}}
  <p class="notice">마지막 엔진 기동 결과:</p>
  <pre>{{.AutostartNote}}</pre>
  {{end}}
</section>

<section data-settings-entry="market-schedule">
  <h2>시장·일정</h2>
  <p><a href="/strategy-runtime/market-schedule">시장·일정 화면 열기</a></p>
  <p class="muted">정규장·연장 시간대와 그 경계에서 무엇이 달라지는지를 읽는다.
  이 화면에서 바꾸는 값은 없다.</p>
</section>
{{end}}

{{/* --- 당일 — 가역 · 일 단위 ---------------------------------------------------- */}}

{{define "settings-daily"}}
{{/*
  The anchor a bookmark to /settings#adoption lands on. The fragment never
  reaches the server, so the browser brings it here — to the default tab, which
  does not have that section. This is the one thing that can meet it.
*/}}
<p id="adoption" class="notice">편입 규칙은 <a href="/settings/standing#adoption">상시 탭</a>으로
옮겼다 — 비가역이고 주 1회 미만으로 만지는 설정이라 당일 조정과 같은 화면에 두지 않는다.</p>

<section id="limits" data-settings-card="limits">
  <h2>Guardian 한도</h2>
  <p>자동 진입 하나가 얼마나 커질 수 있는지, 계좌가 얼마나 열릴 수 있는지, 하루에 얼마까지
  잃으면 진입이 멈추는지. 다섯 값은 <strong>진입</strong>에만 적용된다 — 청산은 RISK_REDUCING이라
  이 한도의 대상이 아니므로, 여기를 아무리 조여도 손절·비상 청산은 느려지지 않는다.</p>

  {{with $.NoticeFor "limits"}}<p class="notice" data-save-result="limits">{{.}}</p>{{end}}
  {{if .LimitsLoadErr}}<p class="danger">한도를 읽을 수 없다: <code>{{.LimitsLoadErr}}</code></p>{{end}}

  <div class="table-scroll" role="region" aria-label="Guardian 한도" tabindex="0"><table id="guardian-current">
    <tr><th>한도</th><th>설정값</th></tr>
    {{range .LimitRows}}<tr><td>{{.Label}}</td><td>{{.Value}}{{if .Unit}} {{.Unit}}{{end}}</td></tr>
    {{end}}
    <tr><td>계좌 기준 통화</td><td>{{.LimitCurrencyText}}</td></tr>
  </table></div>

  {{if not .LimitsUnset}}{{if not .LimitsPartlyConfigured}}
  <p class="muted">현재 값은 {{with .MatchingTier}}<strong>{{.}}</strong>와 일치한다{{else}}
  <strong>사용자 지정값</strong>이다{{end}} — 레지스트리와 대조한 결과이고, 누가 골랐는지에 대한
  주장은 아니다(같은 숫자를 손으로 적었을 수도 있다).</p>
  {{end}}{{end}}

  <p class="notice" data-currency-consequence="true">{{.LimitCurrencyConsequence}}</p>

  {{/*
    The gate's state, read-only, with a link to the tab that owns the switch. The
    "one control, one tab" rule bans the SWITCH from appearing twice; it requires
    this, because an operator setting a ceiling has to be able to see whether the
    thing the ceiling constrains is even on.
  */}}
  <p class="muted">automation gate는 지금 <strong>{{if .Gate.Enabled}}ON{{else}}OFF{{end}}</strong>이다.
  스위치는 <a href="/settings/standing#gate">상시 탭</a>에 있고 이 카드에서는
  <strong>콘솔에서 편집할 수 없다</strong> — 이 카드의 저장은 그 키의 바이트를 건드리지 않는다.</p>

  {{if .LimitGuard.Blocked}}{{template "cardblocked" .LimitGuard}}
  {{else}}
  <h3>프리셋</h3>
  <p class="muted">StockOS <code>risk_profiles.py</code>의 티어를 옮긴 것이다. 클릭 한 번이 다섯 값과
  통화를 함께 기록한다. <strong>등록된 티어 위로는 올릴 수 없다</strong> — 상한은 그 통화에 등록된
  티어의 필드별 최대이고, 그보다 큰 값은 서버가 거부한다.</p>
  {{range .LimitTierCards}}
  <div class="tier-card" data-tier="{{.ID}}">
    <form method="post" action="/settings/limits/preset">
      <input type="hidden" name="csrf" value="{{$.CSRF}}">
      <input type="hidden" name="tier" value="{{.ID}}">
      {{/*
        The preview replaces a dead confirm(). This is the one settings card where
        the server knows both sides before the click, so it is the one card that
        can say 현재 → 변경 rather than 현재 alone.
      */}}
      <dl class="card-preview" aria-label="{{.Label}} 적용 후">
        {{range .Preview.Rows}}
        <dt>{{.Label}}</dt>
        <dd>{{.From}} → <strong>{{.To}}</strong>{{if .Axis}}
          <span class="status-pill" data-limit-axis="{{.Axis}}">{{.Axis}}</span>{{end}}</dd>
        {{end}}
        {{if .Preview.FirstTime}}
        <dt>방향</dt><dd data-limit-axis="최초 설정"><strong>최초 설정</strong> — 이전 값이 없으므로
        강화도 완화도 아니다</dd>
        {{else if .Preview.CurrencyChange}}
        <dt>통화</dt><dd data-limit-axis="통화 변경"><strong>{{.Preview.CurrencyFrom}} → {{.Preview.CurrencyTo}} 통화 변경</strong>
        — 강화도 완화도 아닌 별개의 변경이다. 숫자가 작아져도 조인 것이 아니다</dd>
        {{end}}
      </dl>
      <p class="notice" data-currency-consequence="{{.ID}}">{{.Preview.Consequence}}</p>
      <p><button type="submit">{{.Label}}{{if .Recommended}} · 권장{{end}} 기록</button>
      <span class="muted">[{{.Limits.Currency}}] {{.Summary}}</span></p>
    </form>
  </div>
  {{end}}

  <details class="explain">
    <summary>고급 — 개별 값 직접 기입 (보통은 위의 프리셋으로 관리한다)</summary>
    <p class="muted">낮추는 방향은 양수인 한 자유롭다. 다섯 값 중 하나라도 비거나 0이면 저장이
    거부된다 — 엔진이 기동을 거부할 블록은 기록하지 않는다. 이 폼은 제출 전에는 미리보기를
    낼 수 없다: 서버는 아직 입력값을 모른다.</p>
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
      <p><label>계좌 기준 통화 (KRW 또는 USD)<br>
      <input type="text" name="limit_currency" value="{{.FieldCurrency}}"></label></p>
      <p><button type="submit">한도 저장</button></p>
    </form>
  </details>
  {{end}}
  {{template "cardcautions" .LimitGuard}}
  <p class="notice" data-effect-timing="limits">반영 시점 — {{.EffectNotice}}</p>
</section>

<section id="trading" data-settings-card="trading">
  <h2>거래 정책</h2>
  <p>엔진의 청산 경로가 실제로 쓰는 네 개다. <code>amend</code>·<code>conditional</code>·
  <code>fractional</code>은 이 빌드의 어느 루프도 쓰지 않으므로 여기 없고, 저장은 그 세 값을
  파일에 적힌 그대로 둔다.</p>

  {{with $.NoticeFor "trading"}}<p class="notice" data-save-result="trading">{{.}}</p>{{end}}
  {{if .TradingLoadErr}}<p class="danger">거래 정책을 읽을 수 없다: <code>{{.TradingLoadErr}}</code></p>{{end}}

  <div class="table-scroll" role="region" aria-label="거래 정책 현재값" tabindex="0"><table>
    <tr><th>토글</th><th>현재</th></tr>
    {{range .TradingRows}}<tr><td>{{.Label}}</td><td>{{if .On}}켜짐{{else}}꺼짐{{end}}</td></tr>{{end}}
  </table></div>

  {{if .TradingGuard.Blocked}}{{template "cardblocked" .TradingGuard}}
  {{else}}
  <form method="post" action="/settings/trading">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    {{range .TradingRows}}
    <p><label><input type="checkbox" name="{{.Key}}" {{if .On}}checked{{end}}>
    <strong>{{.Label}}</strong></label><br>
    <span class="muted">{{.Why}}</span></p>
    {{end}}
    <dl class="card-preview" aria-label="거래 정책 적용 후">
      <dt>적용 후</dt><dd>네 값이 체크 상태대로 기록된다. 네 개가 모두 켜져야 기동 인터록 3절을 통과한다</dd>
      <dt>승인의 의미</dt><dd>이 조합은 "엔진이 내 주식을 팔아도 된다"는 승인이다 — 켜는 것은 사람이고 저장은 audit에 남는다</dd>
    </dl>
    <p><button type="submit">거래 정책 저장</button></p>
  </form>
  {{end}}
  {{template "cardcautions" .TradingGuard}}
  <p class="notice" data-effect-timing="trading">반영 시점 — {{.EffectNotice}}</p>
</section>
{{end}}

{{/* --- 전략 — 규칙 자체, 진입점으로 ---------------------------------------------- */}}

{{define "settings-strategy"}}
<section>
  <h2>전략 규칙</h2>
  <p>아래 세 화면이 규칙을 소유한다. 이 탭은 그 화면들을 <strong>흡수하지 않는다</strong> —
  각 화면의 경로는 그대로이고 다른 요구사항이 그 경로를 참조한다. 각 줄의 요약은 그 화면이
  이미 계산한 값을 옮긴 것이고 여기서 다시 계산하지 않는다.</p>
  {{template "settingsentries" .}}
</section>
{{end}}

{{/* --- 도구 — 진단, 드물게 -------------------------------------------------------- */}}

{{define "settings-tools"}}
<section>
  <h2>진단 도구</h2>
  <p>드물게 열고, 열었을 때 상태부터 읽는다. 아래 요약은 각 화면이 읽는 것과 같은 기록에서 온다.</p>
  {{template "settingsentries" .}}
</section>

<section id="system-update" data-settings-card="system-update">
  <h2>시스템 업데이트</h2>
  <p>서버가 고정한 공식 저장소·현재 플랫폼의 최신 stable archive만 확인한다.
  archive SHA-256, GitHub Actions/Sigstore 서명, 정확한 release workflow·tag·source commit,
  transparency log를 모두 검증한 뒤에만 후보를 만든다.</p>

  {{with $.NoticeFor "system-update"}}<p class="notice" data-save-result="system-update">{{.}}</p>{{end}}

  <h3>서명된 릴리스 확인</h3>
  {{if .DownloadGuard.Blocked}}{{template "cardblocked" .DownloadGuard}}
  {{else}}
  <form method="post" action="/settings/system-update/download">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <dl class="card-preview" aria-label="릴리스 확인 적용 후">
      <dt>적용 후</dt><dd>서명을 검증한 archive를 fixed candidate로 staging한다</dd>
      <dt>실행되지 않는 것</dt><dd>설치와 재기동 — 다운로드는 둘 중 어느 것도 하지 않는다</dd>
    </dl>
    <p><button type="submit">서명된 최신 릴리스 확인·다운로드</button></p>
  </form>
  {{end}}

  <p class="muted">개발 세션의 임시 경로를 직접 설치하지 않는다. 이 화면이 볼 수 있는 후보는
  실행 중인 파일 바로 옆의 <code>.candidate</code> 하나뿐이고, 후보 코드는 검사 중 실행되지 않는다.</p>

  {{if .UpdateWired}}
  <h3>현재 실행 파일</h3>
  <div class="table-scroll" role="region" aria-label="현재 실행 파일" tabindex="0"><table>
    <tr><th>경로</th><td><code>{{.Update.Current.Path}}</code></td></tr>
    <tr><th>크기·수정 시각</th><td>{{.Update.Current.Size}} bytes · {{.UpdateCurrentTime}}</td></tr>
    <tr><th>SHA-256</th><td><code>{{.Update.Current.SHA256}}</code></td></tr>
  </table></div>

  <h3>staged candidate</h3>
  <p><code>{{.Update.CandidatePath}}</code></p>
  {{if .Update.Installable}}
  {{if .ReleaseVerified}}
  <p class="notice"><strong>이 프로세스에서 서명 검증됨.</strong>
  {{.ReleaseReceipt.Tag}} · <code>{{.ReleaseReceipt.AssetName}}</code> · archive
  <code>{{.ReleaseReceipt.ArchiveSHA256}}</code><br>
  signer <code>{{.ReleaseReceipt.Signer}}</code><br>
  source commit <code>{{.ReleaseReceipt.SourceCommit}}</code>
  {{if .ReleaseReceipt.Bootstrap}}<br>현재 빌드는 release semver가 없어 bootstrap으로 처리했다.
  버전 순서를 주장하지 않는다.{{end}}</p>
  {{end}}
  <div class="table-scroll" role="region" aria-label="staged candidate" tabindex="0"><table>
    <tr><th>크기·수정 시각</th><td>{{.Update.Candidate.Size}} bytes · {{.UpdateCandidateTime}}</td></tr>
    <tr><th>SHA-256</th><td><code>{{.Update.Candidate.SHA256}}</code></td></tr>
    <tr><th>Go</th><td>{{.Update.Candidate.GoVersion}}</td></tr>
    <tr><th>module</th><td><code>{{.Update.Candidate.ModulePath}}</code> {{.Update.Candidate.ModuleVersion}}</td></tr>
    <tr><th>command</th><td><code>{{.Update.Candidate.CommandPath}}</code></td></tr>
    <tr><th>platform</th><td>{{.Update.Candidate.GOOS}}/{{.Update.Candidate.GOARCH}}</td></tr>
    <tr><th>VCS revision</th><td><code>{{.Update.Candidate.VCSRevision}}</code>
    {{if .Update.Candidate.VCSModified}}(modified){{else}}(clean){{end}}</td></tr>
  </table></div>
  {{end}}
  {{end}}

  {{if .UpdateGuard.Blocked}}{{template "cardblocked" .UpdateGuard}}
  {{else}}
  <form method="post" action="/settings/system-update/install">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <input type="hidden" name="reviewed_sha256" value="{{.Update.Candidate.SHA256}}">
    <dl class="card-preview" aria-label="설치 적용 후">
      <dt>적용 후</dt><dd>위 SHA-256의 candidate를 설치하고 이 콘솔을 같은 포트로 재기동한다</dd>
      <dt>보존</dt><dd>기존 실행 파일은 <code>.rollback</code>으로 남는다</dd>
      <dt>선행 조건</dt><dd>엔진과 실계좌 검증이 모두 멈춘 상태여야 한다 — 서버가 확인하고 거부한다</dd>
    </dl>
    <p><button type="submit">검토한 candidate 설치 및 재기동</button></p>
  </form>
  {{end}}
  {{template "cardcautions" .UpdateGuard}}
</section>
{{end}}

{{/* "settingsentries" is the entry-point list both 전략 and 도구 render. */}}
{{define "settingsentries"}}
<dl class="entry-list">
  {{range .Entries}}
  <dt><a href="{{.Href}}">{{.Label}}</a></dt>
  <dd><strong data-entry-summary="{{.Label}}">{{.Summary}}</strong><br>
  <span class="muted">{{.Note}}</span></dd>
  {{end}}
</dl>
{{end}}
`
