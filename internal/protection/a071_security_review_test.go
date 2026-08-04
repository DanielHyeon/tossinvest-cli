package protection

import (
	"os"
	"strings"
	"testing"
)

func TestProtectionPackageExportsNoArbitraryGatewayFactoryAuthority(t *testing.T) {
	data, err := os.ReadFile("supervisor.go")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatal(err)
	}
	for _, forbidden := range []string{"type GatewayFactory", "Gateway GatewayFactory", "gateway GatewayFactory"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("arbitrary gateway authority remains exported or stored: %q", forbidden)
		}
	}
}
