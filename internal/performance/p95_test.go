package performance

import (
	"slices"
	"testing"
	"time"
)

// p95Samples 는 nearest-rank p95 가 최댓값과 갈라지는 최소 표본 수임.
// n=7 에서는 p95 인덱스가 6(최댓값)이라 부하 중 단발 이상치 하나에 실패함 — 표본 20 으로 통계 정의를 맞춤.
const p95Samples = 20

// p95Index 는 정렬된 n 개 표본에서 nearest-rank p95 의 0 기반 인덱스(ceil(0.95n)-1)임.
func p95Index(n int) int {
	return (n*95+99)/100 - 1
}

// nearestRankP95 는 durations 를 제자리 정렬한 뒤 nearest-rank p95 를 반환함.
// 표본이 p95Samples 미만이면 p95 가 곧 최댓값이 되므로 시험 작성 오류로 보고 panic 함.
func nearestRankP95(durations []time.Duration) time.Duration {
	if len(durations) < p95Samples {
		panic("nearest-rank p95 needs at least p95Samples durations")
	}
	slices.Sort(durations)
	return durations[p95Index(len(durations))]
}

func TestNearestRankP95IndexMakesTwentyTheSmallestNonMaximumSampleCount(t *testing.T) {
	// p95Samples 미만의 모든 n 에서 p95 는 최댓값이어야 하고, p95Samples 에서 처음으로 최댓값을 벗어나야 함
	for n := 1; n < p95Samples; n++ {
		if got := p95Index(n); got != n-1 {
			t.Fatalf("p95Index(%d)=%d want %d (maximum)", n, got, n-1)
		}
	}
	if got := p95Index(p95Samples); got >= p95Samples-1 {
		t.Fatalf("p95Index(%d)=%d still selects the maximum", p95Samples, got)
	}
}

func TestNearestRankP95AbsorbsOneOutlierButNotTwo(t *testing.T) {
	samples := func(n, outliers int) []time.Duration {
		durations := make([]time.Duration, n)
		for index := range durations {
			durations[index] = time.Millisecond
		}
		// 이상치를 앞쪽에 두어 정렬을 거쳐야만 올바른 값이 나오게 함
		for index := range outliers {
			durations[index] = time.Hour
		}
		return durations
	}
	// 1M-row 시험이 쓰는 25 도 같은 성질을 가져야 함
	for _, n := range []int{p95Samples, 25} {
		if got := nearestRankP95(samples(n, 1)); got != time.Millisecond {
			t.Fatalf("one outlier in %d samples: p95=%s want 1ms", n, got)
		}
		if got := nearestRankP95(samples(n, 2)); got != time.Hour {
			t.Fatalf("two outliers in %d samples: p95=%s want 1h", n, got)
		}
	}
	// 표본 부족은 조용히 최댓값을 내지 않고 거절해야 함
	defer func() {
		if recover() == nil {
			t.Fatal("nearestRankP95 accepted fewer than p95Samples durations")
		}
	}()
	nearestRankP95(make([]time.Duration, p95Samples-1))
}
