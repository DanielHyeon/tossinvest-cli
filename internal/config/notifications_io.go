package config

// notifications_io.go is the console's fifth config write surface: where critical
// alerts are sent (change a065).
//
// # Two savers, two closed member lists, and the reason there are two
//
// limits_io.go established the rule: a save emits bytes for its own keys and for
// nothing else, because re-emitting a block means re-emitting keys the screen did
// not edit, sourced from a read taken outside the file lock.
//
//	notificationMembersOf   enabled, base_url, topic   turning alerts ON
//	notificationSwitch      enabled                    turning them OFF
//
// The two lists are not the same length on purpose. Turning alerts off must not
// erase the channel: the operator's phone is subscribed to that identifier, and a
// save that deleted it would turn "mute this for an hour" into "re-subscribe
// every device". Leaving it is safe because the engine's assembly already refuses
// to build a transport from a topic sitting under `enabled: false` — "a topic
// left behind in the file by an operator who turned notifications off does not
// turn them back on (§0.7)".
//
// # `rejected` is not in either list
//
// It is a diagnostic the parser attaches to a block it refused, not a setting an
// operator writes. A save that emitted it would write this process's opinion into
// the file, and the next load would read that opinion back as configuration.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// LoadRawNotifications reads the engine.notifications block as written.
//
// A missing file is the zero block and no error, for the reason LoadRawEngineGate
// gives: the screen renders 꺼짐, which is the truth, rather than an error nobody
// can act on.
//
// It returns what the file spells and applies no validation verdict. The
// screen's question is "what is configured", and mergeNotifications' zeroing of a
// refused block must not round-trip through this reader and erase the channel an
// operator is subscribed to.
func (s *Service) LoadRawNotifications() (Notifications, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Notifications{}, nil
	}
	if err != nil {
		return Notifications{}, err
	}
	var doc struct {
		Engine struct {
			Notifications rawNotifications `json:"notifications"`
		} `json:"engine"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return Notifications{}, fmt.Errorf("config: parsing %s: %w", s.path, err)
	}
	raw := doc.Engine.Notifications
	out := Notifications{BaseURL: raw.BaseURL, Topic: raw.Topic}
	if raw.Enabled != nil {
		out.Enabled = *raw.Enabled
	}
	return out, nil
}

// SaveNotifications writes the three keys that configure delivery.
//
// There is no token parameter and no token key, which is the same structural
// answer notifications.go gives for the parse side: a token has nowhere to land,
// so it cannot land here by accident either.
func (s *Service) SaveNotifications(n Notifications) error {
	return s.spliceInto(func(data []byte) ([]byte, error) {
		members, err := notificationMembersOf(n)
		if err != nil {
			return nil, err
		}
		return spliceMembers(data, notificationsValueSpan, insertEmptyNotifications, members)
	})
}

// SaveNotificationsEnabled writes the switch, and that one key.
//
// Separate from SaveNotifications for the reason SaveEngineGateEnabled is
// separate from SaveGuardianLimits: keeping the channel out of the off path makes
// "turning alerts off cannot lose the subscription" a property of the member list
// rather than a promise in a comment.
func (s *Service) SaveNotificationsEnabled(on bool) error {
	return s.spliceInto(func(data []byte) ([]byte, error) {
		return spliceMembers(data, notificationsValueSpan, insertEmptyNotifications,
			notificationSwitch(on))
	})
}

// notificationMembersOf is the complete, closed list of what an ON save writes.
func notificationMembersOf(n Notifications) ([]gateMember, error) {
	base, err := json.Marshal(n.BaseURL)
	if err != nil {
		return nil, err
	}
	topic, err := json.Marshal(n.Topic)
	if err != nil {
		return nil, err
	}
	return []gateMember{
		{"enabled", []byte(boolLiteral(n.Enabled))},
		{"base_url", base},
		{"topic", topic},
	}, nil
}

// notificationSwitch is the complete, closed list of what an OFF save writes.
func notificationSwitch(on bool) []gateMember {
	return []gateMember{{"enabled", []byte(boolLiteral(on))}}
}

// notificationsValueSpan locates engine.notifications.
func notificationsValueSpan(data []byte) (start, end int64, found bool, err error) {
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil || !eFound {
		return 0, 0, false, err
	}
	nStart, nEnd, nFound, err := valueSpan(data[eStart:eEnd], "notifications")
	if err != nil || !nFound {
		return 0, 0, false, err
	}
	return eStart + nStart, eStart + nEnd, true, nil
}

// insertEmptyNotifications creates the block, and the engine block above it when
// that is missing too. Empty on purpose: the members are spliced afterwards by
// the same loop that handles an existing block, so exactly one code path decides
// which keys a save writes.
func insertEmptyNotifications(data []byte) ([]byte, error) {
	eStart, eEnd, eFound, err := valueSpan(data, "engine")
	if err != nil {
		return nil, err
	}
	if eFound {
		return insertKey(data, eStart, eEnd, "notifications", []byte("{}"))
	}
	out, err := insertEmptyEngine(data)
	if err != nil {
		return nil, err
	}
	eStart, eEnd, eFound, err = valueSpan(out, "engine")
	if err != nil {
		return nil, err
	}
	if !eFound {
		return nil, errors.New("config: the engine block could not be created")
	}
	return insertKey(out, eStart, eEnd, "notifications", []byte("{}"))
}
