package cached_caller

import (
	"github.com/allegro/bigcache/v3"
	"github.com/davidhacking/cached_caller/errors"
)

type bigcacheIter struct {
	iter *bigcache.EntryInfoIterator
}

func (b *bigcacheIter) Next() (key string, value []byte, err error) {
	if !b.iter.SetNext() {
		return "", nil, errors.ErrCacheIterStop
	}
	entry, err := b.iter.Value()
	if err != nil {
		return "", nil, err
	}
	return entry.Key(), entry.Value(), nil
}
