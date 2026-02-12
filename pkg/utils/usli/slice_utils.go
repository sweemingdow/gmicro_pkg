package usli

import (
	"fmt"
	"strconv"
)

type Predicate[T any] func(val T) bool

func FindFirstIf[T any](src []T, pred Predicate[T]) (T, bool) {
	for _, t := range src {
		if pred(t) {
			return t, true
		}
	}

	var zero T
	return zero, false
}

func RemoveFirstIf[T any](src []T, pred Predicate[T]) []T {
	for idx, t := range src {
		if pred(t) {
			return Remove[T](src, idx)
		}
	}

	return src
}

func Remove[T any](src []T, idx int) []T {
	return append(src[:idx], src[idx+1:]...)
}

// 用最后一个元素替代, 原切片的顺序会改变
func RemoveFast[T any](src []T, idx int) []T {
	src[idx] = src[len(src)-1]
	return src[:len(src)-1]
}

type Converter[T, R any] func(T) R

// eg: []int => []string
func Conv[T, R any](src []T, ct Converter[T, R]) []R {
	if len(src) == 0 {
		return []R{}
	}

	dest := make([]R, len(src))
	for i, t := range src {
		dest[i] = ct(t)
	}

	return dest
}

func A2iConvMust(src []string) []int {
	return Conv(src, func(t string) int {
		i, err := strconv.Atoi(t)
		if err != nil {
			panic(err)
		}

		return i
	})
}

func A2iConv(src []string) ([]int, error) {
	if len(src) == 0 {
		return []int{}, nil
	}

	result := make([]int, len(src))
	for i, s := range src {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid int value:%s: %w", s, err)
		}

		result[i] = v
	}

	return result, nil
}

func I2aConv(src []int) []string {
	return Conv(src, func(t int) string {
		return strconv.Itoa(t)
	})
}

type KeyMapper[T any, K comparable] func(T) K
type ValueMapper[T any, K comparable, V any] func(T, K) V

func ToMap[T any, K comparable, V any](src []T, km KeyMapper[T, K], vm ValueMapper[T, K, V]) map[K]V {
	m := make(map[K]V, len(src))
	for _, t := range src {
		k := km(t)
		m[k] = vm(t, k)
	}

	return m
}

func ToMapAir[T comparable](src []T) map[T]struct{} {
	m := make(map[T]struct{}, len(src))
	for _, t := range src {
		m[t] = struct{}{}
	}

	return m
}

func ToItMap[T any, K comparable](src []T, km KeyMapper[T, K]) map[K]T {
	return ToMap(
		src,
		km,
		func(t T, k K) T {
			return t
		},
	)
}

type FilterFunc[T any] func(T) bool

func Filter[T any](src []T, f FilterFunc[T]) []T {
	if len(src) == 0 {
		return make([]T, 0)
	}

	newSli := make([]T, 0, len(src))

	for _, t := range src {
		if f(t) {
			newSli = append(newSli, t)
		}
	}

	return newSli
}

func GroupBy[T any, K comparable, V any](src []T, km KeyMapper[T, K], vm ValueMapper[T, K, V]) map[K][]V {
	if len(src) == 0 {
		return make(map[K][]V)
	}

	m := make(map[K][]V)
	for _, t := range src {
		k := km(t)
		v := vm(t, k)

		if _, ok := m[k]; !ok {
			m[k] = []V{v}
		} else {
			m[k] = append(m[k], v)
		}
	}

	return m
}

func GroupByIt[T any, K comparable](src []T, km KeyMapper[T, K]) map[K][]T {
	return GroupBy(src, km, func(t T, k K) T {
		return t
	})
}

func Distinct[T comparable](src []T) []T {
	n := len(src)
	if n == 0 {
		return []T{}
	} else if n == 1 {
		return []T{src[0]}
	}

	m := make(map[T]bool, n)
	estimate := n * 2 / 3
	if estimate == 0 {
		estimate = 1
	}
	result := make([]T, 0, estimate)

	for _, t := range src {
		if !m[t] {
			m[t] = true
			result = append(result, t)
		}
	}

	// shrink capacity
	if len(result) > 128 && cap(result) > len(result)*2 {
		trimmed := make([]T, len(result))
		copy(trimmed, result)
		return trimmed
	}

	return result
}

type DiffFilter[T any, K comparable] func(item T) K

func Diff[T any, K comparable](src []T, keys []K, df DiffFilter[T, K]) []T {
	exclude := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		exclude[k] = struct{}{}
	}

	newItems := make([]T, 0, len(src)-len(keys))
	for _, item := range src {
		if _, ok := exclude[df(item)]; !ok {
			newItems = append(newItems, item)
		}
	}

	return newItems
}
