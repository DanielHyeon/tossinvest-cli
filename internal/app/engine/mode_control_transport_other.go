//go:build !unix

package engine

// 모드 완화 엔드포인트는 Unix 소켓(파일 권한이 접근 통제)이라 다른 대상에서는 없음 — CLI 는 descriptor 를 못 찾음.
// ⛔ 원장만 고치는 대체 완화 경로를 두지 않음: 산 게이트는 이 프로세스 안에 있고, 원장만 고친 완화는 재시작 전까지 진입을 못 풂.

type ModeControlServer struct{}

func StartModeControlServer(string, *ModeOperations) (*ModeControlServer, error) {
	return &ModeControlServer{}, nil
}

func (*ModeControlServer) Close() error { return nil }
