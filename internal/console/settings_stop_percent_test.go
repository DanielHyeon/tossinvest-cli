package console

import (
	"net/url"
	"strconv"
	"testing"
)

func TestInvalidStopPercentWritesNothing(t *testing.T) {
	for _, raw := range []string{
		"", "not-a-number", "NaN", "+Inf", "-Inf",
		"1.5", "20.5", "7.4999999999", "7.5000000001", "7.6", "0.05",
	} {
		t.Run(raw, func(t *testing.T) {
			seam := &fakeSettings{}
			h := settingsHarness(t, seam)
			h.authenticate(t)
			h.post(t, "/settings/save", url.Values{
				"csrf":                 {h.csrf},
				"default_stop_percent": {raw},
				"include_symbols":      {"005930"},
			})
			if _, saves := seam.saved(); saves != 0 {
				t.Errorf("invalid percentage %q wrote the settings block", raw)
			}
		})
	}
}

func TestEveryAllowedStopPercentConvertsToFraction(t *testing.T) {
	for halfPercentTicks := 4; halfPercentTicks <= 40; halfPercentTicks++ {
		raw := strconv.FormatFloat(float64(halfPercentTicks)/2, 'f', -1, 64)
		t.Run(raw, func(t *testing.T) {
			seam := &fakeSettings{}
			h := settingsHarness(t, seam)
			h.authenticate(t)
			h.post(t, "/settings/save", url.Values{
				"csrf":                 {h.csrf},
				"default_stop_percent": {raw},
			})
			block, saves := seam.saved()
			want := float64(halfPercentTicks) / 200
			if saves != 1 || block.DefaultStopPct != want {
				t.Errorf("%s%% saved %+v after %d saves, want fraction %v", raw, block, saves, want)
			}
		})
	}
}
