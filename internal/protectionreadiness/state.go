package protectionreadiness

import (
	"crypto/sha256"
	"sort"
	"time"
)

type serialScope struct {
	AccountID string
	ProfileID string
	Market    Market
}

type durableState struct {
	TrustedTimeFloor time.Time
	Serials          map[serialScope]uint64
	seal             [32]byte
}

func newDurableState() durableState {
	return newDurableStateWith(time.Time{}, nil)
}

func newDurableStateWith(floor time.Time, serials map[serialScope]uint64) durableState {
	state := durableState{TrustedTimeFloor: floor.UTC(), Serials: make(map[serialScope]uint64, len(serials))}
	for scope, serial := range serials {
		state.Serials[scope] = serial
	}
	state.seal = durableStateSeal(state)
	return state
}

func cloneDurableState(state durableState) durableState {
	return newDurableStateWith(state.TrustedTimeFloor, state.Serials)
}

func validDurableState(state durableState) bool {
	if state.Serials == nil || state.seal != durableStateSeal(state) {
		return false
	}
	for scope, serial := range state.Serials {
		if scope.AccountID == "" || scope.ProfileID == "" || !validMarket(scope.Market) || serial == 0 {
			return false
		}
	}
	return true
}

func durableStateSeal(state durableState) [32]byte {
	hash := sha256.New()
	writeString(hash, formatTime(state.TrustedTimeFloor))
	scopes := make([]serialScope, 0, len(state.Serials))
	for scope := range state.Serials {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool {
		left, right := scopes[i], scopes[j]
		if left.AccountID != right.AccountID {
			return left.AccountID < right.AccountID
		}
		if left.ProfileID != right.ProfileID {
			return left.ProfileID < right.ProfileID
		}
		return left.Market < right.Market
	})
	for _, scope := range scopes {
		writeString(hash, scope.AccountID)
		writeString(hash, scope.ProfileID)
		writeString(hash, string(scope.Market))
		writeUint64(hash, state.Serials[scope])
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

type trustedTime struct {
	Now       time.Time
	Source    string
	available bool
	seal      [32]byte
}

func newTrustedTime(now time.Time, source string) trustedTime {
	if now.IsZero() || source == "" {
		return trustedTime{}
	}
	value := trustedTime{Now: now.UTC(), Source: source, available: true}
	value.seal = trustedTimeSeal(value)
	return value
}

func trustedTimeSeal(value trustedTime) [32]byte {
	available := "unavailable"
	if value.available {
		available = "available"
	}
	return hashStrings(formatTime(value.Now), value.Source, available)
}

func validTrustedTime(value trustedTime) bool {
	return value.available && !value.Now.IsZero() && value.Source != "" && value.seal == trustedTimeSeal(value)
}
