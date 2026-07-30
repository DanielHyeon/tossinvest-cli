package console

func (p settingsPage) UpdateCurrentTime() string {
	if p.Update.Current.ModTime.IsZero() {
		return "(알 수 없음)"
	}
	return p.Update.Current.ModTime.UTC().Format("2006-01-02 15:04:05Z")
}

func (p settingsPage) UpdateCandidateTime() string {
	if p.Update.Candidate.ModTime.IsZero() {
		return "(알 수 없음)"
	}
	return p.Update.Candidate.ModTime.UTC().Format("2006-01-02 15:04:05Z")
}
