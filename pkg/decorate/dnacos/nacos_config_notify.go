package dnacos

import "sync"

type ConfigureObserveFunc func(dataId string, val any)

var (
	rwMu      sync.RWMutex
	observers = make(map[string][]ConfigureObserveFunc)
)

func RegisterObserver(dataId string, cof ConfigureObserveFunc) {
	rwMu.Lock()
	observers[dataId] = append(observers[dataId], cof)
	rwMu.Unlock()
}

func Notify(dataId string, val any) {
	rwMu.RLock()
	defer rwMu.RUnlock()

	if obsSli, ok := observers[dataId]; ok {
		for _, observer := range obsSli {
			observer(dataId, val)
		}
	}
}

func UnregisterAll() {
	rwMu.Lock()

	for k := range observers {
		delete(observers, k)
	}

	rwMu.Unlock()
}
