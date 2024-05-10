package slices

import "k8s.io/apimachinery/pkg/util/sets"

func Contains[T comparable](xs []T, v T) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func ContainsFunc[T any](xs []T, f func(T) bool) bool {
	for _, x := range xs {
		if f(x) {
			return true
		}
	}
	return false
}

func Filter[T any](array []T, keepFunc func(T) bool) []T {
	res := make([]T, 0, len(array))

	for i := range array {
		if keepFunc(array[i]) {
			res = append(res, array[i])
		}
	}

	return res
}

func Find[T any](xs []T, f func(T) bool) (T, int, bool) {
	for i := range xs {
		if f(xs[i]) {
			return xs[i], i, true
		}
	}

	return *new(T), 0, false
}

func Map[T any, U any, S ~[]T](xs S, f func(T) U) []U {
	ys := make([]U, 0, len(xs))

	for _, x := range xs {
		ys = append(ys, f(x))
	}

	return ys
}

func Nub[T comparable, S ~[]T](xs S) S {
	return sets.New[T](xs...).UnsortedList()
}
