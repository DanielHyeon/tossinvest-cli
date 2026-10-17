package main

// screen.go 는 1.6 Jev 선별(최종 표본 선택 전 판단)을 담당함 — shadow 측정 전용, 주문 경로 없음.
//
// 규칙 jev_screen<K>_of_random<POOL>_of_top100_le_order_cap:
//  1. 상한 필터를 통과한 상위 100 후보에서 무작위 POOL(기본 30, 후보가 POOL 이하이면 전부)을 뜸.
//  2. 풀의 모든 종목에 J1·J2 를 물음(종목당 요청 하나 — 기존과 같음).
//  3. J2(악재) p ≥ 0.5 인 종목은 제외, 나머지를 J1(상승) p 내림차순(동률은 랭킹 순)으로 세워 상위 K(기본 10)를
//     결선으로 표시함. 생존이 K 미만이면 결선도 그만큼만.
//
// 측정 설계: 결선만이 아니라 풀 전체의 판단을 원장에 쓰고 라벨링함 — 비결선이 같은 사이클·같은 질문의
// 대조군이 되어, 보고서가 결선과 비결선의 실현 적중을 나란히 비교할 수 있게 하기 위함.
// 판단 실패 종목은 gap 행이고 선별 대상이 아님(p 가 없으므로).

import "sort"

const (
	// screenUpQuestion·screenRedFlagQuestion 은 선별이 읽는 두 질문 id(probeQuestions 와 같음 — 시험이 고정).
	screenUpQuestion      = "j1_up_within_horizon"
	screenRedFlagQuestion = "j2_hold_off_red_flag"
	// redFlagExclusionP 이상인 J2 는 악재로 보고 결선에서 제외함(경계 0.5 자체도 제외).
	redFlagExclusionP = 0.5
)

// screenInput 은 풀에서 판단에 성공한 종목 하나의 두 p — 랭킹 순서로 줌(동률 판정이 그 순서를 씀).
type screenInput struct {
	J1, J2 float64
}

// screenRecord 는 judgment 행에 붙는 선별 결과(1.6 이후). 같은 종목의 J1·J2 행은 같은 값을 가짐.
type screenRecord struct {
	Finalist bool `json:"finalist"`
	// Rank 는 악재 제외 생존 종목 중 J1 순위(1부터). 악재로 제외된 종목은 0.
	Rank            int  `json:"screen_rank"`
	RedFlagExcluded bool `json:"red_flag_excluded"`
}

// screenFinalists 는 입력과 같은 순서·길이로 선별 결과를 돌려주는 순수 함수.
// 생존 종목을 J1 내림차순으로 안정 정렬함 — 같은 J1 이면 입력(랭킹) 순서가 앞선 쪽이 위.
func screenFinalists(judged []screenInput, k int) []screenRecord {
	out := make([]screenRecord, len(judged))
	var survivors []int
	for index, in := range judged {
		if in.J2 >= redFlagExclusionP {
			out[index].RedFlagExcluded = true
			continue
		}
		survivors = append(survivors, index)
	}
	sort.SliceStable(survivors, func(a, b int) bool { return judged[survivors[a]].J1 > judged[survivors[b]].J1 })
	for position, index := range survivors {
		out[index].Rank = position + 1
		out[index].Finalist = position < k
	}
	return out
}
