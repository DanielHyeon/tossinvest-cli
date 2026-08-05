package engine

// instrumentnames.go names a stock the way its owner does (change a085).
//
// # Why the engine did not know the names
//
// The alert bodies said "the exit policy could not judge 032820". The console
// says 에이치엘비 for the same holding, because it reads the account through the
// float path (`domain.Position.Name`). The engine reads the same endpoint through
// the raw path, which existed to preserve the broker's cost-basis decimal and
// dropped everything it did not need — including the name that arrives in the
// same response.
//
// So this costs no request. `GET /api/v1/holdings` has carried `name` all along;
// a085 stopped throwing it away, and this is where the reconciliation loop leaves
// it for the alert composer to find.
//
// # Why nothing is forgotten
//
// Entries are never removed. The alerts that most need a name are the ones about
// a holding that has just stopped being one — "closed outside the engine",
// "could not judge" — and a registry that expired with the position would lose
// the name at exactly the moment it matters. The bound is the number of distinct
// symbols an account has held while this process ran.
//
// # Why nothing is guessed
//
// A symbol the registry has not seen renders as its code. The alternative is
// inferring a name from somewhere else, and a wrong name in an alert about a stop
// that is not being evaluated points an operator at the wrong holding.

import (
	"strings"
	"sync"
)

// InstrumentNames maps a symbol to the broker's display name for it.
//
// The zero value is usable and answers "" for everything, which renders as the
// bare code. An engine assembled without a reconciliation loop therefore alerts
// exactly as it did before a085 rather than failing to alert.
type InstrumentNames struct {
	mu    sync.RWMutex
	names map[string]string
}

// Learn records the names in one holdings snapshot.
//
// Empty names are ignored rather than stored: "the payload carried no name" must
// not overwrite a name an earlier read did carry.
func (n *InstrumentNames) Learn(symbol, name string) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	name = strings.TrimSpace(name)
	if symbol == "" || name == "" {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.names == nil {
		n.names = map[string]string{}
	}
	n.names[symbol] = name
}

// Name returns the broker's display name, or "" when none is known.
func (n *InstrumentNames) Name(symbol string) string {
	if n == nil {
		return ""
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.names[strings.ToUpper(strings.TrimSpace(symbol))]
}

// Label renders a symbol for a human: `이름(코드)`, or the bare code when the
// name is not known.
//
// The code is always present. An operator reading an alert on a phone has to be
// able to type what they see into the app, and a display name is not what the
// app searches on.
func (n *InstrumentNames) Label(symbol string) string {
	code := strings.ToUpper(strings.TrimSpace(symbol))
	name := n.Name(code)
	if name == "" || name == code {
		return code
	}
	return name + "(" + code + ")"
}
