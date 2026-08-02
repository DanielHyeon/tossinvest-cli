package console

// templates_signals.go is the discovery screen's markup (change
// add-candidate-discovery, tasks 5.5–5.7). It joins the same template set as
// everything else, so it shares "head", "foot" and the one stylesheet.
//
// There is no <form> in this file, no <input> and no <button>. That is asserted
// rather than promised: signals_test.go renders the page and fails on any of them,
// and on any word that would be a confirmation prompt.
//
// # The order of the page is the argument
//
// Degradation first, because it says how much of the market was actually looked
// at and every number below it is conditional on that. Then the veto tally with
// the unmeasured count ABOVE the vetoed and passed counts. Then the shadow
// records, labelled as records. The table last, because a reader who has not seen
// the first three blocks will read an empty veto column as "nothing wrong".
//
// # Two words that must never be swapped
//
// 후보 counts candidates. 계열 counts (market, symbol, source) series (D9, D21).
// The acceleration tally's denominator is the second one and it is larger than
// the first, because a source that carries no trading value still occupies a slot
// every tick. Rendering the series number under the word 후보 would turn "계열
// 900개 중 12개" into "후보 300개 중 12개", which is a different and much more
// alarming sentence.

const signalsTemplates = `
{{define "signals"}}{{template "head" .}}
<h1>신호 <span class="muted">발굴 후보</span></h1>
<p class="muted">순위 원천이 시간에 걸쳐 무엇을 말했는지, 그리고 <strong>무엇을 확인하지
못했는지</strong>를 본다. 읽기 전용 — 이 화면은 주문을 내지 않고, 원천을 부르지도 않는다
(저장소 읽기뿐이다). {{.Snap.NowText}} 기준.</p>

{{template "signalsread" .Snap}}
{{range .Snap.Markets}}{{template "signalsmarket" .}}{{end}}

{{template "foot" .}}
{{end}}

{{/*
  "signalsread" — 화면 전체가 읽히지 않은 경우. seam 미배선과 저장소 판독 실패는
  다른 사유이고 운영자의 대응이 다르다.
*/}}
{{define "signalsread"}}
{{if .Read.Known}}
  {{if .TakenAt}}
  <p class="muted">평가 시각 {{.TakenAt}} ({{.AgeSeconds}}초 전). 이 화면은 마지막 스캔이
  저장한 것을 읽는다 — 새 스캔을 일으키지 않는다.</p>
  {{end}}
  {{if not .Markets}}
  <p class="notice"><strong>시장이 하나도 없다.</strong> 저장소는 읽혔는데 평가된 시장이
  없다 — "조용하다"가 아니라 배선을 봐야 한다.</p>
  {{end}}
{{else}}
<p class="notice"><strong>미측정 ({{.Read.Code}})</strong> — {{.Read.Why}}
<br>0건이 아니다. 아무것도 읽지 못했다는 뜻이다.</p>
{{end}}
{{end}}

{{define "signalsmarket"}}
<section>
  <h2>{{.Market}}</h2>
  {{if not .Read.Known}}
  <p class="notice"><strong>미측정 ({{.Read.Code}})</strong> — {{.Read.Why}}
  <br>이 시장의 후보가 0건이라는 뜻이 아니다.</p>
  {{else}}
  {{template "signalspanel" .Panel}}
  {{template "signalsveto" .Veto}}
  {{template "signalsaccel" .Accel}}
  {{range .Bands}}{{template "signalsband" .}}{{end}}
  {{template "signalssightings" .}}
  {{template "signalstable" .}}
  {{end}}
</section>
{{end}}

{{/*
  "signalspanel" is task 5.7. 사유 없는 불리언은 대응할 수 없는 표시다 — 만료된 WTS
  세션과 rate limit에 걸린 공식 랭킹은 같은 boolean이고 정반대의 대응이다. 그래서
  이름과 사유를 함께 적고, 강등인데 이름이 없으면 그 사실을 결함으로 적는다.
*/}}
{{define "signalspanel"}}
<h3>원천</h3>
{{if not .Read.Known}}
<p class="notice"><strong>원천 상태 미측정 ({{.Read.Code}})</strong> — {{.Read.Why}}
<br>"강등 없음"이 아니다. 어느 원천이 빠졌는지 이 화면이 모른다는 뜻이다.</p>
{{else}}
<dl>
  <dt>응답 / 시도</dt>
  <dd>{{.AnsweredText}}{{if .TakenAt}} <span class="muted">— {{.TakenAt}} 스캔</span>{{end}}</dd>
  <dt>강등</dt>
  <dd>{{if .Degraded}}<strong class="bad">있음</strong>{{else if .NothingAttempted}}<strong>판단 불가 — 시도한 원천이 없다</strong>{{else}}<span class="ok">없음 — 전 원천 응답</span>{{end}}</dd>
</dl>
{{if .NothingAttempted}}
<p class="notice"><strong>이 스캔은 원천을 하나도 부르지 않았다.</strong> 일정상 아직 부를
때가 아닌 원천만 있었다는 뜻이다(엔진 양보로 간격이 두 배가 됐거나, tick이 모든 원천의
간격보다 짧거나, 시각이 뒤로 갔을 때). 강등이 아니므로 찾아야 할 고장난 원천은 없지만,
"전 원천 응답"도 아니다 — 아래 숫자들은 이번 통과가 아니라 마지막으로 실제 읽은 통과의
것이다.</p>
{{end}}
{{if .Missing}}
<div class="table-scroll" role="region" aria-label="원천" tabindex="0"><table>
  <thead><tr><th>빠진 원천</th><th>사유</th><th>rate limit</th></tr></thead>
  <tbody>
  {{range .Missing}}
    <tr>
      <td><code>{{.Source}}</code></td>
      <td>{{.Reason}}</td>
      <td>{{if .RateLimited}}<strong class="bad">429</strong>{{else}}<span class="muted">아니다</span>{{end}}</td>
    </tr>
  {{end}}
  </tbody>
</table></div>
<p class="muted">공식 랭킹에는 WTS 우회가 없다(D14) — 429 한 번이면 그 원천이 통째로
사라진다. 결손율이 미문서화 RANKING 한도의 유일한 측정값이다.</p>
{{else if .Unnamed}}
<p class="notice"><strong>강등인데 빠진 원천 이름이 없다.</strong> 스캔이 결손을 보고하면서
어느 원천인지 남기지 않았다 — 사유 없는 불리언은 대응할 수 없는 표시이므로, 없다고
적는 대신 없다는 사실을 적는다.</p>
{{end}}
{{end}}
{{end}}

{{/*
  "signalsveto" is task 5.6. 미측정이 통과처럼 보이지 않아야 한다.

  세 숫자의 순서가 주장이다 — 미측정이 맨 위다. 그리고 passed 0 옆에는 D18 때문에
  구조적으로 0이라는 문장이 붙는다. 그 문장이 없으면 다음 사람이 "고장"으로 읽고,
  D18이 지목한 유일하게 틀린 수리가 THRESHOLD_ABSENT를 통과로 세는 것이다.
*/}}
{{define "signalsveto"}}
<h3>추격 veto — 후보 {{.Candidates}}개</h3>
<dl>
  <dt>미측정 후보</dt>
  <dd><strong>{{.Unmeasured}}</strong> <span class="muted">← 이 화면의 기본 독해는
    "대부분 안전"이 아니라 "대부분 미확인"이다</span></dd>
  <dt>거부된 후보</dt>
  <dd>{{.Vetoed}}</dd>
  <dt>통과한 후보</dt>
  <dd>{{.Passed}} <span class="muted">({{.PassedNote}})</span></dd>
</dl>
{{/*
  집계 항등식 경보. 임계의 유무와 무관하게 계산되므로 임계가 하나도 없는 지금도
  작동한다. 숫자 옆에 붙고 숫자를 대체하지 않는다 — 경보의 근거가 그 숫자다.
*/}}
{{if .Alarms}}
{{range .Alarms}}
<p class="bad">{{.}}</p>
{{end}}
{{end}}
<div class="table-scroll" role="region" aria-label="추격 veto 사유별 후보 수" tabindex="0"><table>
  <thead><tr><th>사유 코드</th><th>위험</th><th>미측정</th></tr></thead>
  <tbody>
  {{range .Codes}}
    <tr><td><code>{{.Code}}</code></td><td>{{.Raised}}</td><td>{{.Unmeasured}}</td></tr>
  {{end}}
  </tbody>
</table></div>
{{if .Reasons}}
<p class="muted">미측정 입력:
{{range $i, $r := .Reasons}}{{if $i}} · {{end}}<code>{{$r.Reason}}</code> {{$r.Count}}{{end}}
<br>NO_DAY_HIGH가 많으면 캔들 예산이 설계대로 도는 것이고, THRESHOLD_ABSENT가 많으면
아무도 veto를 설정하지 않은 것이다. 고칠 대상이 다르다.</p>
{{end}}
{{end}}

{{/*
  "signalsaccel" — D21. 분모가 후보가 아니라 계열이다. 두 단어를 바꿔 쓰면 안 된다.
*/}}
{{define "signalsaccel"}}
<h3>가속도 그림자 교차 — 계열 {{.Series}}개</h3>
<p class="muted">계열 {{.Series}}개 중 {{.Measured}}개 측정. {{.SeriesNote}}</p>
<p class="muted">{{.ShadowNote}}</p>
<p class="muted">교차:
{{range $i, $c := .Crossed}}{{if $i}} · {{end}}{{$c.Reason}}배 {{$c.Count}}{{end}}</p>
{{if .NotComputed}}
<p class="muted">미산출:
{{range $i, $r := .NotComputed}}{{if $i}} · {{end}}<code>{{$r.Reason}}</code> {{$r.Count}}{{end}}</p>
{{end}}
{{end}}

{{/*
  "signalsband" — 그림자 밴드. 붕괴 경보는 건수 위에 붙고 건수를 대체하지 않는다:
  경보가 말하는 대상이 그 건수이고, 숫자를 지우면 근거를 지우는 것이다. 측정된 기록
  전부가 같은 교차 집합을 냈다는 사실은 산술로는 아무 문제가 없어서, 경보가 없으면
  평범한 행으로 읽힌다 — 사람이 같은 줄을 두 번 읽어야 알아차리는 상태다.
*/}}
{{define "signalsband"}}
<h3>그림자 밴드 — {{.Code}} <span class="muted">(veto가 아니다)</span></h3>
<p class="muted">{{.ShadowNote}}</p>
{{if .Alarm}}
<p class="bad">{{.Alarm}}</p>
{{end}}
<p class="muted">후보 {{.Total}}개 중 {{.Measured}}개 측정. 교차:
{{range $i, $c := .Crossed}}{{if $i}} · {{end}}{{$c.Reason}} {{$c.Count}}{{end}}</p>
{{if .NotMeasured}}
<p class="muted">미측정:
{{range $i, $r := .NotMeasured}}{{if $i}} · {{end}}<code>{{$r.Reason}}</code> {{$r.Count}}{{end}}</p>
{{end}}
{{end}}

{{/*
  "signalssightings" — 최초 관측을 원천별로 나눈다.

  veto census는 사유별 건수를 준다. 그것으로는 "공식 거래대금 순위가 짧게 온다"와
  "WTS 인기 목록이 짧게 온다"를 구분할 수 없는데, 둘은 가서 볼 곳이 다르고 그 중
  하나는 seen_late 분포 전체가 비는 원인이다. 공식 순위가 요청한 100행을 실제로
  주는지는 아직 아무도 재지 않았다 — 주지 않는다면 그 원천의 최초 관측은 전부
  READING_TRUNCATED이고, 후속 change가 임계를 고를 분포가 존재하지 않는다.
*/}}
{{define "signalssightings"}}
{{if .Sightings}}
<h3>최초 관측 — 원천별</h3>
<p class="muted">seen_late가 어느 원천의 읽기를 거부하고 있는지. 후보 한 건은 최초
관측을 하나만 가지므로 분모는 후보 수다.</p>
<div class="table-scroll" role="region" aria-label="최초 관측 — 원천별" tabindex="0"><table>
  <thead><tr><th>원천</th><th>측정 / 보유</th><th>거부 사유</th></tr></thead>
  <tbody>
  {{range .Sightings}}
    <tr>
      <td><code>{{.Source}}</code></td>
      <td>{{.Measured}} / {{.Total}}</td>
      <td>{{if .NotMeasured}}{{range $i, $r := .NotMeasured}}{{if $i}} · {{end}}<code>{{$r.Reason}}</code> {{$r.Count}}{{end}}{{else}}<span class="muted">없음</span>{{end}}</td>
    </tr>
  {{end}}
  </tbody>
</table></div>
{{end}}
{{end}}

{{/*
  "signalstable" — 후보 목록. veto 세 칸은 각각 자기 단어를 갖는다: 비어 있는 칸이
  없으므로 "사유가 하나도 없는 행"이라는 상태 자체가 존재하지 않는다.
*/}}
{{define "signalstable"}}
<h3>후보 목록</h3>
{{if .Rows}}
<div class="table-scroll" role="region" aria-label="후보 목록" tabindex="0"><table>
  <thead><tr>
    <th>종목</th><th>상태</th><th>최초 발견(UTC)</th><th>판정</th>
    <th>seen_late</th><th>extended</th><th>near_high</th>
    <th>최초 순위</th><th>확장률</th><th>고가까지 거리</th><th>계열</th><th>원천</th>
  </tr></thead>
  <tbody>
  {{range .Rows}}
    <tr>
      <td><code>{{.Symbol}}</code></td>
      <td>{{.StateText}}</td>
      <td>{{.FirstSeen}}<br><span class="muted">{{.AgeText}}</span></td>
      <td>{{if .Verdict.Vetoed}}<strong class="bad">{{.Verdict.Label}}</strong>{{else if .Verdict.Passed}}<strong class="ok">{{.Verdict.Label}}</strong>{{else}}<strong>{{.Verdict.Label}}</strong>{{end}}
        <br><span class="muted">{{.Verdict.Detail}}</span></td>
      {{range .Vetoes}}
      <td class="{{.Class}}">{{.Label}}{{if .Reason}}<br><code>{{.Reason}}</code>{{end}}</td>
      {{end}}
      <td>{{.Sighting.Value}}{{if not .Sighting.Measured}}<br><code>{{.Sighting.Reason}}</code>{{end}}</td>
      <td>{{.Expansion.Value}}{{if not .Expansion.Measured}}<br><code>{{.Expansion.Reason}}</code>{{end}}</td>
      <td>{{.Distance.Value}}{{if not .Distance.Measured}}<br><code>{{.Distance.Reason}}</code>{{end}}</td>
      <td>{{.SeriesMeasured}} / {{.Series}}{{if .Crossed}}<br><span class="muted">{{.Crossed}}</span>{{end}}</td>
      <td>{{.SourcesText}}<br><span class="muted">{{.Completeness}}{{if .Degraded}} · <span class="bad">강등</span>{{end}}</span></td>
    </tr>
  {{end}}
  </tbody>
</table></div>
<p class="muted"><strong>미측정은 통과가 아니다</strong>(D10). veto 칸이 "미측정"인 행은
그 사유를 확인한 적이 없다는 뜻이고, 상세조회는 초당 5회·종목당 1회이므로 그것이
평시다 — 예외가 아니다. "측정·안전"은 실제로 재고 위험하지 않았다는 뜻이며, 두 상태는
같은 칸에 같은 말로 적히지 않는다.</p>
{{else}}
<p class="muted">이 시장에 살아 있는 후보가 없다. 위의 원천 블록이 전 원천 응답이라고
말할 때에만 "조용하다"로 읽어도 된다.</p>
{{end}}
{{end}}
`
