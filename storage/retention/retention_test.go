package retention

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/platform"
	"github.com/influxdata/platform/mock"
)

func TestService_Open(t *testing.T) {
	t.Run("negative interval", func(t *testing.T) {
		service := NewService(nil, nil, -1, 0)
		defer service.Close()
		if err := service.Open(); err == nil {
			t.Fatal("didn't get error, expected one")
		}
	})

	t.Run("disabled service", func(t *testing.T) {
		service := NewService(mock.NewStore(), mock.NewBucketService(), 0, 0)
		defer service.Close()
		if err := service.Open(); err != nil {
			t.Fatalf("got error %v", err)
		}

		// Since service disabled the wait group shouldn't be waiting on anything.
		done := make(chan struct{})
		go func() {
			service.wg.Wait()
			close(done)
		}()

		timeout := time.NewTimer(time.Second)
		select {
		case <-timeout.C:
			t.Fatal("test timed out waiting for wait group")
		case <-done:
			// Pass
		}
	})

	t.Run("idempotency", func(t *testing.T) {
		service := NewService(mock.NewStore(), mock.NewBucketService(), 0, 0)
		defer service.Close()
		if err := service.Open(); err != nil {
			t.Fatalf("got error %v", err)
		}

		// Re-opening an opened service is a no-op.
		if err := service.Open(); err != nil {
			t.Fatalf("got error %v", err)
		}
	})
}

func TestService_Close(t *testing.T) {
	t.Run("idempotency", func(t *testing.T) {
		service := NewService(mock.NewStore(), mock.NewBucketService(), 1, 0)
		if err := service.Open(); err != nil {
			t.Fatalf("got error %v", err)
		}

		if err := service.Close(); err != nil {
			t.Fatalf("got error %v", err)
		}

		// Re-closing an closed service is a no-op.
		if err := service.Close(); err != nil {
			t.Fatalf("got error %v", err)
		}
	})
}

func TestService_expireData(t *testing.T) {
	service := NewService(nil, mock.NewBucketService(), 0, 0)
	shard := mock.NewShard()
	now := time.Date(2018, 4, 10, 23, 12, 33, 0, time.UTC)

	t.Run("no rpByBucketID", func(t *testing.T) {
		if err := service.expireData(shard, nil, now); err != nil {
			t.Error(err)
		}

		if err := service.expireData(shard, map[string]time.Duration{}, now); err != nil {
			t.Error(err)
		}
	})

	// Generate some measurement names
	var names [][]byte
	rpByBucketID := map[string]time.Duration{}
	expMatchedFrequencies := map[string]int{}  // To be used for verifying test results.
	expRejectedFrequencies := map[string]int{} // To be used for verifying test results.
	for i := 0; i < 10; i++ {
		repeat := rand.Intn(10) + 1 // [1, 10]
		name := genMeasurementName()
		for j := 0; j < repeat; j++ {
			names = append(names, name)
		}

		_, bucketID, err := platform.ReadMeasurement(name)
		if err != nil {
			t.Fatal(err)
		}

		// Put half the rpByBucketID into the set to delete and half into the set
		// to not delete.
		if i%2 == 0 {
			rpByBucketID[string(bucketID)] = 3 * time.Hour
			expMatchedFrequencies[string(name)] = repeat
		} else {
			expRejectedFrequencies[string(name)] = repeat
		}
	}

	// Add a badly formatted measurement.
	for i := 0; i < 5; i++ {
		names = append(names, []byte("zyzwrong"))
	}
	expRejectedFrequencies["zyzwrong"] = 5

	gotMatchedFrequencies := map[string]int{}
	gotRejectedFrequencies := map[string]int{}
	shard.DeleteSeriesRangeWithPredicateFn = func(_ tsdb.SeriesIterator, fn func([]byte, models.Tags) (int64, int64, bool)) error {

		// Iterate over the generated names updating the frequencies by which
		// the predicate function in expireData matches or rejects them.
		for _, name := range names {
			from, to, shouldDelete := fn(name, nil)
			if shouldDelete {
				gotMatchedFrequencies[string(name)]++
				if from != math.MinInt64 {
					return fmt.Errorf("got from %d, expected %d", from, math.MinInt64)
				}
				wantTo := now.Add(-3 * time.Hour).UnixNano()
				if to != wantTo {
					return fmt.Errorf("got to %d, expected %d", to, wantTo)
				}
			} else {
				gotRejectedFrequencies[string(name)]++
			}
		}
		return nil
	}

	t.Run("multiple bucket", func(t *testing.T) {
		if err := service.expireData(shard, rpByBucketID, now); err != nil {
			t.Error(err)
		}

		// Verify that the correct series were marked to be deleted.
		t.Run("matched", func(t *testing.T) {
			if !reflect.DeepEqual(gotMatchedFrequencies, expMatchedFrequencies) {
				t.Fatalf("got\n%#v\nexpected\n%#v", gotMatchedFrequencies, expMatchedFrequencies)
			}
		})

		t.Run("rejected", func(t *testing.T) {
			// Verify that badly formatted measurements were rejected.
			if !reflect.DeepEqual(gotRejectedFrequencies, expRejectedFrequencies) {
				t.Fatalf("got\n%#v\nexpected\n%#v", gotRejectedFrequencies, expRejectedFrequencies)
			}
		})
	})
}

// genMeasurementName generates a random measurement name or panics.
func genMeasurementName() []byte {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
