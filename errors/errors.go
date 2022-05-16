package errors

import (
	"fmt"
)

var (
	// ErrCacheIterStop 缓存遍历终止
	ErrCacheIterStop = fmt.Errorf("cache iter stop")
)
