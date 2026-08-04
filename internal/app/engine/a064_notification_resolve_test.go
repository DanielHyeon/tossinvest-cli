package engine

// a064 task 5: file ⊕ environment → transport.
//
// This is the seam the config package deliberately does not have. internal/config
// parses a file and never reads the environment, so the two halves of a
// notification channel — what an operator wants to diff, and what an operator
// must not commit — are joined here.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func fixedEnv(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestNoPublisherWhenNotificationsAreOff(t *testing.T) {
	// §0.2: this is every config that exists today, and the answer has to be the
	// same nil the build had before a064.
	publisher, resolution := resolveNotificationPublisher(config.Notifications{}, fixedEnv(nil))
	if publisher != nil {
		t.Fatalf("an unconfigured block produced a transport: %#v", publisher)
	}
	if resolution.TopicConfigured || resolution.TokenConfigured {
		t.Fatalf("resolution = %+v, want nothing configured", resolution)
	}
	if resolution.Refused != "" {
		t.Fatalf("an absent block was refused: %s", resolution.Refused)
	}
}

func TestAConfiguredChannelBecomesAPublisher(t *testing.T) {
	publisher, resolution := resolveNotificationPublisher(config.Notifications{
		Enabled: true, BaseURL: "https://ntfy.example.internal", Topic: "tossos-alerts",
	}, fixedEnv(map[string]string{envNtfyToken: "s3cret"}))

	ntfy, ok := publisher.(*obs.Ntfy)
	if !ok {
		t.Fatalf("publisher = %#v, want an *obs.Ntfy", publisher)
	}
	if ntfy.Topic != "tossos-alerts" || ntfy.BaseURL != "https://ntfy.example.internal" {
		t.Fatalf("ntfy = %+v", *ntfy)
	}
	if ntfy.Token != "s3cret" {
		t.Fatal("the token from the environment did not reach the transport")
	}
	if !resolution.TopicConfigured || !resolution.TokenConfigured {
		t.Fatalf("resolution = %+v", resolution)
	}
	if resolution.PublicService {
		t.Fatal("a self-hosted base url was reported as the public service")
	}
}

func TestTheEnvironmentTopicWinsOverTheFile(t *testing.T) {
	// On ntfy.sh the topic name is the only access control, so an operator has to
	// be able to keep it out of a file that anything on the box can read.
	publisher, resolution := resolveNotificationPublisher(config.Notifications{
		Enabled: true, Topic: "in-the-file",
	}, fixedEnv(map[string]string{envNtfyTopic: "from-the-environment"}))

	ntfy, ok := publisher.(*obs.Ntfy)
	if !ok {
		t.Fatalf("publisher = %#v", publisher)
	}
	if ntfy.Topic != "from-the-environment" {
		t.Fatalf("topic = %q, want the environment's", ntfy.Topic)
	}
	if !resolution.PublicService {
		t.Fatal("an empty base url targets ntfy.sh and must be reported as public")
	}
}

func TestAnEnabledBlockWithNoChannelIsRefused(t *testing.T) {
	publisher, resolution := resolveNotificationPublisher(config.Notifications{
		Enabled: true,
	}, fixedEnv(nil))

	if publisher != nil {
		t.Fatalf("a channel-less block produced a transport: %#v", publisher)
	}
	if resolution.Refused == "" {
		t.Fatal("a channel-less block was accepted silently")
	}
	if resolution.TopicConfigured {
		t.Fatalf("resolution = %+v", resolution)
	}
}

func TestARefusedConfigBlockStaysRefused(t *testing.T) {
	// internal/config already zeroed it and recorded why. The resolver must carry
	// that reason forward rather than inventing a second one.
	_, resolution := resolveNotificationPublisher(config.Notifications{
		Rejected: "notification base_url must be an http or https URL",
	}, fixedEnv(map[string]string{envNtfyTopic: "would-have-worked"}))

	if !strings.Contains(resolution.Refused, "http or https") {
		t.Fatalf("refusal = %q, want the config's own reason", resolution.Refused)
	}
}

func TestADisabledBlockWithAChannelStaysOff(t *testing.T) {
	// The toggle is the toggle. A topic left in the file by an operator who turned
	// notifications off must not turn them back on (§0.2, §0.7).
	publisher, resolution := resolveNotificationPublisher(config.Notifications{
		Topic: "tossos-alerts",
	}, fixedEnv(map[string]string{envNtfyToken: "s3cret"}))

	if publisher != nil {
		t.Fatalf("a disabled block produced a transport: %#v", publisher)
	}
	if resolution.Refused != "" {
		t.Fatalf("a disabled block was reported as an error: %s", resolution.Refused)
	}
}

func TestTheResolutionCarriesNoSecret(t *testing.T) {
	// §0.5 asks for the settings change to be traceable; §0.8 forbids the values.
	// The resolution is what reaches the audit trail, so it answers "configured?"
	// and never "configured as what".
	_, resolution := resolveNotificationPublisher(config.Notifications{
		Enabled: true, BaseURL: "https://ntfy.example.internal", Topic: "top-secret-topic",
	}, fixedEnv(map[string]string{envNtfyToken: "top-secret-token"}))

	rendered := strings.ToLower(resolution.BaseURL + "|" + resolution.Refused)
	if strings.Contains(rendered, "top-secret") {
		t.Fatalf("the resolution carries a secret: %+v", resolution)
	}
}

func TestANilGetenvFallsBackToTheProcessEnvironment(t *testing.T) {
	t.Setenv(envNtfyTopic, "from-the-process")
	publisher, _ := resolveNotificationPublisher(config.Notifications{Enabled: true}, nil)
	ntfy, ok := publisher.(*obs.Ntfy)
	if !ok {
		t.Fatalf("publisher = %#v", publisher)
	}
	if ntfy.Topic != "from-the-process" {
		t.Fatalf("topic = %q", ntfy.Topic)
	}
}
