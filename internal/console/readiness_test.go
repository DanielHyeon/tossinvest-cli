package console

import (
	"errors"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestStartupReadinessAppearsOnThreeScreensWithoutControls(t *testing.T) {
	settings := fullSettingsHarness(t)
	settings.authenticate(t)
	for _, path := range []string{pathSettingsStanding, pathSettingsDaily} {
		section := readinessSection(t, body(t, settings.get(t, path)))
		assertReadinessShape(t, section)
		for _, href := range []string{
			`href="/settings/standing#gate"`,
			`href="/settings/standing#autostart"`,
			`href="/settings/daily#limits"`,
			`href="/settings/daily#trading"`,
		} {
			if !strings.Contains(section, href) {
				t.Errorf("%s 준비도에 %s 소유 탭 링크가 없다:\n%s", path, href, section)
			}
		}
		for _, forbidden := range []string{"기동 가능", "시작 가능", "준비 완료"} {
			if strings.Contains(section, forbidden) {
				t.Errorf("%s 준비도에 합산 판정 %q가 있다:\n%s", path, forbidden, section)
			}
		}
	}

	dashboard := newEngineHarness(t, func(o *Options) {
		o.Settings = &fakeSettings{}
		o.Limits = &fakeLimits{gate: fullLimits()}
		o.TradingPolicy = &fakeTrading{block: fullPolicy()}
		o.Gate = &fakeGate{}
		o.EngineBoot = &fakeAutostart{}
	})
	dashboard.authenticate(t)
	page := body(t, dashboard.get(t, pathVerifyConsole))
	section := readinessSection(t, page)
	assertReadinessShape(t, section)
	if strings.Index(page, section) > strings.Index(page, `action="/engine/start"`) {
		t.Errorf("검증 콘솔 준비도가 엔진 시작 폼 뒤에 있다:\n%s", page)
	}

	for _, want := range []string{
		limitSummary(fullLimits().Limits()),
		"청산에 필요한 네 토글이 모두 켜져 있다",
		"파일 경로·만료·계좌 일치는 엔진이 기동 시점에 판정한다",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("GatePreflight 이동 결과에 %q가 없다:\n%s", want, section)
		}
	}

	for _, tc := range []struct {
		name  string
		tweak func(*Options)
		want  string
	}{
		{
			name: "미배선",
			tweak: func(o *Options) {
				o.Limits = nil
				o.TradingPolicy = nil
				o.EngineBoot = nil
			},
			want: "배선되지 않았다",
		},
		{
			name: "읽기 실패",
			tweak: func(o *Options) {
				o.Limits = &fakeLimits{loadErr: errors.New("한도 읽기 실패")}
				o.TradingPolicy = &fakeTrading{loadErr: errors.New("정책 읽기 실패")}
				o.EngineBoot = &fakeAutostart{loadErr: errors.New("자동 시작 읽기 실패")}
			},
			want: "읽을 수 없다",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newEngineHarness(t, tc.tweak)
			h.authenticate(t)
			got := readinessSection(t, body(t, h.get(t, pathVerifyConsole)))
			if strings.Count(got, `data-readiness-state="unknown"`) == 0 || !strings.Contains(got, tc.want) {
				t.Errorf("%s seam이 unknown 사유로 보이지 않는다:\n%s", tc.name, got)
			}
		})
	}
}

func TestStartupReadinessMakesEachBrokenSeamUnknownOnEverySurface(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tweak  func(*Options)
		labels []string
		reason string
	}{
		{
			name:   "한도 미배선",
			tweak:  func(o *Options) { o.Limits = nil },
			labels: []string{"자동화 게이트", "1. 한도 다섯 개와 통화"},
			reason: "설정 seam이 배선되지 않았다",
		},
		{
			name:   "한도 읽기 실패",
			tweak:  func(o *Options) { o.Limits = &fakeLimits{loadErr: errors.New("한도 읽기 실패")} },
			labels: []string{"자동화 게이트", "1. 한도 다섯 개와 통화"},
			reason: "설정을 읽을 수 없다: 한도 읽기 실패",
		},
		{
			name:   "거래 정책 미배선",
			tweak:  func(o *Options) { o.TradingPolicy = nil },
			labels: []string{"3. 거래 정책"},
			reason: "설정 seam이 배선되지 않았다",
		},
		{
			name:   "거래 정책 읽기 실패",
			tweak:  func(o *Options) { o.TradingPolicy = &fakeTrading{loadErr: errors.New("정책 읽기 실패")} },
			labels: []string{"3. 거래 정책"},
			reason: "설정을 읽을 수 없다: 정책 읽기 실패",
		},
		{
			name:   "자동 시작 미배선",
			tweak:  func(o *Options) { o.EngineBoot = nil },
			labels: []string{"엔진 자동 시작"},
			reason: "설정 seam이 배선되지 않았다",
		},
		{
			name:   "자동 시작 읽기 실패",
			tweak:  func(o *Options) { o.EngineBoot = &fakeAutostart{loadErr: errors.New("자동 시작 읽기 실패")} },
			labels: []string{"엔진 자동 시작"},
			reason: "설정을 읽을 수 없다: 자동 시작 읽기 실패",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings := fullSettingsHarness(t, tc.tweak)
			settings.authenticate(t)
			pages := map[string]string{
				pathSettingsStanding: body(t, settings.get(t, pathSettingsStanding)),
				pathSettingsDaily:    body(t, settings.get(t, pathSettingsDaily)),
			}
			console := newEngineHarness(t, func(o *Options) {
				o.Settings = &fakeSettings{}
				o.Limits = &fakeLimits{gate: fullLimits()}
				o.TradingPolicy = &fakeTrading{block: fullPolicy()}
				o.Gate = &fakeGate{}
				o.EngineBoot = &fakeAutostart{}
			}, tc.tweak)
			console.authenticate(t)
			pages[pathVerifyConsole] = body(t, console.get(t, pathVerifyConsole))

			for path, page := range pages {
				section := readinessSection(t, page)
				for _, label := range tc.labels {
					row := readinessRowFor(t, section, label)
					if !strings.Contains(row, `data-readiness-state="unknown"`) || !strings.Contains(row, tc.reason) {
						t.Errorf("%s %s의 %q 행이 unknown 사유를 그대로 보이지 않는다:\n%s", path, tc.name, label, row)
					}
				}
			}
		})
	}
}

func TestGuardianPresetsRenderAsOneMatrixPerCurrency(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, pathSettingsDaily))
	cards := settingsPage{Gate: fullLimits()}.LimitTierCards()
	currencies := map[string]bool{}
	for _, card := range cards {
		currencies[card.Limits.Currency] = true
	}
	if got := strings.Count(page, `class="preset-matrix"`); got != len(currencies) {
		t.Fatalf("프리셋 비교표 = %d, want 통화 수 %d:\n%s", got, len(currencies), page)
	}
	if got := strings.Count(page, `data-preset-row=`); got != len(currencies)*6 {
		t.Errorf("비교표 행 = %d, want 통화별 한도 다섯+통화 %d", got, len(currencies)*6)
	}
	if got := strings.Count(page, `data-currency-consequence=`); got != len(currencies) {
		t.Errorf("통화 결과 설명 = %d, want 통화별 1회 %d", got, len(currencies))
	}
	for _, card := range cards {
		if got := strings.Count(page, `action="/settings/limits/preset" data-preset-tier="`+card.ID+`"`); got != 1 {
			t.Errorf("%s 프리셋 폼 = %d, want 1", card.ID, got)
		}
		if !strings.Contains(page, `name="tier" value="`+card.ID+`"`) {
			t.Errorf("%s 프리셋 hidden tier가 없다", card.ID)
		}
		for _, row := range card.Preview.Rows {
			current := `data-preset-current="` + row.Label + `">` + row.From
			if !strings.Contains(page, current) {
				t.Errorf("%s/%s 현재값이 LimitTierCards Preview.From과 다르다", card.ID, row.Label)
			}
			want := `data-preset-tier="` + card.ID + `" data-preset-limit="` + row.Label + `">` + row.To
			if !strings.Contains(page, want) {
				t.Errorf("%s/%s 값이 LimitTierCards Preview.To와 다르다", card.ID, row.Label)
			}
		}
	}
}

func TestTradingPolicyRowsAppearOnce(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, pathSettingsDaily))
	if strings.Contains(page, `aria-label="거래 정책 현재값"`) {
		t.Error("거래 정책 현재 상태 표가 남아 있다")
	}
	if got := strings.Count(page, `class="toggle-row"`); got != 4 {
		t.Fatalf("toggle-row = %d, want 4:\n%s", got, page)
	}
	wantFields := []string{"allow_live_order_actions", "cancel", "place", "sell"}
	gotFields := make([]string, 0, len(wantFields))
	for _, key := range wantFields {
		row := toggleRowFragment(t, page, key)
		if !strings.Contains(row, `type="checkbox"`) || !strings.Contains(row, `class="status-pill`) || !strings.Contains(row, `class="muted"`) {
			t.Errorf("%s 줄에 체크박스·상태 배지·설명이 함께 있지 않다:\n%s", key, row)
		}
		gotFields = append(gotFields, key)
	}
	sort.Strings(gotFields)
	if !reflect.DeepEqual(gotFields, wantFields) {
		t.Errorf("거래 정책 필드 = %v, want %v", gotFields, wantFields)
	}
}

func TestSettingsFormsAreByteIdenticalAfterLayout(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	standing := body(t, h.get(t, pathSettingsStanding))
	daily := body(t, h.get(t, pathSettingsDaily))
	if !strings.Contains(daily, `class="preset-matrix"`) {
		t.Fatal("폼 계약 비교 전에 프리셋 비교표가 렌더되지 않았다")
	}
	want := a173SettingsFormContracts(h.csrf)
	got := append(settingsFormContracts(t, standing), settingsFormContracts(t, daily)...)
	sort.Slice(got, func(i, j int) bool { return got[i].key() < got[j].key() })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("설정 폼 계약이 a173 기준선과 다르다:\n got: %#v\nwant: %#v", got, want)
	}
}

func readinessSection(t *testing.T, page string) string {
	t.Helper()
	const marker = `<section class="readiness"`
	i := strings.Index(page, marker)
	if i < 0 {
		t.Fatalf("기동 준비 section이 없다:\n%s", page)
	}
	if strings.Count(page, marker) != 1 {
		t.Fatalf("기동 준비 section 수 = %d, want 1", strings.Count(page, marker))
	}
	j := strings.Index(page[i:], "</section>")
	if j < 0 {
		t.Fatalf("기동 준비 section이 닫히지 않았다:\n%s", page[i:])
	}
	return page[i : i+j+len("</section>")]
}

func assertReadinessShape(t *testing.T, section string) {
	t.Helper()
	for _, forbidden := range []string{"<form", "<input", "<button"} {
		if strings.Contains(section, forbidden) {
			t.Errorf("준비도 안에 %s 제어가 있다:\n%s", forbidden, section)
		}
	}
	if got := strings.Count(section, `class="readiness-row"`); got != 6 {
		t.Errorf("준비도 행 = %d, want 절 단위 6:\n%s", got, section)
	}
	for _, want := range []string{
		"자동화 게이트", "엔진 자동 시작", "1. 한도 다섯 개와 통화", "3. 거래 정책",
		"2·6. 능력 증명서", "4·5·7·8. Guardian 주입 · 게이트웨이 · 전송의 키 운반",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("준비도에 %q가 없다:\n%s", want, section)
		}
	}
}

func readinessRowFor(t *testing.T, section, label string) string {
	t.Helper()
	marker := "<strong>" + label + "</strong>"
	i := strings.Index(section, marker)
	if i < 0 {
		t.Fatalf("준비도에 %q 행이 없다:\n%s", label, section)
	}
	start := strings.LastIndex(section[:i], `<li class="readiness-row"`)
	end := strings.Index(section[i:], "</li>")
	if start < 0 || end < 0 {
		t.Fatalf("준비도 %q 행 경계를 찾지 못했다", label)
	}
	return section[start : i+end+len("</li>")]
}

func toggleRowFragment(t *testing.T, page, key string) string {
	t.Helper()
	marker := `data-toggle-key="` + key + `"`
	i := strings.Index(page, marker)
	if i < 0 {
		t.Fatalf("%s 거래 정책 줄이 없다:\n%s", key, page)
	}
	start := strings.LastIndex(page[:i], `<div class="toggle-row"`)
	end := strings.Index(page[i:], "</div>")
	if start < 0 || end < 0 {
		t.Fatalf("%s 거래 정책 줄 경계를 찾지 못했다", key)
	}
	return page[start : i+end+len("</div>")]
}

type settingsFormContract struct {
	Method string
	Action string
	Fields []string
	Hidden []string
}

func (c settingsFormContract) key() string {
	return c.Method + " " + c.Action + " " + strings.Join(c.Hidden, ",")
}

var (
	formPattern  = regexp.MustCompile(`(?s)<form\s+([^>]*)>(.*?)</form>`)
	fieldPattern = regexp.MustCompile(`(?s)<(input|select|textarea)\s+([^>]*)>`)
	attrPattern  = regexp.MustCompile(`([[:alnum:]_-]+)="([^"]*)"`)
)

func settingsFormContracts(t *testing.T, page string) []settingsFormContract {
	t.Helper()
	var out []settingsFormContract
	for _, form := range formPattern.FindAllStringSubmatch(page, -1) {
		formAttrs := attrs(form[1])
		contract := settingsFormContract{Method: strings.ToLower(formAttrs["method"]), Action: formAttrs["action"]}
		for _, field := range fieldPattern.FindAllStringSubmatch(form[2], -1) {
			fieldAttrs := attrs(field[2])
			if name := fieldAttrs["name"]; name != "" {
				contract.Fields = append(contract.Fields, name)
			}
			if field[1] == "input" && fieldAttrs["type"] == "hidden" {
				contract.Hidden = append(contract.Hidden, fieldAttrs["name"]+"="+fieldAttrs["value"])
			}
		}
		sort.Strings(contract.Fields)
		sort.Strings(contract.Hidden)
		out = append(out, contract)
	}
	return out
}

func attrs(source string) map[string]string {
	out := map[string]string{}
	for _, pair := range attrPattern.FindAllStringSubmatch(source, -1) {
		out[pair[1]] = pair[2]
	}
	return out
}

func a173SettingsFormContracts(csrf string) []settingsFormContract {
	contracts := []settingsFormContract{
		{Method: "post", Action: "/settings/save", Fields: []string{"csrf", "default_stop_percent", "enabled", "exclude_symbols", "include_symbols"}, Hidden: []string{"csrf=" + csrf}},
		{Method: "post", Action: "/settings/gate", Fields: []string{"csrf", "enabled"}, Hidden: []string{"csrf=" + csrf}},
		{Method: "post", Action: "/settings/autostart", Fields: []string{"csrf", "enabled"}, Hidden: []string{"csrf=" + csrf}},
		{Method: "post", Action: "/settings/limits", Fields: []string{"csrf", "limit_currency", "max_daily_buy_notional", "max_daily_loss_amount", "max_daily_loss_ratio", "max_open_positions", "max_order_notional", "max_order_quantity", "max_total_exposure"}, Hidden: []string{"csrf=" + csrf}},
		{Method: "post", Action: "/settings/trading", Fields: []string{"allow_live_order_actions", "cancel", "csrf", "place", "sell"}, Hidden: []string{"csrf=" + csrf}},
	}
	for _, id := range []string{"kr-smoke", "kr-small-live", "us-smoke", "us-small-live", "us-single-name"} {
		contracts = append(contracts, settingsFormContract{Method: "post", Action: "/settings/limits/preset", Fields: []string{"csrf", "tier"}, Hidden: []string{"csrf=" + csrf, "tier=" + id}})
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].key() < contracts[j].key() })
	return contracts
}
