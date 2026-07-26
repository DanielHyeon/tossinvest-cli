package execgw

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// wirebody.go renders the exact request body a place will send, so the journal
// can store it at RECORDED (design D2, task 1.3/1.4).
//
// # Why the bytes and not the fields
//
// A replay resends what was sent. If it rebuilt the body from structured fields
// instead, a change in the serialisation rules — a field order, a decimal
// spelling, a newly omitted empty value — would send *different bytes under the
// same idempotency key*, and the broker's answer to that is
// "422 idempotency-key-conflict" (openapi), which says nothing about whether the
// original order exists. Storing the bytes removes that failure mode instead of
// documenting it.
//
// # Why this file exists at all
//
// The bytes are built by internal/official, whose builder is unexported and
// outside this change's edit scope. Rather than widen that package's API, the
// rules are transcribed here and pinned by a drift guard: a test drives a place
// through the real official client against httptest and asserts the captured
// request body is byte-identical to what this file produced and the journal
// stored. If official's builder changes, that test fails — which is the point.

// WireSerializerVersion names the rules PlaceWireBody applies. It is stored
// beside every body so a later build can tell whether it still understands the
// bytes it is about to resend.
const WireSerializerVersion = "official/order-create/v1"

// placeWireV0 is the quantity-based OrderCreateRequest variant; placeWireV1 is
// the amount-based one. Field order matters: the stored text is compared with
// what the client actually sent, byte for byte.
type placeWireV0 struct {
	Symbol                string `json:"symbol"`
	Side                  string `json:"side"`
	OrderType             string `json:"orderType"`
	Quantity              string `json:"quantity"`
	Price                 string `json:"price,omitempty"`
	TimeInForce           string `json:"timeInForce,omitempty"`
	ClientOrderID         string `json:"clientOrderId,omitempty"`
	ConfirmHighValueOrder bool   `json:"confirmHighValueOrder"`
}

type placeWireV1 struct {
	Symbol                string `json:"symbol"`
	Side                  string `json:"side"`
	OrderType             string `json:"orderType"`
	OrderAmount           string `json:"orderAmount"`
	ClientOrderID         string `json:"clientOrderId,omitempty"`
	ConfirmHighValueOrder bool   `json:"confirmHighValueOrder"`
}

// PlaceWireBody renders the JSON body a place with this intent will send.
func PlaceWireBody(intent orderintent.PlaceIntent) (string, error) {
	side := strings.ToUpper(intent.Side)

	if intent.Fractional && strings.EqualFold(intent.Side, "buy") {
		if intent.Amount <= 0 {
			return "", fmt.Errorf("a fractional buy needs a positive amount, got %v", intent.Amount)
		}
		return marshalWire(placeWireV1{
			Symbol:        intent.Symbol,
			Side:          side,
			OrderType:     "MARKET",
			OrderAmount:   wireDecimal(intent.Amount),
			ClientOrderID: intent.ClientOrderID,
		})
	}

	body := placeWireV0{Symbol: intent.Symbol, Side: side, ClientOrderID: intent.ClientOrderID}
	if intent.Fractional && strings.EqualFold(intent.Side, "sell") {
		if intent.Quantity <= 0 {
			return "", fmt.Errorf("a fractional sell needs a positive quantity, got %v", intent.Quantity)
		}
		body.OrderType = "MARKET"
		body.Quantity = wireDecimal(intent.Quantity)
		return marshalWire(body)
	}
	if intent.Quantity <= 0 {
		return "", fmt.Errorf("an order needs a positive quantity, got %v", intent.Quantity)
	}
	body.OrderType = strings.ToUpper(intent.OrderType)
	body.Quantity = wireDecimal(intent.Quantity)
	if body.OrderType == "LIMIT" {
		if intent.Price <= 0 {
			return "", fmt.Errorf("a limit order needs a positive price, got %v", intent.Price)
		}
		body.Price = wireDecimal(intent.Price)
		body.TimeInForce = "DAY"
	}
	return marshalWire(body)
}

func marshalWire(v any) (string, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func wireDecimal(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
