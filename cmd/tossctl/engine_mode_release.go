package main

// engine_mode_release.go 는 a092 착지 단위 ④ 의 사람 모드 완화 명령 — `tossctl engine mode-release`.
//
// 운영 모드 투영이 배선되면 모든 자동 강화(손실 한도 · 자격증명 · exit 관측 두절 · 대사 · 체결 감지 · 알림 전달 실패)가 산
// 엔진의 신규 진입을 실제로 막음. 그것을 푸는 생산 경로가 이 명령 하나임(델타 ADDED — 투영을 배선하는 빌드는 완화 경로를 함께 가짐).
//
// - `mutating: "true"` — 계좌의 진입 허용을 바꿈. 대화형 에이전트는 자동 실행하지 않음(안전 불변식 2).
// - `--to` · `--operator` · `--approval` · `--reason` 전부 필수, 기본값 없음. 타이핑 확인 · 추가 승인 마찰 없음(사용자 지시).
// - 엔진 프로세스 안에서 일어남(모드 전용 제어 소켓) — 원장만 고치면 산 게이트는 안 풀림. 엔진이 없으면 거절.
// - 출력은 엔진이 **다시 읽은** 모드 · 남은 진입 차단 사유 · 통지 행 상태.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newEngineModeReleaseCmd(root *rootOptions) *cobra.Command {
	var req engine.ModeReleaseRequest
	cmd := &cobra.Command{
		Use:   "mode-release",
		Short: "Relax the account's operating mode with a human approval (audited)",
		Long: strings.TrimSpace(`
Change the account's operating mode to a less conservative one — typically
ENTRY_BLOCKED back to NORMAL after an automatic tightening — as a named operator
with an approval reference.

The running engine performs the transition through its own mode control socket,
with its own journal and audit log: the audit line is written before the journal
commit, and an audit failure changes nothing. The live entry gate is released in
the same call. When the engine is not running the command refuses — a ledger-only
change would not release the running gate.

Only the operating-mode block is released. Every other entry block stays and is
printed, because the mode being NORMAL does not mean trading has resumed.

The release is announced to the operator channel. The notice is recorded here and
sent by the engine's delivery executor; when the record itself fails the release
still stands and the failure is shown.

--to, --operator, --approval and --reason are required and have no defaults. The
command places no order and does not stop the engine.`),
		Annotations:  map[string]string{"source": "local", "mutating": "true"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineModeRelease(cmd, root, req)
		},
	}
	cmd.Flags().StringVar(&req.To, "to", "", "Target mode: NORMAL or ENTRY_BLOCKED")
	cmd.Flags().StringVar(&req.Operator, "operator", "", "Operator identity recorded in the journal and the audit log")
	cmd.Flags().StringVar(&req.Approval, "approval", "", "Human approval reference (ticket, runbook step)")
	cmd.Flags().StringVar(&req.Reason, "reason", "", "Why the mode is being relaxed")
	for _, name := range []string{"to", "operator", "approval", "reason"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}

func runEngineModeRelease(cmd *cobra.Command, root *rootOptions, req engine.ModeReleaseRequest) error {
	// 검사만 하고 채워 주지 않음 — 원장 · audit 에 아무도 고르지 않은 값이 들어가면 안 됨.
	for name, value := range map[string]string{"--to": req.To, "--operator": req.Operator,
		"--approval": req.Approval, "--reason": req.Reason} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("engine mode-release: %s is required", name)
		}
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return err
	}
	dir, err := engineJournalDir(root)
	if err != nil {
		return err
	}
	client, err := dialModeControl(ctx, dir)
	if err != nil {
		return err
	}
	result, err := client.Release(ctx, req)
	if err != nil {
		return err
	}
	return writeModeReleaseResult(cmd.OutOrStdout(), format, result)
}

func writeModeReleaseResult(w io.Writer, format output.Format, r engine.ModeReleaseResult) error {
	if format == output.FormatJSON {
		return output.WriteJSON(w, r)
	}
	// 재조회 실패를 먼저 가름 — 읽지 못한 것은 추정하지 않음(26라운드 codex 재확인 R2). 모드 재조회가 실패하면 모드 · 사유 ·
	// 통지 상태 모두 모름. 통지 목록만 실패하면 읽은 모드 · 사유는 그대로 말하고 통지 상태만 모름.
	mode := r.Mode
	if r.ReReadError != "" {
		mode = "(재조회 실패로 확인하지 못함)"
	}
	if !r.Changed {
		fmt.Fprintf(w, "변화 없음 — 전이 행을 남기지 않았다. 현재 모드 %s\n", mode)
	} else {
		head := "완화됨"
		if r.NotifyError != "" {
			head = "완화됨 · 통지 기록 실패"
		}
		fmt.Fprintf(w, "%s — 전이 %s, 현재 모드 %s\n  통지: %s\n", head, r.TransitionID, mode, modeReleaseNoticeState(r))
	}
	if r.ReReadError != "" {
		_, err := fmt.Fprintf(w, "재조회 실패 — 현재 모드와 남은 사유를 읽지 못했다(추정하지 않음): %s\n", r.ReReadError)
		return err
	}
	if len(r.EntryBlocks) == 0 {
		_, err := fmt.Fprintln(w, "남은 진입 차단 사유: 없음")
		return err
	}
	_, err := fmt.Fprintf(w, "남은 진입 차단 사유: %s — 모드가 풀려도 거래가 재개된 것은 아니다\n", strings.Join(r.EntryBlocks, ", "))
	return err
}

// modeReleaseNoticeState 는 완화 통지 행의 상태 한 줄. PENDING 목록에 없다는 것은 「이미 전달됨」이 아님 — 전달 뒤 정산됐거나
// 운영자가 승인했을 수 있음(26라운드 보이스 B #6).
func modeReleaseNoticeState(r engine.ModeReleaseResult) string {
	switch {
	case r.NotifyError != "":
		return r.NotifyError
	case r.ReReadError != "":
		return "기록됨(전달 상태는 재조회 실패로 확인하지 못함)"
	case r.NoticeReadError != "":
		return "기록됨(전달 상태는 확인하지 못함 — " + r.NoticeReadError + ")"
	case r.NoticePending:
		return "기록됨(배달 실행자가 보냄)"
	default:
		return "기록됨(대기 목록에 없음 — 전달 뒤 정산됐거나 승인됨)"
	}
}

// modeControlClient 는 엔진의 모드 제어 소켓 연결 하나. 토큰은 출력하지 않음(안전 불변식 8).
type modeControlClient struct {
	baseURL string
	token   string
	http    *http.Client
}

type modeControlError struct {
	status        int
	code, message string
}

func (e *modeControlError) Error() string {
	if strings.TrimSpace(e.message) == "" {
		return fmt.Sprintf("engine mode-release: HTTP %d", e.status)
	}
	return fmt.Sprintf("engine mode-release: %s: %s", e.code, e.message)
}

func (c *modeControlClient) Release(ctx context.Context, req engine.ModeReleaseRequest) (engine.ModeReleaseResult, error) {
	var result engine.ModeReleaseResult
	encoded, err := json.Marshal(req)
	if err != nil {
		return result, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+engine.ModeControlReleasePath,
		bytes.NewReader(encoded))
	if err != nil {
		return result, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(httpReq)
	if err != nil {
		return result, fmt.Errorf("engine mode-release: calling the engine: %w", err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode/100 != 2 {
		var remote struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(limited).Decode(&remote)
		return result, &modeControlError{status: response.StatusCode, code: remote.Code,
			message: strings.TrimSpace(remote.Message)}
	}
	if err := json.NewDecoder(limited).Decode(&result); err != nil {
		return result, errors.New("engine mode-release: decoding the response failed")
	}
	return result, nil
}

// engineModeControlEndpoint 는 모드 완화 표면의 강등 보고 모양임. 이 표면이 없으면 모드 사유 차단을 사람이 풀 길이 없음 —
// 신규 진입이 계속 막히는 쪽이므로 보수 방향(청산 무관).
func engineModeControlEndpoint(dir string) engineEndpoint {
	return engineEndpoint{
		surface: "mode control",
		lost:    "운영자의 운영 모드 완화 표면 (모드 사유 차단은 유지되어 신규 진입이 계속 막힌다)",
		control: engine.ModeControlDirectory(dir),
		event:   engineControlEndpointDegradedEvent, title: "MODE_CONTROL_UNAVAILABLE",
	}
}
