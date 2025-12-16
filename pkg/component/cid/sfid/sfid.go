package sfid

var (
	needPerformance bool
)

func Setting(performance bool) {
	if performance {
		// 150ns, but more resource
		settingSonyFlake(32)
	} else {
		// 250ns, less resource
		settingSnowflake(0)
	}

	needPerformance = performance
}

func Next() int64 {
	if needPerformance {
		return sonyFlakeNext()
	}

	return snowflakeNext()
}
