package cached_caller

import (
	"fmt"

	"github.com/davidhacking/cached_caller/utils"
)

type defaultLogger struct {
}

func (d *defaultLogger) Debugf(format string, args ...interface{}) {
	fmt.Println(utils.NowStr() + "debug: " + fmt.Sprintf(format, args...))
}

func (d *defaultLogger) Errorf(format string, args ...interface{}) {
	fmt.Println(utils.NowStr() + "error: " + fmt.Sprintf(format, args...))
}
