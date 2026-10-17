package engine

// mode_projection_wiring.go 는 a092 착지 단위 ④ — 원장의 운영 모드를 산 엔진의 진입 게이트에 묶는 조립 한 조각임.
//
// 정본 risk-management 「모드의 강제 지점은 EntryGate 투영이다」는 투영기(SetModeProjector)가 생산에서 묶이지 않아
// 산 프로세스에서는 성립하지 않았음(a092 발견). 여기서 묶고, 첫 진입 점검보다 먼저 현재 모드를 복원함 — buildGateway
// 가 반환한 뒤에야 루프가 뜨므로 그 반환 전에 부름.

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// operatingModeRestoreFailed 는 복원 실패 래치의 설명임. 원장 오류 원문과 계좌는 싣지 않음(게이트 설명은 상태 출력이 어디서나 읽음 — 불변식 8).
const operatingModeRestoreFailed = "the operating mode could not be read from the journal at start-up; " +
	"new entries stay blocked until a restart after the journal is repaired (exits are unaffected)"

// bindOperatingModeProjection 은 게이트를 원장의 모드 투영기로 묶고 현재 모드를 복원함.
//
// 투영기 묶기 실패는 배선 결함(이미 묶임)이라 오류로 돌려 기동을 거부함. 복원(모드 행 읽기) 실패는 기동을 거부하지
// 않음(a092 C15 (ㄴ)): 보호가 배선되지 않은 구성에서 기동 거부는 손절 부재라 안전 불변식 4 가 앞섬. 대신 모드 사유로
// 진입을 막고 로그를 남김 — 이 래치는 뒤이은 성공 투영이 교체하고(M13), 원장 수리 뒤 재시작이 풂(K8).
func bindOperatingModeProjection(ctx context.Context, j *journal.Journal, gate *execgw.EntryGate,
	accountRef string, log *obs.Logger) error {
	if err := j.SetModeProjector(gate); err != nil {
		return err
	}
	if _, err := j.RestoreOperatingModeProjection(ctx, accountRef); err != nil {
		gate.Block(execgw.ReasonOperatingModeBlocked, operatingModeRestoreFailed)
		if log != nil {
			log.Error(obs.EventOperatingMode, err, obs.FieldDetail, operatingModeRestoreFailed)
		}
	}
	return nil
}
