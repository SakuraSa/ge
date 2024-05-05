package gslice

// IsUniqe checks if the slice is unique.
func IsUniqe[T comparable](s []T) bool {
	m := make(map[T]struct{})
	for _, v := range s {
		if _, ok := m[v]; ok {
			return false
		}
		m[v] = struct{}{}
	}
	return true
}
