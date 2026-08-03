package console

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// InstrumentRef is one journal-owned instrument identity. Market and Symbol are
// both required because the same symbol spelling must never cross-attach between
// KR and US.
type InstrumentRef struct {
	Market string
	Symbol string
}

// InstrumentName is optional, current display metadata for one InstrumentRef.
// It is never frozen trade evidence and is never written to the journal.
type InstrumentName struct {
	Market string
	Symbol string
	Name   string
}

// InstrumentNameReader is the console's complete instrument-metadata
// capability: one batch read that returns plain strings. No broker, credential,
// order, config, or journal handle is reachable through this seam.
type InstrumentNameReader interface {
	Names(context.Context, []InstrumentRef) ([]InstrumentName, error)
}

const (
	instrumentNameTTL         = 24 * time.Hour
	instrumentNameTimeout     = 10 * time.Second
	instrumentNameRetryAfter  = time.Minute
	instrumentNameCacheLimit  = 2048
	historyInstrumentRefLimit = 400
	maxInstrumentNameRunes    = 80
)

type instrumentNameEntry struct {
	name    string
	expires time.Time
}

// instrumentNameCache is a bounded, single-flight cache. Only accepted names are
// cached for the full TTL. A successful-but-partial response gets the short
// failure backoff instead: treating an omitted or rejected row as an authoritative
// 24-hour negative result would hide transient upstream truncation.
type instrumentNameCache struct {
	gateOnce       sync.Once
	gate           chan struct{}
	entries        map[string]instrumentNameEntry
	failureUntil   time.Time
	failureMessage string
}

func (c *instrumentNameCache) lock(ctx context.Context) error {
	c.gateOnce.Do(func() {
		c.gate = make(chan struct{}, 1)
		c.gate <- struct{}{}
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.gate:
		return nil
	}
}

func (c *instrumentNameCache) unlock() { c.gate <- struct{}{} }

func historyInstrumentRefs(v historyView) []InstrumentRef {
	refs := make([]InstrumentRef, 0, len(v.Trips)+len(v.Events))
	for _, row := range v.Trips {
		refs = append(refs, InstrumentRef{Market: row.Market, Symbol: row.Symbol})
	}
	for _, row := range v.Events {
		refs = append(refs, InstrumentRef{Market: row.Market, Symbol: row.Symbol})
	}
	return refs
}

func (c *instrumentNameCache) get(ctx context.Context, reader InstrumentNameReader,
	refs []InstrumentRef, now time.Time, hold bool, holdReason string) (map[string]string, string) {
	unique := canonicalInstrumentRefs(refs)
	if len(unique) == 0 {
		return nil, ""
	}

	lookupCtx, cancel := context.WithTimeout(ctx, instrumentNameTimeout)
	defer cancel()
	if err := c.lock(lookupCtx); err != nil {
		return nil, "종목명 조회 시간 초과 — 심볼과 동결 거래 값은 그대로 표시한다: " + err.Error()
	}
	defer c.unlock()
	if c.entries == nil {
		c.entries = make(map[string]instrumentNameEntry)
	}
	c.evictExpired(now)

	resolved, missing := c.cached(unique, now)
	if len(missing) == 0 {
		return resolved, ""
	}
	if hold {
		return resolved, "실계좌 검증 중이라 종목명 조회 보류 — 심볼은 그대로 표시한다 (" + holdReason + ")"
	}
	if reader == nil {
		return resolved, "종목명 조회 기능이 연결되지 않아 심볼만 표시한다."
	}
	if now.Before(c.failureUntil) {
		return resolved, c.failureMessage
	}

	notice := ""
	if len(missing) > historyInstrumentRefLimit {
		missing = missing[:historyInstrumentRefLimit]
		notice = fmt.Sprintf("종목명 조회 한도 %d개를 넘어 일부 행은 심볼만 표시한다.", historyInstrumentRefLimit)
	}
	rows, err := reader.Names(lookupCtx, missing)
	if err != nil {
		c.failureUntil = now.Add(instrumentNameRetryAfter)
		c.failureMessage = "종목명 조회 실패 — 잠시 재시도하지 않으며 심볼과 동결 거래 값은 그대로 표시한다: " + err.Error()
		return resolved, c.failureMessage
	}
	c.failureUntil = time.Time{}
	c.failureMessage = ""
	accepted := acceptedInstrumentNames(missing, rows)
	c.storeAccepted(accepted, resolved, now)
	if len(accepted) != len(missing) {
		c.failureUntil = now.Add(instrumentNameRetryAfter)
		c.failureMessage = "종목명 일부를 확인하지 못해 심볼만 표시한다. 잠시 후 자동으로 다시 조회한다."
		if notice == "" {
			notice = c.failureMessage
		} else {
			notice += " " + c.failureMessage
		}
	}
	return resolved, notice
}

func (c *instrumentNameCache) cached(refs []InstrumentRef, now time.Time) (map[string]string, []InstrumentRef) {
	resolved := make(map[string]string, len(refs))
	missing := make([]InstrumentRef, 0, len(refs))
	for _, ref := range refs {
		key := symbolKey(ref.Market, ref.Symbol)
		entry, ok := c.entries[key]
		if !ok || !now.Before(entry.expires) {
			missing = append(missing, ref)
			continue
		}
		if entry.name != "" {
			resolved[key] = entry.name
		}
	}
	return resolved, missing
}

func acceptedInstrumentNames(requestedRefs []InstrumentRef, rows []InstrumentName) map[string]string {
	requested := make(map[string]bool, len(requestedRefs))
	for _, ref := range requestedRefs {
		requested[symbolKey(ref.Market, ref.Symbol)] = true
	}
	accepted := make(map[string]string, len(rows))
	conflicted := make(map[string]bool)
	for _, row := range rows {
		ref, ok := canonicalInstrumentRef(InstrumentRef{Market: row.Market, Symbol: row.Symbol})
		if !ok {
			continue
		}
		key := symbolKey(ref.Market, ref.Symbol)
		if !requested[key] || conflicted[key] {
			continue
		}
		name, ok := safeInstrumentName(row.Name)
		if !ok {
			continue
		}
		if prior, exists := accepted[key]; exists && prior != name {
			delete(accepted, key)
			conflicted[key] = true
			continue
		}
		accepted[key] = name
	}
	return accepted
}

func (c *instrumentNameCache) storeAccepted(accepted, resolved map[string]string, now time.Time) {
	for key, name := range accepted {
		c.put(key, instrumentNameEntry{name: name, expires: now.Add(instrumentNameTTL)})
		resolved[key] = name
	}
}

func canonicalInstrumentRefs(refs []InstrumentRef) []InstrumentRef {
	seen := make(map[string]bool, len(refs))
	out := make([]InstrumentRef, 0, len(refs))
	for _, raw := range refs {
		ref, ok := canonicalInstrumentRef(raw)
		if !ok {
			continue
		}
		key := symbolKey(ref.Market, ref.Symbol)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool {
		return symbolKey(out[i].Market, out[i].Symbol) < symbolKey(out[j].Market, out[j].Symbol)
	})
	return out
}

func canonicalInstrumentRef(ref InstrumentRef) (InstrumentRef, bool) {
	ref.Market = strings.ToLower(strings.TrimSpace(ref.Market))
	ref.Symbol = strings.ToUpper(strings.TrimSpace(ref.Symbol))
	if (ref.Market != "kr" && ref.Market != "us") || ref.Symbol == "" {
		return InstrumentRef{}, false
	}
	return ref, true
}

func safeInstrumentName(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	runes := []rune(name)
	if name == "" || len(runes) > maxInstrumentNameRunes {
		return "", false
	}
	for _, r := range runes {
		if unicode.IsControl(r) || isBidiControl(r) {
			return "", false
		}
	}
	return name, true
}

func isBidiControl(r rune) bool {
	return r == '\u061c' || r == '\u200e' || r == '\u200f' ||
		(r >= '\u202a' && r <= '\u202e') || (r >= '\u2066' && r <= '\u2069')
}

func (c *instrumentNameCache) evictExpired(now time.Time) {
	for key, entry := range c.entries {
		if !now.Before(entry.expires) {
			delete(c.entries, key)
		}
	}
}

func (c *instrumentNameCache) put(key string, entry instrumentNameEntry) {
	if _, exists := c.entries[key]; !exists && len(c.entries) >= instrumentNameCacheLimit {
		oldestKey := ""
		var oldest time.Time
		for candidate, value := range c.entries {
			if oldestKey == "" || value.expires.Before(oldest) {
				oldestKey, oldest = candidate, value.expires
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[key] = entry
}
