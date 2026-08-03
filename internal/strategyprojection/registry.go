package strategyprojection

type FieldDescriptor struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	JSONPointer string `json:"jsonPointer"`
}

var goldenRegistry = [...]FieldDescriptor{
	{Key: "market", Label: "시장", JSONPointer: "/market"},
	{Key: "status", Label: "시장 상태", JSONPointer: "/status"},
	{Key: "lane", Label: "Lane desired/effective", JSONPointer: "/lane"},
	{Key: "evidence", Label: "Evidence ID/digest/freshness", JSONPointer: "/evidence"},
	{Key: "campaign", Label: "Campaign/leg", JSONPointer: "/campaign"},
	{Key: "horizonRisk", Label: "Horizon risk", JSONPointer: "/horizonRisk"},
	{Key: "scheduler", Label: "Scheduler/calendar", JSONPointer: "/scheduler"},
	{Key: "activation", Label: "Activation", JSONPointer: "/activation"},
	{Key: "protection", Label: "ProtectionReady", JSONPointer: "/protection"},
	{Key: "reconciliation", Label: "Reconciliation", JSONPointer: "/reconciliation"},
	{Key: "firstRefusal", Label: "First typed refusal", JSONPointer: "/firstRefusal"},
	{Key: "observedAt", Label: "Observed at", JSONPointer: "/observedAt"},
}

func Registry() []FieldDescriptor {
	return append([]FieldDescriptor(nil), goldenRegistry[:]...)
}
