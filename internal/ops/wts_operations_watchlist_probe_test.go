package ops

import "testing"

func TestWatchlistProbeDedicatedReadContract(t *testing.T) {
	op, ok := NewCatalog().Get("watchlist")
	if !ok || op.Probe == nil {
		t.Fatal("watchlist probe missing")
	}
	if op.Probe.Method != "GET" || op.Probe.URL != probeCert+"/api/v1/new-watchlists?includePrice=true&lazyLoad=false" || op.Probe.Body != "" {
		t.Fatalf("probe=%+v, want dedicated GET without body", op.Probe)
	}
	for name, body := range map[string][]byte{
		"empty":               []byte(`{"result":{"watchlists":[]}}`),
		"empty items":         []byte(`{"result":{"watchlists":[{"items":[]}]}}`),
		"missing watchlists":  []byte(`{"result":{}}`),
		"null watchlists":     []byte(`{"result":{"watchlists":null}}`),
		"object watchlists":   []byte(`{"result":{"watchlists":{}}`),
		"missing items":       []byte(`{"result":{"watchlists":[{}]}}`),
		"null items":          []byte(`{"result":{"watchlists":[{"items":null}]}}`),
		"object items":        []byte(`{"result":{"watchlists":[{"items":{}}]}}`),
		"later missing items": []byte(`{"result":{"watchlists":[{"items":[]},{}]}}`),
		"invalid json":        []byte(`{`),
	} {
		t.Run(name, func(t *testing.T) {
			err := op.Probe.Check(200, body)
			if name == "empty" || name == "empty items" {
				if err != nil {
					t.Fatalf("err=%v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("want collection contract error")
			}
		})
	}
	if err := op.Probe.Check(500, []byte(`{"result":{"watchlists":[]}}`)); err == nil {
		t.Fatal("want status error")
	}
}
