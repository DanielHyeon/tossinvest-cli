package strategyruntime

import (
	"crypto/sha256"
	"encoding/binary"
	"time"
	"unicode"
	"unicode/utf8"
)

const maximumIdentityBytes = 256

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

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func validIdentity(value string) bool {
	if value == "" || len(value) > maximumIdentityBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validOptionalIdentity(value string) bool { return value == "" || validIdentity(value) }

func hexBytes(data []byte) string {
	const alphabet = "0123456789abcdef"
	encoded := make([]byte, len(data)*2)
	for index, value := range data {
		encoded[index*2], encoded[index*2+1] = alphabet[value>>4], alphabet[value&15]
	}
	return string(encoded)
}
