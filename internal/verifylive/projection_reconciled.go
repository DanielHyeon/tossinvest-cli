package verifylive

// projection_reconciled.go — a121 보고 투영: 대사 줄로 종결된 artifact 를 status·report 가 따로 보이게 함.
// 대사 경로(reconcile*.go)가 아니라 투영 쪽이므로 종결 술어 terminal() 을 그대로 씀(술어는 한 곳).

// ReconciledArtifacts 는 기록이 대사 줄로 종결한 artifact 들임 — status·report 가 "reconciled absent" 로 보임.
// 종결은 단조라 outstandingLines 와 같은 규칙(종결 뒤의 비-종결 줄은 되살리지 않음)으로 마지막 상태를 정함.
func ReconciledArtifacts(entries []Entry) []Artifact {
	order := []string{}
	latest := map[string]Artifact{}
	for _, e := range entries {
		for _, a := range e.Artifacts {
			key := a.Kind + "\x00" + a.ID
			prev, seen := latest[key]
			if !seen {
				order = append(order, key)
			}
			if seen && prev.terminal() && !a.terminal() {
				continue
			}
			latest[key] = a
		}
	}
	var out []Artifact
	for _, key := range order {
		if a := latest[key]; a.ReconciledAbsent {
			out = append(out, a)
		}
	}
	return out
}
