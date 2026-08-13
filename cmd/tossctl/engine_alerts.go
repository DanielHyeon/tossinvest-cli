package main

// engine_alerts.go 는 a098 4.4 의 운영자 표면이다 — 밀린 critical 알림을 **읽는**
// 명령 하나와 **승인하는** 명령 하나.
//
// # 이 명령이 왜 원장을 직접 안 여는가
//
// 진입 게이트는 엔진 프로세스 메모리 안의 맵이다 (execgw/retry.go). CLI 가 원장만
// 고치면 운영자는 승인을 마쳤는데 진입은 **재시작까지** 막힌 채로 남는다. 그래서
// 두 명령 다 엔진이 연 소켓에 붙고, 승인은 엔진 안의 `obs.Notifier.Acknowledge`
// 가 한다 (design D7.1).
//
// # 확인 절차를 안 넣는다
//
// 승인 명령은 `--operator` 말고 아무것도 더 묻지 않는다. 타이핑 확인이나 추가
// 승인 마찰을 넣지 않는다 (사용자 지시 2026-07-27 · task 4.4 R6). 승인은 위험을
// 늘리는 동작이 아니라 **막힌 진입을 정상으로 되돌리는** 동작이고, 그 앞에 마찰을
// 두면 사람이 급할 때 우회할 것을 찾는다.
//
// `engine reconcile-resolve` 에는 `--confirm` 이 있고 여기에는 없다. 그쪽은 엔진이
// **멈춰 있어야** 하는 감사 동작이고, 이쪽은 엔진이 **돌고 있어야** 하는 복구
// 동작이다 — 같은 마찰을 쓸 이유가 없다.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newEngineAlertsCmd(root *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Read and acknowledge the undelivered critical alert backlog",
	}
	cmd.AddCommand(newEngineAlertsListCmd(root), newEngineAlertsAckCmd(root))
	return cmd
}

func newEngineAlertsListCmd(root *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the critical alerts the engine has not delivered",
		Long: strings.TrimSpace(`
List the critical alerts recorded in the outbox that have not been delivered yet.

Each row shows its id, its event type, how long it has been waiting, how many
send attempts it has taken, and which sender currently holds the delivery lease.
A row with a lease holder is being sent right now; a row whose lease has expired
is one the next cycle will take over.

The listing deliberately carries no alert title, body, payload or account
identifier: it exists to count and age rows, and the contents are in the alert
the operator already received.

It reads through the running engine's private local socket. With no engine
running there is nothing to read and the command says so.`),
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineAlertsList(cmd, root)
		},
	}
}

func newEngineAlertsAckCmd(root *rootOptions) *cobra.Command {
	var operator string
	cmd := &cobra.Command{
		Use:   "ack [ID...]",
		Short: "Acknowledge undelivered critical alerts and release the entry latch",
		Long: strings.TrimSpace(`
Record that a person has seen the undelivered critical alerts, and let the engine
release the entry latch those alerts hold.

With no ids the whole backlog is acknowledged, which is the engine's own
convention for this operation. Naming ids acknowledges only those rows.

--operator is required and is recorded in the journal against every row. Nothing
supplies it for you: the ledger's refusal of a blank name is the audit trail's
only guard, so a name this command invented would pass every check and still
leave "who looked at this" unanswered.

Acknowledging clears exactly one entry block — the undelivered-backlog one. When
entry is also blocked for another reason, the remaining reasons are printed:
the backlog being clear does not mean trading has resumed.

This command places no order, changes no setting and cannot start or stop the
engine.`),
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		Args:         cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEngineAlertsAck(cmd, root, operator, args)
		},
	}
	cmd.Flags().StringVar(&operator, "operator", "",
		"Operator identity recorded in the journal against every acknowledged alert")
	_ = cmd.MarkFlagRequired("operator")
	return cmd
}

func runEngineAlertsList(cmd *cobra.Command, root *rootOptions) error {
	ctx, format, client, err := engineAlertsDial(cmd, root)
	if err != nil {
		return err
	}
	rows, err := client.Pending(ctx)
	if err != nil {
		return err
	}
	return writePendingAlerts(cmd.OutOrStdout(), format, rows)
}

func runEngineAlertsAck(cmd *cobra.Command, root *rootOptions, operator string, args []string) error {
	// 이름은 여기서 **검사만** 한다. 채워 주지 않는다 (R5).
	named := strings.TrimSpace(operator)
	if named == "" {
		return errors.New("engine alerts ack: --operator is required — " +
			"누가 봤는지 없는 승인은 승인이 아니다")
	}
	ids, err := parseAlertIDs(args)
	if err != nil {
		return err
	}
	ctx, format, client, err := engineAlertsDial(cmd, root)
	if err != nil {
		return err
	}
	// 원문 그대로 보낸다 — trim 한 값이 아니라. 원장이 저장하는 것과 운영자가 친
	// 것이 다르면 audit trail 은 있는데 대조가 안 된다.
	result, err := client.Acknowledge(ctx, engine.AcknowledgeRequest{IDs: ids, Operator: operator})
	if err != nil {
		return err
	}
	return writeAcknowledgeResult(cmd.OutOrStdout(), format, result)
}

// engineAlertsDial 은 두 명령이 공유하는 앞부분이다 — 형식 파싱, 엔진 디렉터리
// 해석, 소켓 연결. `engineJournalDir` 를 쓰므로 `--config-dir` 를 준 프로파일은
// 자기 엔진의 알림만 본다. 진짜 엔진의 것을 실수로 승인할 수 없다.
func engineAlertsDial(cmd *cobra.Command, root *rootOptions) (context.Context,
	output.Format, *alertControlClient, error) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return nil, "", nil, err
	}
	dir, err := engineJournalDir(root)
	if err != nil {
		return nil, "", nil, err
	}
	client, err := dialAlertControl(ctx, dir)
	if err != nil {
		return nil, "", nil, err
	}
	return ctx, format, client, nil
}

func parseAlertIDs(args []string) ([]int64, error) {
	if len(args) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(args))
	for _, arg := range args {
		id, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("engine alerts ack: %q is not an alert id", arg)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func writePendingAlerts(w io.Writer, format output.Format, rows []engine.PendingAlert) error {
	if format == output.FormatJSON {
		if rows == nil {
			rows = []engine.PendingAlert{}
		}
		return output.WriteJSON(w, rows)
	}
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "밀린 critical 알림이 없다.")
		return err
	}
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "ID\tTYPE\tAGE(s)\tATTEMPTS\tCLAIMED BY\tCLAIM AGE(s)\tLEASE")
	for _, row := range rows {
		var lease string
		switch {
		case row.ClaimedBy == "":
			lease = "unclaimed"
		case row.ClaimExpired:
			lease = "EXPIRED"
		default:
			lease = "held"
		}
		claimedBy := row.ClaimedBy
		if claimedBy == "" {
			claimedBy = "-"
		}
		fmt.Fprintf(table, "%d\t%s\t%d\t%d\t%s\t%d\t%s\n", row.ID, row.Type, row.AgeSeconds,
			row.Attempts, claimedBy, row.ClaimAgeSeconds, lease)
	}
	if err := table.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n%d개가 밀려 있다.\n", len(rows))
	return err
}

func writeAcknowledgeResult(w io.Writer, format output.Format, result engine.AcknowledgeResult) error {
	if format == output.FormatJSON {
		return output.WriteJSON(w, result)
	}
	if _, err := fmt.Fprintf(w, "승인했다. 남은 미전달 critical 알림 %d개\n",
		result.RemainingUndelivered); err != nil {
		return err
	}
	if len(result.EntryBlocks) == 0 {
		_, err := fmt.Fprintln(w, "진입 차단 없음.")
		return err
	}
	// ⛔ 이 줄이 R17 의 운영자 쪽 절반이다. 「승인했다」만 찍으면 운영자는 끝난 줄
	// 알고, 실제로는 다른 사유로 진입이 막힌 채다 — 그러면 재시작을 찾는다.
	_, err := fmt.Fprintf(w, "⚠ 진입은 아직 막혀 있다: %s\n", strings.Join(result.EntryBlocks, ", "))
	return err
}
