package cached_caller

import (
	"time"
)

type Config struct {
	reqCodec                    ReqCodec
	rspCodec                    RspCodec
	itemCodec                   ItemCodec
	cacheDumpDuration           time.Duration
	cache                       Caching
	backgroundUpdater           BackgroundUpdater
	backgroundUpdateBatchNum    int
	backgroundUpdateParallelNum int
	backgroundUpdateTimeout     time.Duration
	backgroundUpdateDuration    time.Duration
	monitor                     Monitor
	transCtrl                   TransCtrl
	transCtrlConfig             TransCtrlConfig
	logger                      Logger
}

type TransCtrlConfig struct {
	ctrlWindow         int // 控制窗口大小
	errThreshold       float64
	timeoutThreshold   float64
	errDegradeRate     float64 // 例如失败率为0.1则 失败降级率 = errDegradeRate*0.1
	timeoutDegradeRate float64 // 整体降级率为 失败降级率+超时降级率
}

func DefaultConfig() *Config {
	transCtrl := &defaultTransCtrl{}
	_ = transCtrl.Init(TransCtrlConfig{
		ctrlWindow:         1e5,
		errThreshold:       0.01,
		timeoutThreshold:   0.01,
		errDegradeRate:     2,
		timeoutDegradeRate: 2,
	})
	return &Config{
		itemCodec:                   &defaultItemCodec{},
		cache:                       NewBigCaching(),
		backgroundUpdater:           &defaultBackgroundUpdater{},
		backgroundUpdateBatchNum:    100,
		backgroundUpdateParallelNum: 5,
		backgroundUpdateTimeout:     10 * time.Second,
		backgroundUpdateDuration:    10 * time.Minute,
		monitor:                     &defaultMonitor{},
		transCtrl:                   transCtrl,
		logger:                      &defaultLogger{},
	}
}

// Option is a function that takes a config struct and modifies it
type Option func(cfg *Config) error

func WithReqCodec(reqCodec ReqCodec) Option {
	return func(cfg *Config) error {
		cfg.reqCodec = reqCodec
		return nil
	}
}

func WithRspCodec(rspCodec RspCodec) Option {
	return func(cfg *Config) error {
		cfg.rspCodec = rspCodec
		return nil
	}
}

func WithCache(cache Caching) Option {
	return func(cfg *Config) error {
		cfg.cache = cache
		return nil
	}
}

func WithBackgroundUpdater(backgroundUpdater BackgroundUpdater) Option {
	return func(cfg *Config) error {
		cfg.backgroundUpdater = backgroundUpdater
		return nil
	}
}

func WithBackgroundUpdateDuration(backgroundUpdateDuration time.Duration) Option {
	return func(cfg *Config) error {
		cfg.backgroundUpdateDuration = backgroundUpdateDuration
		return nil
	}
}
