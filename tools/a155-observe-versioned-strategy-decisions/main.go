// Offline replay only. It has no data collector or order authority.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyobserve"
)

const maxInput = 64 << 20

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("a155-offline-replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input := fs.String("input", "", "local JSONL fixture")
	destination := fs.String("out", "", "new or owned output directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || *destination == "" || fs.NArg() != 0 {
		return errors.New("explicit -input and -out required")
	}
	f, err := os.Open(*input)
	if err != nil {
		return err
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxInput+1))
	_ = f.Close()
	if err != nil {
		return err
	}
	if len(raw) > maxInput {
		return errors.New("input exceeds 64 MiB")
	}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), 1<<20+1)
	records := make([][]byte, 0)
	for scanner.Scan() {
		if len(scanner.Bytes()) > 1<<20 {
			return errors.New("input line exceeds 1 MiB")
		}
		var v strategyobserve.Input
		d := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		d.DisallowUnknownFields()
		if err = d.Decode(&v); err != nil {
			return err
		}
		var extra any
		if err = d.Decode(&extra); err != io.EOF {
			return errors.New("trailing JSON")
		}
		diagnostic, err := strategyobserve.Calculate(v)
		if err != nil {
			return err
		}
		line, err := strategyobserve.Encode(diagnostic)
		if err != nil {
			return err
		}
		records = append(records, line)
		if len(records) > 65536 {
			return errors.New("input record count limit")
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	if len(records) == 0 {
		return errors.New("empty input")
	}
	s, err := strategyobserve.Open(*destination)
	if err != nil {
		return err
	}
	defer s.Close()
	q := strategyobserve.NewQueue()
	done := make(chan error, 1)
	go func() { done <- q.Run(s.Append) }()
	for _, line := range records {
		q.Enqueue(line)
	}
	q.CloseInput()
	err = <-done
	c := q.Counts()
	fmt.Fprintf(out, "offline=unverified output=%s accepted=%d durable=%d duplicate=%d full=%d invalid=%d oversize=%d failed=%d pending=%d\n", *destination, c.Accepted, c.Durable, c.Duplicate, c.Full, c.Invalid, c.Oversize, c.Failed, c.Pending)
	if err != nil {
		return err
	}
	if c.Full+c.Invalid+c.Oversize+c.Failed+c.Pending > 0 {
		return errors.New("replay incomplete")
	}
	return nil
}
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
