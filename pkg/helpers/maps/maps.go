package maps

func Merge[M ~map[K]V, K comparable, V any](ms ...M) M {
	c := 0
	for _, m := range ms {
		c += len(m)
	}

	res := make(M, c)
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}

	return res
}
