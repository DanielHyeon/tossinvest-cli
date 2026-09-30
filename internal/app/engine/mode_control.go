package engine

// mode_control.go 는 a092 착지 단위 ④ 의 운영자 모드 완화 표면 — 엔드포인트 이름과 요청 · 결과 모양임.
//
// # 왜 알림 제어 엔드포인트에 얹지 않는가 (design D0.3g 8 C12)
//
// 알림 제어 descriptor 를 가진 사람은 밀린 알림을 읽고 승인할 수 있음. 모드 완화는 계좌의 신규 진입 허용 자체를
// 바꾸는 다른 힘임 — 같은 토큰에 얹으면 그 토큰을 이미 쥔 모두의 권한이 넓어짐(alert_control.go 의 「다른 힘이면
// 다른 엔드포인트」). 그래서 자기 디렉터리 · 소켓 · 토큰을 가짐.

import (
	"path/filepath"
	"strings"
)

const (
	modeControlDirectoryName      = ".mode-control"
	modeControlDescriptorFileName = "endpoint.json"
	// 11자 이상 — 형제 공용 staging 이름(11자)이 최종 이름보다 길면 sun_path 상한 직전 배포에서 bind 만 죽음(a109).
	modeControlSocketFileName = "modectl.sock"

	// ModeControlReleasePath 는 사람의 모드 완화(`tossctl engine mode-release`) 라우트임.
	ModeControlReleasePath = "/v1/mode/release"
)

// ModeControlDescriptor 는 로컬 클라이언트가 소켓을 찾고 권한을 증명하는 방법임(알림 제어와 같은 모양, 다른 토큰).
type ModeControlDescriptor struct {
	Socket string `json:"socket"`
	Token  string `json:"token"`
	PID    int    `json:"pid"`
}

// ModeReleaseRequest 는 완화 요청임. 네 칸 모두 필수이고 이 패키지 어디에도 기본값이 없음 — 아무도 고르지 않은 값이
// 원장 · audit 에 들어가면 「누가 · 왜 · 무슨 승인으로」가 사라짐.
type ModeReleaseRequest struct {
	To       string `json:"to"`
	Operator string `json:"operator"`
	Approval string `json:"approval"`
	Reason   string `json:"reason"`
}

// ModeReleaseResult 는 완화 뒤 **다시 읽은** 상태임(델타 ADDED — 같은 호출 안의 다른 경로가 방금 푼 모드를 다시 조일 수 있음).
type ModeReleaseResult struct {
	// Changed 는 이 요청이 전이 행을 남겼는지. false 면 이미 그 모드였음(원장의 무변화 규칙).
	Changed      bool   `json:"changed"`
	TransitionID string `json:"transition_id,omitempty"`
	// Mode · Seq 는 원장에서 다시 읽은 현재 모드와 그 커밋 순번.
	Mode string `json:"mode"`
	Seq  int64  `json:"seq,omitempty"`
	// EntryBlocks 는 다시 읽은 남은 진입 차단 사유 코드(설명은 싣지 않음 — alertops 와 같은 이유).
	EntryBlocks []string `json:"entry_blocks,omitempty"`
	// Notified 는 완화 통지 행이 **기록**됐는지. 전송은 배달 실행자의 일이라 이 호출 안에서 드러나지 않음(델타 「통지 실패 = 기록 실패」).
	Notified    bool   `json:"notified"`
	NotifyError string `json:"notify_error,omitempty"`
	// NoticePending 은 다시 읽은 통지 행이 아직 전달 전(PENDING)인지.
	NoticePending bool `json:"notice_pending"`
}

func ModeControlDirectory(engineDir string) string {
	return filepath.Join(strings.TrimSpace(engineDir), modeControlDirectoryName)
}

func ModeControlDescriptorPath(engineDir string) string {
	return filepath.Join(ModeControlDirectory(engineDir), modeControlDescriptorFileName)
}

func ModeControlSocketPath(engineDir string) string {
	return filepath.Join(ModeControlDirectory(engineDir), modeControlSocketFileName)
}

// ModeControlSocketFileName 은 descriptor 의 Socket 칸이 같아야 하는 값임.
func ModeControlSocketFileName() string { return modeControlSocketFileName }
