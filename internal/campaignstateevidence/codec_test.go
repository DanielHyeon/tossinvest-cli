package campaignstateevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixture() Evidence {
	return Evidence{Schema: "campaign-state-evidence:v1", Family: "continuation", CapturedAt: "2026-10-19T00:00:00Z", Purpose: "inspection-only", Origin: "caller-provided-snapshot", Currentness: "ledger-currentness-unverified", Plan: Plan{Scope: Scope{Account: "acct", Market: "KR", Symbol: "A", Campaign: "c", Lane: "kr_short_flow_continuation_v1", Version: "v1", Generation: 1}, Budget: "10", PerShareRisk: "1", Quantity: 14, Ceilings: []uint64{8, 4, 2}, Weights: []uint64{8, 4, 2}, PolicyDigest: "policy", ConfigDigest: "config", AccountCurrency: "KRW", QuoteCurrency: "KRW", Digest: "plan"}, Risk: Risk{PlanDigest: "plan", Budget: "10", Filled: "12", Held: "3", Latches: map[string]bool{"UNKNOWN_ACTUAL_RISK": true, "CAMPAIGN_RISK_OVERAGE": false}, ContinuationFills: map[string]ContinuationFill{"f": {Applied: true, Known: false, Risk: "", Fingerprint: "fp"}}, ContinuationCancels: map[string]ContinuationCancel{"c": {Applied: true, Fingerprint: "cp"}}, CancelStatus: "captured", ReservationStatus: "not_applicable"}}
}
func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func TestOfflineRoundTripAndScope(t *testing.T) {
	e := fixture()
	b, err := Encode(e)
	if err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "snapshot.json")
	if err = os.WriteFile(f, b, 0600); err != nil {
		t.Fatal(err)
	}
	reopened, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(reopened, sum(b), e.Plan.Scope)
	if err != nil {
		t.Fatal(err)
	}
	if got.Risk.Filled != "12" || got.Risk.Held != "3" || got.Risk.ContinuationFills["f"].Risk != "" || !got.Risk.Latches["UNKNOWN_ACTUAL_RISK"] || got.Risk.Latches["CAMPAIGN_RISK_OVERAGE"] {
		t.Fatalf("risk lost: %+v", got.Risk)
	}
	wrong := e.Plan.Scope
	wrong.Symbol = "B"
	if result, err := Decode(b, sum(b), wrong); err == nil || result.Schema != "" {
		t.Fatal("foreign scope accepted")
	}
}
func TestStrictCanonicalBoundaries(t *testing.T) {
	e := fixture()
	b, err := Encode(e)
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]byte{append(bytes.Clone(b), ' '), bytes.Replace(b, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1), bytes.Replace(b, []byte(`"family":`), []byte(`"unexpected":1,"family":`), 1), bytes.Replace(b, []byte(`"schema":"campaign-state-evidence:v1"`), []byte(`"schema":"future"`), 1), b[:len(b)-1], bytes.Repeat([]byte("x"), 4*1024*1024+1)}
	for _, v := range cases {
		if got, err := Decode(v, sum(v), e.Plan.Scope); err == nil || got.Schema != "" {
			t.Fatalf("accepted invalid %q", string(v[:min(len(v), 60)]))
		}
	}
	changed := bytes.Replace(b, []byte(`"filled":"12"`), []byte(`"filled":"13"`), 1)
	if bytes.Equal(changed, b) {
		t.Fatal("mutation failed")
	}
	if _, err := Decode(changed, sum(b), e.Plan.Scope); err == nil {
		t.Fatal("changed input accepted")
	}
	if _, err := Decode(b, strings.Repeat("0", 64), e.Plan.Scope); err == nil {
		t.Fatal("foreign digest accepted")
	}
}

func TestRejectsMismatchedLaneWeightsAndFX(t *testing.T) {
	original := fixture()
	b, err := Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ old, new string }{
		{`"lane":"kr_short_flow_continuation_v1"`, `"lane":"us_short_participation_continuation_v1"`},
		{`"version":"v1"`, `"version":"v9"`},
		{`"weights":[8,4,2]`, `"weights":[8,2,4]`},
		{`"ceilings":[8,4,2]`, `"ceilings":[8,4,3]`},
		{`"quote_currency":"KRW"`, `"quote_currency":"USD"`},
	}
	for _, tc := range tests {
		changed := bytes.Replace(b, []byte(tc.old), []byte(tc.new), 1)
		if bytes.Equal(changed, b) {
			t.Fatalf("mutation missing %s", tc.old)
		}
		scope := original.Plan.Scope
		if tc.new == `"lane":"us_short_participation_continuation_v1"` {
			scope.Lane = "us_short_participation_continuation_v1"
		}
		if tc.new == `"version":"v9"` {
			scope.Version = "v9"
		}
		if _, err := Decode(changed, sum(changed), scope); err == nil {
			t.Fatalf("accepted %s", tc.new)
		}
	}
}

func TestMutableEvidenceMustRevalidate(t *testing.T) {
	cases := []func(*Evidence){func(e *Evidence) { e.Plan.Scope.Lane = "other" }, func(e *Evidence) { e.Plan.Weights[0] = 7 }, func(e *Evidence) { e.Risk.PlanDigest = "other" }, func(e *Evidence) { e.Plan.QuoteCurrency = "USD" }, func(e *Evidence) { e.Risk.Filled = "-1" }, func(e *Evidence) { e.Currentness = "verified" }}
	for i, mutate := range cases {
		e := fixture()
		mutate(&e)
		if _, err := Encode(e); err == nil {
			t.Fatalf("invalid mutation %d", i)
		}
	}
}

func TestMalformedFXValuesRejected(t *testing.T) {
	e := fixture()
	e.Plan.QuoteCurrency = "USD"
	yes := true
	e.Plan.FX = &FX{QuoteID: "quote", AsOf: "2026-10-19T00:00:00Z", Direction: "QUOTE_TO_ACCOUNT", Rate: "2", Haircut: "1.1", Digest: "digest", Seal: strings.Repeat("a", 64), Official: &yes, Frozen: &yes}
	if _, err := Encode(e); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"invalid", "1"}, {"0", "1"}, {"1", "0.9"}, {"1p10000000", "1"}} {
		changed := e
		fx := *e.Plan.FX
		changed.Plan.FX = &fx
		fx.Rate, fx.Haircut = pair[0], pair[1]
		if _, err := Encode(changed); err == nil {
			t.Fatalf("accepted FX rate %q haircut %q", pair[0], pair[1])
		}
	}
}

func TestNilAndEmptyMapsRemainDistinct(t *testing.T) {
	for _, empty := range []bool{false, true} {
		e := fixture()
		e.Risk.Latches = nil
		e.Risk.ContinuationFills = nil
		e.Risk.ContinuationCancels = nil
		if empty {
			e.Risk.Latches = map[string]bool{}
			e.Risk.ContinuationFills = map[string]ContinuationFill{}
			e.Risk.ContinuationCancels = map[string]ContinuationCancel{}
		}
		b, err := Encode(e)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(b, sum(b), e.Plan.Scope)
		if err != nil {
			t.Fatal(err)
		}
		if (got.Risk.Latches == nil) == empty || (got.Risk.ContinuationFills == nil) == empty || (got.Risk.ContinuationCancels == nil) == empty {
			t.Fatal("nil/empty distinction lost")
		}
		wrong := e.Plan.Scope
		wrong.Account = "other"
		got, err = Decode(b, sum(b), wrong)
		if err == nil || !reflect.DeepEqual(got, Evidence{}) {
			t.Fatal("error leaked partial result")
		}
	}
}

func TestRejectsRedistributedCeilings(t *testing.T) {
	e := fixture()
	e.Plan.Ceilings = []uint64{14, 0, 0}
	if _, err := Encode(e); err == nil {
		t.Fatal("sum-preserving redistribution accepted")
	}
}

func TestRejectsMissingZeroFieldsAndNestedDuplicate(t *testing.T) {
	e := fixture()
	e.Risk.Latches = nil
	b, err := Encode(e)
	if err != nil {
		t.Fatal(err)
	}
	changes := []struct{ old, new string }{
		{`"known":false,`, ``},
		{`"latches":null,`, ``},
		{`"known":false,`, `"known":false,"known":false,`},
	}
	for _, tc := range changes {
		changed := bytes.Replace(b, []byte(tc.old), []byte(tc.new), 1)
		if bytes.Equal(changed, b) {
			t.Fatalf("fixture missing %s", tc.old)
		}
		got, err := Decode(changed, sum(changed), e.Plan.Scope)
		if err == nil || !reflect.DeepEqual(got, Evidence{}) {
			t.Fatalf("accepted %s", tc.old)
		}
	}
	changed := append(bytes.Clone(b), []byte(`{}`)...)
	if got, err := Decode(changed, sum(changed), e.Plan.Scope); err == nil || !reflect.DeepEqual(got, Evidence{}) {
		t.Fatal("accepted trailing object")
	}
}

// 각 문자열 필드와 맵 키를 따로 손상해 검증 대상 누락을 감지한다.
func TestTextValidationCoversEveryStringField(t *testing.T) {
	e := fixture()
	e.Plan.FX = &FX{}
	e.Risk.ReversalFills = map[string]ReversalFill{"x": {}}
	e.Risk.ReversalCancels = map[string]string{"x": "x"}
	e.Risk.WeeklyFills = map[string]WeeklyFill{"x": {}}
	var variants func(reflect.Value) []reflect.Value
	variants = func(v reflect.Value) []reflect.Value {
		switch v.Kind() {
		case reflect.String:
			out := reflect.New(v.Type()).Elem()
			out.SetString(string([]byte{0xff}))
			return []reflect.Value{out}
		case reflect.Struct:
			var result []reflect.Value
			for i := 0; i < v.NumField(); i++ {
				for _, replacement := range variants(v.Field(i)) {
					clone := reflect.New(v.Type()).Elem()
					clone.Set(v)
					clone.Field(i).Set(replacement)
					result = append(result, clone)
				}
			}
			return result
		case reflect.Pointer:
			if v.IsNil() {
				return nil
			}
			var result []reflect.Value
			for _, replacement := range variants(v.Elem()) {
				clone := reflect.New(v.Type().Elem())
				clone.Elem().Set(replacement)
				result = append(result, clone)
			}
			return result
		case reflect.Map:
			var result []reflect.Value
			iter := v.MapRange()
			for iter.Next() {
				k, value := iter.Key(), iter.Value()
				for _, replacement := range variants(k) {
					clone := reflect.MakeMap(v.Type())
					iter2 := v.MapRange()
					for iter2.Next() {
						clone.SetMapIndex(iter2.Key(), iter2.Value())
					}
					clone.SetMapIndex(k, reflect.Value{})
					clone.SetMapIndex(replacement, value)
					result = append(result, clone)
				}
				for _, replacement := range variants(value) {
					clone := reflect.MakeMap(v.Type())
					iter2 := v.MapRange()
					for iter2.Next() {
						clone.SetMapIndex(iter2.Key(), iter2.Value())
					}
					clone.SetMapIndex(k, replacement)
					result = append(result, clone)
				}
			}
			return result
		}
		return nil
	}
	cases := variants(reflect.ValueOf(e))
	if len(cases) < 40 {
		t.Fatalf("insufficient string cases: %d", len(cases))
	}
	for i, v := range cases {
		if validText(v.Interface().(Evidence)) {
			t.Fatalf("string case %d ignored", i)
		}
	}
}
