// Package httpapi defines the versioned, transport-only operator API contract.
// It accepts already-authorized read models and deliberately owns no broker,
// journal writer, engine command, gate, kill-switch, or LIVE-order capability.
package httpapi

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

const (
	SchemaVersion      = "tossos.operator-api/v1"
	ErrorSchemaVersion = "tossos.operator-api.error/v1"
)

type Envelope struct {
	SchemaVersion string    `json:"schemaVersion"`
	Resource      string    `json:"resource"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Data          any       `json:"data"`
}

type ErrorResponse struct {
	SchemaVersion string     `json:"schemaVersion"`
	Error         ErrorModel `json:"error"`
}

type ErrorModel struct {
	Code          string        `json:"code"`
	Message       string        `json:"message"`
	Details       []ErrorDetail `json:"details"`
	RequestID     string        `json:"requestId"`
	Timestamp     time.Time     `json:"timestamp"`
	Documentation string        `json:"documentation"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func errorResponseBody(code, message string, details []ErrorDetail, now time.Time) []byte {
	if details == nil {
		details = []ErrorDetail{}
	}
	body, err := json.Marshal(ErrorResponse{SchemaVersion: ErrorSchemaVersion, Error: ErrorModel{
		Code: code, Message: message, Details: details, RequestID: newRequestID(), Timestamp: now.UTC(),
		Documentation: "/docs/api/openapi-v1.json#" + strings.ToLower(code),
	}})
	if err != nil {
		panic("httpapi: static error response cannot be encoded: " + err.Error())
	}
	return body
}

type EngineResource struct {
	Status      string     `json:"status"`
	Running     bool       `json:"running"`
	PID         int        `json:"pid"`
	StartedAt   *time.Time `json:"startedAt"`
	RefreshedAt *time.Time `json:"refreshedAt"`
	BuildAt     *time.Time `json:"buildAt"`
	Stale       bool       `json:"stale"`
	Source      string     `json:"source"`
}

type PositionsResource struct {
	ObservedAt *time.Time `json:"observedAt"`
	Stale      bool       `json:"stale"`
	Source     string     `json:"source"`
	Items      []Position `json:"items"`
}

type Position struct {
	AccountLabel       string              `json:"accountLabel"`
	PositionID         string              `json:"positionId"`
	Market             string              `json:"market"`
	Symbol             string              `json:"symbol"`
	Name               string              `json:"name"`
	Quantity           string              `json:"quantity"`
	AveragePrice       string              `json:"averagePrice"`
	LastPrice          string              `json:"lastPrice"`
	MarketValue        string              `json:"marketValue"`
	UnrealizedPnL      string              `json:"unrealizedPnl"`
	ProfitRate         string              `json:"profitRate"`
	InBroker           bool                `json:"inBroker"`
	InJournal          bool                `json:"inJournal"`
	Eligible           bool                `json:"eligible"`
	ManagementStatus   string              `json:"managementStatus"`
	AdoptionStatus     AdoptionStatus      `json:"adoptionStatus"`
	StatusKnown        bool                `json:"statusKnown"`
	AdoptionLabel      string              `json:"adoptionLabel"`
	AdoptionReason     AdoptionReason      `json:"adoptionReason"`
	Included           bool                `json:"included"`
	Excluded           bool                `json:"excluded"`
	Candidate          bool                `json:"candidate"`
	DesignationKnown   bool                `json:"designationKnown"`
	CoveringBlock      *ReconcileBlock     `json:"coveringBlock"`
	ExitLine           ExitLine            `json:"exitLine"`
	StoredExitEvidence *StoredExitEvidence `json:"storedExitEvidence"`
}

// AdoptionStatus and AdoptionReason are named transport scalars. They preserve
// the positionpolicy projector's stable values without exposing a callable
// engine or reconcile capability through the HTTP contract.
type AdoptionStatus string
type AdoptionReason string

// ReconcileBlock is intentionally narrower than positionpolicy.ReconcileBlock:
// account identity, free-form detail, permanence implementation details and
// command authority never cross the public read boundary.
type ReconcileBlock struct {
	Scope     string     `json:"scope"`
	Market    string     `json:"market"`
	Symbol    string     `json:"symbol"`
	Reason    string     `json:"reason"`
	StartedAt *time.Time `json:"startedAt"`
}

// StoredExitEvidence reports raw ledger facts separately from the actionable
// effective ExitLine. It must never be interpreted as current protection when
// EffectiveKnown is false.
type StoredExitEvidence struct {
	EntryPrice     string `json:"entryPrice"`
	InitialStop    string `json:"initialStop"`
	Baseline       string `json:"baseline"`
	HighWater      string `json:"highWater"`
	EffectiveKnown bool   `json:"effectiveKnown"`
	Label          string `json:"label"`
}

// ExitLine is a JSON spelling of operatorview.ExitLineView. The values are
// copied verbatim; the HTTP adapter never recomputes a protection price,
// projected quantity, action, or freshness verdict.
type ExitLine struct {
	Status             string `json:"status"`
	StatusText         string `json:"statusText"`
	Reason             string `json:"reason"`
	EntryPrice         string `json:"entryPrice"`
	InitialStop        string `json:"initialStop"`
	CurrentProtection  string `json:"currentProtection"`
	NextTarget         string `json:"nextTarget"`
	NextProtection     string `json:"nextProtection"`
	ProjectedQuantity  string `json:"projectedQuantity"`
	ObservedPrice      string `json:"observedPrice"`
	HighWater          string `json:"highWater"`
	Stage              string `json:"stage"`
	ActionText         string `json:"actionText"`
	Policy             string `json:"policy"`
	DecisionID         string `json:"decisionId"`
	SnapshotID         string `json:"snapshotId"`
	ObservationID      string `json:"observationId"`
	ObservationSource  string `json:"observationSource"`
	EvaluatedAt        string `json:"evaluatedAt"`
	EffectiveSource    string `json:"effectiveSource"`
	PositionGeneration string `json:"positionGeneration"`
	OneShare           bool   `json:"oneShare"`
	OneShareText       string `json:"oneShareText"`
	FinalExitText      string `json:"finalExitText"`
}

func ExitLineFrom(v operatorview.ExitLineView) ExitLine {
	return ExitLine{
		Status: v.Status, StatusText: v.StatusText, Reason: v.Reason,
		EntryPrice: v.EntryPrice, InitialStop: v.InitialStop, CurrentProtection: v.CurrentProtection,
		NextTarget: v.NextTarget, NextProtection: v.NextProtection, ProjectedQuantity: v.ProjectedQuantity,
		ObservedPrice: v.ObservedPrice, HighWater: v.HighWater, Stage: v.Stage, ActionText: v.ActionText,
		Policy: v.Policy, DecisionID: v.DecisionID, SnapshotID: v.SnapshotID, ObservationID: v.ObservationID,
		ObservationSource: v.ObservationSource, EvaluatedAt: v.EvaluatedAt, EffectiveSource: v.EffectiveSource,
		PositionGeneration: v.PositionGeneration, OneShare: v.OneShare, OneShareText: v.OneShareText,
		FinalExitText: v.FinalExitText,
	}
}

type OrdersResource struct {
	ObservedAt *time.Time `json:"observedAt"`
	Stale      bool       `json:"stale"`
	Source     string     `json:"source"`
	Items      []Order    `json:"items"`
}

type Order struct {
	ID                 string    `json:"id"`
	AccountLabel       string    `json:"accountLabel"`
	Market             string    `json:"market"`
	Symbol             string    `json:"symbol"`
	Side               string    `json:"side"`
	Kind               string    `json:"kind"`
	Status             string    `json:"status"`
	Currency           string    `json:"currency"`
	Quantity           string    `json:"quantity"`
	Price              string    `json:"price"`
	FilledQuantity     string    `json:"filledQuantity"`
	AverageFilledPrice string    `json:"averageFilledPrice"`
	OrderedAt          string    `json:"orderedAt"`
	CanceledAt         string    `json:"canceledAt"`
	Origin             string    `json:"origin"`
	ExitLine           *ExitLine `json:"exitLine"`
}

type CandidatesResource struct {
	ObservedAt *time.Time  `json:"observedAt"`
	Stale      bool        `json:"stale"`
	Source     string      `json:"source"`
	Items      []Candidate `json:"items"`
}

type Candidate struct {
	Market           string     `json:"market"`
	Symbol           string     `json:"symbol"`
	Name             string     `json:"name"`
	Verdict          string     `json:"verdict"`
	ReasonCodes      []string   `json:"reasonCodes"`
	Rank             int        `json:"rank"`
	FirstSeenAt      *time.Time `json:"firstSeenAt"`
	LastObservedAt   *time.Time `json:"lastObservedAt"`
	EvidenceDigest   string     `json:"evidenceDigest"`
	ThresholdVersion string     `json:"thresholdVersion"`
}

type SettingsResource struct {
	Version          uint64    `json:"version"`
	EffectiveVersion uint64    `json:"effectiveVersion"`
	Items            []Setting `json:"items"`
}

type Setting struct {
	Key         string                      `json:"key"`
	Label       string                      `json:"label"`
	Description string                      `json:"description"`
	Unit        string                      `json:"unit"`
	Default     State                       `json:"default"`
	Desired     State                       `json:"desired"`
	Effective   State                       `json:"effective"`
	ApplyTiming settingmeta.ApplyTiming     `json:"applyTiming"`
	Safety      settingmeta.SafetyDirection `json:"safetyDirection"`
	Provenance  Provenance                  `json:"provenance"`
}

type State struct {
	Kind     string `json:"kind"`
	OptionID string `json:"optionId"`
	Value    string `json:"value"`
	Display  string `json:"display"`
}

type Provenance struct {
	OwnerChange    string `json:"ownerChange"`
	PolicyID       string `json:"policyId"`
	PolicyVersion  string `json:"policyVersion"`
	PolicyDigest   string `json:"policyDigest"`
	EvidenceDigest string `json:"evidenceDigest"`
}
