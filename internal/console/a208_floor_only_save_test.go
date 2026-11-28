package console

// a208 design 8.2 T9 — 콘솔 편입 설정 저장은 폼에 없는 engine.adoption.floor_only_exit 를 이월함.
// 폼에는 그 칸이 없으므로(사람 결정 Q3 — 콘솔 체크박스 없음), 저장이 값을 이월하지 않으면 손절폭만 고쳐도 토글이 조용히 꺼짐.

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestA208ConsoleSaveKeepsFloorOnly(t *testing.T) {
	for _, on := range []bool{true, false} {
		seam := &fakeSettings{block: config.Adoption{Enabled: true, DefaultStopPct: 0.05, FloorOnlyExit: on}}
		h := settingsHarness(t, seam)
		h.authenticate(t)

		resp := h.post(t, "/settings/save", url.Values{
			"csrf":                 {h.csrf},
			"enabled":              {"on"},
			"default_stop_percent": {"7.5"},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("save = %d", resp.StatusCode)
		}
		block, saves := seam.saved()
		if saves != 1 || block.DefaultStopPct != 0.075 {
			t.Fatalf("saved block = %+v after %d saves", block, saves)
		}
		if block.FloorOnlyExit != on {
			t.Fatalf("floor_only_exit was %v before the save and %v after — the form has no such field, "+
				"so the save must carry it over", on, block.FloorOnlyExit)
		}
	}
}
