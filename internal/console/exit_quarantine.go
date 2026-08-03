package console

// exit_quarantine.go is change a063's operator surface: lifting an exit snapshot
// quarantine from the position-management screen.
//
// # What the operator is deciding
//
// A quarantined generation is not judged at all — the engine refuses it before
// any evaluation, so neither the stop nor the take-profit is being checked. The
// release does not repair anything and does not weaken any rule: it fills in one
// column, and the next observation runs the very same recovery selection that
// quarantined the position. If the cause still holds, the position is quarantined
// again immediately, under a new version.
//
// That is why this screen argues from what is *kept* rather than what is fixed:
// the entry, the initial stop and the stored protection line all survive the
// release untouched, which is precisely what "자동관리 해제 → 재편입" would destroy.
//
// # No typed confirmation
//
// The danger acknowledgement is a checkbox behind a mandatory delay, and the
// ledger evidence is composed by the server. Asking an operator to retype a
// phrase is friction this console does not use, and it would be less accurate
// than the fields the engine already holds.

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
)

const exitQuarantineTokenTTL = 5 * time.Minute

// ExitQuarantineCommander is the console's complete authority over quarantine
// releases. It is discovered on the position-policy commander rather than
// injected as a separate option: the engine serves both from one control plane,
// and a console wired to an engine that predates a063 must degrade to "no
// button" rather than to a missing dependency.
type ExitQuarantineCommander interface {
	Quarantines(context.Context) ([]exitquarantine.Row, error)
	PreviewQuarantineRelease(context.Context, exitquarantine.Request) (exitquarantine.Preview, error)
	ReleaseQuarantine(context.Context, exitquarantine.ApplyRequest) (exitquarantine.Result, error)
}

// exitQuarantines reports the release capability, if this console has one.
func (c *Console) exitQuarantines() (ExitQuarantineCommander, bool) {
	commander, ok := c.opts.PositionPolicies.(ExitQuarantineCommander)
	return commander, ok
}

// quarantineBadgeView is what a quarantined row shows and offers.
type quarantineBadgeView struct {
	Present       bool
	Version       int64
	Reason        string
	ReasonText    string
	Evidence      string
	QuarantinedAt string
	Age           string
	Protection    string
	Token         string
}

func (v quarantineBadgeView) ProtectionText() string {
	if strings.TrimSpace(v.Protection) == "" {
		return "알 수 없음"
	}
	return v.Protection
}

// quarantineReasonText names the three reasons the engine writes. An unknown
// code is shown as itself rather than guessed at: a reason this console does not
// recognise is exactly the case where the raw evidence matters most.
func quarantineReasonText(reason string) string {
	switch strings.TrimSpace(reason) {
	case "ambiguous_recovery":
		return "복구 후보를 하나로 정할 수 없음"
	case "stored_snapshot_corrupt":
		return "저장된 snapshot을 읽을 수 없음"
	case "legacy_policy_identity_unknown":
		return "정책 정체성을 해석할 수 없음"
	case "":
		return "사유 미기록"
	default:
		return reason
	}
}

type quarantineTokenPayload struct {
	Purpose    string `json:"p"`
	PositionID string `json:"i"`
	Generation int64  `json:"g"`
	Version    int64  `json:"v"`
	NotBefore  int64  `json:"n"`
	ExpiresAt  int64  `json:"e"`
}

type quarantineReleasePreviewPage struct {
	chrome
	CSRF     string
	Row      exitquarantine.Row
	Token    string
	WaitSecs int
	Reason   string
}

// newQuarantineBadge builds the row view for one active quarantine.
func (c *Console) newQuarantineBadge(row exitquarantine.Row, now time.Time) quarantineBadgeView {
	badge := quarantineBadgeView{
		Present: true, Version: row.Version, Reason: row.Reason,
		ReasonText: quarantineReasonText(row.Reason), Evidence: strings.TrimSpace(row.Evidence),
		QuarantinedAt: strings.TrimSpace(row.QuarantinedAt), Protection: strings.TrimSpace(row.Protection),
		Age: quarantineAge(row.QuarantinedAt, now),
	}
	token, err := c.signQuarantineToken(quarantineTokenPayload{
		Purpose: "quarantine-select", PositionID: row.PositionID, Generation: row.Generation,
		Version: row.Version, NotBefore: now.UnixNano(),
		ExpiresAt: now.Add(exitQuarantineTokenTTL).UnixNano(),
	})
	if err == nil {
		badge.Token = token
	}
	return badge
}

// quarantineAge is how long this position has been unjudged. It is the number an
// operator reacts to, so an unparseable or future stamp says so rather than
// rendering a plausible-looking zero.
func quarantineAge(stamp string, now time.Time) string {
	at, err := time.Parse(time.RFC3339, strings.TrimSpace(stamp))
	if err != nil {
		return "경과 시간 알 수 없음"
	}
	d := now.Sub(at)
	if d < 0 {
		return "경과 시간 알 수 없음"
	}
	switch {
	case d < time.Minute:
		return "1분 미만"
	case d < 24*time.Hour:
		return fmt.Sprintf("%d시간 %d분", int(d/time.Hour), int(d%time.Hour/time.Minute))
	default:
		return fmt.Sprintf("%d일 %d시간", int(d/(24*time.Hour)), int(d%(24*time.Hour)/time.Hour))
	}
}

func (c *Console) handleQuarantineReleasePreview(w http.ResponseWriter, r *http.Request) {
	commander, ok := c.exitQuarantines()
	if !ok {
		c.refuse(w, http.StatusNotImplemented, "판정 격리 command seam 미배선",
			"엔진 control plane이 격리 해제를 제공하지 않는다. 아무것도 변경되지 않았다.")
		return
	}
	payload, err := c.verifyQuarantineToken(r.PostFormValue("quarantine_token"))
	if err != nil {
		c.refuse(w, http.StatusForbidden, "격리 action token 거부", err.Error())
		return
	}
	preview, err := commander.PreviewQuarantineRelease(r.Context(), exitquarantine.Request{
		PositionID: payload.PositionID, Generation: payload.Generation, Version: payload.Version,
		Actor: exitquarantine.ActorLocalOperator, At: c.now().UTC(),
	})
	if err != nil {
		c.writeQuarantineError(w, err)
		return
	}
	if strings.TrimSpace(preview.Capability) == "" {
		c.refuse(w, http.StatusInternalServerError, "격리 해제 token 생성 실패", "아무것도 변경되지 않았다.")
		return
	}
	wait := preview.WaitSeconds
	if wait <= 0 {
		wait = 3
	}
	c.render(w, "quarantine-release-preview", quarantineReleasePreviewPage{
		chrome: c.chromeOnRequest("position-management"), CSRF: c.csrf,
		Row: preview.Row, Token: preview.Capability, WaitSecs: wait,
		Reason: quarantineReasonText(preview.Row.Reason),
	})
}

func (c *Console) handleQuarantineReleaseApply(w http.ResponseWriter, r *http.Request) {
	commander, ok := c.exitQuarantines()
	if !ok {
		c.refuse(w, http.StatusNotImplemented, "판정 격리 command seam 미배선",
			"엔진 control plane이 격리 해제를 제공하지 않는다. 아무것도 변경되지 않았다.")
		return
	}
	capability := strings.TrimSpace(r.PostFormValue("capability"))
	if capability == "" {
		c.refuse(w, http.StatusForbidden, "격리 해제 승인 token 거부",
			"engine capability가 없다. 아무것도 변경되지 않았다.")
		return
	}
	result, err := commander.ReleaseQuarantine(r.Context(), exitquarantine.ApplyRequest{
		Capability: capability, Confirmed: r.PostFormValue("confirm") == "yes",
	})
	if err != nil {
		c.writeQuarantineError(w, err)
		return
	}
	notice := fmt.Sprintf("판정 격리 해제됨 — %s · generation %d · quarantine v%d. "+
		"저장된 기준선은 그대로이고 다음 관측부터 판정이 재개된다. 원인이 남아 있으면 다시 격리된다.",
		result.Row.Symbol, result.Row.Generation, result.Row.Version)
	http.Redirect(w, r, "/position-management?notice="+urlQueryEscape(notice), http.StatusSeeOther)
}

func (c *Console) writeQuarantineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, exitquarantine.ErrUnwired):
		c.refuse(w, http.StatusNotImplemented, "판정 격리 command seam 미배선",
			"엔진 control plane이 격리 해제를 제공하지 않는다. 아무것도 변경되지 않았다.")
	case errors.Is(err, exitquarantine.ErrNotQuarantined):
		c.refuse(w, http.StatusNotFound, "활성 격리 없음",
			"이 generation에는 해제할 격리가 없다. 이미 해제됐거나 다른 세대의 기록이다. 아무것도 변경되지 않았다.")
	case errors.Is(err, exitquarantine.ErrVersionMismatch):
		c.refuse(w, http.StatusPreconditionFailed, "격리 version 충돌",
			"보고 있던 사이에 격리가 다시 기록되거나 해제됐다. 현재값을 다시 불러와 preview부터 진행하라. 아무것도 변경되지 않았다.")
	case errors.Is(err, exitquarantine.ErrCapabilityTooEarly):
		c.refuse(w, http.StatusTooEarly, "3초 확인 대기 중",
			"engine preview 후 최소 3초가 지나야 한다. 같은 화면에서 잠시 기다린 뒤 한 번만 다시 적용하라.")
	case errors.Is(err, exitquarantine.ErrCapabilityExpired):
		c.refuse(w, http.StatusGone, "격리 해제 capability 만료",
			"현재값을 다시 불러와 preview부터 다시 진행하라. 아무것도 변경되지 않았다.")
	case errors.Is(err, exitquarantine.ErrCapabilityInvalid):
		c.refuse(w, http.StatusForbidden, "격리 해제 capability 거부",
			"preview 없이 적용했거나 이미 사용된 capability다. 현재값에서 새 preview를 만들어라.")
	case errors.Is(err, exitquarantine.ErrConfirmationRequired):
		c.refuse(w, http.StatusBadRequest, "위험 변경 확인 필요",
			"판정 재개 안내를 확인한 뒤 체크박스를 선택하라. 문구 입력은 필요 없고 아무것도 변경되지 않았다.")
	default:
		c.refuse(w, http.StatusBadRequest, "격리 해제 요청 거부", err.Error()+" 아무것도 변경되지 않았다.")
	}
}

func (c *Console) signQuarantineToken(payload quarantineTokenPayload) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, c.policySeal)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (c *Console) verifyQuarantineToken(token string) (quarantineTokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return quarantineTokenPayload{}, errors.New("opaque token 형식이 올바르지 않다")
	}
	mac := hmac.New(sha256.New, c.policySeal)
	_, _ = mac.Write([]byte(parts[0]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return quarantineTokenPayload{}, errors.New("opaque token 서명이 일치하지 않는다")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return quarantineTokenPayload{}, errors.New("opaque token payload를 읽을 수 없다")
	}
	var payload quarantineTokenPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Purpose != "quarantine-select" {
		// The purpose check is what stops a policy selection token from being
		// replayed here: both are sealed with the same key, and only the purpose
		// distinguishes what the operator actually clicked.
		return quarantineTokenPayload{}, errors.New("opaque token 용도가 일치하지 않는다")
	}
	now := c.now().UnixNano()
	if now >= payload.ExpiresAt {
		return quarantineTokenPayload{}, errors.New("opaque token이 만료됐다")
	}
	if now < payload.NotBefore {
		return quarantineTokenPayload{}, errors.New("opaque token이 아직 유효하지 않다")
	}
	if strings.TrimSpace(payload.PositionID) == "" || payload.Version <= 0 {
		return quarantineTokenPayload{}, errors.New("opaque token이 격리를 지목하지 않는다")
	}
	return payload, nil
}
