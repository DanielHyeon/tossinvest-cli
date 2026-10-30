package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// renderProfileComparison 은 정확히 하나씩 존재하는 같은 조건의 판단만 비교함.
func validStateDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
func validComparisonRow(r judgmentRow) bool {
	return r.Cohort != "" && validStateDigest(r.SourceStateDigest) && validStateDigest(r.StateDigest) && r.Market != "" && r.Symbol != "" && r.At != "" && r.QuestionID != "" && r.QuestionRev != "" && r.QuestionDigest != "" && r.ModelRequested != "" && r.ModelAnswered != "" && r.HorizonMinutes > 0 && r.BasePrice > 0
}
func renderProfileComparison(c ledgerContents) string {
	type group struct{ current, initial []judgmentRow }
	groups := map[string]*group{}
	invalid := 0
	for _, row := range c.Judgments {
		if row.InputProfile != "current" && row.InputProfile != "initial" {
			continue
		}
		if !validComparisonRow(row) {
			invalid++
			continue
		}
		key := fmt.Sprintf("%q|%q|%q|%q|%q|%q|%q|%q|%q|%q|%d|%g", row.Cohort, row.SourceStateDigest, row.Market, row.Symbol, row.At, row.QuestionID, row.QuestionRev, row.QuestionDigest, row.ModelRequested, row.ModelAnswered, row.HorizonMinutes, row.BasePrice)
		g := groups[key]
		if g == nil {
			g = &group{}
			groups[key] = g
		}
		if row.InputProfile == "current" {
			g.current = append(g.current, row)
		} else {
			g.initial = append(g.initial, row)
		}
	}
	paired, unpaired, duplicate, crossings, nJ2 := 0, 0, 0, 0, 0
	delta := 0.0
	for _, g := range groups {
		if len(g.current) > 1 || len(g.initial) > 1 {
			duplicate++
			continue
		}
		if len(g.current) != 1 || len(g.initial) != 1 {
			unpaired++
			continue
		}
		paired++
		if g.current[0].QuestionID == screenRedFlagQuestion {
			a, b := g.current[0].P, g.initial[0].P
			delta += b - a
			nJ2++
			if (a >= 0.6) != (b >= 0.6) {
				crossings++
			}
		}
	}
	count := map[string]int{}
	realized, labelGaps, pending := map[string]int{}, map[string]int{}, map[string]int{}
	failures := map[string]int{}
	for _, r := range c.Judgments {
		if r.InputProfile != "" {
			count[r.InputProfile]++
			label, ok := c.Labels[r.ID]
			if !ok {
				pending[r.InputProfile]++
			} else if label.RealizedReturn == nil {
				labelGaps[r.InputProfile]++
			} else {
				realized[r.InputProfile]++
			}
		}
	}
	invalidGaps := 0
	for _, r := range c.Gaps {
		if r.InputProfile != "" {
			failures[r.InputProfile]++
			if !validStateDigest(r.SourceStateDigest) {
				invalidGaps++
			}
		}
	}
	type observation struct{ successful, failed map[string]bool }
	observations := map[string]*observation{}
	add := func(cohort, source, market, symbol, at, profile string, failed bool) {
		if cohort == "" || source == "" || (profile != "current" && profile != "initial") {
			return
		}
		key := fmt.Sprintf("%q|%q|%q|%q|%q", cohort, source, market, symbol, at)
		o := observations[key]
		if o == nil {
			o = &observation{successful: map[string]bool{}, failed: map[string]bool{}}
			observations[key] = o
		}
		if failed {
			o.failed[profile] = true
		} else {
			o.successful[profile] = true
		}
	}
	for _, r := range c.Judgments {
		if validComparisonRow(r) {
			add(r.Cohort, r.SourceStateDigest, r.Market, r.Symbol, r.At, r.InputProfile, false)
		}
	}
	for _, r := range c.Gaps {
		if r.Stage == "judgment" {
			add(r.Cohort, r.SourceStateDigest, r.Market, r.Symbol, r.At, r.InputProfile, true)
		}
	}
	bothFailed, oneFailed := 0, 0
	for _, o := range observations {
		if o.failed["current"] && o.failed["initial"] {
			bothFailed++
		} else if (o.failed["current"] && o.successful["initial"]) || (o.failed["initial"] && o.successful["current"]) {
			oneFailed++
		}
	}
	keys := []string{"current", "initial"}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "# Input profile comparison\n\nPaired question rows: %d; unpaired: %d; duplicate groups: %d.\n", paired, unpaired, duplicate)
	fmt.Fprintf(&b, "Invalid judgment rows: %d; invalid failure sources: %d.\n", invalid, invalidGaps)
	fmt.Fprintf(&b, "Observations with both requests failed: %d; one failed: %d.\n", bothFailed, oneFailed)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: judgments %d; realized %d; label gaps %d; pending %d; failed requests %d.\n", k, count[k], realized[k], labelGaps[k], pending[k], failures[k])
	}
	if nJ2 > 0 {
		fmt.Fprintf(&b, "J2 mean initial-current: %.6f; threshold 0.6 crossings: %d / %d.\n", delta/float64(nJ2), crossings, nJ2)
	} else {
		fmt.Fprintln(&b, "J2 difference: unavailable (no paired J2 rows).")
	}
	return b.String()
}

// verifyComparisonStates 는 보고 전 파일과 payload digest가 일치하지 않는 판단을 제외 대상으로 표시함.
func verifyComparisonStates(c ledgerContents, dir string) ledgerContents {
	cache := map[string]bool{}
	valid := func(digest string) bool {
		if v, ok := cache[digest]; ok {
			return v
		}
		if !validStateDigest(digest) {
			cache[digest] = false
			return false
		}
		raw, err := os.ReadFile(filepath.Join(dir, digest[7:]+".json"))
		if err != nil {
			cache[digest] = false
			return false
		}
		raw = bytes.TrimSuffix(raw, []byte("\n"))
		sum := sha256.Sum256(raw)
		good := hex.EncodeToString(sum[:]) == digest[7:]
		cache[digest] = good
		return good
	}
	c.Gaps = append([]gapRow(nil), c.Gaps...)
	for i := range c.Gaps {
		r := &c.Gaps[i]
		if r.InputProfile != "" && !valid(r.SourceStateDigest) {
			r.SourceStateDigest = ""
		}
	}
	c.Judgments = append([]judgmentRow(nil), c.Judgments...)
	for i := range c.Judgments {
		r := &c.Judgments[i]
		if r.InputProfile != "" && (!valid(r.SourceStateDigest) || !valid(r.StateDigest)) {
			r.StateDigest = ""
		}
	}
	return c
}
