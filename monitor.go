package cached_caller

type defaultMonitor struct {
}

func (d *defaultMonitor) Inc(name string, n ...int) {
	return
}
