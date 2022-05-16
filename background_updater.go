package cached_caller

import (
	"time"

	"github.com/davidhacking/cached_caller/utils"
)

type defaultBackgroundUpdater struct {
}

func (v *defaultBackgroundUpdater) NeedUpdate(item *Item) bool {
	if utils.NowTS()-item.TS > int64((10*time.Minute)/time.Second) {
		return true
	}
	return false
}
