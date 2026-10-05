package main

// verify_reconcile.go 는 `tossctl verify reconcile`(a121) 의 **골격**이다.
//
// RED 로트 산출물: 시그니처만 세우고 모든 진입점은 코드 없는 거절을 돌려준다. newVerifyCmd 에 등록하지 않는다
// (등록은 GREEN 의 newVerifyCmd 편집 — FLM 번들 internal-verifylive… 와 cmd-tossctl--newverifycmd 참조).
// 브로커 획득은 verifyBrokerFactory 를 재사용하지 않는다(design G2 P1-1 — Broker 비반환 전용 생성자).

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

// verifyReconcileOptions 는 대사 명령의 입력임. 기록 경로는 --config-dir 에서만 유도함(G3-3).
type verifyReconcileOptions struct {
	market string
	// record 는 --record override 가 들어오면 채워짐 — 대사는 이것을 거절함(G3-3).
	record string
}

var errVerifyReconcileUnimplemented = errors.New("verify reconcile: not implemented (a121 RED skeleton)")

// newVerifyReconcileCmd 는 대사 명령을 만듦. 골격: 주석·플래그 없이 언제나 거절.
func newVerifyReconcileCmd(root *rootOptions) *cobra.Command {
	opts := &verifyReconcileOptions{}
	return &cobra.Command{
		Use: "reconcile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVerifyReconcile(cmd, root, opts)
		},
	}
}

// runVerifyReconcile 는 사전 검사 → 실행 배제 flock → rate-budget lease → 좁은 읽기 생성 → verifylive.Reconcile 순으로
// 갈 자리임. 골격: 아무것도 하지 않고 거절.
func runVerifyReconcile(cmd *cobra.Command, root *rootOptions, opts *verifyReconcileOptions) error {
	return &verifylive.ReconcileRefusal{Detail: errVerifyReconcileUnimplemented.Error()}
}

// reconcilePreflight 는 목록 읽기·잠금 전에 프로필·자격·기록·시장 조건을 판정하고 유도된 기록 경로를 돌려줌(G3-3·G3-4).
// 골격: 언제나 코드 없는 거절.
func reconcilePreflight(root *rootOptions, opts *verifyReconcileOptions, getenv func(string) string) (string, error) {
	return "", &verifylive.ReconcileRefusal{Detail: errVerifyReconcileUnimplemented.Error()}
}

// verifyReconcileReaderFactory 는 명령이 쓰는 읽기 전용 의존 생성자임 — 시험이 httptest 로 바꿔 끼움
// (verifyBrokerFactory 와 같은 배치, --base-url 플래그 없음).
var verifyReconcileReaderFactory = buildVerifyReconcileReader

// buildVerifyReconcileReader 는 프로필 파일의 자격으로 좁은 읽기를 만듦. 골격: 거절.
func buildVerifyReconcileReader(ctx context.Context, root *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
	return nil, verifylive.ReconcileAccount{}, &verifylive.ReconcileRefusal{Detail: errVerifyReconcileUnimplemented.Error()}
}

// newVerifyReconcileReader 는 Broker 를 절대 반환하지 않는 좁은 생성자임(design G2 P1-1, G3-2 P1-2·R2-2).
// 자체 Accounts() 로 계좌 수·기형 행·seq 를 판정하고 검증한 seq 로 명시 재결속한 읽기만 돌려줄 자리임.
// 골격: 네트워크 호출 없이 거절.
func newVerifyReconcileReader(ctx context.Context, creds official.Credentials, tokenFile string,
	opts ...official.Option) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
	return nil, verifylive.ReconcileAccount{}, &verifylive.ReconcileRefusal{Detail: errVerifyReconcileUnimplemented.Error()}
}

// verifyReconcileTerminal 은 기본 승인이 읽고 쓰는 단말임 — 시험이 바꿔 끼움.
// interactive 가 거짓(파이프·리다이렉트 stdin)이면 기본 승인은 거절해야 함(자동 승인 경로 없음).
// 골격: 언제나 비대화형.
var verifyReconcileTerminal = func() (in io.Reader, out io.Writer, interactive bool) {
	return os.Stdin, os.Stderr, false
}

// verifyReconcileApprove 는 사람 승인 자리임(목록 읽기 전). 기본값은 verifyReconcileTerminal 에서 대화형 y/N 확인을
// 받아야 함 — 타이핑 문구 없음, 기본은 N. 골격: 언제나 승인 거절.
var verifyReconcileApprove = func(ctx context.Context, approval verifylive.ReconcileApproval) error {
	return errVerifyReconcileUnimplemented
}
