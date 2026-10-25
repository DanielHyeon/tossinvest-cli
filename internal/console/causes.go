package console

import (
	"reflect"
	"strings"
)

// causes.go 는 "읽지 못한 원인은 화면당 한 번만 설명한다" 규칙(change a173)의 구현임.
//
// 칸마다 같은 문장을 붙이면 운영자는 다른 원인인지 다시 읽어야 함. 그래서 칸은 "미측정 · 사유 코드"
// 만 남기고, 설명은 화면 머리의 배너 한 곳에 모음. 배너 재료는 페이지 뷰 모델을 훑어 얻음 —
// 핸들러마다 목록을 손으로 만들면 새 칸이 생길 때 배너가 빠지는 자리가 생기기 때문임.

// causeBanner 는 한 사유 코드의 설명 한 덩어리임.
type causeBanner struct {
	Code     string
	Sentence string
	// Details 는 같은 코드 아래 서로 다른 구체 사유(파싱 오류·보유 종목 등). 중복은 제거함.
	Details []string
	// Href 는 그 값을 채우는 화면이 있을 때만 채움.
	Href      string
	HrefLabel string
}

// causeRank 는 배너 순서임(열거 순서 — 시험이 세기 쉽도록 고정).
var causeRank = []unmeasuredReason{
	reasonJournalUnreadable, reasonBrokerReadFailed, reasonVerifySuspended, reasonSeamUnwired,
	reasonNeverFetched, reasonConfigUnreadable, reasonJournalValueUnparsable, reasonDiscoveryUnreadable,
}

// causeLinks 는 사유별로 값을 채워 주는 화면임.
var causeLinks = map[unmeasuredReason][2]string{
	reasonJournalUnreadable: {pathVerifyConsole, "검증 콘솔에서 엔진 시작"},
	reasonBrokerReadFailed:  {"/positions", "포지션 화면에서 마지막 실패 사유 보기"},
	reasonNeverFetched:      {"/positions", "포지션 화면 열기"},
	reasonConfigUnreadable:  {pathSettingsDaily, "설정 화면 열기"},
}

var readingType = reflect.TypeOf(reading{})

// causesOf 는 뷰 모델 안의 미측정 reading 을 사유 코드별로 모아 배너 목록을 돌려줌.
// 뷰 모델은 템플릿이 읽는 값이라 필드가 전부 내보내진(exported) 것이어야 하고, 내보내지지 않은
// 필드는 템플릿도 못 읽으므로 건너뜀.
func causesOf(view any) []causeBanner {
	found := map[unmeasuredReason][]string{}
	walkReadings(reflect.ValueOf(view), 0, map[uintptr]bool{}, func(r reading) {
		if r.Known() || r.reason == "" {
			return
		}
		list := found[r.reason]
		if d := strings.TrimSpace(r.detail); d != "" {
			for _, have := range list {
				if have == d {
					return
				}
			}
			list = append(list, d)
		}
		found[r.reason] = list
	})
	var out []causeBanner
	for _, code := range causeRank {
		details, ok := found[code]
		if !ok {
			continue
		}
		b := causeBanner{Code: string(code), Sentence: unmeasuredSentences[code], Details: details}
		if link, has := causeLinks[code]; has {
			b.Href, b.HrefLabel = link[0], link[1]
		}
		out = append(out, b)
	}
	return out
}

// walkReadings 는 구조체·슬라이스·맵·포인터를 따라 내려가며 reading 을 찾음. 깊이와 포인터 방문으로
// 순환을 막음.
func walkReadings(v reflect.Value, depth int, seen map[uintptr]bool, visit func(reading)) {
	if !v.IsValid() || depth > 12 {
		return
	}
	switch v.Kind() {
	case reflect.Interface:
		walkReadings(v.Elem(), depth+1, seen, visit)
	case reflect.Pointer:
		if v.IsNil() || seen[v.Pointer()] {
			return
		}
		seen[v.Pointer()] = true
		walkReadings(v.Elem(), depth+1, seen, visit)
	case reflect.Struct:
		if v.Type() == readingType {
			if v.CanInterface() {
				visit(v.Interface().(reading))
			}
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				walkReadings(v.Field(i), depth+1, seen, visit)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkReadings(v.Index(i), depth+1, seen, visit)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			walkReadings(v.MapIndex(k), depth+1, seen, visit)
		}
	}
}
