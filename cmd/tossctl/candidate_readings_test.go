package main

// candidate_readings_test.go is the scan report's two new blocks: what each source
// asked for beside what arrived, and the first-sighting record split by the source
// that produced it.
//
// # What they are for
//
// Nobody has measured whether the official rankings endpoint returns the hundred
// rows it is asked for. If it does not, every reading is truncated, every first
// sighting from it is refused, and the distribution a seen_late threshold would be
// chosen from does not exist — a state that until now was visible only as a screen
// full of READING_TRUNCATED, several candidate lives after the fact, with no way to
// tell which of the four ranking sources produced it.
//
// The requested-versus-arrived line answers it in one scan. The per-source block
// answers the same question from the other end, and it is the one that also survives
// onto /signals, where no scan record exists.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
)

// aShortScan is one cycle in which the official ranking asked for a hundred rows and
// received three, and the WTS list got exactly what it asked for.
func aShortScan() candidate.CycleResult {
	res := candidate.CycleResult{Market: candidate.MarketKR, At: reportAt}
	res.Scan.Market = candidate.MarketKR
	res.Scan.At = reportAt
	res.Scan.Attempted, res.Scan.Responded = 2, 2
	res.Scan.Readings = map[candidate.SourceID]candidate.ReadingFacts{
		candidate.SourceOfficialTradingValue: {Requested: 100, Arrived: 3},
		candidate.SourceWTSPopular:           {Requested: 30, Arrived: 30},
	}
	res.Sightings = []candidate.SourceSightings{
		{
			Source: candidate.SourceOfficialTradingValue, Total: 3, Measured: 0,
			NotMeasured: map[candidate.VetoUnmeasured]int{candidate.VetoReadingTruncated: 3},
		},
		{
			Source: candidate.SourceWTSPopular, Total: 2, Measured: 2,
			NotMeasured: map[candidate.VetoUnmeasured]int{},
		},
	}
	return res
}

// TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived.
func TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived(t *testing.T) {
	table := renderReport(t, output.FormatTable, aShortScan())

	if !strings.Contains(table, "100 requested, 3 arrived (short)") {
		t.Errorf("the report does not say that the official ranking came back short. That "+
			"is the open question this change leaves open — if the endpoint never returns "+
			"a hundred rows then seen_late has no distribution at all — and it should take "+
			"one scan to answer, not a screen of refusals:\n%s", table)
	}
	if !strings.Contains(table, "30 requested, 30 arrived (whole)") {
		t.Errorf("a complete reading is not reported beside the short one; without it the "+
			"line above is a warning with no baseline:\n%s", table)
	}
	// Both numbers, not the comparison. A hundred requested and ninety-nine arriving
	// and a hundred requested and three arriving are the same one bit and two very
	// different conversations to have with the API.
	if strings.Contains(table, "truncated: true") {
		t.Errorf("the report renders the comparison in place of the counts:\n%s", table)
	}
}

// TestTheScanReportAttributesTheRefusalsToASource.
func TestTheScanReportAttributesTheRefusalsToASource(t *testing.T) {
	table := renderReport(t, output.FormatTable, aShortScan())
	block := table[strings.Index(table, "first sightings by source"):]
	if !strings.Contains(block, string(candidate.SourceOfficialTradingValue)) ||
		!strings.Contains(block, "0 of 3 measured") {
		t.Errorf("the block does not attribute the three refusals to the source whose "+
			"readings produced them:\n%s", block)
	}
	if !strings.Contains(block, string(candidate.VetoReadingTruncated)+" 3") {
		t.Errorf("the reason is not counted beside the source:\n%s", block)
	}
	if !strings.Contains(block, "2 of 2 measured") {
		t.Errorf("the source that is working is missing, so the block reads as a list of "+
			"faults rather than as a census:\n%s", block)
	}
}

// TestTheJSONReportCarriesBothBlocks. The scan output is a machine surface as well,
// and a measurement that only exists in the table is one nothing can aggregate.
func TestTheJSONReportCarriesBothBlocks(t *testing.T) {
	var got struct {
		Readings []struct {
			Source    string `json:"source"`
			Requested int    `json:"requested"`
			Arrived   int    `json:"arrived"`
			Whole     bool   `json:"whole"`
		} `json:"readings"`
		Sightings []struct {
			Source      string         `json:"source"`
			Total       int            `json:"total"`
			Measured    int            `json:"measured"`
			NotMeasured map[string]int `json:"not_measured"`
		} `json:"sightings"`
	}
	if err := json.Unmarshal([]byte(renderReport(t, output.FormatJSON, aShortScan())), &got); err != nil {
		t.Fatalf("decoding the JSON report: %v", err)
	}
	if len(got.Readings) != 2 || got.Readings[0].Requested != 100 || got.Readings[0].Arrived != 3 ||
		got.Readings[0].Whole {
		t.Errorf("readings = %+v; want the official ranking's 100/3 first and not whole",
			got.Readings)
	}
	if len(got.Sightings) != 2 ||
		got.Sightings[0].NotMeasured[string(candidate.VetoReadingTruncated)] != 3 {
		t.Errorf("sightings = %+v", got.Sightings)
	}
}

// TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition.
//
// `tossctl candidate scan` builds its panel, reads once and exits, so every reading
// it takes is a source's first and none of them can qualify a position. The scan no
// longer stores those — a stored one is permanent, an unstored one is recoverable —
// and the count is what makes the difference visible instead of leaving `0 first
// ranks` to read as "there was nothing to record".
func TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition(t *testing.T) {
	res := aShortScan()
	res.Scan.Candidates = 40
	res.Scan.FirstRanksHeld = 40
	table := renderReport(t, output.FormatTable, res)
	if !strings.Contains(table, "40 position(s) not stored") {
		t.Errorf("the held count is not on the report:\n%s", table)
	}
	if !strings.Contains(table, "written once") {
		t.Errorf("the report says a number and not why it matters; write-once is the whole "+
			"reason declining to store is the safe direction:\n%s", table)
	}

	// And a scan with nothing held says nothing, so the line means something when it
	// appears.
	quiet := aShortScan()
	if got := renderReport(t, output.FormatTable, quiet); strings.Contains(got, "not stored") {
		t.Errorf("the held line is drawn for a scan that held nothing:\n%s", got)
	}
}
