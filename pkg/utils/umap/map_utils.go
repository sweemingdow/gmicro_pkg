package umap

func Flat[K comparable, V any](m map[K]V) []any {
	flat := make([]any, len(m)*2)
	idx := 0

	for k, v := range m {
		flat[idx] = k
		flat[idx+1] = v
		idx += 2
	}

	return flat
}

func ToSli[K comparable, V any](m map[K]V) []V {
	sli := make([]V, len(m))
	idx := 0

	for _, v := range m {
		sli[idx] = v
		idx++
	}

	return sli
}

func ToSliWithMap[K comparable, V, E any](m map[K]V, mapper Mapper[K, V, E]) []E {
	sli := make([]E, len(m))
	idx := 0

	for k, v := range m {
		sli[idx] = mapper(k, v)
		idx++
	}

	return sli
}

func KeyToSli[K comparable, V any](m map[K]V) []K {
	sli := make([]K, len(m))
	idx := 0

	for k, _ := range m {
		sli[idx] = k
		idx++
	}

	return sli
}

type Mapper[K, T, E any] func(key K, val T) E

func Map[K comparable, V, E any](m map[K]V, mp Mapper[K, V, E]) map[K]E {
	newMap := make(map[K]E, len(m))

	for k, v := range m {
		newMap[k] = mp(k, v)
	}

	return newMap
}

type Predicate[K comparable, V any] func(key K, val V) bool

func FindFirstIf[K comparable, V any](m map[K]V, pred Predicate[K, V]) (V, bool) {
	for k, v := range m {
		if pred(k, v) {
			return v, true
		}
	}

	var zero V
	return zero, false
}
