package console

// templates_portfolio.go is the markup for the two dashboard screens (change
// add-operator-dashboard, tasks 2.1 and 2.2).
//
// It is a second const rather than more lines in templates.go so the verification
// console and the dashboard can be read apart; both are parsed into the same
// template set (pages.go), which is why the shared "head", "foot" and "style"
// definitions work here unchanged.
//
// Same rules as templates.go: no script, no external asset, no build step. Every
// state has words rather than an empty table, because the operator opens these
// screens exactly when they do not know what is running.

const portfolioTemplates = `
{{/*
  "journalstate" is the one place the four journal failures are worded. Both
  screens include it, so "엔진 미가동" and "콘솔 업데이트 필요" cannot drift into
  saying different things on different pages.
*/}}
{{define "journalstate"}}
{{if .Unwired}}
<p class="notice"><strong>원장 경로가 배선되지 않았다.</strong> 이 빌드의 콘솔은 엔진 journal을 읽도록
연결되어 있지 않다 — 브로커 보유만 표시된다.</p>
{{else if .Missing}}
<p class="notice"><strong>엔진 미가동 — 원장 파일이 아직 없다.</strong> {{.Path}}
<br>엔진이 한 번도 기동하지 않았거나 다른 데이터 디렉터리를 쓰고 있다. 콘솔은 원장을 만들지 않는다(읽기 전용).
브로커 보유는 그대로 표시된다.</p>
{{else if .TooNew}}
<p class="danger"><strong>콘솔 업데이트 필요.</strong> 원장 스키마 v{{.Version}}는 이 콘솔이 이해하는
v{{.Build}}보다 새롭다. 모르는 컬럼을 "없음"으로 읽으면 화면이 계좌에 대해 거짓을 말하게 되므로 읽지 않는다.
<br><code>{{.Detail}}</code></p>
{{else if .TooOld}}
<p class="danger"><strong>엔진 기동 필요 — 마이그레이션 전이다.</strong> 원장 스키마 v{{.Version}}에는 이 화면이
읽는 테이블이 없다. 콘솔은 마이그레이션을 실행하지 않는다(읽기 전용) — 엔진을 한 번 기동하면 올라간다.
<br><code>{{.Detail}}</code></p>
{{else if .Failed}}
<p class="danger"><strong>원장을 읽을 수 없다.</strong> <code>{{.Detail}}</code></p>
{{end}}
{{end}}

{{/*
  "brokerstate" is the broker half's status line: where the numbers came from and
  how old they are. It is not decoration — 평가손익은 캐시 시세로 계산된 값이고,
  그 시각을 함께 보여주지 않으면 화면이 현재가를 단언하는 것이 된다.
*/}}
{{define "brokerstate"}}
{{if not .Wired}}
<p class="notice"><strong>브로커 조회가 배선되지 않았다.</strong> 수량·평단·현재가·평가손익은 표시되지 않고,
원장 기준 정보만 표시된다.</p>
{{else if and .Held (not .Present)}}
<p class="notice"><strong>검증 중 — 데이터 없음.</strong> {{.HeldReason}}. 실계좌 검증이 rate limit을 쓰고 있으므로
브로커 조회를 보류한다(캐시도 비어 있다). 원장 기준 표시는 계속 동작한다.</p>
{{else if .Held}}
<p class="notice"><strong>검증 중 — 갱신 보류.</strong> {{.HeldReason}}. 아래 브로커 수치는 캐시된
{{.TakenAt}} 시점 값이다({{.AgeSeconds}}초 전).</p>
{{else if .Present}}
<details class="explain"><summary>데이터 기준 · {{.TakenAt}} ({{.AgeSeconds}}초 전)</summary><p class="muted">브로커 보유 캐시 <strong>{{.TakenAt}}</strong> ({{.AgeSeconds}}초 전).
갱신은 화면을 열 때만 일어나고 1회 갱신 = holdings 1콜이다 — 백그라운드 폴러는 없다.
현재가는 holdings 응답의 lastPrice이며 종목별 시세를 따로 조회하지 않는다.</p></details>
{{else}}
<p class="notice"><strong>브로커 보유를 아직 읽지 못했다.</strong></p>
{{end}}
{{if .Error}}<p class="danger">브로커 조회 실패: <code>{{.Error}}</code>{{if .Present}} — 위 표는 실패 이전의
캐시 값이다.{{end}}</p>{{end}}
{{end}}

{{define "positions"}}
{{template "head" .}}
<h1>포지션</h1>
<p class="muted page-intro">보유별 손익과 엔진의 현재 보호선을 한곳에서 확인한다.
<strong>계좌·주문은 변경하지 않으며</strong>, 아래 편입·제외 조작은 자동관리 설정만 갱신한다.</p>

{{with .Snap}}
{{template "journalstate" .Journal}}
{{if .NoManaged}}
<p class="notice"><strong>관리 포지션 없음.</strong> 원장은 정상적으로 읽혔고, 엔진이 관리 중인 미청산
포지션이 하나도 없다. 아래 보유가 있다면 전부 엔진 밖에서 생긴 것이다.</p>
{{end}}
{{template "brokerstate" .Holdings}}

{{if .Rows}}
<section>
{{/*
  두 공지는 상호 배타다: AnyUnknown은 원장이 답하지 않았을 때, AnyJournalAbsent는
  원장이 답했고 그 안에 포지션이 없을 때만 참이다. 같은 상태의 모든 행에 공통인
  사유는 여기서 한 번만 말하고, 행에는 판정 라벨만 남는다 — 사유를 행마다 반복하면
  표가 아니라 같은 문단의 목록이 된다 (change refresh-positions-screen).
*/}}
  {{if .AnyUnknown}}
  <p class="notice"><strong>관리 여부 불명 보유가 있다.</strong> 엔진 원장을 읽지 못했으므로 아래
  해당 보유가 엔진 관리 대상인지 알 수 없다 — 위의 원장 안내를 보라. 관리 중이 아니라고 단정하지 않는다.</p>
  {{end}}
  {{if .AnyJournalAbsent}}
  <p class="notice"><strong>엔진 원장에 포지션이 없는 보유가 있다</strong> — 엔진이 진입한 포지션이
  아니므로 손절·익절 라인도 없다. [관리 편입 예약]은 설정만 기록하며, 지정된 심볼은 다음
  대사 주기(엔진 가동 시)에 편입 후보가 된다.</p>
  {{end}}
  <table class="data-table positions-table">
    <caption>보유 종목과 보호 상태</caption>
    <thead><tr>
      <th scope="col">종목</th><th scope="col">관리 편입</th><th scope="col">손익</th>
      <th scope="col">가격</th><th scope="col">보호·익절</th><th scope="col">관리 동작</th>
    </tr></thead>
    <tbody>
    {{range .Rows}}
    <tr class="position-row" data-symbol="{{.Symbol}}">
      <th scope="row" data-label="종목"><code>{{.Symbol}}</code>{{if .Name}}<span class="submetric">{{.Name}}</span>{{end}}
        {{if $.Snap.Multi}}<span class="submetric">계좌 {{if .AccountRef}}{{.AccountRef}}{{else}}—{{end}}</span>{{end}}
      </th>
      <td data-label="관리"><strong class="status-pill">{{.Label}}</strong>
        {{if .Managed}}<span class="submetric">자격 근거 {{.Basis}}</span>{{end}}
        {{if .Designated}}<span class="submetric ok">편입 예약됨 · 아직 보호 미적용</span>{{end}}
        {{if and $.CanDesignate .InBroker (not .Managed) (not .Unknown) .Excluded}}<span class="submetric">편입하려면 제외를 먼저 해제</span>{{end}}
      </td>
      <td data-label="손익" class="{{if .Gain}}ok{{else}}bad{{end}}">
        <span class="metric">평가손익 {{.PnL}}</span><span class="submetric">수익률 {{.Rate}}</span>
      </td>
      <td data-label="가격"><span class="metric">현재가 {{.Last}}</span>
        {{if .HasExit}}
        <span class="submetric">진입가 <strong>{{.Exit.EntryPrice}}</strong></span>
        {{else}}<span class="submetric">평단 {{.Avg}}</span>{{end}}
      </td>
      <td data-label="보호·익절">
        {{if .HasExit}}
        <span class="metric">현재 보호선 <strong>{{.Exit.Baseline}}</strong></span>
        <span class="submetric">최초 손절 <strong>{{.Exit.InitialStop}}</strong></span>
        <span class="submetric">익절 진행 <strong>{{.Taken}}</strong> · ladder rung {{.Rung}}</span>
        {{else}}<span class="metric muted">보호선 —</span>{{if .Reason}}<span class="submetric">보호 근거 없음</span>{{end}}{{end}}
        {{if .HasDetail}}
        <details class="row-details"><summary>보호·원장 상세</summary><div class="detail-grid">
          <span>수량 <strong>{{.Qty}}</strong></span><span>평가금액 <strong>{{.Value}}</strong></span>
          {{if .Reason}}<span>{{.Reason}}</span>{{end}}
          {{if .HasExit}}<span>워터마크 <strong>{{.Exit.HighWater}}</strong></span>
          <span>래칫 단계 <strong>{{.Exit.RatchetLevel}}</strong></span>
          <span>pending exit <strong>{{.Pending}}</strong></span>
          <span>정책 {{.Exit.PolicyKind}} · 최초 리스크 {{.Exit.InitialRisk}}</span>
          <span>갱신 {{.Exit.UpdatedAt}}</span>{{end}}
          {{if .BrokerMissing}}<span class="bad">브로커 보유에 없다.</span>
          <span>원장은 {{.State}} 상태 수량 {{.JournalQuantity}}로 보고 있다.</span>
          {{else if .InJournal}}<span>원장 수량 {{.JournalQuantity}} · 원장 평단 {{.JournalAvgPrice}}</span>
          <span>상태 {{.State}} · <code>{{.PositionID}}</code></span>{{end}}
        </div></details>
        {{end}}
      </td>
      <td data-label="관리 동작"><div class="actions">
        {{if and $.CanDesignate .InBroker (not .Managed) (not .Unknown) (not .Excluded)}}
        <form method="post" action="/settings/include">
          <input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="symbol" value="{{.Symbol}}">
          {{if .Designated}}<input type="hidden" name="remove" value="1">{{end}}
          <button type="submit" class="secondary" aria-label="{{.Symbol}} {{if .Designated}}편입 예약 해제{{else}}관리 편입 예약{{end}}">{{if .Designated}}편입 예약 해제{{else}}관리 편입 예약{{end}}</button>
        </form>{{end}}
        {{if and $.CanDesignate .InBroker (not .Managed) (not .Unknown)}}
        <form method="post" action="/settings/exclude">
          <input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="symbol" value="{{.Symbol}}">
          {{if .Excluded}}<input type="hidden" name="remove" value="1">{{end}}
          <button type="submit" class="secondary" aria-label="{{.Symbol}} {{if .Excluded}}제외 해제{{else}}자동관리 제외{{end}}">{{if .Excluded}}제외 해제{{else}}자동관리 제외{{end}}</button>
        </form>{{end}}
        {{if or (not $.CanDesignate) .Managed .Unknown}}<span class="muted">—</span>{{end}}
      </div><span class="submetric">설정만 변경 · 계좌/주문 영향 없음</span></td>
    </tr>
    {{end}}
    </tbody>
  </table>
  <details class="explain"><summary>관리 상태와 데이터 기준</summary><p class="muted"><strong>관리 외(미편입)</strong>은 엔진의 exit 정책 대상이 아니라는 뜻이다 — 손절·익절이
  자동으로 걸려 있지 않다. <strong>관리 편입</strong>으로 예약한 보유는 편입 예약 상태다 — 실제 손절·익절은 엔진이
  가동되어 대사 루프가 편입을 완료한 뒤부터 걸린다. 관리 중인 행의 <strong>자격 근거</strong>는 그 보호가 어디서 왔는지를 말한다:
  <em>진입 결정</em>은 엔진이 직접 진입한 포지션, <em>편입 기록</em>은 사용자가 수동 매수한 보유를 엔진이
  편입한 것이다. 편입 자체는 엔진의 대사 루프가 수행하며 이 화면의 기능이 아니다 — 이 화면은 읽기 전용이다.</p>
  </details>
</section>
{{else}}
<section>
  <p class="muted">표시할 보유가 없다.
  {{if .Journal.Readable}}엔진 원장에 미청산 포지션이 없고{{end}}
  {{if .Holdings.Present}} 브로커 보유도 비어 있다.{{else}} 브로커 보유는 읽지 못했다.{{end}}</p>
</section>
{{end}}
{{end}}
{{template "foot" .}}
{{end}}

{{define "history"}}
{{template "head" .}}
<h1>거래 이력</h1>
<p class="muted">원장에 <strong>동결된 값</strong>만 표시한다. 실현손익·실현 R·초기 수량·보유 시간·도달 exit
단계는 청산 트랜잭션에서 얼어붙은 값 그대로이고, 심볼(positions)과 진입가(exit_states)만 조인해서 붙인다.
청산가는 스키마에 없으므로 표시하지 않으며 체결에서 재계산하지도 않는다 — 재계산한 값은 옆에 있는 동결 R과
어긋난다.</p>

{{with .Snap}}
{{template "journalstate" .Journal}}

<section>
  <h2>완결 왕복 거래</h2>
  {{if .Trips}}
  <table>
    <tr>
      {{if .Multi}}<th>계좌</th>{{end}}
      <th>청산 시각</th><th>심볼</th><th>진입가</th><th>실현손익(비용 차감)</th><th>실현 R</th>
      <th>초기 수량</th><th>보유 시간</th><th>도달 exit 단계</th>
    </tr>
    {{range .Trips}}
    <tr>
      {{if $.Snap.Multi}}<td>{{.AccountRef}}</td>{{end}}
      <td>{{.ClosedAt}}</td>
      <td><code>{{.Symbol}}</code></td>
      <td>{{.Entry}}</td>
      <td class="{{if .Win}}ok{{else}}bad{{end}}">{{.PnL}}</td>
      <td class="{{if .Win}}ok{{else}}bad{{end}}">{{.R}}</td>
      <td>{{.Quantity}}</td>
      <td>{{.Held}}</td>
      <td>{{.Stage}}</td>
    </tr>
    {{end}}
  </table>
  {{else if .Journal.Readable}}
  <p class="muted">완결 거래 없음 — 아직 왕복이 끝난 포지션이 없다.</p>
  {{end}}
  <p class="notice"><strong>이 표의 한계.</strong> 성과 행(trade_outcomes)은 엔진이 관리하던 포지션이
  닫힐 때만 동결된다. 외부 매도로 종결된 포지션은 여기에 행이 생기지 않으므로, 그 공백은 아래 exit 이벤트
  스트림이 보완한다. 보유 시간이 <code>—</code>인 행은 원장에 그 값이 NULL이라는 뜻이지 0초가 아니다.</p>
</section>

<section>
  <h2>exit 이벤트</h2>
  {{if .Events}}
  <p class="muted">엔진이 내린 exit 판정 기록이다(시간순). 성과 행이 없는 종결도 여기에는 흔적이 남는다.</p>
  {{if .Truncated}}<p class="notice">최근 기록만 표시한다 — 창을 넘는 오래된 판정은 생략되었다.</p>{{end}}
  <table>
    <tr>
      {{if .Multi}}<th>계좌</th>{{end}}
      <th>시각</th><th>심볼</th><th>관측가</th><th>워터마크</th><th>기준선</th><th>단계</th><th>발의</th>
    </tr>
    {{range .Events}}
    <tr>
      {{if $.Snap.Multi}}<td>{{.AccountRef}}</td>{{end}}
      <td>{{.At}}</td><td><code>{{.Symbol}}</code></td><td>{{.Observed}}</td>
      <td>{{.HighWater}}</td><td>{{.Baseline}}</td><td>{{.Level}}</td><td>{{.Proposal}}</td>
    </tr>
    {{end}}
  </table>
  {{else if .Journal.Readable}}
  <p class="muted">exit 판정 기록 없음.</p>
  {{end}}
</section>
{{end}}
{{template "foot" .}}
{{end}}
`
