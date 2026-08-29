package strategyprojection

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// 이 파일이 재는 것 — a112 결정 54의 「읽을 수 없는 숫자는 적을 수 없다」.
//
// 운영자가 활성화 매니페스트에 옮겨 적어야 하는 두 숫자(config/build digest)가
// projection envelope 로 나가고, 나가는 동안 반쪽이 되거나 형식이 깨지지 않는다.

// runtimeIdentityDigest 는 **스케줄러가 견주는 그 형식**을 만든다. 접두사가 붙어 있다.
func runtimeIdentityDigest(seed string) string { return "sha256:" + digestString(seed) }

// TestRuntimeIdentityIsAttachedOnlyAsAWholeCanonicalPair 는 「붙이거나 안 붙이거나」다.
//
// 반쪽 짜리 진실이 화면에 오르면 운영자는 없는 값을 지어내서 채운다. 그래서 하나라도
// 정규형이 아니면 둘 다 안 붙인다.
func TestRuntimeIdentityIsAttachedOnlyAsAWholeCanonicalPair(t *testing.T) {
	config, build := runtimeIdentityDigest("config"), runtimeIdentityDigest("build")

	attached := WithRuntimeIdentity(DormantSnapshot(projectionNow), config, build)
	if err := Validate(attached); err != nil {
		t.Fatal(err)
	}
	if attached.Runtime.ConfigDigest == nil || *attached.Runtime.ConfigDigest != config ||
		attached.Runtime.BuildDigest == nil || *attached.Runtime.BuildDigest != build {
		t.Fatalf("정규 쌍이 붙지 않았다: %+v", attached.Runtime)
	}

	// 정규형이 아닌 입력들. 셋 다 「실제로 실수하는 모양」이다: sha256: 접두사를 안 벗김,
	// 대문자, 길이 부족.
	// 셋 다 「실제로 실수하는 모양」이다. 특히 **접두사를 벗긴 64자리**가 여기 있다 —
	// 그것이 이 패키지의 다른 digest 필드가 쓰는 형식이라서 가장 그럴듯한 실수이고,
	// 그대로 옮겨 적으면 엔진이 매니페스트를 거절한다.
	for name, pair := range map[string][2]string{
		"bare config":  {strings.TrimPrefix(config, "sha256:"), build},
		"upper build":  {config, strings.ToUpper(build)},
		"short config": {config[:len(config)-1], build},
		"empty build":  {config, ""},
	} {
		t.Run(name, func(t *testing.T) {
			got := WithRuntimeIdentity(DormantSnapshot(projectionNow), pair[0], pair[1])
			if got.Runtime.ConfigDigest != nil || got.Runtime.BuildDigest != nil {
				t.Fatalf("비정규 입력이 화면으로 나갔다: %+v", got.Runtime)
			}
			if err := Validate(got); err != nil {
				t.Fatalf("붙이지 않은 스냅샷이 무효가 됐다: %v", err)
			}
		})
	}
}

// TestValidationRejectsPartialOrNoncanonicalRuntimeIdentity 는 **경계**의 핀이다.
//
// 스냅샷은 Unix transport 를 건너 온다. 상대가 무엇을 보내든 이쪽 화면에 올리기 전에
// 다시 판정해야 한다 — WithRuntimeIdentity 를 거치지 않은 값이 존재할 수 있다.
func TestValidationRejectsPartialOrNoncanonicalRuntimeIdentity(t *testing.T) {
	config, build := runtimeIdentityDigest("config"), runtimeIdentityDigest("build")
	noncanonical := strings.ToUpper(config)
	bare := strings.TrimPrefix(build, "sha256:")

	for name, identity := range map[string]RuntimeIdentityProjection{
		"config only":         {ConfigDigest: &config},
		"build only":          {BuildDigest: &build},
		"noncanonical config": {ConfigDigest: &noncanonical, BuildDigest: &build},
		"bare build":          {ConfigDigest: &config, BuildDigest: &bare},
	} {
		t.Run(name, func(t *testing.T) {
			snapshot := DormantSnapshot(projectionNow)
			snapshot.Runtime = identity
			if err := Validate(snapshot); err == nil {
				t.Fatal("반쪽이거나 비정규인 runtime identity 가 통과했다")
			}
		})
	}
}

// TestCloneCarriesRuntimeIdentityWithoutSharingIt 는 Clone 이 이 필드를 **떨어뜨리지도,
// 공유하지도** 않음을 잡는다. 떨어뜨리면 store 를 거쳐 나가는 순간 값이 사라지고,
// 공유하면 소비자가 store 안의 진실을 바꿀 수 있다.
func TestCloneCarriesRuntimeIdentityWithoutSharingIt(t *testing.T) {
	config, build := runtimeIdentityDigest("config"), runtimeIdentityDigest("build")
	source := WithRuntimeIdentity(DormantSnapshot(projectionNow), config, build)

	cloned := Clone(source)
	if cloned.Runtime.ConfigDigest == nil || *cloned.Runtime.ConfigDigest != config ||
		cloned.Runtime.BuildDigest == nil || *cloned.Runtime.BuildDigest != build {
		t.Fatalf("Clone 이 runtime identity 를 떨어뜨렸다: %+v", cloned.Runtime)
	}
	if cloned.Runtime.ConfigDigest == source.Runtime.ConfigDigest ||
		cloned.Runtime.BuildDigest == source.Runtime.BuildDigest {
		t.Fatal("Clone 이 같은 포인터를 넘겨 store 안의 진실이 밖에서 바뀔 수 있다")
	}
	*cloned.Runtime.ConfigDigest = runtimeIdentityDigest("tampered")
	if *source.Runtime.ConfigDigest != config {
		t.Fatal("소비자의 수정이 원본까지 갔다")
	}
}

// TestDormantSnapshotHasNoRuntimeIdentity 는 「엔진이 아닌 곳이 만든 스냅샷」의 핀이다.
//
// reader 가 아예 없을 때 화면이 만드는 dormant 스냅샷에는 엔진의 build 를 알 방법이
// 없다. 그때 콘솔이 자기 바이너리 값으로 채우면 그 숫자는 엔진이 거절할 매니페스트가 된다.
func TestDormantSnapshotHasNoRuntimeIdentity(t *testing.T) {
	for name, snapshot := range map[string]Snapshot{
		"dormant":     DormantSnapshot(projectionNow),
		"unavailable": UnavailableSnapshot(projectionNow),
	} {
		t.Run(name, func(t *testing.T) {
			if snapshot.Runtime.ConfigDigest != nil || snapshot.Runtime.BuildDigest != nil {
				t.Fatalf("엔진 밖에서 만든 스냅샷이 build 를 지어냈다: %+v", snapshot.Runtime)
			}
			if err := Validate(snapshot); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestRuntimeIdentityCrossesTheWireUnderItsPublishedNames 는 JSON 이름의 핀이다.
// OpenAPI 문서와 콘솔이 같은 이름을 쓴다.
func TestRuntimeIdentityCrossesTheWireUnderItsPublishedNames(t *testing.T) {
	config, build := runtimeIdentityDigest("config"), runtimeIdentityDigest("build")
	raw, err := json.Marshal(WithRuntimeIdentity(DormantSnapshot(projectionNow), config, build))
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Runtime struct {
			ConfigDigest *string `json:"configDigest"`
			BuildDigest  *string `json:"buildDigest"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.Runtime.ConfigDigest == nil || *wire.Runtime.ConfigDigest != config ||
		wire.Runtime.BuildDigest == nil || *wire.Runtime.BuildDigest != build {
		t.Fatalf("published 이름으로 값이 나가지 않았다: %s", raw)
	}
}

// TestMarketFailureOverlayKeepsTheRuntimeIdentity 는 **덮어쓰기 경로**의 핀이다.
//
// 엔진의 Read 는 identity 를 붙인 뒤 latch 된 시장을 WithMarketFailure 로 덮는다.
// 그 덮기가 envelope 를 새로 만들면서 identity 를 떨어뜨리면, 정확히 무언가 잘못된
// 상태에서 운영자가 숫자를 잃는다.
func TestMarketFailureOverlayKeepsTheRuntimeIdentity(t *testing.T) {
	config, build := runtimeIdentityDigest("config"), runtimeIdentityDigest("build")
	snapshot := WithRuntimeIdentity(currentPair(t), config, build)
	overlaid := WithMarketFailure(snapshot, MarketKR, RefusalRuntimeUnavailable, projectionNow.Add(time.Second))
	if err := Validate(overlaid); err != nil {
		t.Fatal(err)
	}
	if overlaid.Runtime.ConfigDigest == nil || *overlaid.Runtime.ConfigDigest != config ||
		overlaid.Runtime.BuildDigest == nil || *overlaid.Runtime.BuildDigest != build {
		t.Fatalf("시장 하나가 실패하자 프로세스 전체의 identity 가 사라졌다: %+v", overlaid.Runtime)
	}
}
