package strategyobserve

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
	"strconv"
	"sync"
	"testing"
	"time"
)

func qty(n uint64) *uint64 { return &n }

func TestDiagnosticIndependentExactRiskAndUnsupportedUS(t *testing.T) {
	in := Input{Parent: "sample", PolicyVersion: "research-v1", Features: strategymarket.SessionFeatureInput{Market: "US", Symbol: "XYZ", Currency: "USD", AsOf: time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)}, Geometry: &Geometry{Currency: "USD", Basis: "unadjusted", Entry: "100", Stop: "99.3", Target: "102.1", Quantity: qty(100)}}
	d, err := Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	if d.Risk != "70" || d.Features.Status != strategymarket.FeatureUnsupported || d.RRStatus != "unsupported" {
		t.Fatalf("unexpected diagnostic: %+v", d)
	}
}

func TestQueueRejectsOversizeAndClose(t *testing.T) {
	q := NewQueue()
	if got := q.Enqueue(make([]byte, 65537)); got != Oversize {
		t.Fatalf("oversize=%s", got)
	}
	q.CloseInput()
	if got := q.Enqueue([]byte("{}\n")); got != Closed {
		t.Fatalf("closed=%s", got)
	}
}

func testLine(t *testing.T, parent string) []byte {
	t.Helper()
	d, err := Calculate(Input{Parent: parent, PolicyVersion: "v1", Features: strategymarket.SessionFeatureInput{Market: "US", Symbol: "XYZ", Currency: "USD", AsOf: time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encode(d)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func TestQueueFullAndImmutable(t *testing.T) {
	q := NewQueue()
	b := testLine(t, "one")
	saved := append([]byte(nil), b...)
	for i := 0; i < 256; i++ {
		if q.Enqueue(b) != Accepted {
			t.Fatal(i)
		}
	}
	if q.Enqueue(b) != Full {
		t.Fatal("queue did not fill")
	}
	b[0] = 'x'
	q.CloseInput()
	if err := q.Run(func(got []byte) (bool, error) {
		if !bytes.Equal(got, saved) {
			t.Fatal("input alias")
		}
		return false, nil
	}); err != nil {
		t.Fatal(err)
	}
	c := q.Counts()
	if c.Durable != 256 || c.Full != 1 || c.Pending != 0 {
		t.Fatalf("counts %+v", c)
	}
}
func TestQueueBlockedWriterDoesNotBlockProducerOrExit(t *testing.T) {
	q := NewQueue()
	start := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	var once sync.Once
	go func() {
		done <- q.Run(func([]byte) (bool, error) { once.Do(func() { close(start) }); <-release; return false, nil })
	}()
	q.Enqueue(testLine(t, "first"))
	<-start
	exit := make(chan struct{})
	go func() { q.Enqueue(testLine(t, "second")); close(exit) }()
	select {
	case <-exit:
	case <-time.After(time.Second):
		t.Fatal("producer blocked by disk")
	}
	q.CloseInput()
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
func TestInputImmutableAndUTCIdentity(t *testing.T) {
	in := Input{Parent: "x", PolicyVersion: "v1", Features: strategymarket.SessionFeatureInput{Market: "US", Symbol: "XYZ", Currency: "USD", AsOf: time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)}}
	a, err := Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	in.Features.AsOf = in.Features.AsOf.In(time.FixedZone("other", 9*3600))
	b, err := Calculate(in)
	if err != nil || a.ID != b.ID {
		t.Fatalf("UTC identity %v", err)
	}
	in.PolicyVersion = "v2"
	c, _ := Calculate(in)
	if c.ID == b.ID {
		t.Fatal("version omitted from ID")
	}
}

func TestGeometryBoundaries(t *testing.T) {
	base := Input{Parent: "p", PolicyVersion: "v", Features: strategymarket.SessionFeatureInput{Market: "US", Symbol: "X", Currency: "USD"}}
	d, err := Calculate(base)
	if err != nil || d.RRStatus != "unsupported" {
		t.Fatalf("US without geometry: %+v %v", d, err)
	}
	base.Geometry = &Geometry{Currency: "USD", Basis: "unadjusted", Entry: "100", Stop: "99", Quantity: qty(1)}
	for _, mutate := range []func(*Geometry){func(g *Geometry) { g.Currency = "KRW" }, func(g *Geometry) { g.Basis = "adjusted" }, func(g *Geometry) { g.Entry = "1p1000000" }, func(g *Geometry) { g.Quantity = qty(1000001) }, func(g *Geometry) { g.LVNSpacePct = string(bytes.Repeat([]byte("1"), 129)) }} {
		v := *base.Geometry
		mutate(&v)
		base.Geometry = &v
		if _, err := Calculate(base); err == nil {
			t.Fatalf("invalid geometry accepted: %+v", v)
		}
		base.Geometry = &Geometry{Currency: "USD", Basis: "unadjusted", Entry: "100", Stop: "99", Quantity: qty(1)}
	}
}

func TestQueueExactRecordBoundaryAndConcurrentClose(t *testing.T) {
	line := testLine(t, "boundary")
	exact := append([]byte(nil), line[:len(line)-1]...)
	exact = append(exact, bytes.Repeat([]byte(" "), MaxRecord-len(exact)-1)...)
	exact = append(exact, '\n')
	q := NewQueue()
	if got := q.Enqueue(exact); got != Accepted {
		t.Fatalf("exact limit: %s", got)
	}
	if got := q.Enqueue(append(exact, ' ')); got != Oversize {
		t.Fatalf("over limit: %s", got)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 50; n++ {
				q.Enqueue(line)
			}
		}()
	}
	q.CloseInput()
	wg.Wait()
	if err := q.Run(func([]byte) (bool, error) { return false, nil }); err != nil {
		t.Fatal(err)
	}
}
func TestQueueWriterFailureKeepsPending(t *testing.T) {
	q := NewQueue()
	line := testLine(t, "x")
	q.Enqueue(line)
	q.Enqueue(line)
	q.CloseInput()
	if err := q.Run(func([]byte) (bool, error) { return false, errors.New("disk") }); err == nil {
		t.Fatal("failure hidden")
	}
	c := q.Counts()
	if c.Failed != 1 || c.Pending != 2 || c.Durable != 0 {
		t.Fatalf("counts %+v", c)
	}
}

func TestKRFeatureAndParkerOracle(t *testing.T) {
	open := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)
	f := strategymarket.SessionFeatureInput{Market: "KR", Symbol: "005930", Currency: "KRW", CalendarVersion: "v1", CalendarSource: "official-open-api", CalendarVerified: true, TradingDate: "2026-05-08", OpenAt: open, CloseAt: open.Add(390 * time.Minute), AsOf: open.Add(50 * time.Minute)}
	for i := 1; i <= 50; i++ {
		p := strconv.Itoa(101 + (i-1)/5)
		at := open.Add(time.Duration(i) * time.Minute)
		b := strategymarket.SessionFeatureBar{Market: "KR", Symbol: f.Symbol, Currency: "KRW", Source: "official-open-api", Basis: "unadjusted", CloseAt: at, ReceivedAt: at, Open: p, High: p, Low: p, Close: p, Volume: "1"}
		f.Minutes = append(f.Minutes, b)
		if i%5 == 0 {
			b.Volume = "5"
			f.Fives = append(f.Fives, b)
		}
	}
	in := Input{Parent: "p", PolicyVersion: "v", Features: f, Geometry: &Geometry{Currency: "KRW", Basis: "unadjusted", Entry: "100", Stop: "99.3", Target: "102.1", LVNSpacePct: "1.2", Quantity: qty(100), LaneID: strategyengine.LaneID, LaneVersion: strategyengine.LaneVersion}}
	before, _ := json.Marshal(in)
	d, err := Calculate(in)
	after, _ := json.Marshal(in)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("input changed or error: %v", err)
	}
	if d.Features.Status != strategymarket.FeatureKnown || d.Features.VWAP.Value != "211/2" || d.Features.EMA9.Value != "106" || d.Risk != "70" || d.RR.Space.Ratio != "12/7" || d.RR.Target.Ratio != "3" {
		t.Fatalf("unexpected %+v", d)
	}
}

func TestReviewGeometryMissingQuantityAndCurrency(t *testing.T) {
	var in Input
	if err := json.Unmarshal([]byte(`{"parent":"p","policy_version":"v","features":{"Market":"US","Symbol":"X","Currency":"USD"},"geometry":{"currency":"USD","basis":"unadjusted","entry":"100","stop":"99"}}`), &in); err != nil {
		t.Fatal(err)
	}
	d, err := Calculate(in)
	if err != nil || d.RiskState != "unknown" {
		t.Fatalf("optional quantity: %+v %v", d, err)
	}
	in.Features.Currency = "KRW"
	if _, err = Calculate(in); err == nil {
		t.Fatal("mixed currency accepted")
	}
}
func TestReviewPhysicalLineBoundary(t *testing.T) {
	line := testLine(t, "line")
	if _, err := ValidateLine(append(line, '\n')); err == nil {
		t.Fatal("second physical newline accepted")
	}
}
func TestReviewMissingBarBasisRejected(t *testing.T) {
	in := Input{Parent: "p", PolicyVersion: "v", Features: strategymarket.SessionFeatureInput{Market: "US", Currency: "USD", Minutes: []strategymarket.SessionFeatureBar{{}}}, Geometry: &Geometry{Currency: "USD", Basis: "unadjusted", Entry: "100", Stop: "99", Quantity: qty(1)}}
	if _, err := Calculate(in); err == nil {
		t.Fatal("missing bar basis accepted")
	}
}

func TestReviewQuantityNullZeroAndCopy(t *testing.T) {
	var g Geometry
	if err := json.Unmarshal([]byte(`{"quantity":null}`), &g); err == nil {
		t.Fatal("null accepted")
	}
	in := Input{Parent: "p", PolicyVersion: "v", Features: strategymarket.SessionFeatureInput{Market: "US", Currency: "USD"}, Geometry: &Geometry{Currency: "USD", Basis: "unadjusted", Entry: "100", Stop: "99", Quantity: qty(0)}}
	if _, err := Calculate(in); err == nil {
		t.Fatal("zero accepted")
	}
	*in.Geometry.Quantity = 5
	d, err := Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	*in.Geometry.Quantity = 7
	if *d.Geometry.Quantity != 5 || d.Risk != "5" {
		t.Fatal("result shares mutable input")
	}
}
