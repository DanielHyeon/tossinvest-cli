package protectionofficial

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
)

type fakeClient struct {
	created      official.ConditionalCreateBody
	modified     official.ConditionalModifyBody
	raw          map[string]official.RawConditionalOrder
	pages        map[string]official.RawConditionalOrderList
	cancel       string
	sellable     string
	readErr      error
	readErrAfter int
	readCalls    int
	createRef    *domain.ConditionalOrderRef
}

func (f *fakeClient) CreateConditionalOrder(_ context.Context, body official.ConditionalCreateBody) (domain.ConditionalOrderRef, error) {
	f.created = body
	if f.createRef != nil {
		return *f.createRef, nil
	}
	return domain.ConditionalOrderRef{ID: "co-1", ClientOrderID: body.ClientOrderID}, nil
}

func (f *fakeClient) ModifyConditionalOrderRef(_ context.Context, id string, body official.ConditionalModifyBody) (domain.ConditionalOrderRef, error) {
	f.modified = body
	return domain.ConditionalOrderRef{ID: id + "-new"}, nil
}

func (f *fakeClient) CancelConditionalOrder(_ context.Context, id string) error {
	f.cancel = id
	return nil
}

func (f *fakeClient) ConditionalOrderRaw(_ context.Context, id string) (official.RawConditionalOrder, error) {
	f.readCalls++
	if f.readErr != nil && (f.readErrAfter == 0 || f.readCalls > f.readErrAfter) {
		return official.RawConditionalOrder{}, f.readErr
	}
	raw, ok := f.raw[id]
	if !ok {
		return official.RawConditionalOrder{}, errors.New("not found")
	}
	return raw, nil
}

func (f *fakeClient) ProtectionConditionalOrdersRaw(_ context.Context, status, _, cursor string, _ int) (official.RawConditionalOrderList, error) {
	return f.pages[status+":"+cursor], nil
}

func (f *fakeClient) SellableQuantityRaw(context.Context, string) (string, time.Time, error) {
	return f.sellable, gatewayNow, nil
}

var gatewayNow = time.Date(2026, 8, 1, 2, 0, 0, 0, time.UTC)

func gatewayScope() protection.Scope {
	return protection.Scope{AccountRef: "acct-1", Profile: "primary", Market: protection.MarketKR, Symbol: "005930"}
}

func rawConditional(id, clientID, status string) official.RawConditionalOrder {
	return official.RawConditionalOrder{
		ID: id, ClientOrderID: clientID, Symbol: "005930", Market: "KR", Type: "SINGLE", Status: status,
		OrderType: "MARKET", OrderSide: "SELL", Quantity: "1", TriggerPrice: "70000", ConditionType: "STOP", ExpireDate: "2026-08-08",
	}
}

func conditionalBody() protection.ConditionalBody {
	return protection.ConditionalBody{
		SerializerVersion: protection.SerializerVersion, ClientOrderID: "client-1", AccountRef: "acct-1",
		Market: "KR", Symbol: "005930", Side: "SELL", ConditionalType: "SINGLE", OrderType: "MARKET",
		TriggerSource: "LAST_TRADE", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08",
	}
}

func TestGatewayCreateUsesOnlyCanonicalOfficialShapeAndConfirmsExactReadback(t *testing.T) {
	client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": rawConditional("co-1", "client-1", "WATCHING")}}
	gateway, err := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	if err != nil {
		t.Fatal(err)
	}
	got, err := gateway.Create(context.Background(), conditionalBody())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "co-1" || got.Quantity != 1 || got.Trigger != 70000 || got.Terminal {
		t.Fatalf("got=%+v", got)
	}
	if client.created.Type != "SINGLE" || client.created.OrderType != "MARKET" || client.created.First.OrderSide != "SELL" || client.created.Second != nil || client.created.ConfirmHighValueOrder {
		t.Fatalf("official body=%+v", client.created)
	}
}

func TestGatewayRejectsLossyOrMismatchedReadback(t *testing.T) {
	for name, raw := range map[string]official.RawConditionalOrder{
		"fractional": func() official.RawConditionalOrder {
			v := rawConditional("co-1", "client-1", "WATCHING")
			v.Quantity = "1.0"
			return v
		}(),
		"wrong-client":   rawConditional("co-1", "other", "WATCHING"),
		"missing-client": rawConditional("co-1", "", "WATCHING"),
		"wrong-type": func() official.RawConditionalOrder {
			v := rawConditional("co-1", "client-1", "WATCHING")
			v.Type = "OCO"
			return v
		}(),
		"wrong-id": rawConditional("other-id", "client-1", "WATCHING"),
		"wrong-side": func() official.RawConditionalOrder {
			v := rawConditional("co-1", "client-1", "WATCHING")
			v.OrderSide = "BUY"
			return v
		}(),
		"wrong-condition": func() official.RawConditionalOrder {
			v := rawConditional("co-1", "client-1", "WATCHING")
			v.ConditionType = "PROFIT_RATE"
			return v
		}(),
		"wrong-expiry": func() official.RawConditionalOrder {
			v := rawConditional("co-1", "client-1", "WATCHING")
			v.ExpireDate = "2026-08-09"
			return v
		}(),
		"unknown-status": rawConditional("co-1", "client-1", "UNKNOWN"),
	} {
		t.Run(name, func(t *testing.T) {
			client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": raw}}
			gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
			if _, err := gateway.Create(context.Background(), conditionalBody()); err == nil {
				t.Fatal("unsafe readback accepted")
			}
		})
	}
}

func TestGatewayCreateRejectsMismatchedMutationReceiptBeforeReadback(t *testing.T) {
	client := &fakeClient{createRef: &domain.ConditionalOrderRef{ID: "co-1", ClientOrderID: "other-client"}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	if _, err := gateway.Create(context.Background(), conditionalBody()); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("mismatched receipt=%v", err)
	}
}

func TestGatewayCancelPreflightFailureNeverDispatchesDelete(t *testing.T) {
	ambiguous := rawConditional("co-1", "client-1", "UNKNOWN")
	mismatch := rawConditional("co-1", "client-1", "WATCHING")
	mismatch.ExpireDate = "2026-08-09"
	for _, tc := range []struct {
		name string
		raw  *official.RawConditionalOrder
		err  error
	}{
		{name: "404", err: errors.New("404")},
		{name: "timeout", err: context.DeadlineExceeded},
		{name: "transport error", err: errors.New("transport")},
		{name: "ambiguous lifecycle", raw: &ambiguous},
		{name: "identity mismatch", raw: &mismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeClient{raw: map[string]official.RawConditionalOrder{}}
			if tc.raw != nil {
				client.raw["co-1"] = *tc.raw
			}
			client.readErr = tc.err
			gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
			if _, err := gateway.Cancel(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"}); !errors.Is(err, ErrAmbiguousConditional) {
				t.Fatalf("cancel preflight=%v", err)
			}
			if client.cancel != "" {
				t.Fatalf("preflight failure dispatched cancel target=%q", client.cancel)
			}
		})
	}
}

func TestGatewayCancelPostDeleteDisappearanceRemainsInDoubt(t *testing.T) {
	active := rawConditional("co-1", "client-1", "WATCHING")
	client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": active}, readErr: errors.New("404"), readErrAfter: 1,
		pages: map[string]official.RawConditionalOrderList{}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	if _, err := gateway.Cancel(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"}); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("post-delete disappearance=%v", err)
	}
	if client.cancel != "co-1" {
		t.Fatalf("verified cancel target=%q", client.cancel)
	}
}

func TestGatewayCancelNeedsTerminalNonTriggeredObservationAndExactSellable(t *testing.T) {
	closed := rawConditional("co-1", "client-1", "CANCELLED")
	active := rawConditional("co-1", "client-1", "WATCHING")
	client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": active}, readErr: errors.New("gone"), readErrAfter: 1, sellable: "1", pages: map[string]official.RawConditionalOrderList{
		"OPEN:": {}, "CLOSED:": {Orders: []official.RawConditionalOrder{closed}},
	}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	cancel, err := gateway.Cancel(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"})
	if err != nil || !cancel.Terminal || cancel.Triggered {
		t.Fatalf("cancel=%+v err=%v", cancel, err)
	}
	sellable, err := gateway.Sellable(context.Background(), gatewayScope(), "co-1")
	if err != nil || sellable.Quantity != 1 || sellable.At != gatewayNow {
		t.Fatalf("sellable=%+v err=%v", sellable, err)
	}
	client.sellable = "1.0"
	if _, err := gateway.Sellable(context.Background(), gatewayScope(), "co-1"); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("lossy sellable=%v", err)
	}
}

func TestGatewayCancelRejectsWrongOrDuplicateTerminalIdentity(t *testing.T) {
	target := protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"}
	t.Run("wrong side", func(t *testing.T) {
		raw := rawConditional("co-1", "client-1", "CANCELLED")
		raw.OrderSide = "BUY"
		client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": raw}}
		gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
		if _, err := gateway.Cancel(context.Background(), target); !errors.Is(err, ErrAmbiguousConditional) {
			t.Fatalf("wrong-side cancel=%v", err)
		}
	})
	t.Run("duplicate client identity", func(t *testing.T) {
		one := rawConditional("co-1", "client-1", "CANCELLED")
		two := rawConditional("co-other", "client-1", "CANCELLED")
		active := rawConditional("co-1", "client-1", "WATCHING")
		client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": active}, readErr: errors.New("gone"), readErrAfter: 1, pages: map[string]official.RawConditionalOrderList{
			"OPEN:": {}, "CLOSED:": {Orders: []official.RawConditionalOrder{one, two}},
		}}
		gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
		if _, err := gateway.Cancel(context.Background(), target); !errors.Is(err, ErrAmbiguousConditional) {
			t.Fatalf("duplicate cancel=%v", err)
		}
	})
}

func TestGatewayTargetRequiresExactCurrentExpiry(t *testing.T) {
	raw := rawConditional("co-1", "client-1", "CANCELLED")
	client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": raw}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	base := protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1}
	if _, err := gateway.Get(context.Background(), base); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("empty expiry get=%v", err)
	}
	base.ExpireDate = "2026-08-09"
	if _, err := gateway.Get(context.Background(), base); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("mismatched expiry get=%v", err)
	}
	if _, err := gateway.Cancel(context.Background(), base); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("mismatched expiry cancel=%v", err)
	}
	if client.cancel != "" {
		t.Fatalf("mismatched expiry dispatched cancel for %q", client.cancel)
	}
}

func TestGatewayCancelFallbackIgnoresOnlyAttestedRetiredReplaceRows(t *testing.T) {
	target := protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-current", ClientOrderID: "client-1", Trigger: 72000, Quantity: 1, ExpireDate: "2026-08-08",
		Retired: []protection.RetiredBrokerTarget{
			{BrokerID: "co-old-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"},
			{BrokerID: "co-old-2", ClientOrderID: "client-1", Trigger: 71000, Quantity: 1, ExpireDate: "2026-08-08"},
		}}
	current := rawConditional("co-current", "client-1", "CANCELLED")
	current.TriggerPrice = "72000"
	activeCurrent := rawConditional("co-current", "client-1", "WATCHING")
	activeCurrent.TriggerPrice = "72000"
	old1 := rawConditional("co-old-1", "client-1", "COMPLETED")
	old2 := rawConditional("co-old-2", "client-1", "EXPIRED")
	old2.TriggerPrice = "71000"
	client := &fakeClient{raw: map[string]official.RawConditionalOrder{"co-current": activeCurrent}, readErr: errors.New("gone"), readErrAfter: 1, pages: map[string]official.RawConditionalOrderList{
		"OPEN:": {}, "CLOSED:": {Orders: []official.RawConditionalOrder{old1, current, old2}},
	}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	got, err := gateway.Cancel(context.Background(), target)
	if err != nil || got.BrokerID != "co-current" || !got.Terminal || got.Triggered {
		t.Fatalf("cancel=%+v err=%v", got, err)
	}
	client.readCalls = 0
	got, err = gateway.Cancel(context.Background(), target)
	if err != nil || got.BrokerID != "co-current" || !got.Terminal || got.Triggered {
		t.Fatalf("repeated cancel=%+v err=%v", got, err)
	}
	client.readCalls = 0
	client.pages["CLOSED:"].Orders[0].TriggerPrice = "69999"
	if _, err := gateway.Cancel(context.Background(), target); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("mismatched retired cancel=%v", err)
	}
}

func TestGatewayCancelFallbackRequiresCompleteOpenAndClosedScans(t *testing.T) {
	target := protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"}
	active := rawConditional("co-1", "client-1", "WATCHING")
	closed := rawConditional("co-1", "client-1", "CANCELLED")
	clientFor := func(pages map[string]official.RawConditionalOrderList) *fakeClient {
		return &fakeClient{raw: map[string]official.RawConditionalOrder{"co-1": active}, readErr: errors.New("post-delete detail unavailable"), readErrAfter: 1, pages: pages}
	}
	assertAmbiguous := func(t *testing.T, pages map[string]official.RawConditionalOrderList) {
		t.Helper()
		client := clientFor(pages)
		gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
		if _, err := gateway.Cancel(context.Background(), target); !errors.Is(err, ErrAmbiguousConditional) {
			t.Fatalf("cancel fallback err=%v", err)
		}
		if client.cancel != "co-1" {
			t.Fatalf("preflight-verified delete target=%q", client.cancel)
		}
	}

	t.Run("target early but duplicate could exist on page eleven", func(t *testing.T) {
		pages := map[string]official.RawConditionalOrderList{"CLOSED:": {}}
		cursor := ""
		for page := 0; page < 10; page++ {
			next := fmt.Sprintf("cursor-%d", page+1)
			orders := []official.RawConditionalOrder(nil)
			if page == 0 {
				orders = []official.RawConditionalOrder{closed}
			}
			pages["OPEN:"+cursor] = official.RawConditionalOrderList{Orders: orders, HasNext: true, NextCursor: next}
			cursor = next
		}
		pages["OPEN:"+cursor] = official.RawConditionalOrderList{Orders: []official.RawConditionalOrder{closed}}
		assertAmbiguous(t, pages)
	})

	t.Run("has next with empty cursor", func(t *testing.T) {
		assertAmbiguous(t, map[string]official.RawConditionalOrderList{
			"OPEN:":   {},
			"CLOSED:": {Orders: []official.RawConditionalOrder{closed}, HasNext: true},
		})
	})

	t.Run("repeated cursor", func(t *testing.T) {
		assertAmbiguous(t, map[string]official.RawConditionalOrderList{
			"OPEN:":     {Orders: []official.RawConditionalOrder{closed}, HasNext: true, NextCursor: "same"},
			"OPEN:same": {HasNext: true, NextCursor: "same"},
			"CLOSED:":   {},
		})
	})

	t.Run("cursor cycle", func(t *testing.T) {
		assertAmbiguous(t, map[string]official.RawConditionalOrderList{
			"OPEN:":   {Orders: []official.RawConditionalOrder{closed}, HasNext: true, NextCursor: "a"},
			"OPEN:a":  {HasNext: true, NextCursor: "b"},
			"OPEN:b":  {HasNext: true, NextCursor: "a"},
			"CLOSED:": {},
		})
	})

	t.Run("open complete closed incomplete", func(t *testing.T) {
		assertAmbiguous(t, map[string]official.RawConditionalOrderList{
			"OPEN:":              {},
			"CLOSED:":            {Orders: []official.RawConditionalOrder{closed}, HasNext: true, NextCursor: "missing-end"},
			"CLOSED:missing-end": {HasNext: true, NextCursor: "missing-end"},
		})
	})

	t.Run("open incomplete closed complete", func(t *testing.T) {
		assertAmbiguous(t, map[string]official.RawConditionalOrderList{
			"OPEN:":   {HasNext: true},
			"CLOSED:": {Orders: []official.RawConditionalOrder{closed}},
		})
	})

	t.Run("both scans complete", func(t *testing.T) {
		client := clientFor(map[string]official.RawConditionalOrderList{
			"OPEN:":   {},
			"CLOSED:": {Orders: []official.RawConditionalOrder{closed}},
		})
		gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
		got, err := gateway.Cancel(context.Background(), target)
		if err != nil || got.BrokerID != "co-1" || !got.Terminal || got.Triggered {
			t.Fatalf("complete cancel=%+v err=%v", got, err)
		}
	})
}

func TestGatewayListIsBoundedAndRejectsMixedScope(t *testing.T) {
	client := &fakeClient{pages: map[string]official.RawConditionalOrderList{
		"OPEN:": {Orders: []official.RawConditionalOrder{rawConditional("co-1", "client-1", "WATCHING")}},
		"CLOSED:": {Orders: []official.RawConditionalOrder{func() official.RawConditionalOrder {
			closed := rawConditional("co-2", "client-2", "COMPLETED")
			closed.TriggeredOrderID = "plain-sell-1"
			return closed
		}()}},
	}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	got, err := gateway.List(context.Background(), gatewayScope())
	if err != nil || len(got) != 2 || !got[1].Terminal || !got[1].Triggered {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	other := gatewayScope()
	other.Symbol = "000660"
	if _, err := gateway.List(context.Background(), other); !errors.Is(err, protection.ErrMixedScope) {
		t.Fatalf("mixed scope=%v", err)
	}
}

func TestGatewayTriggeredOrderIsTerminalForReconciliation(t *testing.T) {
	raw := rawConditional("co-triggered", "client-triggered", "ORDERED")
	raw.TriggeredOrderID = "plain-sell-1"
	client := &fakeClient{raw: map[string]official.RawConditionalOrder{raw.ID: raw}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	got, err := gateway.Get(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: raw.ID, ClientOrderID: raw.ClientOrderID, Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"})
	if err != nil || !got.Triggered || !got.Terminal {
		t.Fatalf("triggered=%+v err=%v", got, err)
	}
}
