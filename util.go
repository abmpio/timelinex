package timelinex

func removeSliceByIndex[T any](v []T, index int) []T {
	return append(v[:index], v[index+1:]...)
}
