package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const initialProfile = "initial-fields-v1"

// initialState 는 초기 입력 종류만 독립된 값으로 투영함.
func initialState(source PublicState) PublicState {
	s := PublicState{Market: source.Market, Symbol: source.Symbol, Name: source.Name, ListingMarket: source.ListingMarket, Currency: source.Currency, ObservedAt: source.ObservedAt, HorizonMinutes: source.HorizonMinutes, BasePrice: source.BasePrice, ReferencePrice: source.ReferencePrice, RankingType: source.RankingType, RankingRank: source.RankingRank, TradingAmount: source.TradingAmount, TradingVolume: source.TradingVolume, MinutesSinceOpen: source.MinutesSinceOpen, MinutesToClose: source.MinutesToClose, SessionSource: source.SessionSource, News: newsAbsent}
	if source.ChangePctVsReference != nil {
		v := *source.ChangePctVsReference
		s.ChangePctVsReference = &v
	}
	if source.TopOfBook != nil {
		v := *source.TopOfBook
		s.TopOfBook = &v
	}
	for _, gap := range source.MissingInputs {
		if strings.HasPrefix(gap, "stocks_meta:") || strings.HasPrefix(gap, "top_of_book:") {
			s.MissingInputs = append(s.MissingInputs, gap)
		}
	}
	s.MissingInputs = append(s.MissingInputs, "news: intentionally omitted by initial-fields-v1")
	return s
}

// validateProfileOutput 는 실험 자료가 운영 원장에 섞이지 않도록 먼저 경로를 검사함.
func validateProfileOutput(profile, out string, explicit bool) error {
	if profile != "current" && profile != "initial" && profile != "compare" {
		return fmt.Errorf("invalid input profile %q", profile)
	}
	if profile == "current" {
		if explicit {
			contents, err := readLedger(filepath.Join(out, ledgerFileName))
			if err != nil {
				return err
			}
			for _, row := range contents.Judgments {
				if row.Cohort != "" {
					return errors.New("current output contains experimental judgments")
				}
			}
			for _, row := range contents.Gaps {
				if row.Cohort != "" {
					return errors.New("current output contains experimental gaps")
				}
			}
		}
		return nil
	}
	if !explicit || out == "" {
		return errors.New("initial and compare require an explicit --out")
	}
	path, err := resolvedOutputPath(out)
	if err != nil {
		return err
	}
	base, err := resolvedOutputPath(defaultLedgerDir)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return err
	}
	if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return errors.New("experimental output must be outside the default ledger directory")
	}
	contents, err := readLedger(filepath.Join(path, ledgerFileName))
	if err != nil {
		return fmt.Errorf("checking experimental ledger: %w", err)
	}
	for _, row := range contents.Judgments {
		if row.InputProfile == "" || row.Cohort == "" {
			return errors.New("experimental output contains legacy judgments")
		}
	}
	for _, row := range contents.Gaps {
		if row.Cohort == "" {
			return errors.New("experimental output contains legacy gaps")
		}
	}
	if len(contents.Judgments) == 0 && len(contents.Labels) > 0 {
		return errors.New("experimental output contains unassociated labels")
	}
	return nil
}

func resolvedOutputPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	suffix := []string{}
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if info, lerr := os.Lstat(path); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("unresolved output symlink %s", path)
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		suffix = append(suffix, filepath.Base(path))
		path = parent
	}
}

// resolveProfileOutput 은 자동 실행을 입력군별 별도 원장으로 보냄.
func resolveProfileOutput(profile, out string, explicit bool) (string, error) {
	if !explicit {
		switch profile {
		case "current":
			out = defaultLedgerDir
		case "initial":
			out = defaultLedgerDir + "-initial"
		case "compare":
			out = defaultLedgerDir + "-compare"
		default:
			return "", fmt.Errorf("invalid input profile %q", profile)
		}
	}
	if err := validateProfileOutput(profile, out, true); err != nil {
		return "", err
	}
	return out, nil
}
