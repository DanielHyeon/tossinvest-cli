package strategyprojection

import (
	"errors"
	"strings"
)

// RuntimeIdentityProjection 은 「이 스냅샷을 만든 엔진 프로세스가 **어떤 build 로,
// 어떤 lane 설정으로** 돌고 있는가」다.
//
// # 왜 시장별이 아니라 envelope 인가
//
// 두 값 모두 프로세스 하나의 사실이다. 시장별로 다를 수 없고, 무엇보다 사람이 이
// 숫자를 가장 필요로 하는 순간이 **두 시장 다 UNKNOWN 일 때**다. 시장 레코드 안에
// 두면 바로 그때 사라진다.
//
// # 왜 필요한가
//
// 활성화 매니페스트에 사람이 손으로 적어야 하는 숫자가 이 두 개다. 지금까지 엔진은
// 이 값을 스냅샷 내부에 채우고도 **어떤 운영 화면으로도 내보내지 않았다.** 읽을 수
// 없는 숫자는 적을 수 없고, 적지 못하면 스케줄러는 영원히 안 깨어난다 — a112 결정 54.
//
// # 콘솔이 스스로 계산하면 안 되는 이유
//
// 같은 저장소의 상수로 만드는 값이라 콘솔도 자기 바이너리에서 똑같이 계산할 수 있다.
// 하지만 콘솔과 엔진의 build 가 다르면 그렇게 만든 숫자는 **엔진이 거절할 매니페스트**를
// 낳는다. 이 값의 권위는 오직 엔진 프로세스다.
type RuntimeIdentityProjection struct {
	ConfigDigest *string `json:"configDigest"`
	BuildDigest  *string `json:"buildDigest"`
}

// runtimeIdentityPrefix 는 이 두 값이 **반드시** 달고 있어야 하는 접두사다.
//
// # 왜 벗기지 않는가 — 이 문자열은 사람이 그대로 옮겨 적는다
//
// 활성화 매니페스트 검증은 `body.ConfigVersion != binding.ConfigVersion` 과
// `body.BuildDigest != binding.BuildDigest` 의 **정확한 문자열 비교**다
// (`internal/scheduler/production_activation.go` 의 validateProductionActivationManifest).
// 그리고 binding 쪽 값은 `strategyRuntimeConfigDigest()`/`strategyRuntimeBuildDigest()` 가
// 만든 `sha256:<64hex>` 다.
//
// 이 패키지의 다른 digest 필드(evidence, activation)는 벗긴 64자리를 쓴다. 그것들은
// **보여 주는** 값이라 그래도 된다. 이 두 개는 **받아 적는** 값이라 안 된다 —
// 접두사를 벗겨서 보여 주면 운영자가 옮겨 적은 매니페스트를 엔진이 거절한다.
const runtimeIdentityPrefix = "sha256:"

// WithRuntimeIdentity 는 정규 digest 두 개를 envelope 에 붙인다.
//
// 하나라도 정규형(`sha256:` + 소문자 64자리 hex)이 아니면 **아무것도 붙이지 않는다.**
// 반쪽이나 형식이 깨진 숫자를 화면에 올리는 것보다 「관측 없음」이 정직하다 — 운영자는
// 화면의 숫자를 그대로 파일에 옮겨 적기 때문이다.
func WithRuntimeIdentity(snapshot Snapshot, configDigest, buildDigest string) Snapshot {
	next := Clone(snapshot)
	if !validRuntimeIdentityDigest(configDigest) || !validRuntimeIdentityDigest(buildDigest) {
		return next
	}
	config, build := configDigest, buildDigest
	next.Runtime = RuntimeIdentityProjection{ConfigDigest: &config, BuildDigest: &build}
	return next
}

// cloneRuntimeIdentity 는 포인터 두 개를 값으로 복사한다. 스냅샷을 받은 쪽이 문자열을
// 바꿔도 store 안의 진실이 흔들리지 않아야 한다.
func cloneRuntimeIdentity(identity RuntimeIdentityProjection) RuntimeIdentityProjection {
	return RuntimeIdentityProjection{ConfigDigest: cloneString(identity.ConfigDigest), BuildDigest: cloneString(identity.BuildDigest)}
}

// validateRuntimeIdentity 는 이 패키지의 다른 짝 필드와 같은 규칙을 쓴다: 함께 있거나
// 함께 없고, 있으면 정규 digest 여야 한다. 스냅샷은 Unix transport 를 건너 오므로
// 상대가 무엇을 보내든 경계에서 다시 판정한다.
func validateRuntimeIdentity(identity RuntimeIdentityProjection) error {
	if !pairedIdentity(identity.ConfigDigest, identity.BuildDigest) {
		return errors.New("partial runtime identity")
	}
	if identity.ConfigDigest != nil && (!validRuntimeIdentityDigest(*identity.ConfigDigest) || !validRuntimeIdentityDigest(*identity.BuildDigest)) {
		return errors.New("noncanonical runtime identity digest")
	}
	return nil
}

// validRuntimeIdentityDigest 는 스케줄러가 실제로 견주는 그 형식이다 — 접두사까지 포함해서.
func validRuntimeIdentityDigest(value string) bool {
	return strings.HasPrefix(value, runtimeIdentityPrefix) &&
		validDigest(strings.TrimPrefix(value, runtimeIdentityPrefix))
}
