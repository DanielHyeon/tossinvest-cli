package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const (
	DefaultStreamMaxClients = 32
	DefaultStreamQueueSize  = 64
	DefaultStreamHeartbeat  = 15 * time.Second
	MaxLastEventIDBytes     = 160

	StreamEventSnapshot = "snapshot"
)

var (
	ErrStreamClientLimit = errors.New("stream client limit reached")
	ErrStreamClosed      = errors.New("stream closed")
	ErrStreamExhausted   = errors.New("stream sequence exhausted")
)

// SnapshotFunc returns the complete, already-serialized read model used when a
// client cannot safely resume from the in-memory event history.
type SnapshotFunc func(context.Context) ([]byte, error)

type StreamOptions struct {
	Epoch      string
	MaxClients int
	QueueSize  int
	Heartbeat  time.Duration
	Clock      clock.Clock
}

type StreamEvent struct {
	ID       string
	Sequence uint64
	Type     string
	Data     []byte
}

// Stream is a bounded, process-local SSE event hub. It deliberately provides
// no durable replay promise: an epoch mismatch or history gap converges through
// a fresh snapshot.
type Stream struct {
	mu sync.Mutex

	epoch      string
	sequence   uint64
	maxClients int
	queueSize  int
	heartbeat  time.Duration
	clock      clock.Clock
	snapshot   SnapshotFunc
	history    []StreamEvent
	clients    map[uint64]*streamClient
	nextClient uint64
	closed     bool
}

type streamClient struct {
	id     uint64
	events chan StreamEvent
	done   chan struct{}
	closed bool
}

type streamSubscription struct {
	stream  *Stream
	client  *streamClient
	initial []StreamEvent
	events  <-chan StreamEvent
	done    <-chan struct{}
	once    sync.Once
}

func NewStream(options StreamOptions, snapshot SnapshotFunc) (*Stream, error) {
	if snapshot == nil {
		return nil, errors.New("stream snapshot provider is required")
	}

	epoch := options.Epoch
	if epoch == "" {
		var err error
		epoch, err = newStreamEpoch()
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(epoch) != epoch || strings.ContainsAny(epoch, ":\r\n") {
		return nil, errors.New("stream epoch must be a non-empty SSE-safe token")
	}

	maxClients := options.MaxClients
	if maxClients == 0 {
		maxClients = DefaultStreamMaxClients
	}
	if maxClients < 1 || maxClients > DefaultStreamMaxClients {
		return nil, fmt.Errorf("stream max clients must be between 1 and %d", DefaultStreamMaxClients)
	}
	queueSize := options.QueueSize
	if queueSize == 0 {
		queueSize = DefaultStreamQueueSize
	}
	if queueSize < 1 || queueSize > DefaultStreamQueueSize {
		return nil, fmt.Errorf("stream queue size must be between 1 and %d", DefaultStreamQueueSize)
	}
	heartbeat := options.Heartbeat
	if heartbeat == 0 {
		heartbeat = DefaultStreamHeartbeat
	}
	if heartbeat < 1 || heartbeat > DefaultStreamHeartbeat {
		return nil, fmt.Errorf("stream heartbeat must be between 1ns and %s", DefaultStreamHeartbeat)
	}
	streamClock := options.Clock
	if streamClock == nil {
		streamClock = clock.System()
	}

	return &Stream{
		epoch:      epoch,
		maxClients: maxClients,
		queueSize:  queueSize,
		heartbeat:  heartbeat,
		clock:      streamClock,
		snapshot:   snapshot,
		history:    make([]StreamEvent, 0, queueSize),
		clients:    make(map[uint64]*streamClient, maxClients),
	}, nil
}

func newStreamEpoch() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate stream epoch: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// Publish records an event and offers it to every connected client without
// waiting. A full client queue evicts only that client.
func (s *Stream) Publish(eventType string, data []byte) (StreamEvent, error) {
	if eventType == "" || strings.ContainsAny(eventType, "\r\n") {
		return StreamEvent{}, errors.New("stream event type must be SSE-safe")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return StreamEvent{}, ErrStreamClosed
	}
	if s.sequence == math.MaxUint64 {
		return StreamEvent{}, ErrStreamExhausted
	}

	s.sequence++
	event := StreamEvent{
		ID:       streamEventID(s.epoch, s.sequence),
		Sequence: s.sequence,
		Type:     eventType,
		Data:     append([]byte(nil), data...),
	}
	s.history = append(s.history, event)
	if len(s.history) > s.queueSize {
		copy(s.history, s.history[len(s.history)-s.queueSize:])
		s.history = s.history[:s.queueSize]
	}

	for id, client := range s.clients {
		select {
		case client.events <- cloneStreamEvent(event):
		default:
			s.closeClientLocked(id, client)
		}
	}
	return cloneStreamEvent(event), nil
}

func streamEventID(epoch string, sequence uint64) string {
	return epoch + ":" + strconv.FormatUint(sequence, 10)
}

func (s *Stream) subscribe(ctx context.Context, lastEventID string) (*streamSubscription, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, ErrStreamClosed
	}
	if len(s.clients) >= s.maxClients {
		s.mu.Unlock()
		return nil, ErrStreamClientLimit
	}

	baseSequence := s.sequence
	initial, resumable := s.resumeEventsLocked(lastEventID)
	s.nextClient++
	client := &streamClient{
		id:     s.nextClient,
		events: make(chan StreamEvent, s.queueSize),
		done:   make(chan struct{}),
	}
	s.clients[client.id] = client
	s.mu.Unlock()

	subscription := &streamSubscription{
		stream:  s,
		client:  client,
		initial: initial,
		events:  client.events,
		done:    client.done,
	}
	if !resumable {
		data, err := s.snapshot(ctx)
		if err != nil {
			subscription.close()
			return nil, fmt.Errorf("read stream snapshot: %w", err)
		}
		subscription.initial = []StreamEvent{{
			ID:       streamEventID(s.epoch, baseSequence),
			Sequence: baseSequence,
			Type:     StreamEventSnapshot,
			Data:     append([]byte(nil), data...),
		}}
	}

	if ctx.Done() != nil {
		go func() {
			select {
			case <-ctx.Done():
				subscription.close()
			case <-subscription.done:
			}
		}()
	}
	return subscription, nil
}

func (s *Stream) resumeEventsLocked(lastEventID string) ([]StreamEvent, bool) {
	epoch, sequence, ok := parseStreamEventID(lastEventID)
	if !ok || epoch != s.epoch || sequence > s.sequence {
		return nil, false
	}
	if sequence == s.sequence {
		return nil, true
	}
	want := sequence + 1
	if len(s.history) == 0 || s.history[0].Sequence > want {
		return nil, false
	}

	replay := make([]StreamEvent, 0, s.sequence-sequence)
	for _, event := range s.history {
		if event.Sequence < want {
			continue
		}
		if event.Sequence != want {
			return nil, false
		}
		replay = append(replay, cloneStreamEvent(event))
		want++
	}
	if want != s.sequence+1 {
		return nil, false
	}
	return replay, true
}

func parseStreamEventID(id string) (string, uint64, bool) {
	epoch, sequenceText, found := strings.Cut(id, ":")
	if !found || epoch == "" || sequenceText == "" || strings.Contains(sequenceText, ":") {
		return "", 0, false
	}
	sequence, err := strconv.ParseUint(sequenceText, 10, 64)
	if err != nil || strconv.FormatUint(sequence, 10) != sequenceText {
		return "", 0, false
	}
	return epoch, sequence, true
}

func cloneStreamEvent(event StreamEvent) StreamEvent {
	event.Data = append([]byte(nil), event.Data...)
	return event
}

func (s *Stream) ClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

func (s *Stream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	for id, client := range s.clients {
		s.closeClientLocked(id, client)
	}
}

func (s *Stream) closeClientLocked(id uint64, client *streamClient) {
	if client.closed {
		return
	}
	client.closed = true
	delete(s.clients, id)
	close(client.done)
}

func (s *streamSubscription) close() {
	s.once.Do(func() {
		s.stream.mu.Lock()
		defer s.stream.mu.Unlock()
		s.stream.closeClientLocked(s.client.id, s.client)
	})
}

func (s *Stream) heartbeats(ctx context.Context) <-chan struct{} {
	beats := make(chan struct{}, 1)
	go func() {
		defer close(beats)
		for {
			if err := s.clock.Sleep(ctx, s.heartbeat); err != nil {
				return
			}
			select {
			case beats <- struct{}{}:
			default:
			}
		}
	}()
	return beats
}

// ServeHTTP serves one exact SSE connection. Route matching remains the
// router's responsibility; this handler owns only stream protocol behavior.
func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		writeStreamError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "The stream accepts GET or HEAD only.", s.clock.Now())
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeStreamError(w, http.StatusInternalServerError, "STREAMING_UNSUPPORTED", "The response transport does not support streaming.", s.clock.Now())
		return
	}

	if r.Method == http.MethodHead {
		setStreamHeaders(w.Header())
		w.WriteHeader(http.StatusOK)
		return
	}

	lastEventID, ok := exactLastEventID(r.Header)
	if !ok {
		writeStreamError(w, http.StatusBadRequest, "INVALID_LAST_EVENT_ID", "Last-Event-ID must be one exact bounded value.", s.clock.Now())
		return
	}
	subscription, err := s.subscribe(r.Context(), lastEventID)
	if err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, ErrStreamClientLimit) {
			w.Header().Set("Retry-After", "15")
		}
		writeStreamError(w, status, "STREAM_UNAVAILABLE", "The bounded stream cannot be opened; reconnect after Retry-After.", s.clock.Now())
		return
	}
	defer subscription.close()

	setStreamHeaders(w.Header())
	w.WriteHeader(http.StatusOK)
	for _, event := range subscription.initial {
		if err := writeSSEEvent(w, event); err != nil {
			return
		}
	}
	flusher.Flush()

	heartbeatContext, stopHeartbeats := context.WithCancel(r.Context())
	defer stopHeartbeats()
	beats := s.heartbeats(heartbeatContext)
	for {
		select {
		case <-subscription.done:
			return
		default:
		}
		select {
		case <-r.Context().Done():
			return
		case <-subscription.done:
			return
		case event := <-subscription.events:
			if err := writeSSEEvent(w, event); err != nil {
				return
			}
			flusher.Flush()
		case _, ok := <-beats:
			if !ok {
				return
			}
			if _, err := io.WriteString(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func exactLastEventID(header http.Header) (string, bool) {
	values, present := header[http.CanonicalHeaderKey("Last-Event-ID")]
	if !present {
		return "", true
	}
	if len(values) != 1 || len(values[0]) > MaxLastEventIDBytes || strings.TrimSpace(values[0]) != values[0] ||
		strings.ContainsAny(values[0], "\r\n,") {
		return "", false
	}
	return values[0], true
}

func writeStreamError(w http.ResponseWriter, status int, code, message string, now time.Time) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, _ = w.Write(errorResponseBody(code, message, nil, now))
}

func setStreamHeaders(header http.Header) {
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache, no-transform")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
}

func writeSSEEvent(w io.Writer, event StreamEvent) error {
	if _, err := fmt.Fprintf(w, "id: %s\nevent: %s\n", event.ID, event.Type); err != nil {
		return err
	}
	for _, line := range strings.Split(string(event.Data), "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", strings.TrimSuffix(line, "\r")); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}
