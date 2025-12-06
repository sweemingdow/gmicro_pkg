package observer

type ConfigurationChangeListener[T any] func(val T)

var (
	configurationChangeListeners map[string]int
)
