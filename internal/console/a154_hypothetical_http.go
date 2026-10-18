package console

import (
	"net/http"
)

type hypotheticalField struct{ Key, Label, Value, Error string }
type hypotheticalPage struct {
	chrome
	CSRF   string
	Market string
	Fields []hypotheticalField
	Result *hypotheticalResult
}

func (hypotheticalPage) Refresh() bool { return false }
func hypotheticalFields(in hypotheticalInput, errors map[string]string) []hypotheticalField {
	return []hypotheticalField{{"entry", "가정 진입가", in.Entry, errors["entry"]}, {"low", "신호봉 저가 (선택)", in.Low, errors["low"]}, {"atr", "ATR14 (선택)", in.ATR, errors["atr"]}, {"tick", "가정 호가 간격", in.Tick, errors["tick"]}, {"min", "가격 범위 하한", in.Min, errors["min"]}, {"max", "가격 범위 상한", in.Max, errors["max"]}, {"budget", "가격 위험 예산", in.Budget, errors["budget"]}, {"max_shares", "최대 수량", in.MaxShares, errors["max_shares"]}, {"max_notional", "최대 매수금액", in.MaxNotional, errors["max_notional"]}}
}
func (c *Console) hypotheticalNoStore(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Cache-Control", "no-store"); next(w, r) }
}
func (c *Console) handleHypothetical(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		http.Error(w, "query parameters are not accepted", http.StatusBadRequest)
		return
	}
	allowed := map[string]bool{"csrf": true, "market": true, "entry": true, "low": true, "atr": true, "tick": true, "min": true, "max": true, "budget": true, "max_shares": true, "max_notional": true}
	for k, v := range r.PostForm {
		if !allowed[k] || len(v) != 1 {
			http.Error(w, "invalid form fields", http.StatusBadRequest)
			return
		}
	}
	val := r.PostForm.Get
	in := hypotheticalInput{Market: val("market"), Entry: val("entry"), Low: val("low"), ATR: val("atr"), Tick: val("tick"), Min: val("min"), Max: val("max"), Budget: val("budget"), MaxShares: val("max_shares"), MaxNotional: val("max_notional")}
	result := calculateHypothetical(in)
	c.render(w, "hypothetical-result", hypotheticalPage{chrome: c.chromeOnRequest("optimization"), CSRF: c.csrf, Market: in.Market, Fields: hypotheticalFields(in, result.Errors), Result: &result})
}

func (optimizationPage) HypotheticalFields() []hypotheticalField {
	return hypotheticalFields(hypotheticalInput{}, nil)
}
func (optimizationPage) HypotheticalMarket() string                { return "" }
func (optimizationPage) HypotheticalResult() *hypotheticalResult   { return nil }
func (p hypotheticalPage) HypotheticalFields() []hypotheticalField { return p.Fields }
func (p hypotheticalPage) HypotheticalMarket() string              { return p.Market }
func (p hypotheticalPage) HypotheticalResult() *hypotheticalResult { return p.Result }

func (optimizationPage) HypotheticalMarketError() string   { return "" }
func (p hypotheticalPage) HypotheticalMarketError() string { return p.Result.Errors["market"] }
