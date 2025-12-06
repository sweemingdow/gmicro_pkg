package utils

import "testing"

func BenchmarkRandStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStr(32)
	}
}

func BenchmarkRandStrParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandStr(32)
		}
	})
}

func BenchmarkRandInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandInt(1, 100)
	}
}

func BenchmarkRandIntParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandInt(1, 100)
		}
	})
}
