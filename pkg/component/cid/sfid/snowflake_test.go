package sfid

import "testing"

func BenchmarkSnowflakeNext(b *testing.B) {
	settingSnowflake(0)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			snowflakeNext()
		}
	})
}
