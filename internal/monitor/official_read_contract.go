package monitor

import "net/http"

// OfficialReadContractProbes returns schema probes for official Open API reads
// that the strategy runtime depends on. These probes are contract-only: Run is
// a WTS session-cookie runner and must not send them without a separately
// designed OAuth transport. Keeping the registry separate makes an accidental
// unauthenticated live request mechanically testable while still pinning the
// upstream schema and one-request rate budget.
func OfficialReadContractProbes() []Probe {
	return []Probe{{
		Name:              "official-exchange-rate",
		Method:            http.MethodGet,
		URL:               "https://openapi.tossinvest.com/api/v1/exchange-rate?baseCurrency=USD&quoteCurrency=KRW",
		Source:            "official-open-api",
		ContractOnly:      true,
		MaxRequestsPerRun: 1,
		Check: func(status int, body []byte) error {
			if err := expectStatus(status, http.StatusOK); err != nil {
				return err
			}
			for _, path := range []string{
				"result.baseCurrency",
				"result.quoteCurrency",
				"result.rate",
				"result.midRate",
				"result.validFrom",
				"result.validUntil",
			} {
				if err := expectPath(body, path, "string"); err != nil {
					return err
				}
			}
			return nil
		},
	}}
}
