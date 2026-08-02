package console

// explain.go is how a disclosure stays open on a screen that reloads itself
// (change a055-console-settings-cadence §6).
//
// # Why not <details>
//
// A native <details> on a screen with a meta refresh opens when clicked and
// closes on the next tick. That is worse than having no fold at all: the
// operator learns the control does not work, and the prose they wanted is now
// behind a control they have stopped trusting.
//
// The first design put the open id in the URL and kept the <details>. That is
// the same bug wearing a hat — the native triangle does not change the URL, so
// clicking it opens a panel the next reload closes, while the link beside it
// opens one that persists. Two controls for one thing, disagreeing.
//
// So the reloading screens get no native toggle at all. A link sets ?explain=id,
// the server renders that one panel open, and the reload carries the parameter
// with it because it is in the address bar. Back and forward work. A link is
// shareable. No JavaScript, which this console does not have anyway.
//
// # It is display only
//
// The parameter reaches no judgement, no save and no audit record. An id nobody
// recognises is ignored and everything renders folded: an unknown value is not an
// error, because the only thing it can possibly be wrong about is which
// paragraph is visible.

import (
	"net/http"
	"net/url"
	"strings"
)

// explainQueryKey is the parameter, spelled once.
const explainQueryKey = "explain"

// explainState is which disclosure is open on this render, and how to link to
// the others.
type explainState struct {
	// Open is the id the URL asked for. It is whatever the request said — this
	// type does not hold a registry of valid ids, because an id that matches
	// nothing simply matches nothing and every panel stays folded.
	Open string
	// base is the screen's own path with every query parameter EXCEPT explain,
	// so a link can flip the fold without dropping a filter the operator set.
	base string
	// query is those surviving parameters, already encoded.
	query string
}

// explainFrom reads the state out of a request.
func explainFrom(r *http.Request) explainState {
	if r == nil || r.URL == nil {
		return explainState{}
	}
	values := r.URL.Query()
	state := explainState{Open: strings.TrimSpace(values.Get(explainQueryKey))}
	values.Del(explainQueryKey)
	// A notice is a one-shot answer to something that already happened. Carrying
	// it into every fold link would make it reappear each time a panel is opened,
	// long after the save it describes.
	values.Del("notice")
	state.base = r.URL.Path
	state.query = values.Encode()
	return state
}

// Is reports that this disclosure is the open one.
func (e explainState) Is(id string) bool { return e.Open != "" && e.Open == id }

// Href is the link that opens this disclosure, or closes it when it is already
// the open one. One parameter means one panel at a time, which is the point: the
// fold exists because the screen was too long to read.
func (e explainState) Href(id string) string {
	values := url.Values{}
	if e.query != "" {
		values, _ = url.ParseQuery(e.query)
	}
	if !e.Is(id) {
		values.Set(explainQueryKey, id)
	}
	path := e.base
	if path == "" {
		path = "."
	}
	if encoded := values.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}

// Toggle is the word on the link, so a reader knows which way it goes before
// clicking.
func (e explainState) Toggle(id string) string {
	if e.Is(id) {
		return "접기"
	}
	return "펼치기"
}
