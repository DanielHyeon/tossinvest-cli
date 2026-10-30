package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/JungHoonGhae/tossinvest-cli/internal/i18n"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
	"github.com/spf13/cobra"
)

func main() {
	i18n.SetLang(i18n.Resolve(os.Args[1:], os.Getenv))

	opts := &rootOptions{budgetOwner: newRootBudgetOwner(os.LookupEnv, os.Stderr)}
	root := newRootCmdWithOptions(opts)
	// a131: cobra 가 실패 시 루트의 stderr 에 쓰는 `Error:` 줄과 아래 마지막 오류 줄이 계좌 규칙을 지남(engine.log · soak.log · journald).
	// stdout 은 감싸지 않음 — 사람이 요청한 표시(계좌 목록 등)임.
	root.SetErr(redact.Writer(os.Stderr))
	if err := executeRoot(root, opts); err != nil {
		if err != errMonitorFailed {
			fmt.Fprintln(redact.Writer(os.Stderr), err)
		}
		os.Exit(1)
	}
}

func executeRoot(root *cobra.Command, opts *rootOptions) error {
	err := root.Execute()
	var closeErr error
	if opts != nil && opts.budgetOwner != nil {
		closeErr = opts.budgetOwner.Close()
	}
	if closeErr == nil {
		return err
	}
	if err == errMonitorFailed {
		return closeErr
	}
	return errors.Join(err, closeErr)
}
