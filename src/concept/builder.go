package concept

// Builder is an interface that defines a builder concept.
type Builder[T any] interface {
	Build() (T, error)
}
