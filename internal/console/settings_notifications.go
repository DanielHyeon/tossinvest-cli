package console

// settings_notifications.go is the button that turns critical alerts on (change
// a075).
//
// # What this card exists to delete
//
// a074 gave the engine a configurable transport and stopped there, so turning
// alerts on meant hand-editing `engine.notifications` into config.json and
// inventing a topic name. Nothing on any screen said what a good topic was, and
// the one place that knew — a source comment — said the topic name is the only
// access control the public service has. That is the same shape
// console-owns-the-operating-toggles found on the automation gate: the setting
// with the largest consequence was on the path with the least protection.
//
// # The console picks nothing and knows nothing
//
// Every value this card writes is chosen by the seam, not here. `Enable` takes no
// argument: the console can press it, and what gets written — a fresh
// cryptographic channel, the server address, the audit line — is the seam's
// business. That is why this package still names no config service, no random
// source, no transport and no audit log, which the static guards already require
// of it.
//
// # The channel is rendered and never redirected
//
// redirectSettings answers through a URL query parameter, and URLs live in
// browser history and proxy logs. The channel identifier is an access control, so
// no notice on this card carries it (design D5). The card body renders it from
// Load(), which is the server writing into a response body rather than into a
// location header.

import (
	"context"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// NotificationSettings is the console's entire write surface for alert delivery.
//
// Four methods, and the shape of the first is the safety property: Enable takes
// no parameters, so there is no value this screen can put into the file. A form
// field for a channel — or worse, for a token — has nowhere to arrive.
type NotificationSettings interface {
	// Load returns the block as written in the file.
	Load() (config.Notifications, error)
	// Enable turns delivery on, creating a channel when there is none, and
	// returns what it wrote so the screen can show the subscribe address without
	// a second read that might fail on its own.
	Enable() (config.Notifications, error)
	// Disable writes the switch and keeps the channel.
	Disable() error
	// Test sends one message to the configured channel, outside the outbox.
	Test(context.Context) error
}

// NotificationOn reports the file's switch.
func (p settingsPage) NotificationOn() bool { return p.Notifications.Enabled }

// NotificationServer is the address alerts are sent to, as the file spells it.
func (p settingsPage) NotificationServer() string {
	return strings.TrimSpace(p.Notifications.BaseURL)
}

// NotificationChannel is the channel identifier, or "" when none is configured.
func (p settingsPage) NotificationChannel() string {
	return strings.TrimSpace(p.Notifications.Topic)
}

// NotificationSubscribeURL is the one address the operator has to open.
//
// It is built here rather than imported from internal/obs because the console has
// no business knowing this transport's defaults: it renders what the file says,
// and a blank server means nothing has been configured yet, which the card says
// in words instead of guessing a host.
func (p settingsPage) NotificationSubscribeURL() string {
	server := strings.TrimRight(p.NotificationServer(), "/")
	channel := p.NotificationChannel()
	if server == "" || channel == "" {
		return ""
	}
	return server + "/" + channel
}

// NotificationGuard is the alert card's reasons.
//
// A refused block is a caution rather than a block: the file may hold a base_url
// the engine will not accept, and the one path that can fix it is this card's
// button, which overwrites all three keys.
func (p settingsPage) NotificationGuard() cardGuard {
	var g cardGuard
	if !p.NotificationsWired {
		g.Blocks = append(g.Blocks, saveBlock{"알림 설정 미배선",
			"알림 설정이 배선되지 않았다 — 이 빌드의 콘솔에는 알림 seam이 주입되지 않았다. " +
				"표시는 되지만 켤 수 없다."})
	}
	if p.NotificationsLoadErr != "" {
		g.Blocks = append(g.Blocks, saveBlock{"알림 설정 읽기 실패", p.NotificationsLoadErr})
	}
	if strings.TrimSpace(p.Notifications.Rejected) != "" {
		g.Cautions = append(g.Cautions, saveBlock{"엔진이 거부할 블록",
			p.Notifications.Rejected + " — [알림 켜기]가 세 키를 다시 써서 고친다."})
	}
	g.Cautions = append(g.Cautions, p.engineRunningCaution()...)
	return g
}

// handleSettingsNotificationsOn turns delivery on and proves it in one press.
//
// Order matters and it is the order of durability: the setting is written first
// and the test message second. A test that fails after a successful save leaves
// alerts configured — which is design D8's most deliberate decision, because the
// alternative makes whether alerts are configured depend on whether the network
// was up for one second.
func (c *Console) handleSettingsNotificationsOn(w http.ResponseWriter, r *http.Request) {
	if c.opts.Notifications == nil {
		c.refuse(w, http.StatusNotImplemented, "알림 설정이 배선되지 않았다",
			"이 빌드의 콘솔에는 알림 설정 seam이 주입되지 않았다.")
		return
	}
	if _, err := c.opts.Notifications.Enable(); err != nil {
		c.redirectSettings(w, r, "알림 켜기 실패 — "+err.Error())
		return
	}
	// Nothing from the returned block goes into this sentence. The subscribe
	// address is rendered by the card from its own read, for the reason the file
	// header gives.
	notice := "알림 켜짐 — 아래 구독 주소를 열면 이 채널을 받는다. " +
		"주소는 방금 만들어졌고 이 화면 밖 어디에도 기록되지 않는다."
	if err := c.opts.Notifications.Test(r.Context()); err != nil {
		c.redirectSettings(w, r, notice+" 테스트 발송은 실패했다: "+err.Error()+
			" — 설정은 켜진 채로 남아 있다. 구독한 뒤 [테스트 한 통 더]로 다시 확인한다.")
		return
	}
	c.redirectSettings(w, r, notice+" 테스트 알림 한 통을 방금 보냈다 — "+
		"구독 전에 보낸 것이라 지금은 도착하지 않을 수 있다. 구독한 뒤 [테스트 한 통 더]를 누른다. "+
		effectNotice(c.engineRunning()))
}

// handleSettingsNotificationsTest sends one message and changes nothing.
func (c *Console) handleSettingsNotificationsTest(w http.ResponseWriter, r *http.Request) {
	if c.opts.Notifications == nil {
		c.refuse(w, http.StatusNotImplemented, "알림 설정이 배선되지 않았다",
			"이 빌드의 콘솔에는 알림 설정 seam이 주입되지 않았다.")
		return
	}
	if err := c.opts.Notifications.Test(r.Context()); err != nil {
		c.redirectSettings(w, r, "테스트 발송 실패 — "+err.Error()+
			". 설정은 그대로다: 이 버튼은 보내기만 하고 아무것도 바꾸지 않는다.")
		return
	}
	c.redirectSettings(w, r, "테스트 알림 한 통을 보냈다. 도착하지 않으면 구독 주소를 다시 확인한다 — "+
		"이 발송은 outbox를 거치지 않으므로 도착 여부가 진입 게이트에 영향을 주지 않는다.")
}

// handleSettingsNotificationsOff writes the switch and keeps the channel.
func (c *Console) handleSettingsNotificationsOff(w http.ResponseWriter, r *http.Request) {
	if c.opts.Notifications == nil {
		c.refuse(w, http.StatusNotImplemented, "알림 설정이 배선되지 않았다",
			"이 빌드의 콘솔에는 알림 설정 seam이 주입되지 않았다.")
		return
	}
	if err := c.opts.Notifications.Disable(); err != nil {
		c.redirectSettings(w, r, "알림 끄기 실패 — "+err.Error())
		return
	}
	c.redirectSettings(w, r, "알림 꺼짐 — critical 알림은 다시 outbox에 쌓이기만 한다. "+
		"구독 주소는 지우지 않았으므로 다시 켜면 같은 주소로 돌아온다. "+effectNotice(c.engineRunning()))
}
