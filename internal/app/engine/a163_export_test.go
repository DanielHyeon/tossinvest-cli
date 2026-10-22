package engine

// a163 시험 전용 노출 — 보고 래치 맵(포지션 id → 마지막 구간)의 사본. 시험이 정리(held 기준)와 OFF 의 「래치 항목 0」 을 잼.
func (o *ExitObserver) SourceAgeLatchForTest() map[string]string {
	out := make(map[string]string, len(o.sourceAge))
	for id, rec := range o.sourceAge {
		out[id] = rec.bucket
	}
	return out
}
