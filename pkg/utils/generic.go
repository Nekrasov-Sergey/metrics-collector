package utils

func Ptr[T any](v T) *T {
	return &v
}

func Deref[T any](p *T) (v T) {
	if p == nil {
		return v
	}
	return *p
}
