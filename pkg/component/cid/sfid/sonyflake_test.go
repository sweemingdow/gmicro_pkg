package sfid

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkSonyFlakeNext(b *testing.B) {
	settingSonyFlake(16)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sonyFlakeNext()
		}
	})
}

func BenchmarkSonyFlakeNextRepeat(b *testing.B) {
	settingSonyFlake(32)
	b.ResetTimer()

	var dupCount atomic.Int64
	seen := make(map[int64]struct{}, b.N)
	var mu sync.Mutex

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		localSeen := make(map[int64]struct{})

		for pb.Next() {
			id := sonyFlakeNext()

			// 先本地检查
			if _, exists := localSeen[id]; exists {
				dupCount.Add(1)
				continue
			}
			localSeen[id] = struct{}{}

			// 全局检查
			mu.Lock()
			if _, exists := seen[id]; exists {
				dupCount.Add(1)
			} else {
				seen[id] = struct{}{}
			}
			mu.Unlock()
		}
	})

	b.ReportMetric(float64(dupCount.Load())/float64(b.N)*100, "%duplicate")
}
