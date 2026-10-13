package main

// engine_heartbeat_check.go — `tossctl engine heartbeat-check`: 엔진 밖에서 엔진의 침묵을 판정하는 명령(openspec change a134).
//
// 호스트 systemd user 타이머(deploy/systemd/tossos-heartbeat.timer)가 매 분 부름. 엔진 · 저널 · 브로커 · 세션을 열지 않고,
// 읽는 것은 config.json(알림 설정 · engine.autostart)과 엔진 활성 마커의 수정 시각뿐. 쓰는 것은 감시자 상태 파일 하나.
// 판정 · 상태 · 한 번 전송의 규칙은 internal/enginewatch 가 가짐 — 이 파일은 설정을 읽어 그 입력을 채우는 배선임.

import (
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginewatch"
	"github.com/spf13/cobra"
)

func newEngineHeartbeatCheckCmd(root *rootOptions) *cobra.Command {
	var staleAfter time.Duration
	cmd := &cobra.Command{
		Use:   "heartbeat-check",
		Short: "Alert once when the engine's liveness marker goes silent",
		Long: strings.TrimSpace(`
Judge, from outside the engine, whether the engine process is alive, and send one
ntfy alert when that answer changes.

The engine rewrites its active marker (engine-run.json in the journal directory)
every minute and removes it on a clean stop. This command reads only the marker's
modification time and config.json:

  marker older than --stale-after          engine down (a process that died cannot
                                           remove its marker)
  marker absent, engine.autostart on       engine down (it is supposed to run)
  marker absent, engine.autostart off      not watched (clean stop, manual operation)

The first run that sees the engine down sends one critical alert; later runs of the
same outage send nothing; the first run that sees it back sends one recovery alert.
The state lives in engine-heartbeat-watch.json beside the marker and moves only
after a send succeeds, so a failed send is retried by the next run.

It uses the engine's own notification settings. When notifications are off or have
no channel it exits non-zero and says so: a watcher that cannot alert must not look
healthy. It opens neither the engine, the journal, the broker nor a session.

Meant to be run by deploy/systemd/tossos-heartbeat.timer every minute.`),
		// local: 브로커 API 를 부르지 않음. mutating=false: 계좌 · 원장 · 설정 · 운영 토글을 바꾸지 않음 — 쓰는 것은 자기 상태 파일과 알림 한 건뿐.
		Annotations:  map[string]string{"source": "local", "mutating": "false"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineHeartbeatCheck(cmd, root, staleAfter)
		},
	}
	cmd.Flags().DurationVar(&staleAfter, "stale-after", enginewatch.DefaultStaleAfter,
		"how old the engine's active marker may be before the engine is judged down")
	return cmd
}

// runEngineHeartbeatCheck 는 설정을 읽어 enginewatch.Check 를 채우고 한 번 돌린 뒤 판정 한 줄을 출력함.
func runEngineHeartbeatCheck(cmd *cobra.Command, root *rootOptions, staleAfter time.Duration) error {
	ctx := cmd.Context()
	// 엔진 run 과 같은 디렉터리 해석 — 격리 프로파일(--config-dir)은 자기 엔진만 감시함.
	dir, err := engineJournalDir(root)
	if err != nil {
		return err
	}
	svc := configServiceFor(root)
	if svc == nil {
		return fmt.Errorf("engine heartbeat: 설정 파일 위치를 정할 수 없음")
	}
	cfg, err := svc.Load(ctx)
	if err != nil {
		return fmt.Errorf("engine heartbeat: 설정을 읽을 수 없어 판정하지 않음: %w", err)
	}
	// 전송 수단은 엔진 조립과 같은 규칙으로 얻음(규칙 사본 없음, a074 · a133 경계 그대로).
	publisher, why := engine.NotificationPublisher(cfg.Engine.Notifications, nil)
	if publisher == nil {
		return fmt.Errorf("engine heartbeat: 알림을 보낼 수 없어 엔진 정지를 알릴 수 없음 — %s: %w", why, enginewatch.ErrNoPublisher)
	}

	out, err := enginewatch.Check{
		Dir:        dir,
		StaleAfter: staleAfter,
		Expected:   cfg.Engine.Autostart,
		Publisher:  publisher,
	}.Run(ctx)
	if err != nil {
		return fmt.Errorf("engine heartbeat: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), describeHeartbeatOutcome(out))
	return nil
}

// describeHeartbeatOutcome 은 journalctl 한 줄로 읽히는 판정 요약을 만듦(계좌 정보 없음).
func describeHeartbeatOutcome(out enginewatch.Outcome) string {
	o := out.Observation
	line := "engine heartbeat  " + string(o.Verdict)
	if o.Reason != "" {
		line += " (" + o.Reason + ")"
	}
	if !o.LastBeat.IsZero() {
		line += fmt.Sprintf(" · 마지막 갱신 %s 전", o.Age.Round(time.Second))
	}
	switch {
	case out.Sent != "":
		line += " · 알림 보냄 " + string(out.Sent)
	case o.Verdict == enginewatch.VerdictDown:
		line += " · 이미 알림(중복 억제)"
	}
	return line
}
