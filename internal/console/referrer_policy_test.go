package console

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConsoleDocumentsUseSameOriginReferrerPolicy(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     any
	}{
		{
			name:     "shared head",
			template: "head",
			data:     dashboardPage{Nav: "overview"},
		},
		{
			name:     "restart interstitial",
			template: "restart",
			data:     restartPage{Target: "/", Delay: restartRefreshDelay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			(&Console{}).render(w, tt.template, tt.data)

			if w.Code != http.StatusOK {
				t.Fatalf("render status = %d, body=%s", w.Code, w.Body.String())
			}
			if got := w.Header().Get("Referrer-Policy"); got != "same-origin" {
				t.Errorf("Referrer-Policy = %q, want same-origin", got)
			}
			if body := w.Body.String(); !strings.Contains(body,
				`<meta name="referrer" content="same-origin">`) {
				t.Errorf("rendered document has no same-origin referrer meta: %s", body)
			} else if strings.Contains(body, `content="no-referrer"`) {
				t.Errorf("rendered document retains a no-referrer override: %s", body)
			}
		})
	}
}
