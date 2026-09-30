//go:build unix

package main

// 게이트 준비 gstack 리뷰(testing): CLI 는 토큰을 보내기 **전에** descriptor 모양을 검사함 — 소켓 이름(다른 경로로 토큰을
// 보내지 않게) · pid · 토큰 길이 · 모르는 필드 · 크기 상한. 각 거절은 자기 문구로 멈추고(소켓 검사까지 가지 않음),
// 올바른 모양은 소켓 검사까지 감(양성 대조 — 소켓이 없으니 거기서 실패).

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
)

func a092WriteModeDescriptor(t *testing.T, body string) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "a092-desc-*") // sun_path 한도 — a108 하니스와 같은 이유
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(engine.ModeControlDirectory(dir), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(engine.ModeControlDescriptorPath(dir), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func a092Descriptor(socket, token string, pid int) string {
	b, _ := json.Marshal(engine.ModeControlDescriptor{Socket: socket, Token: token, PID: pid})
	return string(b)
}

func TestA092TheModeReleaseClientChecksTheDescriptorBeforeSendingTheToken(t *testing.T) {
	good := strings.Repeat("t", 32)
	sock := engine.ModeControlSocketFileName()
	cases := []struct{ name, body, want string }{
		{"foreign-socket", a092Descriptor("../x.sock", good, 42), "descriptor fields are invalid"},
		{"no-pid", a092Descriptor(sock, good, 0), "descriptor fields are invalid"},
		{"short-token", a092Descriptor(sock, strings.Repeat("t", 31), 42), "descriptor fields are invalid"},
		{"unknown-field", `{"socket":"` + sock + `","token":"` + good + `","pid":42,"extra":1}`, "descriptor JSON is invalid"},
		{"oversized", `{"socket":"` + sock + `","token":"` + strings.Repeat("t", 4<<10) + `","pid":42}`, "unreadable or oversized"},
		{"control", a092Descriptor(sock, good, 42), "validating the socket"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := a092WriteModeDescriptor(t, c.body)
			_, err := dialModeControl(context.Background(), dir)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
}
