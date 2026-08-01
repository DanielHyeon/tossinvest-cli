package engine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
)

func TestClosedMarketDoesNotSuppressSafetySupervisorLoops(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.FixedZone("KST", 9*60*60))
	nextOpen := time.Date(2026, 8, 3, 9, 0, 0, 0, time.FixedZone("KST", 9*60*60))
	calendar, err := scheduler.AdaptOfficialCalendar(marketclock.MarketKR, official.MarketCalendarResponse{
		PreviousBusinessDay: official.MarketCalendarDay{Date: "2026-07-31", Integrated: &official.MarketCalendarSessions{
			RegularMarket: &official.MarketCalendarSession{
				StartTime: time.Date(2026, 7, 31, 9, 0, 0, 0, time.FixedZone("KST", 9*60*60)),
				EndTime:   time.Date(2026, 7, 31, 15, 30, 0, 0, time.FixedZone("KST", 9*60*60)),
			},
		}},
		Today: official.MarketCalendarDay{Date: "2026-08-01"},
		NextBusinessDay: official.MarketCalendarDay{Date: "2026-08-03", Integrated: &official.MarketCalendarSessions{
			RegularMarket: &official.MarketCalendarSession{StartTime: nextOpen, EndTime: nextOpen.Add(6*time.Hour + 30*time.Minute)},
		}},
	}, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if calendar.Today.Regular != nil || calendar.NextBusinessDay.Regular == nil {
		t.Fatalf("official holiday calendar = %+v", calendar)
	}

	started := make(chan string, 3)
	var stopped sync.WaitGroup
	stopped.Add(3)
	loop := func(name string) engine.SupervisedLoop {
		return engine.SupervisedLoop{Name: name, Run: func(ctx context.Context) error {
			started <- name
			defer stopped.Done()
			<-ctx.Done()
			return ctx.Err()
		}}
	}
	runtime, err := engine.NewRuntime(engine.RuntimeOptions{Loops: []engine.SupervisedLoop{
		loop("reconcile"), loop("emergency-exit"), loop("fill-detection"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runtime.Run(ctx) }()
	seen := map[string]bool{}
	for len(seen) < 3 {
		select {
		case name := <-started:
			seen[name] = true
		case <-time.After(time.Second):
			t.Fatalf("safety loops started = %v", seen)
		}
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("graceful runtime stop: %v", err)
	}
	stopped.Wait()
}
