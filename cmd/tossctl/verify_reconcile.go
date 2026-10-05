package main

// verify_reconcile.go 는 `tossctl verify reconcile`(a121) 이다.
//
// 이전 검증이 남긴 stale 조건주문 하나를, 공식 목록을 **읽기만** 해서 그 부재를 좁혀 확인한 뒤 기록에 대사 사건 한 줄을
// 더한다. 주문을 내거나 취소·정정하지 않는다 — 브로커 변이 경로가 없고(봉인: 좁은 생성자가 GET 셋만 노출하는 읽기를
// 돌려줌), 실행 전 사람 승인(대화형 y/N)이 필요하며, mutating=true 로 등재해 대화형 에이전트가 자동 실행하지 않는다.
//
// 흐름: 사전 검사(프로필·자격·기록·시장) → 실행 배제 flock → rate-budget lease → 좁은 생성자(자체 Accounts() 판정·
// seq 명시 재결속) → verifylive.Reconcile(가드 순서 계약 그대로). verifyBrokerFactory 를 재사용하지 않는다(design G2 P1-1).

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// verifyReconcileOptions 는 대사 명령의 입력임. 기록 경로는 --config-dir 에서만 유도함(G3-3).
type verifyReconcileOptions struct {
	market string
	// record 는 --record override 가 들어오면 채워짐 — 대사는 이것을 거절함(G3-3).
	record string
}

func reconcileRefusal(code verifylive.ReconcileRefusalCode, detail string) error {
	return &verifylive.ReconcileRefusal{Code: code, Detail: detail}
}

// newVerifyReconcileCmd 는 대사 명령을 만듦.
func newVerifyReconcileCmd(root *rootOptions) *cobra.Command {
	opts := &verifyReconcileOptions{}
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Record a read-backed 'reconciled absent' event for one stale verification conditional",
		Long: strings.TrimSpace(`
Reconcile the one conditional order an earlier verification left outstanding
after its cleanup DELETE failed (for example with conditional-order-not-found).

It never places, cancels or amends anything. It reads the official conditional
and plain order lists for the artifact's symbol twice, refuses on any sign of a
live successor, a fired child, an incomplete or malformed read, and appends one
'reconciled absent' line to the record only when every condition holds. A DELETE
404 by itself is never treated as an ending.

It requires an explicit --config-dir profile (credentials from that profile's
file, never the environment), an explicit --market, and an interactive y/N
approval after it shows the record's and the credentials' masked accounts.`),
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVerifyReconcile(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.market, "market", "", "Market whose record to reconcile: KR or US (required)")
	cmd.Flags().StringVar(&opts.record, "record", "",
		"Refused: the record is always derived from --config-dir so it matches the credentials' profile")
	return cmd
}

// runVerifyReconcile 는 사전 검사 → flock → lease → 좁은 생성자 → verifylive.Reconcile 순으로 감.
func runVerifyReconcile(cmd *cobra.Command, root *rootOptions, opts *verifyReconcileOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	out := cmd.OutOrStdout()

	recordPath, err := reconcilePreflight(root, opts, os.Getenv)
	if err != nil {
		return err
	}
	lock, err := acquireVerifyExecutionLock(root)
	if err != nil {
		return reconcileRefusal(verifylive.RefuseExecutionLock, err.Error())
	}
	defer lock.Release()
	lease, err := acquireVerifyRateBudget(ctx, out, root)
	if err != nil {
		return reconcileRefusal(verifylive.RefuseRateBudget, err.Error())
	}
	defer lease.Release()

	reader, account, err := verifyReconcileReaderFactory(ctx, root)
	if err != nil {
		return err
	}
	res, err := verifylive.Reconcile(ctx, verifylive.ReconcileParams{
		RecordPath: recordPath,
		Market:     verifylive.NormalizeMarket(opts.market),
		Account:    account,
		Reader:     reader,
		Out:        out,
		Approve:    verifyReconcileApprove,
		Now:        time.Now,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "\nreconciled absent: %s %s (%s) — 기록 %s\n  근거 %s\n  취소·체결·성공한 요청이 아니다. 실패한 정리 줄은 기록에 그대로 남는다.\n",
		res.Artifact.Kind, res.Artifact.ID, res.Artifact.Symbol, recordPath, res.Basis)
	return nil
}

// reconcilePreflight 는 목록 읽기·잠금 전에 프로필·자격·기록·시장 조건을 판정하고 유도된 기록 경로를 돌려줌(G3-3·G3-4).
// 환경 변수 자격은 **하나라도** 설정돼 있으면 거절(어느 자격으로 읽었는지 모호해지는 반쪽 상태 포함).
func reconcilePreflight(root *rootOptions, opts *verifyReconcileOptions, getenv func(string) string) (string, error) {
	if root == nil || strings.TrimSpace(root.configDir) == "" {
		return "", reconcileRefusal(verifylive.RefuseNoConfigDir,
			"--config-dir is required: the record and the credentials must come from the same profile")
	}
	if getenv("TOSSCTL_OPENAPI_KEY") != "" || getenv("TOSSCTL_OPENAPI_SECRET") != "" {
		return "", reconcileRefusal(verifylive.RefuseEnvCredentials,
			"TOSSCTL_OPENAPI_KEY/TOSSCTL_OPENAPI_SECRET is set: reconciliation reads only the profile's credential file")
	}
	if strings.TrimSpace(opts.record) != "" {
		return "", reconcileRefusal(verifylive.RefuseRecordOverride,
			"--record is not accepted: the record is derived from --config-dir")
	}
	market := strings.TrimSpace(opts.market)
	if !strings.EqualFold(market, verifylive.MarketKR) && !strings.EqualFold(market, verifylive.MarketUS) {
		return "", reconcileRefusal(verifylive.RefuseNoMarket, "--market KR or --market US is required")
	}
	return filepath.Join(root.configDir, verifylive.RecordFileName(market)), nil
}

// verifyReconcileReaderFactory 는 명령이 쓰는 읽기 전용 의존 생성자임 — 시험이 httptest 로 바꿔 끼움
// (verifyBrokerFactory 와 같은 배치, --base-url 플래그 없음).
var verifyReconcileReaderFactory = buildVerifyReconcileReader

// buildVerifyReconcileReader 는 프로필 파일의 자격(환경 변수 무시)으로 좁은 읽기를 만듦.
func buildVerifyReconcileReader(ctx context.Context, root *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseNoCredentials, err.Error())
	}
	noEnv := func(string) string { return "" }
	creds, err := official.LoadCredentials(noEnv, credFile)
	if err != nil {
		return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseNoCredentials,
			"the profile's credential file cannot be read: "+err.Error())
	}
	if creds == nil {
		return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseNoCredentials,
			"no Open API credentials in this profile — run `tossctl openapi login --config-dir …`")
	}
	return newVerifyReconcileReader(ctx, *creds, tokenFile)
}

// newVerifyReconcileReader 는 Broker 를 절대 반환하지 않는 좁은 생성자임(design G2 P1-1, G3-2 P1-2·R2-2).
// 자체 Accounts() 로 기형 행·계좌 수·seq 를 판정하고, 검증한 seq 로 명시 재결속한 클라이언트를 GET 셋만 노출하는
// 래퍼에 담아 돌려줌. 계좌 판정용 첫 클라이언트는 밖으로 나가지 않음.
func newVerifyReconcileReader(ctx context.Context, creds official.Credentials, tokenFile string,
	opts ...official.Option) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
	accounts, err := official.New(creds, tokenFile, opts...).Accounts(ctx)
	if err != nil {
		return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountsRead,
			"the account list could not be read: "+err.Error())
	}
	// 기형 행(빈 계좌번호)은 계수에서 빼지 않고 그 자체로 거절 — 첫 행이면 transport 가 그 seq 를 캐시함(R2-2).
	for _, a := range accounts {
		if strings.TrimSpace(a.DisplayName) == "" {
			return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountMalformedRow,
				"the account list holds a row without an account number")
		}
	}
	if len(accounts) != 1 {
		return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountCount,
			fmt.Sprintf("the credentials list %d accounts — reconciliation needs exactly one (Q5)", len(accounts)))
	}
	seq, err := strconv.Atoi(strings.TrimSpace(accounts[0].ID))
	if err != nil || seq <= 0 {
		return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountSeqZero,
			"the account's sequence is not a positive number — the request header cannot be pinned")
	}
	bound := official.New(creds, tokenFile, append(append([]official.Option(nil), opts...), official.WithAccountSeq(seq))...)
	return &reconcileReads{client: bound},
		verifylive.ReconcileAccount{Ref: strings.TrimSpace(accounts[0].DisplayName), Seq: seq, Count: len(accounts)}, nil
}

// reconcileReads 는 대사가 받는 유일한 브로커 의존 값임 — 공식 GET 셋만 위임함. 클라이언트는 비공개 필드라
// type assertion 으로도 쓰기 메서드에 닿지 않음.
type reconcileReads struct {
	client *official.Client
}

func (r *reconcileReads) ReconcileConditionalOrdersPage(ctx context.Context, status, symbol, cursor string, limit int) (official.ReconcileConditionalPage, error) {
	return r.client.ReconcileConditionalOrdersPage(ctx, status, symbol, cursor, limit)
}

func (r *reconcileReads) ReconcileOpenOrdersPage(ctx context.Context, symbol, cursor string, limit int) (official.ReconcileOrderPage, error) {
	return r.client.ReconcileOpenOrdersPage(ctx, symbol, cursor, limit)
}

func (r *reconcileReads) ReconcileInstrument(ctx context.Context, symbol string) (string, error) {
	return r.client.ReconcileInstrument(ctx, symbol)
}

// verifyReconcileTerminal 은 기본 승인이 읽고 쓰는 단말임 — 시험이 바꿔 끼움. 대화형 판정은 실제 tty 여부
// (term.IsTerminal)임: 파이프·리다이렉트·/dev/null(문자 장치지만 tty 아님)은 비대화형.
var verifyReconcileTerminal = func() (in io.Reader, out io.Writer, interactive bool) {
	return os.Stdin, os.Stderr, term.IsTerminal(int(os.Stdin.Fd()))
}

// verifyReconcileApprove 는 사람 승인 자리임(목록 읽기 전). 비대화형이면 거절하고, 대화형이면 y/N 을 물어 명시 y 만
// 진행함 — 기본은 N, 타이핑 확인 문구 없음.
var verifyReconcileApprove = func(ctx context.Context, approval verifylive.ReconcileApproval) error {
	in, out, interactive := verifyReconcileTerminal()
	if !interactive {
		return errors.New("stdin is not an interactive terminal — reconciliation is approved by a person at a terminal, never by a pipe")
	}
	fmt.Fprintf(out, "%s %s 를 reconciled absent 로 기록할까? [y/N] ", approval.Kind, approval.ID)
	line, _ := bufio.NewReader(in).ReadString('\n')
	switch strings.TrimSpace(line) {
	case "y", "Y":
		return nil
	}
	return errors.New("the operator did not approve")
}
