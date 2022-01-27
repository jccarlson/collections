package gopherbox

type Container[T any] interface {
	Len() int
	Has(T) bool
}
