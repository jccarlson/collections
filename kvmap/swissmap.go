package kvmap

type swissMapEntry[K any, V any] struct {
	key   K
	value V
}

type swissMetadata uint8

const (
	presentMask = 0x80
	hashMask    = 0x7F
)

func (md swissMetadata) isPresent() bool {
	return md&presentMask != 0
}

func (md swissMetadata) hashMatch(h uint64) bool {
	return md&hashMask == swissMetadata(h)&hashMask
}

type swissMetadataTable [16]swissMetadata
