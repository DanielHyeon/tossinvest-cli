package accountbudget

import (
	"errors"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// ObservationKey는 caller가 계좌 결속을 확인한 뒤 제공하는 비원문 slot임.
type ObservationKey [32]byte

type ObservationKind uint8

const (
	ObservationPoll ObservationKind = iota + 1
	ObservationToken
	ObservationOther
)

type ObservationFinish uint8

const (
	ObservationNoResponse ObservationFinish = iota + 1
	ObservationResponse
	ObservationRateLimited
)

var ErrObservationPoisoned = errors.New("observation poll slot poisoned")

type ObservationPending struct {
	Sequence uint64
	Kind     ObservationKind
}
type observationPending struct {
	Sequence uint64          `json:"sequence"`
	Kind     ObservationKind `json:"kind"`
	Key      ObservationKey  `json:"key"`
}
type ObservationCompletion struct {
	Attempt uint64            `json:"attempt"`
	Kind    ObservationKind   `json:"kind"`
	Result  ObservationFinish `json:"result"`
}
type ObservationEvent struct {
	Sequence uint64          `json:"sequence"`
	Attempt  uint64          `json:"attempt"`
	Kind     ObservationKind `json:"kind"`
	Stamp    clock.BootStamp `json:"stamp"`
}
type ObservationGap struct {
	Sequence uint64          `json:"sequence"`
	Attempt  uint64          `json:"attempt"`
	Kind     ObservationKind `json:"kind"`
}
type ObservationPendingStatus uint8

const (
	ObservationPendingNone ObservationPendingStatus = iota
	ObservationPendingBusy
	ObservationPendingOrphan
)

type ObservationCursor struct {
	Scope    string
	Instance string
	Event    uint64
	Gap      uint64
}
type ObservationView struct {
	Last          *ObservationCompletion
	Cursor        ObservationCursor
	Generation    uint64
	Events        []ObservationEvent
	Gaps          []ObservationGap
	EventOverrun  bool
	GapOverrun    bool
	Pending       *ObservationPending
	PendingStatus ObservationPendingStatus
}
type observationHint struct {
	Stamp    clock.BootStamp  `json:"stamp"`
	Duration time.Duration    `json:"duration"`
	Anchor   *clock.BootStamp `json:"anchor"`
}
type observationSlot struct {
	Key      ObservationKey   `json:"key"`
	Poisoned bool             `json:"poisoned"`
	Hint     *observationHint `json:"hint"`
}
type observationState struct {
	Version    uint8                  `json:"version"`
	Instance   string                 `json:"instance"`
	Scope      string                 `json:"scope"`
	Generation uint64                 `json:"generation"`
	Attempt    uint64                 `json:"attempt"`
	Event      uint64                 `json:"event"`
	Gap        uint64                 `json:"gap"`
	Pending    *observationPending    `json:"pending"`
	Last       *ObservationCompletion `json:"last"`
	Events     []ObservationEvent     `json:"events"`
	Gaps       []ObservationGap       `json:"gaps"`
	Slots      []observationSlot      `json:"slots"`
}
