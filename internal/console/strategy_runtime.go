package console

import (
	"context"
	"net/http"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

// StrategyRuntimeControl is an authority-owned projection. The console must
// display these states independently and must not derive one control from
// another (for example, by ANDing lane and LIVE approval).
type StrategyRuntimeControl struct {
	Default   strategyengine.RuntimeState
	Desired   strategyengine.RuntimeState
	Effective strategyengine.RuntimeState
	Reason    strategyengine.RuntimeRefusal
}

type StrategyRuntimeEntryCapability struct {
	Default      strategyengine.RuntimeState
	Desired      strategyengine.RuntimeState
	Effective    strategyengine.RuntimeState
	FirstRefusal strategyengine.RuntimeRefusal
}

type StrategyRuntimeReading struct {
	Descriptor      strategyengine.RuntimeDescriptor
	GeneratedAt     time.Time
	ObservedAt      time.Time
	Freshness       strategyengine.RuntimeState
	Lane            StrategyRuntimeControl
	AutoStart       StrategyRuntimeControl
	GateApproval    StrategyRuntimeControl
	LiveApproval    StrategyRuntimeControl
	EntryCapability StrategyRuntimeEntryCapability
	Blockers        []strategyengine.RuntimeBlocker
}

type StrategyRuntimeReader interface {
	Read(context.Context) (StrategyRuntimeReading, error)
}

type runtimeStateView struct {
	Value string
	Class string
}

type runtimeControlView struct {
	Default   runtimeStateView
	Desired   runtimeStateView
	Effective runtimeStateView
	Reason    string
}

type runtimeFieldView struct {
	Key, Label, Help, Default, Desired, Effective, Unit, Range, Provenance, ApplyTiming string
}

type runtimeBlockerView struct {
	Key, Label                    string
	Desired, Effective, Freshness runtimeStateView
	Reason                        string
}

type strategyRuntimePage struct {
	Nav, ParameterSection, LaneSection, AutoStartSection, LiveSection string
	LoadErr, Unwired                                                  bool
	GeneratedAt, ObservedAt                                           string
	Freshness                                                         runtimeStateView
	Fields                                                            []runtimeFieldView
	Lane, AutoStart, GateApproval, LiveApproval                       runtimeControlView
	EntryDefault, EntryDesired, EntryEffective                        runtimeStateView
	FirstRefusal                                                      string
	Blockers                                                          []runtimeBlockerView
}

func (strategyRuntimePage) Refresh() bool { return false }

func (c *Console) handleStrategyRuntime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다", "전략 상태는 GET/HEAD만 허용한다. 아무것도 전송되지 않았다.")
		return
	}

	reading := dormantStrategyRuntimeReading(time.Now())
	page := strategyRuntimePage{Nav: "optimization", Unwired: c.opts.StrategyRuntime == nil}
	if c.opts.StrategyRuntime != nil {
		value, err := c.opts.StrategyRuntime.Read(r.Context())
		if err != nil || !validStrategyRuntimeReading(value) {
			page.LoadErr = true
			reading.EntryCapability.FirstRefusal = strategyengine.RuntimeRefusalReadFailed
		} else {
			reading = value
		}
	}

	page.project(reading)
	c.render(w, "strategy-runtime", page)
}

func dormantStrategyRuntimeReading(generatedAt time.Time) StrategyRuntimeReading {
	descriptor := strategyengine.DormantRuntimeDescriptor()
	return StrategyRuntimeReading{
		Descriptor:  descriptor,
		GeneratedAt: generatedAt,
		Freshness:   strategyengine.RuntimeStateUnobserved,
		Lane: StrategyRuntimeControl{
			Default: strategyengine.RuntimeStateOff, Desired: strategyengine.RuntimeStateOff,
			Effective: strategyengine.RuntimeStateOff, Reason: strategyengine.RuntimeRefusalSourceManifest,
		},
		AutoStart: StrategyRuntimeControl{
			Default: strategyengine.RuntimeStateOff, Desired: strategyengine.RuntimeStateOff,
			Effective: strategyengine.RuntimeStateOff, Reason: strategyengine.RuntimeRefusalActivationManifestAbsent,
		},
		GateApproval: StrategyRuntimeControl{
			Default: strategyengine.RuntimeStateUnapproved, Desired: strategyengine.RuntimeStateUnapproved,
			Effective: strategyengine.RuntimeStateUnapproved, Reason: strategyengine.RuntimeRefusalGuardianUnapproved,
		},
		LiveApproval: StrategyRuntimeControl{
			Default: strategyengine.RuntimeStateUnapproved, Desired: strategyengine.RuntimeStateUnapproved,
			Effective: strategyengine.RuntimeStateUnapproved, Reason: strategyengine.RuntimeRefusalActivationManifestAbsent,
		},
		EntryCapability: StrategyRuntimeEntryCapability{
			Default: strategyengine.RuntimeStateOff, Desired: strategyengine.RuntimeStateOff,
			Effective: strategyengine.RuntimeStateOff, FirstRefusal: strategyengine.RuntimeRefusalSourceManifest,
		},
		Blockers: append([]strategyengine.RuntimeBlocker(nil), descriptor.Blockers[:]...),
	}
}

func validStrategyRuntimeReading(reading StrategyRuntimeReading) bool {
	if reading.Descriptor.Category != "strategy-runtime" || !runtimeFreshnessValid(reading.Freshness) ||
		reading.GeneratedAt.IsZero() ||
		(reading.ObservedAt.IsZero() && reading.Freshness != strategyengine.RuntimeStateUnobserved) ||
		(!reading.ObservedAt.IsZero() && reading.GeneratedAt.Before(reading.ObservedAt)) {
		return false
	}
	expected := strategyengine.DormantRuntimeDescriptor()
	for i, section := range reading.Descriptor.Sections {
		if section.ID != expected.Sections[i].ID || section.Label == "" || section.ActionOwner == "" {
			return false
		}
	}
	for i, field := range reading.Descriptor.Fields {
		if field.Key != expected.Fields[i].Key || field.Label == "" || field.Help == "" || field.Default == "" ||
			field.Desired == "" || field.Effective == "" || field.Unit == "" || field.Range == "" ||
			field.Provenance == "" || field.ApplyTiming == "" {
			return false
		}
	}
	if !runtimeToggleControlValid(reading.Lane) || !runtimeToggleControlValid(reading.AutoStart) ||
		!runtimeApprovalControlValid(reading.GateApproval) || !runtimeApprovalControlValid(reading.LiveApproval) {
		return false
	}
	if !runtimeToggleStateValid(reading.EntryCapability.Default) ||
		!runtimeToggleStateValid(reading.EntryCapability.Desired) ||
		!runtimeToggleStateValid(reading.EntryCapability.Effective) ||
		!reading.EntryCapability.FirstRefusal.Valid() {
		return false
	}
	if reading.EntryCapability.Effective == strategyengine.RuntimeStateOn {
		if reading.EntryCapability.FirstRefusal != strategyengine.RuntimeRefusalNone {
			return false
		}
	} else if reading.EntryCapability.FirstRefusal == strategyengine.RuntimeRefusalNone {
		return false
	}

	if len(reading.Blockers) != len(expected.Blockers) {
		return false
	}
	for i, blocker := range reading.Blockers {
		if blocker.Key != expected.Blockers[i].Key || blocker.Label == "" || !blocker.Desired.Valid() ||
			!blocker.Effective.Valid() || !runtimeFreshnessValid(blocker.Freshness) || !blocker.Reason.Valid() {
			return false
		}
	}
	return true
}

func runtimeToggleControlValid(control StrategyRuntimeControl) bool {
	return runtimeToggleStateValid(control.Default) && runtimeToggleStateValid(control.Desired) &&
		runtimeToggleStateValid(control.Effective) && control.Reason.Valid()
}

func runtimeApprovalControlValid(control StrategyRuntimeControl) bool {
	return runtimeApprovalStateValid(control.Default) && runtimeApprovalStateValid(control.Desired) &&
		runtimeApprovalStateValid(control.Effective) && control.Reason.Valid()
}

func runtimeToggleStateValid(state strategyengine.RuntimeState) bool {
	return state == strategyengine.RuntimeStateOff || state == strategyengine.RuntimeStateOn
}

func runtimeApprovalStateValid(state strategyengine.RuntimeState) bool {
	return state == strategyengine.RuntimeStateUnapproved || state == strategyengine.RuntimeStateVerified
}

func runtimeFreshnessValid(state strategyengine.RuntimeState) bool {
	return state == strategyengine.RuntimeStateUnobserved || state == strategyengine.RuntimeStateVerified ||
		state == strategyengine.RuntimeStateStale
}

func (page *strategyRuntimePage) project(reading StrategyRuntimeReading) {
	page.ParameterSection = reading.Descriptor.Sections[0].Label
	page.LaneSection = reading.Descriptor.Sections[1].Label
	page.AutoStartSection = reading.Descriptor.Sections[2].Label
	page.LiveSection = reading.Descriptor.Sections[3].Label
	page.GeneratedAt = runtimeTime(reading.GeneratedAt)
	page.ObservedAt = runtimeTime(reading.ObservedAt)
	page.Freshness = runtimeState(reading.Freshness)
	for _, field := range reading.Descriptor.Fields {
		page.Fields = append(page.Fields, runtimeFieldView{
			Key: field.Key, Label: field.Label, Help: field.Help, Default: field.Default,
			Desired: field.Desired, Effective: field.Effective, Unit: field.Unit, Range: field.Range,
			Provenance: field.Provenance, ApplyTiming: field.ApplyTiming,
		})
	}
	page.Lane = runtimeControl(reading.Lane)
	page.AutoStart = runtimeControl(reading.AutoStart)
	page.GateApproval = runtimeControl(reading.GateApproval)
	page.LiveApproval = runtimeControl(reading.LiveApproval)
	page.EntryDefault = runtimeState(reading.EntryCapability.Default)
	page.EntryDesired = runtimeState(reading.EntryCapability.Desired)
	page.EntryEffective = runtimeState(reading.EntryCapability.Effective)
	page.FirstRefusal = runtimeRefusal(reading.EntryCapability.FirstRefusal)
	for _, blocker := range reading.Blockers {
		page.Blockers = append(page.Blockers, runtimeBlockerView{
			Key: blocker.Key, Label: blocker.Label, Desired: runtimeState(blocker.Desired),
			Effective: runtimeState(blocker.Effective), Freshness: runtimeState(blocker.Freshness),
			Reason: runtimeRefusal(blocker.Reason),
		})
	}
}

func runtimeControl(control StrategyRuntimeControl) runtimeControlView {
	return runtimeControlView{
		Default: runtimeState(control.Default), Desired: runtimeState(control.Desired),
		Effective: runtimeState(control.Effective), Reason: runtimeRefusal(control.Reason),
	}
}

func runtimeState(state strategyengine.RuntimeState) runtimeStateView {
	class := "muted"
	switch state {
	case strategyengine.RuntimeStateOn, strategyengine.RuntimeStateVerified, strategyengine.RuntimeStateReady,
		strategyengine.RuntimeStateHealthy, strategyengine.RuntimeStateLive, strategyengine.RuntimeStateValid:
		class = "ok"
	case strategyengine.RuntimeStateUnwired, strategyengine.RuntimeStateMissing, strategyengine.RuntimeStateStale,
		strategyengine.RuntimeStateRefused, strategyengine.RuntimeStateUnknown:
		class = "bad"
	}
	return runtimeStateView{Value: string(state), Class: class}
}

func runtimeRefusal(refusal strategyengine.RuntimeRefusal) string {
	if refusal == strategyengine.RuntimeRefusalNone {
		return "entry_permitted"
	}
	return string(refusal)
}

func runtimeTime(value time.Time) string {
	if value.IsZero() {
		return "관측 없음"
	}
	return value.UTC().Format(time.RFC3339)
}
