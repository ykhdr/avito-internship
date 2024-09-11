package store

func trimSlice[T any](slice []T, limit, offset int) []T {
	if len(slice) < offset {
		return make([]T, 0)
	}
	end := offset + limit
	if end > len(slice) {
		end = len(slice)
	}
	return slice[offset:end]
}
