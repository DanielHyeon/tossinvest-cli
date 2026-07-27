package verifylive

// korean_test.go is the drift guard task 1.8 ③ asks for.
//
// The failure it exists to prevent is quiet: a step, verdict or mutation class
// added to the catalogue without a Korean label renders as a bare English
// identifier on the screen where a person decides whether to send live orders. That
// is not a crash and no other test would notice it.

import (
	"strings"
	"testing"
	"unicode"
)

// hasHangul reports that a string contains Korean, which is how the test tells a
// label from an identifier that was left as-is.
func hasHangul(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

// TestEveryCatalogueStepHasAKoreanLabel.
func TestEveryCatalogueStepHasAKoreanLabel(t *testing.T) {
	labels := StepLabels()
	for _, step := range Steps() {
		label, ok := labels[step.ID]
		switch {
		case !ok:
			t.Errorf("step %s has no Korean label; it would render as a bare identifier on the approval "+
				"screen", step.ID)
		case strings.TrimSpace(label) == "":
			t.Errorf("step %s has an empty Korean label", step.ID)
		case !hasHangul(label):
			t.Errorf("step %s is labelled %q, which contains no Korean", step.ID, label)
		}
		if StepLabel(step.ID) != label {
			t.Errorf("StepLabel(%s) and the map disagree", step.ID)
		}
	}

	// And nothing is labelled that is not in the catalogue: a leftover entry for a
	// removed step is a label nobody will notice is wrong.
	catalogue := map[StepID]bool{}
	for _, step := range Steps() {
		catalogue[step.ID] = true
	}
	for id := range labels {
		if !catalogue[id] {
			t.Errorf("%s is labelled but is not a step in the catalogue", id)
		}
	}
}

// TestTheEnglishTitleIsUntouched.
//
// Step.Title is copied verbatim into the record's `title` field, and records
// already written are in English. The Manager's ruling for 1.8: comparability with
// those records outranks the display language, so the Korean lives in the label map
// and Title stays as it is.
func TestTheEnglishTitleIsUntouched(t *testing.T) {
	for _, step := range Steps() {
		if strings.TrimSpace(step.Title) == "" {
			t.Errorf("step %s has no title", step.ID)
			continue
		}
		if hasHangul(step.Title) {
			t.Errorf("step %s has a Korean Title (%q). Title becomes the record's `title` field and the "+
				"records already on disk are in English; the screen label is StepLabel", step.ID, step.Title)
		}
	}
}

// TestEveryVerdictIsGlossedAndStillShowsItsRecordedValue.
func TestEveryVerdictIsGlossedAndStillShowsItsRecordedValue(t *testing.T) {
	for _, v := range Verdicts() {
		rendered := VerdictLabel(v)
		if !strings.HasPrefix(rendered, string(v)) {
			t.Errorf("VerdictLabel(%s) = %q; the recorded value has to lead, because that is what is in the "+
				"file and in a report", v, rendered)
		}
		if !hasHangul(rendered) {
			t.Errorf("verdict %s has no Korean gloss: %q", v, rendered)
		}
	}
	if got := VerdictLabel(""); got != "-" {
		t.Errorf("VerdictLabel(\"\") = %q, want a dash", got)
	}
	if got := VerdictLabel("something-new"); got != "something-new" {
		t.Errorf("an unknown verdict rendered as %q; it must show itself rather than vanish", got)
	}
}

// TestEveryMutationClassHasAKoreanVerb. The approval summary's action column.
func TestEveryMutationClassHasAKoreanVerb(t *testing.T) {
	for _, k := range MutationKinds() {
		if !hasHangul(k.VerbKO()) {
			t.Errorf("mutation class %s has no Korean verb: %q", k, k.VerbKO())
		}
		if strings.TrimSpace(k.Verb()) == "" {
			t.Errorf("mutation class %s lost its English verb", k)
		}
	}
	// And the catalogue declares nothing this list has not seen.
	known := map[MutationKind]bool{}
	for _, k := range MutationKinds() {
		known[k] = true
	}
	for _, step := range Steps() {
		for _, m := range step.Mutations {
			if !known[m.Kind] {
				t.Errorf("step %s declares mutation class %s, which MutationKinds() does not list — so it "+
					"has no Korean verb and nothing noticed", step.ID, m.Kind)
			}
		}
	}
}

// TestTheBrokerVocabularyIsNotTranslated.
//
// The evidence has to be checkable against the broker's own answers. A screen that
// said "주문가능시간 아님" where the API said `order-hours-closed` would make the
// record and the page two different claims.
func TestTheBrokerVocabularyIsNotTranslated(t *testing.T) {
	verbatim := []string{
		"clientOrderId", "sellableQuantity", "triggeredOrderId",
		"POST /api/v1/orders", "GET /api/v1/holdings", "WATCHING", "SINGLE",
	}
	var everything strings.Builder
	for _, step := range Steps() {
		everything.WriteString(step.Proves)
		everything.WriteString("\n")
		for _, line := range step.Procedure {
			everything.WriteString(line)
			everything.WriteString("\n")
		}
	}
	body := everything.String()
	for _, want := range verbatim {
		if !strings.Contains(body, want) {
			t.Errorf("the catalogue no longer names %q anywhere; the broker's own vocabulary is what makes "+
				"the evidence checkable", want)
		}
	}
}

// TestTheDisplayLanguageCannotMoveThePlanDigest.
//
// approval.plan_digest is the record's proof that a person was shown a list of
// exactly this shape, and the operator has records on disk carrying digests
// computed by earlier builds (measurements.md M3: the same digest across three
// processes). The Korean is therefore carried alongside the English rather than
// instead of it, on json:"-" fields — and this is the test that says so, because
// dropping the tag would be a one-character change with no other symptom.
func TestTheDisplayLanguageCannotMoveThePlanDigest(t *testing.T) {
	bare := []PlannedMutation{{
		Ordinal: 1, Step: StepOrderCancel, Kind: MutatePlaceOrder,
		Symbol: "005930", Side: "buy", Quantity: "1 share", MaxQuantity: 1,
		Pricing: PriceFarBuy.Describe(DefaultOffset, MarketKR),
		Ends:    "cancelled inside this step",
		Note:    "a note",
	}}
	translated := append([]PlannedMutation(nil), bare...)
	translated[0].EndsKO = "이 단계 안에서 취소된다"
	translated[0].NoteKO = "비고"
	translated[0].PricingKO = PriceFarBuy.DescribeKO(DefaultOffset, MarketKR)

	before := Plan{Mutations: bare}.Digest()
	after := Plan{Mutations: translated}.Digest()
	if before != after {
		t.Fatalf("the Korean moved the plan digest (%s -> %s). Records already written carry digests from "+
			"builds that had no translation; the display language must not be hashed", before, after)
	}
	if before == "sha256:unencodable" {
		t.Fatal("the digest could not be computed at all")
	}

	// And the two really are different renderings of the same line.
	if translated[0].HeadlineKO() == translated[0].Headline() {
		t.Error("HeadlineKO and Headline produced the same text; nothing is being translated")
	}
	for _, part := range []string{"BUY", "005930", "KR"} {
		if !strings.Contains(translated[0].HeadlineKO(), part) {
			t.Errorf("the Korean headline dropped %q — the request's identity is not translatable", part)
		}
	}
}

// TestTheApprovalSummaryIsRenderedInKorean walks the real catalogue rather than a
// fixture, so a step whose Ends never got a translation shows up here.
func TestTheApprovalSummaryIsRenderedInKorean(t *testing.T) {
	for _, step := range Steps() {
		for i, m := range step.Mutations {
			if strings.TrimSpace(m.EndsKO) == "" {
				t.Errorf("%s mutation %d (%s) has no Korean for how its exposure ends; the approval "+
					"summary would show English on the line that matters most", step.ID, i, m.Kind)
			}
			if strings.TrimSpace(m.Note) != "" && strings.TrimSpace(m.NoteKO) == "" {
				t.Errorf("%s mutation %d (%s) has a note with no translation", step.ID, i, m.Kind)
			}
			// And the English is still there: it is what the digest is computed over.
			if strings.TrimSpace(m.Ends) == "" {
				t.Errorf("%s mutation %d (%s) lost its English Ends, which the plan digest is hashed from",
					step.ID, i, m.Kind)
			}
		}
	}
	for _, b := range []PricingBasis{PriceFarBuy, PriceFarSell, PriceFarStop, PriceOneTickFurther, PriceIdenticalBody} {
		if !hasHangul(b.DescribeKO(DefaultOffset, MarketKR)) {
			t.Errorf("pricing basis %s has no Korean description", b)
		}
		if strings.TrimSpace(b.Describe(DefaultOffset, MarketKR)) == "" {
			t.Errorf("pricing basis %s lost its English description, which the digest is hashed from", b)
		}
	}
	if got := PriceNone.DescribeKO(DefaultOffset, MarketKR); got != "" {
		t.Errorf("PriceNone describes itself as %q; a request with no price must say nothing", got)
	}
}
