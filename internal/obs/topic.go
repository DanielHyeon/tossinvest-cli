package obs

// topic.go makes the notification channel rather than asking for one (change
// a075, engine-safety "알림 채널 식별자는 기계가 만든다").
//
// # Why there is no input field anywhere for this
//
// ntfy.go already says it: "the topic name is the only access control ntfy.sh
// has, and an engine publishing account events to a guessable public topic is an
// information leak." a074 left the topic to the operator, in the config file or
// in an environment variable, which means the access control was whatever a
// person typed at 2am — `tossos`, the account nickname, the product name.
//
// So the console does not ask. It presses a button and this function answers.
// The whole design decision is that there is no path on which a human picks a
// bad one, which is stronger than validating the one they picked: a validator
// has to explain what it rejects, that explanation is "it must be random
// enough", and at that point the machine should have made it.
//
// # The prefix is not entropy and is not counted as any
//
// `tossos-` exists so an operator scanning a subscription list in the ntfy app
// can tell what this channel is. The secret is the 26 characters after it.

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

// TopicPrefix is the human-readable marker in front of the random part.
const TopicPrefix = "tossos-"

// topicEntropyBytes is 128 bits. ntfy topics are guessed by asking the public
// server for them, not by an offline attack, so this is far past what the threat
// needs — and a topic is typed once and then lives in an app's subscription
// list, so there is no cost to being past it.
const topicEntropyBytes = 16

// NewTopic returns a fresh channel identifier.
//
// An error is returned rather than falling back to a weaker source. A channel
// this function could not make random is a channel that must not exist: the
// caller's answer is to say the button did not work, not to publish account
// events somewhere guessable.
func NewTopic() (string, error) {
	buf := make([]byte, topicEntropyBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("obs: generating a notification channel: %w", err)
	}
	// Lowercase base32 without padding: ntfy accepts [A-Za-z0-9_-], and a
	// lowercase alphabet with no `=` is what a person can read off a screen and
	// type into a phone without asking which characters are significant.
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	return TopicPrefix + strings.ToLower(encoded), nil
}

// SubscribeURL is the address a person opens to receive this channel.
//
// It is the same URL the publisher POSTs to, which is not a coincidence — ntfy
// serves the subscriber UI and the publish endpoint at one path — and it is why
// this is the only thing the operator has to be told.
func SubscribeURL(baseURL, topic string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = DefaultNtfyBaseURL
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ""
	}
	return base + "/" + topic
}
