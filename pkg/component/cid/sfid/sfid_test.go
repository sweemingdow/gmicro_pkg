package sfid

import "testing"

func BenchmarkNext(b *testing.B) {
	Setting(true)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Next()
		}
	})
}

func TestNext(t *testing.T) {
	Setting(false)

	for i := 0; i < 10; i++ {
		println(Next())
	}
}
