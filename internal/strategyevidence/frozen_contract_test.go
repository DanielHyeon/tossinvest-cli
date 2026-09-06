package strategyevidence

// testdata/official_contracts.json 을 **읽는 유일한 코드**다.
//
// 그 파일은 tasks.md 1.3 과 review.md 가 "공식 계약 동결"이라고 부르는 산출물인데,
// 저장소 어디에서도 읽히지 않았다(`grep -rn official_contracts --include='*.go'` → 없음).
// 값은 Go 상수로 다시 적혀 자기 자신과 비교되고 있었으므로, JSON 을 고쳐도 아무것도
// 깨지지 않고 상수를 고쳐도 JSON 과 대조되지 않았다. 동결이 아니라 장식이었다.
//
// 이제 계약의 숫자와 문장은 **영수증에서 읽어** 코드와 맞춘다.

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

type frozenOfficialContracts struct {
	FrozenAt string `json:"frozen_at"`
	SEC      struct {
		ContractID                string `json:"contract_id"`
		Endpoint                  string `json:"endpoint"`
		Schema                    string `json:"schema"`
		MaximumRequestsPerSecond  int    `json:"maximum_requests_per_second"`
		DeclaredUserAgentRequired bool   `json:"declared_user_agent_required"`
	} `json:"sec_edgar"`
	OpenDART struct {
		ContractID          string `json:"contract_id"`
		Endpoint            string `json:"endpoint"`
		Schema              string `json:"schema"`
		CredentialParameter string `json:"credential_parameter"`
		MaximumPageCount    int    `json:"maximum_page_count"`
	} `json:"opendart"`
	KRX struct {
		ContractID string `json:"contract_id"`
		Frozen     bool   `json:"official_programmatic_contract_frozen"`
	} `json:"krx"`
}

func frozenContracts(t *testing.T) frozenOfficialContracts {
	t.Helper()
	body, err := os.ReadFile("testdata/official_contracts.json")
	if err != nil {
		t.Fatal(err)
	}
	var contracts frozenOfficialContracts
	if err := json.Unmarshal(body, &contracts); err != nil {
		t.Fatal(err)
	}
	if contracts.FrozenAt == "" || contracts.SEC.ContractID == "" || contracts.OpenDART.ContractID == "" {
		t.Fatalf("frozen contract file did not decode: %+v", contracts)
	}
	return contracts
}

func TestMintedPoliciesMatchTheFrozenContractFile(t *testing.T) {
	t.Parallel()
	contracts := frozenContracts(t)

	if secContractID != contracts.SEC.ContractID {
		t.Fatalf("secContractID=%q, frozen contract file says %q", secContractID, contracts.SEC.ContractID)
	}
	if dartContractID != contracts.OpenDART.ContractID {
		t.Fatalf("dartContractID=%q, frozen contract file says %q", dartContractID, contracts.OpenDART.ContractID)
	}
	if maxOfficialPageSize != contracts.OpenDART.MaximumPageCount {
		t.Fatalf("maxOfficialPageSize=%d, frozen OpenDART page cap is %d", maxOfficialPageSize, contracts.OpenDART.MaximumPageCount)
	}

	sec, err := MintSourcePolicy(validSECConfig())
	if err != nil {
		t.Fatal(err)
	}
	if sec.EndpointIdentity != contracts.SEC.Endpoint || sec.SchemaVersion != contracts.SEC.Schema {
		t.Fatalf("SEC policy endpoint/schema = %q/%q, frozen file says %q/%q",
			sec.EndpointIdentity, sec.SchemaVersion, contracts.SEC.Endpoint, contracts.SEC.Schema)
	}
	dart, err := MintSourcePolicy(validDARTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if dart.EndpointIdentity != contracts.OpenDART.Endpoint || dart.SchemaVersion != contracts.OpenDART.Schema {
		t.Fatalf("OpenDART policy endpoint/schema = %q/%q, frozen file says %q/%q",
			dart.EndpointIdentity, dart.SchemaVersion, contracts.OpenDART.Endpoint, contracts.OpenDART.Schema)
	}
	if !dart.CredentialRequired || sec.CredentialRequired {
		t.Fatalf("credential requirement does not follow the frozen contracts: sec=%v dart=%v", sec.CredentialRequired, dart.CredentialRequired)
	}
}

func TestSECFairAccessRateComesFromTheFrozenContractFile(t *testing.T) {
	t.Parallel()
	contracts := frozenContracts(t)

	atCap := validSECConfig()
	atCap.MaxCalls = contracts.SEC.MaximumRequestsPerSecond
	atCap.AbsoluteCallWindow = time.Second
	if _, err := MintSourcePolicy(atCap); err != nil {
		t.Fatalf("SEC policy at the frozen fair-access rate (%d/s) was refused: %v", atCap.MaxCalls, err)
	}
	overCap := atCap
	overCap.MaxCalls = contracts.SEC.MaximumRequestsPerSecond + 1
	if _, err := MintSourcePolicy(overCap); !errors.Is(err, ErrSourceDisabled) {
		t.Fatalf("SEC policy above the frozen fair-access rate (%d/s) was minted: %v", overCap.MaxCalls, err)
	}
}

func TestSECDeclaredIdentityIsRequiredWhenTheContractSaysSo(t *testing.T) {
	t.Parallel()
	contracts := frozenContracts(t)
	if !contracts.SEC.DeclaredUserAgentRequired {
		t.Skip("frozen SEC contract no longer requires a declared identity")
	}
	// 빈 문자열이 아니라 **신원이 아닌 문자열**을 준다. 빈 문자열은 일반 필수-문자열
	// 검사가 먼저 거부하므로, 그것으로는 `validSECRequestIdentity` 를 지워도 시험이
	// 초록으로 남는다(2026-09-07 독립 리뷰 S7).
	anonymous := validSECConfig()
	anonymous.RequestIdentity = "anonymous-bot"
	if _, err := MintSourcePolicy(anonymous); !errors.Is(err, ErrSourceDisabled) {
		t.Fatalf("SEC policy without the contractually required declared identity was minted: %v", err)
	}
	blank := validSECConfig()
	blank.RequestIdentity = "   "
	if _, err := MintSourcePolicy(blank); !errors.Is(err, ErrSourceDisabled) {
		t.Fatalf("SEC policy with a blank declared identity was minted: %v", err)
	}
}

func TestOpenDARTCredentialParameterIsForbiddenInsideEvidencePayloads(t *testing.T) {
	t.Parallel()
	contracts := frozenContracts(t)
	if contracts.OpenDART.CredentialParameter == "" {
		t.Fatal("frozen OpenDART contract declares no credential parameter")
	}
	header := validHeader(marketclock.MarketKR, KindKRNetFlow, "credential-payload", "rev-1")
	header.Authority = AuthorityOpenDART
	header.Currency = "KRW"
	payload := `{"` + contracts.OpenDART.CredentialParameter + `":"leaked"}`
	if _, err := NewEnvelope(header, []byte(payload)); err == nil {
		t.Fatalf("evidence payload accepted the OpenDART credential parameter %q", contracts.OpenDART.CredentialParameter)
	}
}

func TestKRXStaysUnavailableWhileItsContractIsNotFrozen(t *testing.T) {
	t.Parallel()
	contracts := frozenContracts(t)
	if contracts.KRX.Frozen {
		t.Skip("KRX now has a frozen official programmatic contract")
	}
	krx := validSECConfig()
	krx.Authority = AuthorityKRX
	krx.ContractID = contracts.KRX.ContractID
	if _, err := MintSourcePolicy(krx); !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("KRX policy minted while its contract is unfrozen: %v", err)
	}
}
