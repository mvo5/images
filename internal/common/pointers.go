package common

func ToPtr[T any](x T) *T {
	return &x
}

func UnrefOrDefault[T any](p *T) T {
	if p == nil {
		var defaulT T
		return defaulT
	}
	return *p
}

func UnrefOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}
