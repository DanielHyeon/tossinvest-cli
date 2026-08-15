package console

// a109_a_degraded_surface_is_not_an_old_build_test.go — a109 §2.5 (design D3a-2).
//
// a108 의 강등 교리는 「콘솔·httpapi 의 전략 화면이 dormant 로 뜬다」는 세 번째 표면을
// 전제했다. 형제 endpoint 에는 그 표면이 없고, 대신 기존 소비자 메시지가 강등을 **다른
// 원인으로 단정**했다: 「엔진 control plane이 격리 해제를 제공하지 않는다」는 사실상
// 「빌드가 낡았다」는 뜻이다.
//
// 운영자는 그 말을 따라 엉뚱한 것을 고치고, 강등 원인은 결정적이므로 같은 상태가 그대로
// 재현된다. 수정은 문구의 정직화다 — 새 화면도, 새 디스크 상태도 만들지 않는다.
//
// # 왜 문구를 테스트가 잡는가
//
// 이 change 가 실제로 바꾼 것이 문구뿐이므로, 문구를 안 재면 이 task 는 **아무것도
// 재지 않는다.** 그리고 같은 문장이 세 경로에 있었다 — 셋 중 하나만 고치는 정정을
// 통과시키지 않는 것이 이 파일의 두 번째 일이다.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
)

// TestTheUnwiredQuarantineMessageDoesNotBlameTheBuild 는 문구 자체의 핀이다.
func TestTheUnwiredQuarantineMessageDoesNotBlameTheBuild(t *testing.T) {
	for _, required := range []string{"강등", "엔진 로그", "아무것도 변경되지 않았다"} {
		if !strings.Contains(quarantineUnwiredDetail, required) {
			t.Errorf("격리 해제 미배선 문구에 %q 가 없다 — 운영자가 강등 가능성과 "+
				"확인할 곳을 알 수 없다:\n%s", required, quarantineUnwiredDetail)
		}
	}
	// 원인 단정의 옛 문장이 남아 있으면 안 된다.
	if strings.Contains(quarantineUnwiredDetail, "엔진 control plane이 격리 해제를 제공하지 않는다") {
		t.Errorf("문구가 아직 원인을 단정한다 (빌드 탓) — 강등 부팅에서 거짓이다:\n%s",
			quarantineUnwiredDetail)
	}
}

// TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage 는 **값 단위 정정**의 핀이다.
//
// 설계는 이 문장을 한 자리(writeQuarantineError)만 인용했지만 사본은 셋이었다.
// 정정의 단위가 줄이면 나머지 둘이 살아남고, 운영자가 먼저 만나는 것은 preview 쪽이다.
func TestEveryQuarantineRefusalPathCarriesTheSameHonestMessage(t *testing.T) {
	console := &Console{}

	// ① writeQuarantineError 의 ErrUnwired 가지
	recorder := httptest.NewRecorder()
	console.writeQuarantineError(recorder, exitquarantine.ErrUnwired)
	assertQuarantineHonest(t, "writeQuarantineError", recorder)

	// ② release preview, ③ release apply — commander 가 없으면 미배선 경로다.
	for name, handler := range map[string]func(http.ResponseWriter, *http.Request){
		"handleQuarantineReleasePreview": console.handleQuarantineReleasePreview,
		"handleQuarantineReleaseApply":   console.handleQuarantineReleaseApply,
	} {
		recorder := httptest.NewRecorder()
		handler(recorder, httptest.NewRequest(http.MethodPost, "/quarantine", nil))
		assertQuarantineHonest(t, name, recorder)
	}
}

func assertQuarantineHonest(t *testing.T, path string, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("%s: status=%d, want %d — 미배선 경로가 아니다",
			path, recorder.Code, http.StatusNotImplemented)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "강등") || !strings.Contains(body, "엔진 로그") {
		t.Errorf("%s: 화면이 강등 가능성을 말하지 않는다 — 이 경로만 옛 문구로 남았다:\n%s",
			path, body)
	}
	// **제목도 값이다** — a109 §2-fix F5 (A2 P2-5).
	//
	// 본문(detail)은 강등을 말하게 고쳤는데 제목은 「판정 격리 command seam 미배선」
	// 그대로였고, 그것은 화면에서 **먼저 읽히는 줄**이다. seam 미배선은 이 저장소에서
	// 「이 빌드에 그 기능이 배선되지 않았다」는 뜻으로 쓰인다(overview.go 의
	// reasonSeamUnwired) — 강등 부팅에서는 거짓이고, 운영자는 빌드를 올리러 간다.
	if strings.Contains(body, "판정 격리 command seam 미배선") {
		t.Errorf("%s: 제목이 아직 빌드를 탓한다 — 본문은 강등을 말하는데 굵은 줄은 "+
			"「이 빌드에 없다」다:\n%s", path, body)
	}
	if !strings.Contains(body, "격리 해제 표면 없음") {
		t.Errorf("%s: 제목이 무엇이 없는지 말하지 않는다:\n%s", path, body)
	}
}
