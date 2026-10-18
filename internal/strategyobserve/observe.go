// Package strategyobserve contains offline, caller-supplied diagnostics only.
package strategyobserve

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
)

const Schema = "strategyobserve-v1"
const Calculation = "diagnostic-v1"
const MaxRecord = 65536

type Geometry struct {
	Currency    string  `json:"currency"`
	Basis       string  `json:"basis"`
	Entry       string  `json:"entry"`
	Stop        string  `json:"stop"`
	Target      string  `json:"target"`
	LVNSpacePct string  `json:"lvn_space_pct"`
	Quantity    *uint64 `json:"quantity,omitempty"`
	LaneID      string  `json:"lane_id"`
	LaneVersion string  `json:"lane_version"`
}

func (g *Geometry) UnmarshalJSON(data []byte) error {
	type plain Geometry
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if q, ok := fields["quantity"]; ok && bytes.Equal(bytes.TrimSpace(q), []byte("null")) {
		return errors.New("null quantity")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode((*plain)(g))
}

type Input struct {
	Parent        string                             `json:"parent"`
	PolicyVersion string                             `json:"policy_version"`
	Features      strategymarket.SessionFeatureInput `json:"features"`
	Geometry      *Geometry                          `json:"geometry,omitempty"`
}
type Diagnostic struct {
	Schema        string                             `json:"schema"`
	Calculation   string                             `json:"calculation"`
	ID            string                             `json:"id"`
	InputDigest   string                             `json:"input_digest"`
	Parent        string                             `json:"parent"`
	PolicyVersion string                             `json:"policy_version"`
	Provenance    string                             `json:"provenance"`
	Features      strategymarket.SessionFeatures     `json:"features"`
	Geometry      *Geometry                          `json:"geometry,omitempty"`
	Risk          string                             `json:"risk"`
	RiskCurrency  string                             `json:"risk_currency"`
	RiskState     string                             `json:"risk_state"`
	RRStatus      string                             `json:"rr_status"`
	RR            strategyengine.ParkerRRDiagnostics `json:"rr"`
}
type Envelope struct {
	ID            string          `json:"id"`
	Payload       json.RawMessage `json:"payload"`
	PayloadDigest string          `json:"payload_digest"`
}

var decimal = regexp.MustCompile(`^[0-9]+(?:\.[0-9]{1,8})?$`)

func textOK(s string) bool {
	if len(s) > 128 {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func number(s string) (*big.Rat, bool) {
	if len(s) == 0 || len(s) > 32 || !decimal.MatchString(s) {
		return nil, false
	}
	v, ok := new(big.Rat).SetString(s)
	return v, ok
}
func digest(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

// Calculate never consumes a sealed decision or execution authority.
func Calculate(in Input) (Diagnostic, error) {
	f := in.Features
	if !textOK(in.Parent) || in.Parent == "" || !textOK(in.PolicyVersion) || in.PolicyVersion == "" {
		return Diagnostic{}, errors.New("invalid identifiers")
	}
	for _, v := range []string{f.Market, f.Symbol, f.Currency, f.TradingDate, f.CalendarSource, f.CalendarVersion} {
		if !textOK(v) {
			return Diagnostic{}, errors.New("invalid metadata")
		}
	}
	if len(f.Minutes) > 390 || len(f.Fives) > 78 {
		return Diagnostic{}, errors.New("bar limit exceeded")
	}
	f.OpenAt = f.OpenAt.UTC()
	f.CloseAt = f.CloseAt.UTC()
	f.AsOf = f.AsOf.UTC()
	f.Minutes = append([]strategymarket.SessionFeatureBar(nil), f.Minutes...)
	f.Fives = append([]strategymarket.SessionFeatureBar(nil), f.Fives...)
	for _, bars := range [][]strategymarket.SessionFeatureBar{f.Minutes, f.Fives} {
		for i := range bars {
			b := &bars[i]
			b.CloseAt = b.CloseAt.UTC()
			b.ReceivedAt = b.ReceivedAt.UTC()
			for _, s := range []string{b.Market, b.Symbol, b.Currency, b.Source, b.Basis, b.Open, b.High, b.Low, b.Close, b.Volume} {
				if !textOK(s) {
					return Diagnostic{}, errors.New("invalid bar text")
				}
			}
		}
	}
	in.Features = f
	if in.Geometry != nil {
		g := *in.Geometry
		if g.Quantity != nil {
			n := *g.Quantity
			g.Quantity = &n
		}
		in.Geometry = &g
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return Diagnostic{}, err
	}
	id := digest(append([]byte(Schema+"\x00"+Calculation+"\x00"), raw...))
	d := Diagnostic{Schema: Schema, Calculation: Calculation, ID: id, InputDigest: digest(raw), Parent: in.Parent, PolicyVersion: in.PolicyVersion, Provenance: "caller-supplied, unverified; offline", Features: strategymarket.CalculateSessionFeatures(f), RiskState: "unknown", RRStatus: "unknown"}
	if f.Market == "US" {
		d.RRStatus = "unsupported"
	}
	if in.Geometry == nil {
		return d, nil
	}
	g := *in.Geometry
	for _, v := range []string{g.Currency, g.Basis, g.LaneID, g.LaneVersion, g.LVNSpacePct} {
		if !textOK(v) {
			return Diagnostic{}, errors.New("invalid geometry metadata")
		}
	}
	want := ""
	switch f.Market {
	case "KR":
		want = "KRW"
	case "US":
		want = "USD"
	default:
		return Diagnostic{}, errors.New("unsupported geometry market")
	}
	if g.Currency != want || f.Currency != want || g.Basis != "unadjusted" {
		return Diagnostic{}, errors.New("geometry currency or basis mismatch")
	}
	for _, bars := range [][]strategymarket.SessionFeatureBar{f.Minutes, f.Fives} {
		for _, b := range bars {
			if b.Basis != "unadjusted" {
				return Diagnostic{}, errors.New("mixed basis")
			}
		}
	}
	e, ok := number(g.Entry)
	if !ok || e.Sign() <= 0 {
		return Diagnostic{}, errors.New("invalid entry")
	}
	s, ok := number(g.Stop)
	if !ok || s.Sign() <= 0 || s.Cmp(e) >= 0 {
		return Diagnostic{}, errors.New("invalid stop")
	}
	if g.Target != "" {
		v, ok := number(g.Target)
		if !ok || v.Cmp(e) <= 0 {
			return Diagnostic{}, errors.New("invalid target")
		}
	}
	if g.Quantity != nil && (*g.Quantity == 0 || *g.Quantity > 1000000) {
		return Diagnostic{}, errors.New("invalid quantity")
	}
	d.Geometry = &g
	if g.Quantity != nil {
		d.Risk = new(big.Rat).Mul(new(big.Rat).Sub(e, s), new(big.Rat).SetInt(new(big.Int).SetUint64(*g.Quantity))).RatString()
		d.RiskCurrency = want
		d.RiskState = "known"
	}
	if f.Market == "US" {
		d.RRStatus = "unsupported"
		return d, nil
	}
	if g.LaneID == strategyengine.LaneID && g.LaneVersion == strategyengine.LaneVersion {
		d.RR = strategyengine.ParkerRRFromDecision(strategyengine.DecisionRecord{LaneID: g.LaneID, LaneVersion: g.LaneVersion, Currency: want, EntryPrice: g.Entry, StopPrice: g.Stop, TargetPrice: g.Target, LVNSpacePct: g.LVNSpacePct})
		d.RRStatus = string(d.RR.Target.State)
	}
	return d, nil
}
func Encode(d Diagnostic) ([]byte, error) {
	payload, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(Envelope{ID: d.ID, Payload: payload, PayloadDigest: digest(payload)})
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	if len(b) > MaxRecord {
		return nil, fmt.Errorf("record exceeds %d bytes", MaxRecord)
	}
	return b, nil
}
func ValidateLine(line []byte) (Envelope, error) {
	var e Envelope
	if len(line) == 0 || len(line) > MaxRecord || line[len(line)-1] != '\n' || bytes.Count(line, []byte{'\n'}) != 1 {
		return e, errors.New("invalid record boundary")
	}
	dec := json.NewDecoder(strings.NewReader(string(line)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return e, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return e, errors.New("trailing record JSON")
	}
	var d Diagnostic
	dec = json.NewDecoder(strings.NewReader(string(e.Payload)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return e, err
	}
	if err := dec.Decode(&extra); err != io.EOF {
		return e, errors.New("trailing payload JSON")
	}
	if _, err := hex.DecodeString(e.ID); err != nil {
		return e, errors.New("invalid record ID")
	}
	if len(e.ID) != 64 || e.ID != d.ID || digest(e.Payload) != e.PayloadDigest || d.Schema != Schema || d.Calculation != Calculation {
		return e, errors.New("record identity or digest mismatch")
	}
	return e, nil
}
