package console

import "strings"

// readinessRow 는 기동 전 화면이 파일에서 읽을 수 있는 사실 하나를 나타냄.
type readinessRow struct {
	Label  string
	State  string
	Detail string
	Href   string
}

// StatusText 는 상태의 짧은 표시 문구를 반환함.
func (r readinessRow) StatusText() string {
	switch r.State {
	case "ok":
		return "확인됨"
	case "warn":
		return "확인 필요"
	default:
		return "판정 불가"
	}
}

// StartupReadiness 는 기동 전 화면이 읽을 수 있는 여섯 절을 옮겨 표시함.
func (p settingsPage) StartupReadiness() []readinessRow {
	limitsKnown := p.LimitsWired && p.LimitsLoadErr == ""
	tradingKnown := p.TradingWired && p.TradingLoadErr == ""
	rows := []readinessRow{
		readinessGate(p, limitsKnown),
		readinessAutostart(p),
	}
	clauses := p.GatePreflight()
	rows = append(rows,
		readinessClause(clauses, "1. 한도 다섯 개와 통화", limitsKnown, p.LimitsWired, p.LimitsLoadErr, "/settings/daily#limits"),
		readinessClause(clauses, "3. 거래 정책", tradingKnown, p.TradingWired, p.TradingLoadErr, "/settings/daily#trading"),
		readinessClause(clauses, "2·6. 능력 증명서", true, true, "", ""),
		readinessClause(clauses, "4·5·7·8. Guardian 주입 · 게이트웨이 · 전송의 키 운반", true, true, "", ""),
	)
	return rows
}

func readinessGate(p settingsPage, known bool) readinessRow {
	if !known {
		return readinessUnknown("자동화 게이트", p.LimitsWired, p.LimitsLoadErr, "/settings/standing#gate")
	}
	if p.Gate.Enabled {
		return readinessRow{Label: "자동화 게이트", State: "ok", Detail: "켜짐", Href: "/settings/standing#gate"}
	}
	return readinessRow{Label: "자동화 게이트", State: "warn", Detail: "꺼짐", Href: "/settings/standing#gate"}
}

func readinessAutostart(p settingsPage) readinessRow {
	if !p.AutostartWired || p.AutostartLoadErr != "" {
		return readinessUnknown("엔진 자동 시작", p.AutostartWired, p.AutostartLoadErr, "/settings/standing#autostart")
	}
	if p.Autostart {
		return readinessRow{Label: "엔진 자동 시작", State: "ok", Detail: "켜짐", Href: "/settings/standing#autostart"}
	}
	return readinessRow{Label: "엔진 자동 시작", State: "warn", Detail: "꺼짐", Href: "/settings/standing#autostart"}
}

func readinessClause(clauses []preflightClause, label string, known, wired bool, loadErr, href string) readinessRow {
	if !known {
		return readinessUnknown(label, wired, loadErr, href)
	}
	for _, clause := range clauses {
		if clause.Name != label {
			continue
		}
		if !clause.Judgeable {
			return readinessRow{Label: clause.Name, State: "unknown", Detail: clause.Detail, Href: href}
		}
		if clause.OK {
			return readinessRow{Label: clause.Name, State: "ok", Detail: clause.Detail, Href: href}
		}
		return readinessRow{Label: clause.Name, State: "warn", Detail: clause.Detail, Href: href}
	}
	return readinessRow{Label: label, State: "unknown", Detail: "사전 판정 항목을 찾을 수 없다", Href: href}
}

func readinessUnknown(label string, wired bool, loadErr, href string) readinessRow {
	if !wired {
		return readinessRow{Label: label, State: "unknown", Detail: "설정 seam이 배선되지 않았다", Href: href}
	}
	return readinessRow{Label: label, State: "unknown", Detail: "설정을 읽을 수 없다: " + strings.TrimSpace(loadErr), Href: href}
}

// startupReadiness 는 검증 콘솔에서 필요한 설정 seam만 읽어 준비도를 만듦.
func (c *Console) startupReadiness() []readinessRow {
	p := settingsPage{}
	if c.opts.Limits != nil {
		p.LimitsWired = true
		gate, err := c.opts.Limits.Load()
		if err != nil {
			p.LimitsLoadErr = err.Error()
		}
		p.Gate = gate
	}
	if c.opts.TradingPolicy != nil {
		p.TradingWired = true
		trading, err := c.opts.TradingPolicy.Load()
		if err != nil {
			p.TradingLoadErr = err.Error()
		}
		p.Trading = trading
	}
	if c.opts.EngineBoot != nil {
		p.AutostartWired = true
		on, err := c.opts.EngineBoot.Load()
		if err != nil {
			p.AutostartLoadErr = err.Error()
		}
		p.Autostart = on
	}
	return p.StartupReadiness()
}

// limitTierMatrix 는 통화 하나의 현재값과 프리셋 값을 나란히 나타냄.
type limitTierMatrix struct {
	Currency    string
	Rows        []limitTierMatrixRow
	Tiers       []tierCard
	Consequence string
	Note        string
}

// limitTierMatrixRow 는 한도 하나의 현재값과 티어별 값을 보존함.
type limitTierMatrixRow struct {
	Label   string
	Current string
	Values  []limitTierMatrixValue
}

// limitTierMatrixValue 는 기존 미리보기에서 옮긴 티어별 셀 값임.
type limitTierMatrixValue struct {
	TierID string
	Value  string
	Axis   string
}

// LimitTierMatrices 는 카드 미리보기 값을 통화별 비교표로 옮김.
func (p settingsPage) LimitTierMatrices() []limitTierMatrix {
	var out []limitTierMatrix
	byCurrency := map[string]int{}
	for _, card := range p.LimitTierCards() {
		currency := strings.TrimSpace(card.Limits.Currency)
		index, exists := byCurrency[currency]
		if !exists {
			index = len(out)
			byCurrency[currency] = index
			matrix := limitTierMatrix{Currency: currency, Consequence: card.Preview.Consequence, Note: limitTierMatrixNote(card)}
			for _, row := range card.Preview.Rows {
				matrix.Rows = append(matrix.Rows, limitTierMatrixRow{Label: row.Label, Current: row.From})
			}
			currentCurrency := card.Preview.CurrencyFrom
			if card.Preview.FirstTime {
				currentCurrency = "미설정"
			}
			matrix.Rows = append(matrix.Rows, limitTierMatrixRow{Label: "계좌 기준 통화", Current: currentCurrency})
			out = append(out, matrix)
		}
		out[index].Tiers = append(out[index].Tiers, card)
		for rowIndex, row := range card.Preview.Rows {
			out[index].Rows[rowIndex].Values = append(out[index].Rows[rowIndex].Values, limitTierMatrixValue{
				TierID: card.ID,
				Value:  row.To,
				Axis:   row.Axis,
			})
		}
		currencyRow := len(out[index].Rows) - 1
		out[index].Rows[currencyRow].Values = append(out[index].Rows[currencyRow].Values, limitTierMatrixValue{
			TierID: card.ID,
			Value:  card.Limits.Currency,
		})
	}
	return out
}

func limitTierMatrixNote(card tierCard) string {
	if card.Preview.FirstTime {
		return "최초 설정 — 이전 값이 없으므로 강화도 완화도 아니다"
	}
	if card.Preview.CurrencyChange {
		return card.Preview.CurrencyFrom + " → " + card.Preview.CurrencyTo + " 통화 변경 — 강화도 완화도 아닌 별개의 변경이다"
	}
	return ""
}
