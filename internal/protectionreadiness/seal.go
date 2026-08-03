package protectionreadiness

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"time"
)

type digestWriter interface{ Write([]byte) (int, error) }

func writeString(writer digestWriter, value string) {
	writeUint64(writer, uint64(len(value)))
	_, _ = writer.Write([]byte(value))
}

func writeUint64(writer digestWriter, value uint64) {
	var buffer [8]byte
	binary.BigEndian.PutUint64(buffer[:], value)
	_, _ = writer.Write(buffer[:])
}

func hashStrings(values ...string) [32]byte {
	hash := sha256.New()
	for _, value := range values {
		writeString(hash, value)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func sortedMarkets[V any](values map[Market]V) []Market {
	markets := make([]Market, 0, len(values))
	for market := range values {
		markets = append(markets, market)
	}
	sort.Slice(markets, func(i, j int) bool { return markets[i] < markets[j] })
	return markets
}
