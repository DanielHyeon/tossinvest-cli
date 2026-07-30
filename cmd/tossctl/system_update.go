package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/localupdate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/releaseupdate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

// assembleConsoleSystemUpdate is the production composition seam shared by the
// CLI and its assembly regression test. Constructing it performs no network
// request and changes neither the running executable nor the fixed candidate.
// A Sigstore setup error returns the still-usable local installer so the
// operator can review an already staged candidate without silently weakening
// the signed download path.
func assembleConsoleSystemUpdate(
	currentPath, sigstoreCacheDir, currentVersion string,
) (*localupdate.Updater, *releaseupdate.Client, error) {
	updater, err := localupdate.New(currentPath)
	if err != nil {
		return nil, nil, err
	}
	downloader, err := releaseupdate.NewProduction(sigstoreCacheDir, currentVersion)
	if err != nil {
		return updater, nil, err
	}
	return updater, downloader, nil
}

// strictVerifyActivity is the system updater's fail-closed reading of the
// external verification marker. runlock.Fresh intentionally fails open because
// its normal consumer is a survey; executable replacement has the opposite
// contract, so unreadable or unclassifiable evidence is a refusal.
func strictVerifyActivity(path string, now time.Time) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("verification marker path is empty")
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading verification marker: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("verification marker %s is not a regular non-symlink file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening verification marker: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing verification marker: %w", err)
	}
	age := now.Sub(info.ModTime())
	if age < 0 || age <= runlock.StaleAfter {
		return fmt.Errorf("verification marker is fresh (last update %s)",
			info.ModTime().UTC().Format(time.RFC3339))
	}
	return nil
}
