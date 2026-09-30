package main

// engine_attempt_resolve.go 는 a094 D−4.5 park 해동 명령의 운영자 표면임(`tossctl engine attempt-resolve`).
//
// park(UNRESOLVED_IN_DOUBT)는 원 주문의 존재를 증명할 수 없어 사람에게 넘겨진 attempt 임. 운영자가 브로커에서 그 주문을
// 확인한 뒤 이 명령으로 닫는다 — 없으면 FAILED_CONFIRMED(그 attempt 가 무장한 발의를 같은 명령 안에서 해제해 손절이 다시
// 발의되게 함), 있으면 CONFIRMED(주문 번호 필수 — 발의는 체결 경로가 끝냄).
//
// 완화 명령 가족 계약(a066 · a092)을 따름: `mutating: true`(대화형 에이전트는 자동 실행하지 않음) · 운영자 · 승인 참조 · note
// 필수 · 엔진 제어 endpoint 경유(엔진이 자기 journal 과 audit 로그로 실행, audit 줄이 원장 전이보다 먼저) · 명령 시점에 park 가
// 아니면 거절 · **엔진 lock 을 잡지 않음**(엔진을 멈추면 손절이 없음) · 콘솔 버튼 없음. 브로커를 호출하지 않음.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attemptthaw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
	"github.com/spf13/cobra"
)

type attemptThawClient interface {
	ResolveParkedAttempt(context.Context, attemptthaw.Request) (attemptthaw.Result, error)
}

type attemptThawDeps struct {
	engineDir func(*rootOptions) (string, error)
	dial      func(ctx context.Context, engineDir string) (attemptThawClient, error)
}

func productionAttemptThawDeps() attemptThawDeps {
	return attemptThawDeps{
		engineDir: engineJournalDir,
		dial: func(ctx context.Context, engineDir string) (attemptThawClient, error) {
			descriptor := positionpolicyrpc.DescriptorPath(engineDir)
			if _, err := os.Stat(descriptor); err != nil {
				return nil, err
			}
			client, err := positionpolicyrpc.Dial(ctx, descriptor)
			if err != nil {
				return nil, err
			}
			return client, nil
		},
	}
}

func newEngineAttemptResolveCmd(root *rootOptions) *cobra.Command {
	return newEngineAttemptResolveCmdWithDeps(root, productionAttemptThawDeps())
}

func newEngineAttemptResolveCmdWithDeps(root *rootOptions, deps attemptThawDeps) *cobra.Command {
	var req attemptthaw.Request
	cmd := &cobra.Command{
		Use:   "attempt-resolve",
		Short: "Close one parked (UNRESOLVED_IN_DOUBT) mutation attempt after checking the broker (audited)",
		Long: strings.TrimSpace(`
Close one attempt the engine parked as UNRESOLVED_IN_DOUBT, after you have checked the
broker yourself. --target FAILED_CONFIRMED says the order does not exist at the broker;
the exit proposal it armed is then released in the same command, so the next
observation can propose the stop again. --target CONFIRMED says it exists and needs
--broker-order-id; the proposal then ends through the fill path.

The running engine performs the resolution through its control endpoint, with its own
journal and audit log; the audit line is written before the ledger transition. The
command is refused when the attempt is no longer parked. It does not stop the engine,
takes no engine lock and places no order. --operator, --approval and --note are required.`),
		Annotations:  map[string]string{"source": "local", "mutating": "true"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineAttemptResolve(cmd, root, deps, req)
		},
	}
	cmd.Flags().StringVar(&req.AttemptID, "attempt", "", "Parked attempt id")
	cmd.Flags().StringVar(&req.Target, "target", "", "FAILED_CONFIRMED (absent at the broker) or CONFIRMED (present)")
	cmd.Flags().StringVar(&req.BrokerOrderID, "broker-order-id", "", "Broker order number found (required for CONFIRMED)")
	cmd.Flags().StringVar(&req.Operator, "operator", "", "Operator identity recorded with the resolution")
	cmd.Flags().StringVar(&req.Approval, "approval", "", "Human approval reference (ticket, runbook step, approver)")
	cmd.Flags().StringVar(&req.Note, "note", "", "What was checked at the broker")
	for _, name := range []string{"attempt", "target", "operator", "approval", "note"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}

func runEngineAttemptResolve(cmd *cobra.Command, root *rootOptions, deps attemptThawDeps, req attemptthaw.Request) error {
	for _, f := range []struct{ flag, value string }{
		{"--attempt", req.AttemptID}, {"--target", req.Target}, {"--operator", req.Operator},
		{"--approval", req.Approval}, {"--note", req.Note},
	} {
		if strings.TrimSpace(f.value) == "" {
			return fmt.Errorf("engine attempt-resolve: %s is required", f.flag)
		}
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if deps.engineDir == nil || deps.dial == nil {
		return errors.New("engine attempt-resolve: wiring is incomplete")
	}
	dir, err := deps.engineDir(root)
	if err != nil {
		return err
	}
	client, err := deps.dial(ctx, dir)
	if err != nil || client == nil {
		return fmt.Errorf("engine attempt-resolve: the engine is not running or its control endpoint is unavailable (%v); "+
			"nothing was changed — start it (tossctl engine run) and retry", err)
	}
	result, err := client.ResolveParkedAttempt(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, attemptthaw.ErrStale), errors.Is(err, attemptthaw.ErrInvalidRequest),
			errors.Is(err, attemptthaw.ErrNotFound), errors.Is(err, attemptthaw.ErrAuditUnavailable),
			errors.Is(err, attemptthaw.ErrUnwired):
			return fmt.Errorf("engine attempt-resolve: refused, nothing was changed: %w", err)
		}
		return fmt.Errorf("engine attempt-resolve: the outcome is unknown (%w); a repeated resolution of an already "+
			"closed attempt is refused as stale", err)
	}
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "attempt %s closed as %s\n", result.AttemptID, result.State); err != nil {
		return err
	}
	switch {
	case result.ProposalReleased:
		_, err = fmt.Fprintf(out, "exit proposal of position %s released — the next observation may propose the stop again\n", result.PositionID)
		return err
	case result.ReleaseError != "":
		return fmt.Errorf("engine attempt-resolve: the attempt is closed, but releasing its exit proposal failed (%s); "+
			"the next engine start releases it", result.ReleaseError)
	}
	return nil
}
