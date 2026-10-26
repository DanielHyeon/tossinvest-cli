package console

import (
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

type optimizationPresetMatrix struct {
	Policies []optimizationPresetMatrixPolicy
	Rows     []optimizationPresetMatrixRow
}

type optimizationPresetMatrixPolicy struct {
	ID, Label, Summary string
	Recommended        bool
	Selected           bool
}

type optimizationPresetMatrixRow struct {
	Label string
	Cells []optimizationPresetMatrixCell
}

type optimizationPresetMatrixCell struct {
	Target, Stop, Partial string
}

func (p optimizationPage) PresetMatrix() optimizationPresetMatrix {
	matrix := optimizationPresetMatrix{Policies: make([]optimizationPresetMatrixPolicy, 0, len(p.Policies))}
	maxRungs := 0
	for _, policy := range p.Policies {
		matrix.Policies = append(matrix.Policies, optimizationPresetMatrixPolicy{
			ID: policy.ID, Label: policy.Label, Summary: policy.Summary,
			Recommended: policy.Recommended, Selected: policy.Selected,
		})
		if len(policy.Ladder.Rungs) > maxRungs {
			maxRungs = len(policy.Ladder.Rungs)
		}
	}
	for index := range maxRungs {
		row := optimizationPresetMatrixRow{Label: "T" + strconv.Itoa(index+1), Cells: make([]optimizationPresetMatrixCell, 0, len(p.Policies))}
		for _, policy := range p.Policies {
			cell := optimizationPresetMatrixCell{Target: "—", Stop: "—", Partial: "—"}
			if index < len(policy.Ladder.Rungs) {
				rung := policy.Ladder.Rungs[index]
				cell = optimizationPresetMatrixCell{
					Target: optimizationPresetTarget(policy.ID, rung.TargetPct),
					Stop:   rung.StopPct + "%", Partial: rung.PartialRatio,
				}
			}
			row.Cells = append(row.Cells, cell)
		}
		matrix.Rows = append(matrix.Rows, row)
	}
	return matrix
}

func optimizationPresetTarget(policyID, target string) string {
	if policyID == "COMMON_LADDER_RUNNER" && target == "999.0" {
		return "고정 목표 없음"
	}
	return target + "%"
}

type candidateFilterMatrix struct {
	Rows []candidateFilterMatrixRow
}

type candidateFilterMatrixRow struct {
	Key, Label string
	Common     bool
	KR, US     candidateFilterMatrixCell
}

type candidateFilterMatrixCell struct {
	Market, Session, Help, DefaultState, DesiredState, EffectiveState string
	Unit, ValidRange, Direction, SampleState, EvidenceState           string
	ApplyTiming, LegacyValue, Provenance, PreviewContract             string
	MissingEvidence                                                   []string
}

func (p optimizationPage) CandidateFilterMatrix() candidateFilterMatrix {
	return candidateFilterMatrixFromMarkets(candidate.CandidateFilterMarkets())
}

func candidateFilterMatrixFromMarkets(markets []candidate.CandidateFilterMarket) candidateFilterMatrix {
	byMarket := make(map[string]map[string]candidate.CandidateFilterDescriptor, len(markets))
	var keys []string
	for _, market := range markets {
		byKey := make(map[string]candidate.CandidateFilterDescriptor, len(market.Filters))
		for _, descriptor := range market.Filters {
			byKey[descriptor.Key] = descriptor
			if market.Market == candidate.MarketKR {
				keys = append(keys, descriptor.Key)
			}
		}
		byMarket[market.Market] = byKey
	}

	matrix := candidateFilterMatrix{Rows: make([]candidateFilterMatrixRow, 0, len(keys))}
	for _, key := range keys {
		kr := byMarket[candidate.MarketKR][key]
		us := byMarket[candidate.MarketUS][key]
		row := candidateFilterMatrixRow{Key: key, Label: kr.Label, KR: candidateFilterMatrixCellFrom(kr), US: candidateFilterMatrixCellFrom(us)}
		row.Common = candidateFilterDescriptorEqual(kr, us)
		if row.Common {
			row.KR.Market = "KR · US 공통"
		}
		matrix.Rows = append(matrix.Rows, row)
	}
	return matrix
}

func candidateFilterMatrixCellFrom(descriptor candidate.CandidateFilterDescriptor) candidateFilterMatrixCell {
	return candidateFilterMatrixCell{
		Market: descriptor.Market, Session: descriptor.Session, Help: descriptor.Help,
		DefaultState: string(descriptor.DefaultState), DesiredState: string(descriptor.DesiredState),
		EffectiveState: string(descriptor.EffectiveState), Unit: descriptor.Unit, ValidRange: descriptor.ValidRange,
		Direction: descriptor.Direction, SampleState: descriptor.SampleState, EvidenceState: descriptor.EvidenceState,
		ApplyTiming: descriptor.ApplyTiming, LegacyValue: descriptor.LegacyValue, Provenance: descriptor.Provenance,
		PreviewContract: descriptor.PreviewContract, MissingEvidence: append([]string(nil), descriptor.MissingEvidence...),
	}
}

func candidateFilterDescriptorEqual(kr, us candidate.CandidateFilterDescriptor) bool {
	if kr.Key != us.Key || kr.Category != us.Category || kr.Label != us.Label || kr.Help != us.Help ||
		kr.Session != us.Session || kr.DefaultState != us.DefaultState || kr.DesiredState != us.DesiredState ||
		kr.EffectiveState != us.EffectiveState || kr.DesiredValue != us.DesiredValue ||
		kr.EffectiveValue != us.EffectiveValue || kr.Unit != us.Unit || kr.ValidRange != us.ValidRange ||
		kr.Direction != us.Direction || kr.SampleState != us.SampleState || kr.EvidenceState != us.EvidenceState ||
		kr.ApplyTiming != us.ApplyTiming || kr.ReadOnly != us.ReadOnly || kr.CASRequired != us.CASRequired ||
		kr.Provenance != us.Provenance || kr.LegacyValue != us.LegacyValue || kr.PreviewContract != us.PreviewContract ||
		len(kr.MissingEvidence) != len(us.MissingEvidence) {
		return false
	}
	for index := range kr.MissingEvidence {
		if kr.MissingEvidence[index] != us.MissingEvidence[index] {
			return false
		}
	}
	return true
}
