package journal

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
)

// idempotency.go derives the broker's idempotency key from the decision that
// authorised the mutation (design D2, task 1.2).
//
// # Why the key comes from the decision and not from the intent
//
// The intent id is minted inside the Gateway, after the issuer has already
// decided and persisted. A key derived from it could not be written on the
// decision row, and a replay — whose only input is an attempt id — would have no
// way to prove the key it is about to resend is the one the decision authorised.
// The two values the issuer owns at issue time are the decision id and the
// generation, so the key is exactly f(decision_id, generation).
//
// # Why a hash and not the id itself
//
// The broker constrains the key: "최대 36자, 영숫자 및 `-`, `_` 허용", pattern
// ^[a-zA-Z0-9\-_]+$ (OrderCreateRequest.clientOrderId,
// docs/migration/openapi.latest.json). Decision ids are ours to choose and may
// not survive that constraint; a hash always does, at a fixed width, and keeps
// the key opaque to anyone reading the wire.

const (
	// clientOrderIDPrefix marks a key as this build's. It is inside the
	// documented character class and costs four of the thirty-six characters.
	clientOrderIDPrefix = "tos-"

	// clientOrderIDBytes is how much of the digest the key carries: 24 bytes →
	// 32 base64url characters → 36 with the prefix, exactly the broker's
	// maximum. 192 bits is far past the point where a collision between two
	// decisions is a failure mode worth designing against.
	clientOrderIDBytes = 24

	// MaxClientOrderIDLen is the broker's documented maximum (openapi).
	MaxClientOrderIDLen = 36

	// clientOrderIDDomain separates this hash from every other use of SHA-256 in
	// the journal, and carries a version so a future scheme cannot be mistaken
	// for this one.
	clientOrderIDDomain = "tossos/idempotency-key/v1"
)

// DeriveClientOrderID returns the broker idempotency key for a decision.
//
// It is deterministic and total: the same pair always yields the same key, and
// every key it can return satisfies the broker's pattern and length. generation
// is the reissue counter and is always 0 in this change (design D1); it is part
// of the preimage anyway, because a reissued decision must address a *new* broker
// order rather than collect the cached result of the previous one ("동일 값으로
// 재요청 시 이전 주문 결과를 그대로 재반환합니다" — openapi).
func DeriveClientOrderID(decisionID string, generation int) string {
	h := sha256.New()
	// Length-prefixed rather than delimiter-joined: a decision id is caller
	// data, and "a|12" must not be constructible from ("a1", 2).
	h.Write([]byte(clientOrderIDDomain))
	h.Write([]byte{0})
	h.Write([]byte(strconv.Itoa(len(decisionID))))
	h.Write([]byte{0})
	h.Write([]byte(decisionID))
	h.Write([]byte{0})
	h.Write([]byte(strconv.Itoa(generation)))
	sum := h.Sum(nil)
	return clientOrderIDPrefix + base64.RawURLEncoding.EncodeToString(sum[:clientOrderIDBytes])
}

// ValidClientOrderID reports whether a key is one the broker documents as
// acceptable: non-empty, at most 36 characters, and only [a-zA-Z0-9\-_].
//
// It is checked at every boundary the key crosses rather than assumed from the
// derivation, because a key read back from a row is not necessarily one this
// build derived.
func ValidClientOrderID(key string) bool {
	if key == "" || len(key) > MaxClientOrderIDLen {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
		default:
			return false
		}
	}
	return true
}
