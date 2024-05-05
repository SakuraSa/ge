package concept

// AOP is an interface that defines an aspect-oriented programming (AOP) concept.
type AOP interface {
	Apply(TaskFunc) TaskFunc
}

// AOPs is a slice of AOP.
type AOPs []AOP

// Apply applies all AOPs to a TaskFunc.
func (a AOPs) Apply(f TaskFunc) TaskFunc {
	for _, aop := range a {
		f = aop.Apply(f)
	}
	return f
}
