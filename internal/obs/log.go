package obs

// log.go is the structured-logging half (harden-execution-base task 4.3).
//
// # Countable, not narrative
//
// The convention the spec asks for ("셀 수 있는 이벤트 규약") is that every line
// carries a stable `event` field from the EventType enum, and that the variable
// parts are *fields* rather than interpolated into the message. A log where the
// symbol is inside the sentence can be read; a log where it is a field can be
// counted, grouped and alerted on. The engine's questions are all counting
// questions — how many IN_DOUBT attempts today, which symbol is refusing to
// reconcile — so the format is chosen for those.
//
// Field names are fixed here as constants for the same reason the event names
// are: a dashboard keyed on `symbol` breaks if half the call sites write `sym`.

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// Field names. Stable contract; append, never rename.
const (
	FieldEvent      = "event"
	FieldSubject    = "subject"
	FieldSeverity   = "severity"
	FieldAccount    = "account"
	FieldSymbol     = "symbol"
	FieldMarket     = "market"
	FieldSide       = "side"
	FieldQuantity   = "quantity"
	FieldPrice      = "price"
	FieldIntentID   = "intent_id"
	FieldAttemptID  = "attempt_id"
	FieldOrderID    = "order_id"
	FieldFromState  = "from_state"
	FieldToState    = "to_state"
	FieldReason     = "reason"
	FieldDetail     = "detail"
	FieldError      = "error"
	FieldDurationMS = "duration_ms"
	FieldCount      = "count"
	FieldScope      = "scope"
	// FieldActor names who caused a change: AUTO or OPERATOR. Appended for the
	// operating-mode transition (add-core-domain task 3.3), where the whole
	// approval asymmetry hangs on which of the two it was — an automatic
	// tightening and an operator's relaxation are different events and must be
	// countable apart.
	FieldActor = "actor"
)

// Logger writes the engine's structured log.
//
// It wraps slog rather than replacing it: the standard library's handler already
// solves the encoding, and what this type adds is the convention — the event
// enum, the fixed field names, and an injected clock so a test can assert on the
// timestamp of a line.
type Logger struct {
	log *slog.Logger
	clk clock.Clock
}

// LogOptions configures NewLogger.
type LogOptions struct {
	// Writer receives the log. Defaults to os.Stderr.
	Writer io.Writer
	// Level is the minimum level. Zero value is slog.LevelInfo.
	Level slog.Level
	// JSON selects the JSON handler over the text one. The engine runs
	// unattended, so JSON is the production choice; text is for a human tailing
	// it during development.
	JSON bool
	// Clock stamps lines. Defaults to clock.System().
	Clock clock.Clock
	// Base are attributes attached to every line (account ref, engine version).
	Base []slog.Attr
}

// NewLogger builds a logger.
func NewLogger(opts LogOptions) *Logger {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}

	handlerOpts := &slog.HandlerOptions{
		Level: opts.Level,
		// The injected clock has to reach the emitted line, and slog stamps
		// records from time.Now() internally. Replacing the time attribute is
		// the supported seam for that.
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, clk.Now().UTC())
			}
			return a
		},
	}

	var handler slog.Handler
	if opts.JSON {
		handler = slog.NewJSONHandler(w, handlerOpts)
	} else {
		handler = slog.NewTextHandler(w, handlerOpts)
	}
	if len(opts.Base) > 0 {
		handler = handler.WithAttrs(opts.Base)
	}
	return &Logger{log: slog.New(handler), clk: clk}
}

// Slog exposes the underlying logger, for code that needs to pass one on.
func (l *Logger) Slog() *slog.Logger {
	if l == nil {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return l.log
}

// With returns a logger carrying extra attributes on every line.
func (l *Logger) With(args ...any) *Logger {
	if l == nil {
		return nil
	}
	return &Logger{log: l.log.With(args...), clk: l.clk}
}

// Event writes one countable line at info level.
func (l *Logger) Event(t EventType, args ...any) {
	l.emit(context.Background(), slog.LevelInfo, t, args...)
}

// Warn writes a countable line at warn level.
func (l *Logger) Warn(t EventType, args ...any) {
	l.emit(context.Background(), slog.LevelWarn, t, args...)
}

// Error writes a countable line at error level. err is attached as the `error`
// field rather than folded into the message, so failures can be grouped by event
// and counted.
func (l *Logger) Error(t EventType, err error, args ...any) {
	if err != nil {
		args = append(args, FieldError, err.Error())
	}
	l.emit(context.Background(), slog.LevelError, t, args...)
}

// OrderTransition is the order-lifecycle line the spec names explicitly.
func (l *Logger) OrderTransition(t EventType, intentID, attemptID, symbol, from, to string, args ...any) {
	l.Event(t, append([]any{
		FieldIntentID, intentID,
		FieldAttemptID, attemptID,
		FieldSymbol, symbol,
		FieldFromState, from,
		FieldToState, to,
	}, args...)...)
}

// Reconcile is the reconciliation line the spec names explicitly.
func (l *Logger) Reconcile(t EventType, matched, mismatched int, took time.Duration, args ...any) {
	l.Event(t, append([]any{
		FieldCount, matched + mismatched,
		"matched", matched,
		"mismatched", mismatched,
		FieldDurationMS, took.Milliseconds(),
	}, args...)...)
}

func (l *Logger) emit(ctx context.Context, level slog.Level, t EventType, args ...any) {
	if l == nil || l.log == nil {
		return
	}
	attrs := append([]any{
		FieldEvent, string(t),
		FieldSubject, t.Subject(),
		FieldSeverity, string(SeverityOf(t)),
	}, args...)
	l.log.Log(ctx, level, string(t), attrs...)
}
