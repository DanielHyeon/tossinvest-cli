package engine

// a134_notification_publisher.go — 엔진 밖 감시자(a134, `tossctl engine heartbeat-check`)가 엔진과 **같은 규칙**으로 알림 전송 수단을 얻는 입구.
// 규칙(enabled · 파일 토픽 ⊕ TOSSCTL_NTFY_TOPIC · 토큰 env · 설정 거절)은 resolveNotificationPublisher 한 곳에만 있고 여기는 부르기만 함 — 사본 금지.

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// NotificationPublisher 는 엔진 조립과 같은 해석으로 알림 전송 수단을 돌려줌. 만들 수 없으면 nil 과 이유 문구(비밀 값 없음)를 돌려줌.
//
// 이유가 비는 경우는 알림이 꺼진 경우뿐이라 그 문구는 여기서 채움 — 감시자는 이유 없이 실패하지 않음(design D4).
// getenv 가 nil 이면 os.Getenv(엔진과 같음).
func NotificationPublisher(cfg config.Notifications, getenv func(string) string) (obs.Publisher, string) {
	publisher, resolution := resolveNotificationPublisher(cfg, getenv)
	if publisher != nil {
		return publisher, ""
	}
	if resolution.Refused != "" {
		return nil, resolution.Refused
	}
	return nil, "notifications are off (engine.notifications.enabled is false or the block is absent)"
}
