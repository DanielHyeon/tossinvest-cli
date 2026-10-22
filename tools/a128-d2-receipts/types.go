package main

// types.go 는 D2 유형 영수증 analysis/d2-receipts/security-types.json 을 만듦(사전등록 7절, tasks 6.6).
//
// 출처: 공식 GET /api/v1/stocks 의 securityType 원문(a169 가 domain.Quote.SecurityType 으로 노출). 갈래 대응
// (STOCK/ETF/ETN/기타/미분류)은 평가기 몫이라 여기서는 원문 그대로 적음. 값이 없거나 응답에서 빠진 종목은 null
// = 평가기에서 「미분류」.
//
// 파일 하나에 날짜마다 덧붙임: 이미 적힌 종목(null 포함)은 다시 묻지 않고 바꾸지 않음 — 호출을 줄이고 앞선 영수증을
// 보존하려는 것. 실행마다 runs 에 무엇을 몇 번 물었는지 남김. 429 는 재시도 없이 멈추고 파일을 바꾸지 않음.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const stocksBatchMax = 200 // GET /api/v1/stocks 의 symbols 상한(문서값)

type stockSource interface {
	Stocks(ctx context.Context, symbols []string) ([]domain.Quote, error)
}

type typesRun struct {
	Market         string   `json:"market"`
	Date           string   `json:"date"`
	FetchedAt      string   `json:"fetched_at"`
	ElapsedSeconds float64  `json:"elapsed_seconds"`
	Calls          int      `json:"calls"`
	Queried        []string `json:"queried"`
	Unclassified   []string `json:"unclassified"` // 응답에서 빠졌거나 securityType 이 빈 종목(둘의 합)
	NotInResponse  []string `json:"not_in_response"`
	EmptyType      []string `json:"empty_type"`
}

// typesReceipt 는 security-types.json 내용. 평가기는 types 만 읽음(값이 문자열인 것만 — null 은 미분류).
type typesReceipt struct {
	Prereg string             `json:"prereg"`
	Source string             `json:"source"`
	Types  map[string]*string `json:"types"`
	Runs   []typesRun         `json:"runs"`
}

// readTypes 는 기존 영수증을 읽음. 없으면 빈 영수증.
func readTypes(path string) (typesReceipt, error) {
	rc := typesReceipt{Types: map[string]*string{}}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return rc, nil
	}
	if err != nil {
		return rc, err
	}
	if err := json.Unmarshal(raw, &rc); err != nil {
		return rc, fmt.Errorf("%s: %w", path, err)
	}
	if rc.Types == nil {
		rc.Types = map[string]*string{}
	}
	return rc, nil
}

// collectTypes 는 known 에 없는 종목만 200 개씩 묶어 물음. 반환한 map 은 새로 물은 종목 전부를 담음(값 없음 = nil).
// 429 는 즉시 errRateLimited, 그 밖의 오류도 멈춤(재시도 없음).
func collectTypes(ctx context.Context, src stockSource, symbols []string, known map[string]*string,
	now func() time.Time) (map[string]*string, typesRun, error) {
	started := now()
	var ask []string
	for _, s := range symbols {
		if _, ok := known[s]; !ok {
			ask = append(ask, s)
		}
	}
	got := map[string]*string{}
	run := typesRun{Queried: ask, Unclassified: []string{}, NotInResponse: []string{}, EmptyType: []string{}}
	seen := map[string]bool{}
	// shortcut: 묶음 사이 쉼 없음 — 목록은 하루 30 종목이라 묶음 1 개; 200 을 넘는 목록이 생기면 --gap 을 여기에 적용.
	for i := 0; i < len(ask); i += stocksBatchMax {
		batch := ask[i:min(i+stocksBatchMax, len(ask))]
		run.Calls++
		quotes, err := src.Stocks(ctx, batch)
		if errors.Is(err, official.ErrRateLimited) {
			return nil, typesRun{}, fmt.Errorf("%w (stocks call %d)", errRateLimited, run.Calls)
		}
		if err != nil {
			return nil, typesRun{}, fmt.Errorf("stocks call %d: %w", run.Calls, err)
		}
		for _, q := range quotes {
			seen[q.Symbol] = true
			if q.SecurityType != "" {
				v := q.SecurityType
				got[q.Symbol] = &v
			}
		}
	}
	for _, s := range ask {
		if got[s] == nil {
			got[s] = nil // 응답에서 빠짐 또는 securityType 빈 값 → 미분류(이후 재질의 안 함 — 영수증 불변)
			run.Unclassified = append(run.Unclassified, s)
			if seen[s] {
				run.EmptyType = append(run.EmptyType, s)
			} else {
				run.NotInResponse = append(run.NotInResponse, s)
			}
		}
	}
	end := now()
	run.FetchedAt = end.UTC().Format(time.RFC3339)
	run.ElapsedSeconds = end.Sub(started).Seconds()
	return got, run, nil
}

// writeTypes 는 임시 파일에 쓴 뒤 rename 함(쓰다 끊겨도 앞선 영수증이 반쯤 덮이지 않게).
func writeTypes(path string, rc typesReceipt) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(rc, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".security-types-*.json")
	if err != nil {
		return err
	}
	_, err = tmp.Write(append(body, '\n'))
	if err == nil {
		err = tmp.Sync() // rename 전에 디스크로 — 크래시 때 빈 파일이 영수증을 덮지 않게
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		os.Remove(tmp.Name())
	}
	return err
}

// updateTypes 는 읽기 → 새 종목만 묻기 → 합치기 → 쓰기. 앞선 값은 바꾸지 않음.
// shortcut: 파일 잠금 없음 — 실행은 Manager 지시로 한 번에 하나(tasks 6.7); 자동 타이머를 붙이면 flock 을 더할 것.
func updateTypes(ctx context.Context, src stockSource, path, market, date string, symbols []string,
	now func() time.Time) (typesRun, error) {
	rc, err := readTypes(path)
	if err != nil {
		return typesRun{}, err
	}
	got, run, err := collectTypes(ctx, src, symbols, rc.Types, now)
	if err != nil {
		return typesRun{}, err
	}
	for s, v := range got {
		rc.Types[s] = v
	}
	run.Market, run.Date = market, date
	sort.Strings(run.Unclassified)
	rc.Runs = append(rc.Runs, run)
	rc.Prereg = preregDoc
	rc.Source = "official GET /api/v1/stocks securityType (raw enum; null = missing from response or empty → 미분류)"
	return run, writeTypes(path, rc)
}
