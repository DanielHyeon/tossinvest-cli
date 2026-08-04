package config

// notifications.go adds the alert transport to the engine's operating settings
// (change a064, engine-safety "알림 전송 경로는 설정 가능하고 그 설정은 audit된다").
//
// # What was missing
//
// engine-safety has always said critical alerts are written to a durable outbox
// and then *sent*. The sending half had no configuration surface, so
// cmd/tossctl's engine assembly passed no Publisher and said why: "configuring a
// transport is an operational setting with an audit trail, which is a change of
// its own." The cost showed up in the operational ledger as six critical alerts
// sitting PENDING with attempts=0 — not failed sends, sends that were never
// attempted.
//
// # There is no field here a token can be written into
//
// That is the design, not an omission (§0.8). The token is read from the
// environment at the assembly point and never passes through this struct, so a
// token pasted into config.json has nowhere to land. The topic may also be
// supplied from the environment, because on ntfy.sh the topic name *is* the
// access control and a config file readable by anything on the box is the wrong
// place for it. A self-hosted instance behind a token has no such problem and
// keeping the topic in the file makes it diffable, so both are allowed.
//
// This package reads no environment variable. It parses a file; resolving
// file ⊕ environment belongs where the engine is assembled, and keeping it out
// of here is what lets these tests be independent of the process they run in.
//
// # Off is the default and off is what everyone has
//
// §0.2: the zero value is `Enabled: false`, which is what every config written
// before this change produces, and it wires no publisher — exactly today's
// behaviour, including today's consequences for an undelivered critical alert.

import "strings"

// Notifications configures where critical alerts are sent.
type Notifications struct {
	// Enabled turns the transport on. Default false, and false is the only value
	// a pre-a064 config can produce.
	Enabled bool `json:"enabled"`

	// BaseURL is the ntfy server. Empty uses the public service, which the engine
	// warns about rather than refusing: an operator may be running a private
	// ntfy.sh topic deliberately.
	BaseURL string `json:"base_url,omitempty"`

	// Topic is the ntfy topic. It may be left empty here and supplied from the
	// environment instead; see the file header.
	Topic string `json:"topic,omitempty"`

	// Rejected explains why the block was refused, and is empty when it was
	// accepted. A refused block is zeroed, so `Enabled` is already false — this
	// field exists so the engine can say *why* rather than silently ignoring what
	// an operator wrote.
	Rejected string `json:"rejected,omitempty"`
}

// validate reports why the block cannot be used, or "" when it can.
//
// It checks only what a file can be wrong about on its own. Whether a channel
// was ultimately resolved is not decidable here — the environment may supply the
// topic — so that judgement belongs to the assembly point and is recorded there.
func (n Notifications) validate() string {
	base := strings.TrimSpace(n.BaseURL)
	if base == "" {
		return ""
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		return "notification base_url must be an http or https URL"
	}
	return ""
}

// rawNotifications is the parse shape. Enabled is a pointer for the same reason
// the gate's and adoption's are: an explicit false and an absent key are
// distinguishable in tests, and both produce false.
//
// There is deliberately no token field. See the file header.
type rawNotifications struct {
	Enabled *bool  `json:"enabled"`
	BaseURL string `json:"base_url"`
	Topic   string `json:"topic"`
}

// mergeNotifications copies a present notifications block onto the defaults, or
// refuses it.
//
// A refused block is *zeroed*, not partially kept, for the same reason adoption's
// is: half a transport configuration is not a safer transport configuration, and
// an operator who wrote a malformed URL should find the transport off with a
// reason rather than on and pointed somewhere unexpected.
func mergeNotifications(cfg *Engine, raw *rawNotifications) {
	if raw == nil {
		return
	}
	next := Notifications{
		BaseURL: strings.TrimSpace(raw.BaseURL),
		Topic:   strings.TrimSpace(raw.Topic),
	}
	if raw.Enabled != nil {
		next.Enabled = *raw.Enabled
	}
	if why := next.validate(); why != "" {
		cfg.Notifications = Notifications{Rejected: why}
		return
	}
	cfg.Notifications = next
}
