package collections

type Iterator[V any] interface {
	Next() (V, bool)
}

type Iterable[V any] interface {
	Iterator() Iterator[V]
}
