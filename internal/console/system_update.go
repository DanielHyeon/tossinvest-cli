package console

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/localupdate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/releaseupdate"
)

var errBinaryCommitted = errors.New(
	"시스템 업데이트가 설치되어 이 콘솔은 종료 대기 중이다. 새 프로세스가 뜬 뒤 다시 시도하라")

type SystemUpdater interface {
	Inspect() localupdate.Inspection
	Install(reviewedSHA256 string, commitGuard func() error) (localupdate.Result, error)
}

type ReleaseDownloader interface {
	Fetch(context.Context) (releaseupdate.Release, error)
}

type ReleaseCandidateStager interface {
	StageCandidate(io.Reader, string) (localupdate.StageResult, error)
}

type signedReleaseReceipt struct {
	Tag           string
	AssetName     string
	ArchiveSHA256 string
	Signer        string
	SourceCommit  string
	CandidateSHA  string
	Bootstrap     bool
}

type AcquireUpdateEngineLock func() (release func(), err error)
type CheckUpdateVerifyActivity func() error

// startExclusive serializes the two work-start routes with installation.
func (c *Console) startExclusive(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.activityMu.Lock()
		defer c.activityMu.Unlock()
		if c.updateCommitted {
			c.redirectDashboard(w, r, errBinaryCommitted.Error())
			return
		}
		next(w, r)
	}
}

func (c *Console) handleSystemUpdateInstall(w http.ResponseWriter, r *http.Request) {
	c.activityMu.Lock()
	defer c.activityMu.Unlock()

	if c.updateCommitted {
		c.redirectSettings(w, r, errBinaryCommitted.Error())
		return
	}
	if c.opts.SystemUpdater == nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 시스템 업데이트 seam이 배선되지 않았다.")
		return
	}
	if c.opts.Relaunch == nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 같은 포트 재기동 seam이 배선되지 않았다.")
		return
	}
	port, ok := c.listeningPort()
	if !ok {
		c.redirectSettings(w, r, "업데이트 안 됨 — 콘솔이 아직 재사용할 포트를 열지 않았다.")
		return
	}
	if run := c.currentRun(); run != nil && !run.finished() {
		c.redirectSettings(w, r, "업데이트 안 됨 — 검증이 진행 중이다. 먼저 검증을 끝내라.")
		return
	}
	if c.opts.AcquireUpdateEngineLock == nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 실제 엔진 exclusion이 배선되지 않았다.")
		return
	}
	release, err := c.opts.AcquireUpdateEngineLock()
	if err != nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 엔진 exclusion을 잡을 수 없다: "+err.Error())
		return
	}
	if release == nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 엔진 exclusion release가 비어 있다.")
		return
	}
	defer release()
	if c.opts.CheckUpdateVerifyActivity == nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 외부 검증 활동 판정이 배선되지 않았다.")
		return
	}
	if err := c.opts.CheckUpdateVerifyActivity(); err != nil {
		c.redirectSettings(w, r, "업데이트 안 됨 — 외부 검증 활동을 배제할 수 없다: "+err.Error())
		return
	}

	view := c.opts.SystemUpdater.Inspect()
	reviewed := strings.ToLower(strings.TrimSpace(r.PostFormValue("reviewed_sha256")))
	if !view.Installable || reviewed == "" || reviewed != strings.ToLower(view.Candidate.SHA256) {
		reason := view.Reason
		if reason == "" {
			reason = "화면에서 검토한 candidate SHA-256과 현재 후보가 다르다"
		}
		c.redirectSettings(w, r, "업데이트 안 됨 — "+reason)
		return
	}

	result, err := c.opts.SystemUpdater.Install(reviewed, c.opts.CheckUpdateVerifyActivity)
	if err != nil {
		c.redirectSettings(w, r, "업데이트 실패 — 기존 콘솔은 계속 실행 중이다: "+err.Error())
		return
	}

	fmt.Fprintf(c.out, "system update    old=%s new=%s rollback=%s\n",
		result.OldSHA256, result.NewSHA256, result.RollbackPath)
	c.updateCommitted = true
	token, handoffNote := c.mintHandoff()
	note := fmt.Sprintf("시스템 업데이트 설치 완료: %s → %s · rollback %s",
		result.OldSHA256, result.NewSHA256, result.RollbackPath)
	if handoffNote != "" {
		note += "\n" + handoffNote
	}
	c.render(w, "restart", restartPage{
		Target: restartTarget(token),
		Note:   note,
		Delay:  restartRefreshDelay,
	})
	c.requestRelaunch(port)
}

func (c *Console) handleSystemUpdateDownload(w http.ResponseWriter, r *http.Request) {
	if c.opts.ReleaseDownloader == nil || c.opts.ReleaseCandidateStager == nil {
		c.redirectSettings(w, r, "서명 릴리스 다운로드 안 됨 — release verifier 또는 fixed candidate publisher가 배선되지 않았다.")
		return
	}

	// This mutex is deliberately not activityMu: TUF and GitHub can be slow, and
	// a download must not delay an engine/verification start. It does cover the
	// entire release operation so the shared TUF cache has one writer and a slow
	// old discovery cannot publish after a newer request.
	c.releaseMu.Lock()
	defer c.releaseMu.Unlock()

	release, err := c.opts.ReleaseDownloader.Fetch(r.Context())
	if err != nil {
		c.redirectSettings(w, r, "서명 릴리스 확인 실패 — 기존 candidate는 유지됐다: "+err.Error())
		return
	}

	// Only the fixed local publish interval shares the install/start exclusion.
	c.activityMu.Lock()
	defer c.activityMu.Unlock()
	if c.updateCommitted {
		c.redirectSettings(w, r, errBinaryCommitted.Error())
		return
	}
	result, err := c.opts.ReleaseCandidateStager.StageCandidate(
		bytes.NewReader(release.Binary), release.SourceCommit)
	if err != nil {
		note := "서명 릴리스 staging 실패 — 기존 candidate는 유지됐다: " + err.Error()
		if result.RecoveryPath != "" {
			note += " · recovery " + result.RecoveryPath
		}
		c.redirectSettings(w, r, note)
		return
	}
	c.releaseReceiptMu.Lock()
	c.releaseReceipt = &signedReleaseReceipt{
		Tag: release.Tag, AssetName: release.AssetName,
		ArchiveSHA256: release.ArchiveSHA256, Signer: release.WorkflowIdentity,
		SourceCommit: release.SourceCommit,
		CandidateSHA: result.Metadata.SHA256, Bootstrap: release.Bootstrap,
	}
	c.releaseReceiptMu.Unlock()

	bootstrap := ""
	if release.Bootstrap {
		bootstrap = " · 현재 빌드는 release semver가 없어 bootstrap으로 처리했으며 버전 순서를 주장하지 않는다"
	}
	c.redirectSettings(w, r, fmt.Sprintf(
		"서명 릴리스 staged: %s · %s · archive SHA-256 %s · signer %s · "+
			"source commit %s · candidate SHA-256 %s%s",
		release.Tag, release.AssetName, release.ArchiveSHA256,
		release.WorkflowIdentity, release.SourceCommit, result.Metadata.SHA256, bootstrap))
}

func (c *Console) signedReleaseReceipt(candidateSHA string) (signedReleaseReceipt, bool) {
	c.releaseReceiptMu.Lock()
	defer c.releaseReceiptMu.Unlock()
	if c.releaseReceipt == nil ||
		!strings.EqualFold(c.releaseReceipt.CandidateSHA, strings.TrimSpace(candidateSHA)) {
		return signedReleaseReceipt{}, false
	}
	return *c.releaseReceipt, true
}
