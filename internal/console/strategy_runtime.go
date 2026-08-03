package console

import "net/http"

func (c *Console) handleStrategyRuntime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다", "전략 상태는 GET/HEAD만 허용한다. 아무것도 전송되지 않았다.")
		return
	}

	// This child route intentionally does not mark a top-level navigation item
	// as current. The projection itself remains server-owned and read-only.
	page := c.buildMultiMarketStrategyRuntimePage(r)
	c.render(w, "strategy-runtime", page)
}
