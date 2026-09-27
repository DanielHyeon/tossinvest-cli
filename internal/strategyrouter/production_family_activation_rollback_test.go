package strategyrouter

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 태스크 8.7.2: 활성화가 없거나 만료·폐기되면 entry worker 를 OFF 로 되돌린다.
//
// 이 패키지가 지는 몫은 둘이다. (1) "배포가 이 시장에 활성화를 선언하지 않았다" 를
// 다른 모든 실패와 **구별되는 답**으로 낸다 — 엔진은 그 답일 때만 기존 경로를 쓰고,
// 나머지 실패에서는 네 가족을 OFF 로 세운다. (2) 주문 lease 가 활성화 수명을 넘어
// 살지 못하게 하는 상한 규칙을 한 곳에 둔다.

// 핀이 비어 있다는 것만이 "미선언" 이다. 다른 결함은 그 답을 낼 수 없다.
//
// 대조군이 요점이다: 핀이 있으면 무엇이 틀렸든(형식·파일·폐기·만료) 미선언이 아니다.
// 그 경계가 흐려지면 사람이 선언한 활성화의 고장이 "기존 경로로 돌아가라" 로 읽힌다 —
// 8.7.2 가 닫는 바로 그 넓힘이다.
func TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketKR)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	valid := fixture.config(MarketKR, body, data) // 파일은 아직 없다
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	var absent context.Context

	for name, pin := range map[string]string{"empty": "", "spaces": "   ", "tab": "\t\n"} {
		config := valid
		config.ManifestDigest = pin
		// ctx 가 없거나 취소돼도, 다른 결속 값이 비어 있어도 답은 같아야 한다 —
		// 오늘 생산(핀 없음)의 답이 주기의 사정에 따라 흔들리면 안 된다.
		broken := config
		broken.CalibrationDigest, broken.RouteManifestDigest = "", ""
		for label, call := range map[string]func() (FamilyActivation, error){
			"background": func() (FamilyActivation, error) { return LoadProductionFamilyActivation(context.Background(), config) },
			"nil ctx":    func() (FamilyActivation, error) { return LoadProductionFamilyActivation(absent, config) },
			"cancelled":  func() (FamilyActivation, error) { return LoadProductionFamilyActivation(cancelled, config) },
			"no bindings": func() (FamilyActivation, error) {
				return LoadProductionFamilyActivation(context.Background(), broken)
			},
		} {
			activation, err := call()
			if !errors.Is(err, ErrProductionFamilyActivationUndeclared) {
				t.Errorf("%s/%s: err=%v, want undeclared", name, label, err)
			}
			if activation.Verified() {
				t.Errorf("%s/%s: an undeclared activation is verified", name, label)
			}
		}
	}

	// 핀이 있으면 미선언이 아니다.
	malformed := valid
	malformed.ManifestDigest = "sha256:not-a-digest"
	declared := map[string]func() error{
		"malformed pin": func() error {
			_, err := LoadProductionFamilyActivation(context.Background(), malformed)
			return err
		},
		"pinned file missing": func() error {
			_, err := LoadProductionFamilyActivation(context.Background(), valid)
			return err
		},
		"revoked": func() error {
			revoked := body
			revoked.Revoked = true
			_, err := LoadProductionFamilyActivation(context.Background(), fixture.write(t, MarketKR, revoked))
			return err
		},
		"expired": func() error {
			expired := body
			expired.IssuedAt = fixture.now.Add(-2 * time.Hour).Format(time.RFC3339Nano)
			expired.ApprovedAt = expired.IssuedAt
			expired.ExpiresAt = fixture.now.Add(-time.Hour).Format(time.RFC3339Nano)
			_, err := LoadProductionFamilyActivation(context.Background(), fixture.write(t, MarketKR, expired))
			return err
		},
	}
	for _, name := range []string{"malformed pin", "pinned file missing", "revoked", "expired"} {
		err := declared[name]()
		if err == nil {
			t.Fatalf("%s: a declared but broken activation loaded", name)
		}
		if errors.Is(err, ErrProductionFamilyActivationUndeclared) {
			t.Errorf("%s: a declared activation answered undeclared (%v)", name, err)
		}
	}
}

// lease 상한은 줄어들기만 한다.
//
// 세 경우를 함께 세운다: 검증 안 된 값(상한 그대로), 수명이 상한보다 긴 값(상한 그대로),
// 수명이 상한보다 짧은 값(수명으로 깎임). 그리고 수명이 다 됐으면 만료 오류다. 하나만
// 세우면 "언제나 상한을 돌려준다" 나 "언제나 수명을 돌려준다" 판본이 통과한다.
func TestTheLeaseCeilingOnlyEverShrinks(t *testing.T) {
	const ceiling = 30 * time.Second
	var zero FamilyActivation
	if got, err := zero.LeaseCeiling(time.Now(), ceiling); err != nil || got != ceiling {
		t.Fatalf("zero activation: ceiling=%v err=%v, want %v unchanged", got, err, ceiling)
	}

	fixture := newFamilyActivationFixture(t)
	activation, err := LoadProductionFamilyActivation(context.Background(), fixture.write(t, MarketKR, fixture.body(MarketKR)))
	if err != nil {
		t.Fatal(err)
	}
	expires := activation.ExpiresAt()
	for name, entry := range map[string]struct {
		at   time.Time
		want time.Duration
	}{
		"수명이 상한보다 길다": {expires.Add(-time.Hour), ceiling},
		"수명이 상한보다 짧다": {expires.Add(-10 * time.Second), 10 * time.Second},
		"수명이 상한과 같다":  {expires.Add(-ceiling), ceiling},
		"수명이 1ns 남았다": {expires.Add(-time.Nanosecond), time.Nanosecond},
	} {
		got, err := activation.LeaseCeiling(entry.at, ceiling)
		if err != nil || got != entry.want {
			t.Errorf("%s: ceiling=%v err=%v, want %v", name, got, err, entry.want)
		}
	}
	// 전수 쓸기: 결과는 입력 상한도, 남은 수명도 넘지 않는다.
	for offset := -2 * ceiling; offset <= 2*ceiling; offset += 250 * time.Millisecond {
		for _, input := range []time.Duration{time.Second, ceiling, time.Minute} {
			at := expires.Add(-offset)
			got, err := activation.LeaseCeiling(at, input)
			if err != nil {
				continue
			}
			if got > input || got > expires.Sub(at) || got <= 0 {
				t.Fatalf("offset=%v input=%v: ceiling=%v grew past the input or the remaining life", offset, input, got)
			}
		}
	}
	for name, at := range map[string]time.Time{"만료 시각": expires, "만료 뒤": expires.Add(time.Second)} {
		if got, err := activation.LeaseCeiling(at, ceiling); !errors.Is(err, ErrProductionFamilyActivationExpired) || got != 0 {
			t.Errorf("%s: ceiling=%v err=%v, want 0 + expired", name, got, err)
		}
	}
}

// 적재 시점의 만료 판정과 lease 시점의 만료 판정은 **같은 경계**다.
//
// 둘이 1ns 라도 갈리면 적재가 받아들인 활성화를 lease 가 거절하거나(무해) 적재가
// 거절한 순간에 lease 가 살아 있다(유해). 규칙이 한 곳이라는 것을 경계 세 점에서 잰다.
func TestLoadingAndLeasingJudgeExpiryAtTheSameInstant(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketKR)
	config := fixture.write(t, MarketKR, body)
	loaded, err := LoadProductionFamilyActivation(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	expires := loaded.ExpiresAt()
	for _, at := range []time.Time{expires.Add(-time.Nanosecond), expires, expires.Add(time.Nanosecond)} {
		observed := config
		observed.ObservedAt = at
		_, loadErr := LoadProductionFamilyActivation(context.Background(), observed)
		_, leaseErr := loaded.LeaseCeiling(at, 30*time.Second)
		if errors.Is(loadErr, ErrProductionFamilyActivationExpired) != errors.Is(leaseErr, ErrProductionFamilyActivationExpired) {
			t.Errorf("at %v: load=%v lease=%v — the two judgements disagree", at, loadErr, leaseErr)
		}
	}
}

// 만료 판정은 이 패키지에 **한 곳**뿐이다.
//
// 행동 시험은 같은 규칙을 두 번 적은 판본을 못 가른다 — 둘이 같은 답을 내는 한 초록이다.
// 그런데 그 상태가 곧 "판정이 둘이면 반증이 죽는다" 이다: 한쪽을 지우는 변이를 다른 쪽이
// 통과시킨다. 그래서 만료 오류를 **돌려주는** 자리를 비시험 파일 전체에서 센다.
func TestExpiryIsJudgedInExactlyOnePlace(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	holders := []string{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				ident, ok := node.(*ast.Ident)
				if ok && ident.Name == "ErrProductionFamilyActivationExpired" {
					holders = append(holders, name+":"+function.Name.Name)
				}
				return true
			})
		}
	}
	if len(holders) != 1 || !strings.HasSuffix(holders[0], ":familyActivationRemaining") {
		t.Fatalf("만료 오류를 내는 자리=%v, want [production_family_activation.go:familyActivationRemaining] 하나", holders)
	}
}
