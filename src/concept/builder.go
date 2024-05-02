package concept

type Builder[T any] interface {
	Build() (T, error)
}
