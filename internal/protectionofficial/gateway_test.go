package protectionofficial

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
)

type fakeClient struct {
	created   official.ConditionalCreateBody
	modified  official.ConditionalModifyBody
	raw       map[string]official.RawConditionalOrder
	pages     map[string]official.RawConditionalOrderList
	cancel    string
	sellable  string
	readErr   error
	createRef *domain.ConditionalOrderRef
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
	if f.readErr != nil {
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

func TestGatewayCancelDisappearanceIsInDoubtInsteadOfAssumedCancelled(t *testing.T) {
	client := &fakeClient{readErr: errors.New("404"), pages: map[string]official.RawConditionalOrderList{}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	if _, err := gateway.Cancel(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1}); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("cancel disappearance=%v", err)
	}
	if client.cancel != "co-1" {
		t.Fatalf("cancel target=%q", client.cancel)
	}
}

func TestGatewayCancelNeedsTerminalNonTriggeredObservationAndExactSellable(t *testing.T) {
	closed := rawConditional("co-1", "client-1", "CANCELLED")
	client := &fakeClient{readErr: errors.New("gone"), sellable: "1", pages: map[string]official.RawConditionalOrderList{
		"OPEN:": {}, "CLOSED:": {Orders: []official.RawConditionalOrder{closed}},
	}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	cancel, err := gateway.Cancel(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1})
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
	target := protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1}
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
		client := &fakeClient{readErr: errors.New("gone"), pages: map[string]official.RawConditionalOrderList{
			"OPEN:": {}, "CLOSED:": {Orders: []official.RawConditionalOrder{one, two}},
		}}
		gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
		if _, err := gateway.Cancel(context.Background(), target); !errors.Is(err, ErrAmbiguousConditional) {
			t.Fatalf("duplicate cancel=%v", err)
		}
	})
}

func TestGatewayCancelFallbackIgnoresOnlyAttestedRetiredReplaceRows(t *testing.T) {
	target := protection.BrokerTarget{Scope: gatewayScope(), BrokerID: "co-current", ClientOrderID: "client-1", Trigger: 72000, Quantity: 1,
		Retired: []protection.RetiredBrokerTarget{
			{BrokerID: "co-old-1", ClientOrderID: "client-1", Trigger: 70000, Quantity: 1, ExpireDate: "2026-08-08"},
			{BrokerID: "co-old-2", ClientOrderID: "client-1", Trigger: 71000, Quantity: 1, ExpireDate: "2026-08-08"},
		}}
	current := rawConditional("co-current", "client-1", "CANCELLED")
	current.TriggerPrice = "72000"
	old1 := rawConditional("co-old-1", "client-1", "COMPLETED")
	old2 := rawConditional("co-old-2", "client-1", "EXPIRED")
	old2.TriggerPrice = "71000"
	client := &fakeClient{readErr: errors.New("gone"), pages: map[string]official.RawConditionalOrderList{
		"OPEN:": {}, "CLOSED:": {Orders: []official.RawConditionalOrder{old1, current, old2}},
	}}
	gateway, _ := New(client, gatewayScope(), func() time.Time { return gatewayNow })
	got, err := gateway.Cancel(context.Background(), target)
	if err != nil || got.BrokerID != "co-current" || !got.Terminal || got.Triggered {
		t.Fatalf("cancel=%+v err=%v", got, err)
	}
	got, err = gateway.Cancel(context.Background(), target)
	if err != nil || got.BrokerID != "co-current" || !got.Terminal || got.Triggered {
		t.Fatalf("repeated cancel=%+v err=%v", got, err)
	}
	client.pages["CLOSED:"].Orders[0].TriggerPrice = "69999"
	if _, err := gateway.Cancel(context.Background(), target); !errors.Is(err, ErrAmbiguousConditional) {
		t.Fatalf("mismatched retired cancel=%v", err)
	}
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
	got, err := gateway.Get(context.Background(), protection.BrokerTarget{Scope: gatewayScope(), BrokerID: raw.ID, ClientOrderID: raw.ClientOrderID, Trigger: 70000, Quantity: 1})
	if err != nil || !got.Triggered || !got.Terminal {
		t.Fatalf("triggered=%+v err=%v", got, err)
	}
}
