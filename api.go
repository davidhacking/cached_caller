package cached_caller

import (
	"context"
	"io"
	"time"
)

// Req 请求体定义
type Req interface{}

// Rsp 返回体定义
type Rsp interface{}

// Caller 请求方法
type Caller interface {
	Call(ctx context.Context, timeout time.Duration, req Req) (rsp Rsp, err error)
}

// CachedCaller 带缓存的调用
type CachedCaller interface {
	Caller
	Init(caller Caller, opts ...Option) error
}

// ReqCodec 请求包编解码
type ReqCodec interface {
	Encode(req Req) (items []*Item, err error)
	Decode(items []*Item) (req Req, err error)
}

// RspCodec 回包编解码
type RspCodec interface {
	Encode(rsp Rsp) (items []*Item, err error)
	Decode(items []*Item) (rsp Rsp, err error)
}

// ItemCodec 缓存item编解码
type ItemCodec interface {
	Encode(item *Item) (key string, value []byte, err error)
	Decode(key string, value []byte) (item *Item, err error)
}

// CacheDumpable 缓存dump到磁盘
type CacheDumpable interface {
	Dump(writer io.Writer) error
	FromDump(reader io.Reader) error
}

// Caching 缓存接口定义
type Caching interface {
	Get(key string) (value []byte, err error)
	Put(key string, value []byte) error
	Del(key string) error
}

// IterableCache 获取缓存遍历接口
type IterableCache interface {
	GetIter() CacheIter
}

// CacheIter 遍历缓存方法
type CacheIter interface {
	Next() (key string, value []byte, err error)
}

// BackgroundUpdater 缓存后台更新方法
type BackgroundUpdater interface {
	NeedUpdate(item *Item) bool
}

// Logger 日志接口定义（TODO 保证日志打印行号正确）
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Monitor 监控接口定义
type Monitor interface {
	Inc(name string, n ...int)
}

// TransType 调用结果
type TransType int

const (
	// TransTypeSuccess 成功
	TransTypeSuccess TransType = iota
	// TransTypeFail 失败
	TransTypeFail
	// TransTypeTimeout 超时
	TransTypeTimeout
)

// TransCtrl 拥塞控制保护下游
type TransCtrl interface {
	Report(t TransType)
	Degrade() bool
	Init(config TransCtrlConfig) error
}

// FindInCacheCaller 提供直接方法缓存获取结果方法
type FindInCacheCaller interface {
	CallInCache(req Req) (rsp Rsp, err error)
}
