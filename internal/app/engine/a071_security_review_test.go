package engine

import (
	"os"
	"strings"
	"testing"
)

func TestUnprovenFillLifecycleKeepsBothProductionAssembliesUnwired(t *testing.T) {
	build, _ := productionProtectionDigests()
	assemblies := productionProtectionAssemblies(build)
	if len(assemblies) != 2 {
		t.Fatalf("assemblies=%+v", assemblies)
	}
	for _, assembly := range assemblies {
		if assembly.Wired {
			t.Fatalf("%s claimed WIRED without a committed-fill lifecycle consumer", assembly.Market)
		}
	}
}

func TestEngineHasNoArbitraryProtectionGatewayFactoryOrProtectionDBStartupDependency(t *testing.T) {
	data, err := os.ReadFile("gateway.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, forbidden := range []string{"protectionofficial.New", "protection.NewSupervisor", "protection.db", "GatewayFactory"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("production engine still contains unproven protection authority %q", forbidden)
		}
	}
}
