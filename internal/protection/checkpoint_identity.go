package protection

// Generation and Identity expose only validated audit coordinates. They do
// not expose the checkpoint seal and cannot be used to construct one.
func (checkpoint Checkpoint) Generation() uint64 {
	if !checkpoint.Valid() {
		return 0
	}
	return checkpoint.generation
}

func (checkpoint Checkpoint) Identity() string {
	if !checkpoint.Valid() {
		return ""
	}
	return checkpoint.identity
}

func (checkpoint Checkpoint) Market() string {
	if !checkpoint.Valid() {
		return ""
	}
	return string(checkpoint.market)
}
