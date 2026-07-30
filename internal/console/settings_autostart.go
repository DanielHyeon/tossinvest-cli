package console

import (
	"net/http"
	"strings"
)

// EngineBootSettings is the console's entire write surface for
// engine.autostart. It intentionally carries neither the automation gate nor
// Guardian limits: lifecycle approval and order capability are separate acts.
type EngineBootSettings interface {
	Load() (bool, error)
	Save(bool) error
}

// handleSettingsAutostart records the lifecycle approval and, when turning it
// on, reuses the exact same engine start seam as the dashboard button.
func (c *Console) handleSettingsAutostart(w http.ResponseWriter, r *http.Request) {
	if c.opts.EngineBoot == nil {
		c.refuse(w, http.StatusNotImplemented, "엔진 자동 시작 저장이 배선되지 않았다",
			"이 빌드의 콘솔에는 엔진 자동 시작 설정 seam이 주입되지 않았다.")
		return
	}

	on := r.PostFormValue("enabled") != ""
	if err := c.opts.EngineBoot.Save(on); err != nil {
		c.redirectSettings(w, r, "엔진 자동 시작 저장 안 됨 — "+err.Error())
		return
	}

	if !on {
		c.redirectSettings(w, r, "엔진 자동 시작 OFF 기록됨. 다음 콘솔 기동부터 자동 시작하지 않는다. "+
			"현재 엔진은 [엔진 정지] 버튼으로 별도 정지해야 한다.")
		return
	}

	if c.opts.StartEngine == nil {
		reason := "엔진 자동 시작 ON 기록됨 — 이 빌드에는 엔진 기동 배선이 없다."
		c.noteEngine(reason)
		c.redirectSettings(w, r, reason)
		return
	}

	note, err := c.opts.StartEngine()
	if err != nil {
		reason := "엔진 자동 시작 ON 기록됨 — 엔진이 기동을 거부했다: " + err.Error()
		if strings.TrimSpace(note) != "" {
			reason += "\n" + note
		}
		c.noteEngine(reason)
		c.redirectSettings(w, r, "엔진 자동 시작 ON 기록됨 — 기동 거부 사유는 아래에 표시한다.")
		return
	}
	if strings.TrimSpace(note) == "" {
		note = "엔진을 시작했다."
	}
	reason := "엔진 자동 시작 ON 기록됨 — " + note
	c.noteEngine(reason)
	c.redirectSettings(w, r, reason)
}
