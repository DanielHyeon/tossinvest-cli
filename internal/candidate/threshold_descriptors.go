package candidate

type ThresholdState string

const ThresholdStateUnapproved ThresholdState = "unapproved"

type CandidateFilterDescriptor struct {
	Key             string
	Category        string
	Label           string
	Help            string
	Market          string
	Session         string
	DefaultState    ThresholdState
	DesiredState    ThresholdState
	EffectiveState  ThresholdState
	DesiredValue    string
	EffectiveValue  string
	Unit            string
	ValidRange      string
	Direction       string
	SampleState     string
	EvidenceState   string
	ApplyTiming     string
	ReadOnly        bool
	CASRequired     bool
	Provenance      string
	LegacyValue     string
	MissingEvidence []string
	PreviewContract string
}

type CandidateFilterMarket struct {
	Market  string
	Session string
	Filters []CandidateFilterDescriptor
}

// CandidateFilterDescriptors is the dormant a046 registry read model. There is
// deliberately no numeric option and no writer: a human evidence activation
// record must add an immutable option before any desired/effective value exists.
func CandidateFilterDescriptors() []CandidateFilterDescriptor {
	definitions := []CandidateFilterDescriptor{
		{
			Key: string(VetoSeenLate), Label: "첫 발견이 너무 늦음",
			Help: "처음 발견했을 때 이미 순위 상단에 있었는지 판정한다.",
			Unit: "percentile points", ValidRange: "0 초과 100 이하", Direction: "> 이면 거부",
		},
		{
			Key: string(VetoExtended), Label: "첫 가격 대비 과도한 상승",
			Help: "저장된 첫 가격 이후 이미 너무 많이 상승했는지 판정한다.",
			Unit: "%", ValidRange: "0 초과", Direction: "> 이면 거부",
		},
		{
			Key: string(VetoNearHigh), Label: "당일 고점에 너무 가까움",
			Help: "현재가와 당일 고점 사이의 남은 거리가 너무 작은지 판정한다.",
			Unit: "%", ValidRange: "0 초과 100 이하", Direction: "< 이면 거부",
			LegacyValue: LegacyNearHighThresholdPct, Provenance: "legacy-unapproved",
		},
	}
	markets := [...]string{MarketKR, MarketUS}
	out := make([]CandidateFilterDescriptor, 0, len(markets)*len(definitions))
	for _, market := range markets {
		for _, definition := range definitions {
			descriptor := definition
			descriptor.Category = "candidate-filters"
			descriptor.Market = market
			descriptor.Session = SessionRegular
			descriptor.DefaultState = ThresholdStateUnapproved
			descriptor.DesiredState = ThresholdStateUnapproved
			descriptor.EffectiveState = ThresholdStateUnapproved
			descriptor.SampleState = "not_measured"
			descriptor.EvidenceState = "not_measured"
			descriptor.ApplyTiming = "다음 candidate 평가부터"
			descriptor.ReadOnly = true
			descriptor.CASRequired = true
			descriptor.MissingEvidence = []string{
				"sample_count", "missing_rate", "evidence_digest", "human_activation_record",
			}
			descriptor.PreviewContract = "registry option 선택 후 before/after, 시장·세션, " +
				"예상 verdict count, evidence version, 적용 시점을 preview하고 base-version CAS를 확인한다"
			out = append(out, descriptor)
		}
	}
	return out
}

func CandidateFilterMarkets() []CandidateFilterMarket {
	markets := []CandidateFilterMarket{
		{Market: MarketKR, Session: SessionRegular},
		{Market: MarketUS, Session: SessionRegular},
	}
	for _, descriptor := range CandidateFilterDescriptors() {
		for index := range markets {
			if markets[index].Market == descriptor.Market && markets[index].Session == descriptor.Session {
				markets[index].Filters = append(markets[index].Filters, descriptor)
				break
			}
		}
	}
	return markets
}
