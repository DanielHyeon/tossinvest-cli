package strategyflow

// a112 태스크 2.1 「rejection of … mismatched bindings」 의 빈칸(2.x 대조 감사): 기존 시험은 불일치 축으로 Desired 하나만 바꿨다. 구조체 등식이
// 나머지를 막지만 그것을 재는 시험이 없었다. 여기서는 Descriptor 의 **모든** 열쇠 아닌 필드를 반사로 열거해 하나씩 바꾼다 — 필드가 늘면 그
// 필드도 자동으로 들어온다(손으로 고른 목록이 아님). LaneID 는 열쇠라 「unknown」 축(기존 시험)이 맡는다.

import (
	"reflect"
	"testing"
)

func TestEveryDescriptorFieldOtherThanTheKeyIsPartOfTheBinding(t *testing.T) {
	full := Descriptors()
	descriptorType := reflect.TypeOf(Descriptor{})
	checked := 0
	for index := range full {
		for field := 0; field < descriptorType.NumField(); field++ {
			name := descriptorType.Field(field).Name
			if name == "LaneID" {
				continue
			}
			values := append([]Descriptor(nil), full...)
			target := reflect.ValueOf(&values[index]).Elem().Field(field)
			if target.Kind() != reflect.String {
				t.Fatalf("Descriptor.%s is a %s — teach this census how to drift it", name, target.Kind())
			}
			target.SetString(target.String() + "-drift")
			if err := ValidateDescriptors(values); err == nil {
				t.Fatalf("descriptor %s with drifted %s was accepted", full[index].LaneID, name)
			}
			checked++
		}
	}
	if want := len(full) * (descriptorType.NumField() - 1); checked != want || checked == 0 {
		t.Fatalf("checked %d field drifts, want %d", checked, want)
	}
}
