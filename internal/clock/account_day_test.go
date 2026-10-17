package clock

import (
	"testing"
	"time"
)

// a136 D1=D: 계좌 단위 하루는 KST 08:00 에 넘어감. 서울 날짜 D 의 KR 정규장과 뉴욕 날짜 D 의 US 정규장이
// 같은 라벨 D 를 받아야 하고, 서머타임 양쪽에서 US 정규장이 둘로 쪼개지지 않아야 함.
func TestA136AccountTradingDayRollsOverAtEightKST(t *testing.T) {
	seoul := mustLoc(t, "Asia/Seoul")
	cases := []struct {
		at   time.Time
		want string
	}{
		{time.Date(2026, 10, 14, 7, 59, 59, 0, seoul), "2026-10-13"},
		{time.Date(2026, 10, 14, 8, 0, 0, 0, seoul), "2026-10-14"},
		{time.Date(2026, 10, 14, 23, 59, 59, 0, seoul), "2026-10-14"},
		{time.Date(2026, 10, 15, 0, 0, 0, 0, seoul), "2026-10-14"},
	}
	for _, c := range cases {
		got, err := AccountTradingDay(c.at)
		if err != nil {
			t.Fatalf("AccountTradingDay(%s): %v", c.at, err)
		}
		if got != c.want {
			t.Fatalf("AccountTradingDay(%s) = %s, want %s", c.at, got, c.want)
		}
	}
}

func TestA136BothRegularSessionsOfDateDShareLabelD(t *testing.T) {
	// 서머타임(EDT, 2026-07-15)과 표준시(EST, 2026-12-15) 각각에서 두 시장 정규장의 첫 순간과 마지막 순간.
	for _, day := range []string{"2026-07-15", "2026-12-15"} {
		date, err := time.Parse("2006-01-02", day)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range []Market{MarketKR, MarketUS} {
			s, err := m.RegularSession(time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, mustLoc(t, tzName[m])))
			if err != nil {
				t.Fatal(err)
			}
			for _, at := range []time.Time{s.Open, s.Close.Add(-time.Second)} {
				got, err := AccountTradingDay(at)
				if err != nil {
					t.Fatal(err)
				}
				if got != day {
					t.Fatalf("%s session instant %s (%s) labelled %s, want %s", m, at, at.UTC(), got, day)
				}
			}
		}
	}
}

func mustLoc(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("LoadLocation(%s): %v", name, err)
	}
	return loc
}
