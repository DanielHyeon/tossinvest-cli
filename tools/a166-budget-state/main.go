// a166-budget-state only prepares or checks a local budget manifest.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
)

func run(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: a166-budget-state provision|check")
	}
	switch args[0] {
	case "provision":
		fs := flag.NewFlagSet("provision", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		root := fs.String("root", "", "")
		manifest := fs.String("manifest", "", "")
		scope := fs.String("scope", "", "")
		instance := fs.String("instance", "", "")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 || *root == "" || *manifest == "" || *scope == "" || *instance == "" {
			return errors.New("invalid provision arguments")
		}
		if err := accountbudgetruntime.ProvisionOffline(ctx, *root, *manifest, *scope, *instance); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "provisioned")
		return err
	case "check":
		fs := flag.NewFlagSet("check", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		manifest := fs.String("manifest", "", "")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 || *manifest == "" {
			return errors.New("invalid check arguments")
		}
		if err := accountbudgetruntime.CheckOffline(ctx, *manifest); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "verified")
		return err
	default:
		return errors.New("usage: a166-budget-state provision|check")
	}
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
