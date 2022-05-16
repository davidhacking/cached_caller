package cached_caller

import (
	"math"
	"sync/atomic"
)

type defaultTransCtrl struct {
	ctrlWindow       uint64
	errThreshold     float64
	timeoutThreshold float64
	errCnt           uint64
	timeoutCnt       uint64
	total            uint64
	degradeCnt       uint64
	errDegrade       uint64
	timeoutDegrade   uint64
}

func (d *defaultTransCtrl) errRateDegradeFlag() bool {
	errCnt := float64(d.errCnt)
	total := float64(d.total)
	if total <= 0 {
		return false
	}
	return errCnt/total > d.errThreshold
}

func (d *defaultTransCtrl) timeoutRateDegradeFlag() bool {
	timeoutCnt := float64(d.timeoutCnt)
	total := float64(d.total)
	if total <= 0 {
		return false
	}
	return timeoutCnt/total > d.timeoutThreshold
}

func (d *defaultTransCtrl) Report(t TransType) {
	atomic.AddUint64(&d.total, 1)
	total := d.total
	switch t {
	case TransTypeFail:
		atomic.AddUint64(&d.errCnt, 1)
	case TransTypeTimeout:
		atomic.AddUint64(&d.timeoutCnt, 1)
	}
	if d.errRateDegradeFlag() {
		atomic.AddUint64(&d.degradeCnt, d.errDegrade)
	}
	if d.timeoutRateDegradeFlag() {
		atomic.AddUint64(&d.degradeCnt, d.timeoutDegrade)
	}
	if total > d.ctrlWindow {
		atomic.StoreUint64(&d.total, 0)
		atomic.StoreUint64(&d.errCnt, 0)
		atomic.StoreUint64(&d.timeoutCnt, 0)
	}
}

func (d *defaultTransCtrl) Degrade() bool {
	dc := atomic.LoadUint64(&d.degradeCnt)
	if dc > 0 {
		return atomic.CompareAndSwapUint64(&d.degradeCnt, dc, dc-1)
	}
	return false
}

func (d *defaultTransCtrl) Init(config TransCtrlConfig) error {
	d.ctrlWindow = uint64(config.ctrlWindow)
	d.errDegrade = uint64(math.Ceil(config.errDegradeRate))
	d.timeoutDegrade = uint64(math.Ceil(config.timeoutDegradeRate))
	d.errThreshold = config.errThreshold
	d.timeoutThreshold = config.timeoutThreshold
	return nil
}
