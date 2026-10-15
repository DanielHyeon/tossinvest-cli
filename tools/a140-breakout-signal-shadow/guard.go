package main

// guard.go 는 도구가 엔진의 생산 evidence.db 에 쓰지 못하게 막음(spec 「그림자 신호 측정은 주문 권위가 없다」 · design §2.1-9).
//
// 생산 제안 재생은 저널 옆 evidence.db 를 읽기 전용으로 엶(strategyproposal.LoadProductionAuthorityBatch). 그 파일에 그림자 봉이
// 섞이면 그림자 측정이 생산 증거 권위가 됨. 그래서 거절 조건을 셋으로 겹침 — 이름 · 이웃 저널 · 파일 신원(하드 링크 · 심링크 포함).

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const productionEvidenceName = "evidence.db"

// productionEvidenceCandidates 는 엔진이 쓸 수 있는 생산 evidence.db 자리 — 저널 기본 자리 옆과 설정 디렉터리 안
// (배포 기록상 컨테이너 저널은 설정 디렉터리에 놓임).
func productionEvidenceCandidates(configDir string) []string {
	var out []string
	if path, err := journal.DefaultPath(); err == nil {
		out = append(out, filepath.Join(filepath.Dir(path), productionEvidenceName))
	}
	if configDir != "" {
		out = append(out, filepath.Join(configDir, productionEvidenceName))
	}
	return out
}

// checkEvidencePath 는 target 이 생산 evidence 일 수 있으면 오류.
func checkEvidencePath(target string, production []string) error {
	if target == "" {
		return errors.New("an evidence store path is required (--evidence-db)")
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	// 디렉터리는 심링크를 풀어서 봄 — 링크한 디렉터리를 거쳐 저널 옆에 닿는 길을 막음.
	dir, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return fmt.Errorf("resolving the directory of %s: %w", target, err)
	}
	resolved := filepath.Join(dir, filepath.Base(abs))
	if info, err := os.Lstat(resolved); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if resolved, err = filepath.EvalSymlinks(resolved); err != nil {
			return fmt.Errorf("resolving %s: %w", target, err)
		}
	}
	if filepath.Base(resolved) == productionEvidenceName || filepath.Base(abs) == productionEvidenceName {
		return fmt.Errorf("refusing %s: the engine's evidence store is named %s; use a private file name", target, productionEvidenceName)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(resolved), journal.DBFileName)); err == nil {
		return fmt.Errorf("refusing %s: it sits next to an engine journal (%s)", target, journal.DBFileName)
	}
	targetInfo, targetErr := os.Stat(resolved)
	for _, candidate := range production {
		if filepath.Clean(candidate) == resolved {
			return fmt.Errorf("refusing %s: it is the production evidence store %s", target, candidate)
		}
		candidateInfo, err := os.Stat(candidate)
		if err == nil && targetErr == nil && os.SameFile(candidateInfo, targetInfo) {
			return fmt.Errorf("refusing %s: it is the same file as the production evidence store %s", target, candidate)
		}
	}
	return nil
}
