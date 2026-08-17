package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	var cfg config
	flags := flag.NewFlagSet("a112-mb-us-source", flag.ExitOnError)
	flags.StringVar(&cfg.tokenCache, "token-cache", "", "absolute cached-token path (required)")
	flags.StringVar(&cfg.goBinary, "go-binary", "", "absolute no-follow Go executable used for identity and rebuild (required)")
	flags.StringVar(&cfg.receiptRoot, "receipt-root", "", "existing owner-only 0700 directory below /tmp (required)")
	flags.StringVar(&cfg.sessionDate, "session-date", "", "exact US session date YYYY-MM-DD (required)")
	before := flags.String("before", "", "exact initial decoded UTF-8 cursor (required; may be empty)")
	_ = flags.Parse(os.Args[1:])
	provided := map[string]bool{}
	flags.Visit(func(item *flag.Flag) { provided[item.Name] = true })
	if !provided["before"] {
		fmt.Fprintln(os.Stderr, "a112 m-b1 measurement HOLD: --before must be explicitly supplied")
		os.Exit(2)
	}
	cfg.before = []byte(*before)
	ctx, stop := signal.NotifyContext(context.Background(), measurementSignals()...)
	defer stop()
	if err := run(ctx, cfg, newProductionDependencies(cfg)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println("A112 M-B1 receipt sealed; independent review and Manager PASS are still required")
}
