package monitor

import (
	"strings"
	"testing"
)

func TestOfficialFXContractProbeIsSeparatedFromWTSSessionRunner(t *testing.T) {
	probes := OfficialReadContractProbes()
	if len(probes) != 1 {
		t.Fatalf("official contract probes=%v, want one exchange-rate probe", probes)
	}
	probe := probes[0]
	if probe.Name != "official-exchange-rate" || probe.Method != "GET" ||
		probe.URL != "https://openapi.tossinvest.com/api/v1/exchange-rate?baseCurrency=USD&quoteCurrency=KRW" ||
		probe.Source != "official-open-api" || !probe.ContractOnly || probe.MaxRequestsPerRun != 1 {
		t.Fatalf("official FX probe=%+v", probe)
	}
	for _, wtsProbe := range Probes() {
		if strings.Contains(wtsProbe.URL, "openapi.tossinvest.com") || wtsProbe.Name == probe.Name || wtsProbe.ContractOnly {
			t.Fatalf("official OAuth probe leaked into WTS session runner: %+v", wtsProbe)
		}
	}
}

func TestOfficialFXContractProbeFailsClosedOnCriticalSchema(t *testing.T) {
	probe := OfficialReadContractProbes()[0]
	valid := []byte(`{"result":{"baseCurrency":"USD","quoteCurrency":"KRW","rate":"1380.5000","midRate":"1375.2500","validFrom":"2026-08-04T09:30:00+09:00","validUntil":"2026-08-04T09:31:00+09:00"}}`)
	if err := probe.Check(200, valid); err != nil {
		t.Fatalf("valid official FX schema refused: %v", err)
	}
	fields := []string{"baseCurrency", "quoteCurrency", "rate", "midRate", "validFrom", "validUntil"}
	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			body := []byte(strings.Replace(string(valid), `"`+field+`":"`, `"missing_`+field+`":"`, 1))
			if err := probe.Check(200, body); err == nil {
				t.Fatalf("probe accepted body missing %s", field)
			}
		})
	}
	pii := []byte(`{"result":{"accountNo":"9999999999","token":"do-not-leak"}}`)
	err := probe.Check(500, pii)
	if err == nil {
		t.Fatal("official FX probe accepted status 500")
	}
	if strings.Contains(err.Error(), "9999999999") || strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("official FX probe leaked response content: %v", err)
	}
}
