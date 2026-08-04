package strategyprojection

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestStorePublishesOnlyValidatedImmutablePairedSnapshots(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	initial := DormantSnapshot(now)
	store, err := NewStore(initial)
	if err != nil {
		t.Fatal(err)
	}
	invalid := Clone(initial)
	delete(invalid.Markets, MarketUS)
	if err := store.Replace(invalid); err == nil {
		t.Fatal("accepted incomplete paired snapshot")
	}
	got, err := store.Read(context.Background())
	if err != nil || !reflect.DeepEqual(got, initial) {
		t.Fatalf("read=%+v err=%v", got, err)
	}
	delete(got.Markets, MarketKR)
	again, _ := store.Read(context.Background())
	if !reflect.DeepEqual(again, initial) {
		t.Fatal("caller mutation escaped into store")
	}

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				if read, readErr := store.Read(context.Background()); readErr != nil || Validate(read) != nil {
					t.Errorf("concurrent read invalid: %+v %v", read, readErr)
					return
				}
			}
		}()
	}
	wg.Wait()
}
