package obs

// a075_topic_test.go pins the one property the generated channel exists for:
// nobody can guess it, including the operator who pressed the button.

import (
	"strings"
	"testing"
)

// TestTwoChannelsAreNeverTheSame. A generator that returned a constant would
// satisfy every other test in this file — the length is right, the prefix is
// right — and would put every TossOS install on one public topic.
func TestTwoChannelsAreNeverTheSame(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 64; i++ {
		topic, err := NewTopic()
		if err != nil {
			t.Fatalf("NewTopic: %v", err)
		}
		if seen[topic] {
			t.Fatalf("NewTopic returned %q twice in %d draws; the channel is not random", topic, i+1)
		}
		seen[topic] = true
	}
}

// TestAChannelCarries128BitsAfterThePrefix.
//
// The prefix is for a human reading a subscription list and is not entropy. 16
// random bytes are 26 base32 characters with no padding, so the assertion is on
// what follows the prefix — a change that shortened the random part while
// lengthening the prefix would keep the total length and lose the security.
func TestAChannelCarries128BitsAfterThePrefix(t *testing.T) {
	topic, err := NewTopic()
	if err != nil {
		t.Fatalf("NewTopic: %v", err)
	}
	if !strings.HasPrefix(topic, TopicPrefix) {
		t.Fatalf("NewTopic returned %q, which does not start with %q; an operator "+
			"cannot tell what this subscription is", topic, TopicPrefix)
	}
	random := strings.TrimPrefix(topic, TopicPrefix)
	if len(random) != 26 {
		t.Errorf("the random part of %q is %d characters; 128 bits of base32 is 26",
			topic, len(random))
	}
	for _, r := range random {
		if (r < 'a' || r > 'z') && (r < '2' || r > '7') {
			t.Errorf("the channel %q carries %q, which is outside lowercase base32; "+
				"ntfy topics accept a limited alphabet and a person has to type this", topic, r)
		}
	}
}

// TestTheSubscribeAddressIsTheServerAndTheChannel.
func TestTheSubscribeAddressIsTheServerAndTheChannel(t *testing.T) {
	for _, c := range []struct {
		name, base, topic, want string
	}{
		{"explicit server", "https://ntfy.example", "tossos-abc", "https://ntfy.example/tossos-abc"},
		{"trailing slash", "https://ntfy.example/", "tossos-abc", "https://ntfy.example/tossos-abc"},
		{"empty server is the public one", "", "tossos-abc", DefaultNtfyBaseURL + "/tossos-abc"},
		{"no channel is no address", "https://ntfy.example", "", ""},
	} {
		if got := SubscribeURL(c.base, c.topic); got != c.want {
			t.Errorf("%s: SubscribeURL(%q, %q) = %q, want %q", c.name, c.base, c.topic, got, c.want)
		}
	}
}
