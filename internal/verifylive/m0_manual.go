package verifylive

// m0ManualReconcileIDs returns the exact parent/child pair checkpointed after a
// trigger. They remain visible in status, but no automated cleanup or abort may
// cancel an unresolved child measurement.
func m0ManualReconcileIDs(entries []Entry) map[string]bool {
	ids := map[string]bool{}
	for _, entry := range entries {
		if entry.StepID == StepConditionalTrigger && m0TriggeredUnresolved(entry) {
			for _, artifact := range entry.Artifacts {
				if artifact.Kind == KindConditional && !artifact.Cancelled && !artifact.Filled {
					ids[KindConditional+"\x00"+artifact.ID] = true
				}
			}
		}
		if entry.Kind != KindM0Checkpoint || entry.M0Checkpoint == nil {
			continue
		}
		if entry.M0Checkpoint.Kind != "parent-created" && entry.M0Checkpoint.Kind != "child-observed" {
			continue
		}
		if entry.M0Checkpoint.ParentConditionalID != "" {
			ids[KindConditional+"\x00"+entry.M0Checkpoint.ParentConditionalID] = true
		}
		if entry.M0Checkpoint.ChildOrderID != "" {
			ids[KindOrder+"\x00"+entry.M0Checkpoint.ChildOrderID] = true
		}
	}
	return ids
}

func m0TriggeredUnresolved(entry Entry) bool {
	if entry.Verdict != VerdictFail {
		return false
	}
	for _, observation := range entry.Observations {
		if (observation.Key == "conditional.trigger_observed" && observation.Value == "true") ||
			(observation.Key == "conditional.trigger.conditional_presumed_fired" && observation.Value == "true") {
			return true
		}
	}
	return false
}

func withoutM0ManualReconcile(entries []Entry, artifacts []Artifact) []Artifact {
	ids := m0ManualReconcileIDs(entries)
	out := make([]Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if !ids[artifact.Kind+"\x00"+artifact.ID] {
			out = append(out, artifact)
		}
	}
	return out
}
