package cached_caller

import (
	"context"
	"fmt"
	"time"

	"github.com/davidhacking/cached_caller/errors"
	"github.com/davidhacking/cached_caller/utils"
)

type cachedCallerImpl struct {
	caller Caller
	config *Config
}

func (c *cachedCallerImpl) realCall(ctx context.Context, timeout time.Duration,
	req Req) (rsp Rsp, delta time.Duration, err error) {
	start := time.Now()
	rsp, err = c.caller.Call(ctx, timeout, req)
	delta = time.Since(start)
	return rsp, delta, err
}

func (c *cachedCallerImpl) Call(ctx context.Context, timeout time.Duration, req Req) (rsp Rsp, err error) {
	cache := c.config.cache
	ctrl := c.config.transCtrl
	mon := c.config.monitor
	reqCodec := c.config.reqCodec
	mon.Inc("Enter")
	items, err := reqCodec.Encode(req)
	if err != nil {
		mon.Inc("EncodeFail")
		return nil, fmt.Errorf("reqCodec Encode failed, err=%v", err)
	}
	mon.Inc("cachedCallerEnterItems", len(items))
	if len(items) <= 0 {
		mon.Inc("EncodeNil")
		return nil, fmt.Errorf("reqCodec Encode return items is nil")
	}
	err = c.findItemByCache(items)
	if err != nil {
		mon.Inc("FindInCacheFail")
		return nil, fmt.Errorf("findItemByCache failed, err=%v", err)
	}
	if ctrl.Degrade() {
		mon.Inc("DegradeEnter")
		if cache == nil {
			mon.Inc("DegradeNoCache")
			return nil, fmt.Errorf("caller degrade with no cache return")
		}
		err = c.setItemsToCache(items, true)
		if err != nil {
			mon.Inc("setCacheFail")
			return nil, fmt.Errorf("setItemsToCache failed, err=%v", err)
		}
		rsp, err = c.items2Rsp(items)
		if err != nil {
			mon.Inc("items2RspFail")
			return nil, err
		}
		return rsp, nil
	}
	var delta time.Duration
	rsp, delta, err = c.realCall(ctx, timeout, req)
	c.reportTransCtrl(delta, timeout, err)
	if err != nil {
		mon.Inc("realCallFail")
		setCacheErr := c.setItemsToCache(items)
		if setCacheErr != nil {
			mon.Inc("setCacheFail")
			return nil, fmt.Errorf("setItemsToCache failed, err=%v", setCacheErr)
		}
		return nil, fmt.Errorf("realCall failed, err=%v", err)
	}
	err = c.rsp2Items(rsp, items)
	if err != nil {
		setCacheErr := c.setItemsToCache(items)
		if setCacheErr != nil {
			mon.Inc("setCacheFail")
			return nil, fmt.Errorf("setItemsToCache failed, err=%v", setCacheErr)
		}
		return nil, fmt.Errorf("rsp2Items failed, err=%v", err)
	}
	err = c.setItemsToCache(items)
	if err != nil {
		mon.Inc("setCacheFail")
		return nil, fmt.Errorf("setItemsToCache failed, err=%v", err)
	}
	return rsp, nil
}

func (c *cachedCallerImpl) CallInCache(req Req) (rsp Rsp, err error) {
	mon := c.config.monitor
	reqCodec := c.config.reqCodec
	mon.Inc("CallInCacheEnter")
	items, err := reqCodec.Encode(req)
	if err != nil {
		mon.Inc("EncodeFail")
		return nil, fmt.Errorf("reqCodec Encode failed, err=%v", err)
	}
	err = c.findItemByCache(items)
	if err != nil {
		mon.Inc("FindInCacheFail")
		return nil, fmt.Errorf("findItemByCache failed, err=%v", err)
	}
	rsp, err = c.items2Rsp(items)
	if err != nil {
		mon.Inc("items2RspFail")
		return nil, err
	}
	return rsp, nil
}

func (c *cachedCallerImpl) rsp2Items(rsp Rsp, items []*Item) error {
	rspCodec := c.config.rspCodec
	mon := c.config.monitor
	newItems, err := rspCodec.Encode(rsp)
	if err != nil {
		return fmt.Errorf("rspCodec Encode failed, err=%v", err)
	}
	m := make(map[string]*Item, len(newItems))
	invalid := 0
	for _, item := range newItems {
		if item.Key == "" {
			invalid++
			continue
		}
		m[item.Key] = item
	}
	mon.Inc("rsp2ItemsInvalidKey", invalid)
	invalidResp := 0
	for i, item := range items {
		newItem, ok := m[item.Key]
		if !ok {
			invalidResp++
			continue
		}
		items[i] = newItem
	}
	mon.Inc("rsp2ItemsInvalidResp", invalidResp)
	return nil
}

func (c *cachedCallerImpl) reportTransCtrl(delta time.Duration, timeout time.Duration, err error) {
	ctrl := c.config.transCtrl
	mon := c.config.monitor
	if delta > timeout {
		mon.Inc("recallCallTimeout")
		ctrl.Report(TransTypeTimeout)
		return
	}
	if err != nil {
		mon.Inc("recallCallFail")
		ctrl.Report(TransTypeFail)
		return
	}
	ctrl.Report(TransTypeSuccess)
}

func (c *cachedCallerImpl) items2Rsp(items []*Item) (rsp Rsp, err error) {
	rspCodec := c.config.rspCodec
	return rspCodec.Decode(items)
}

// findItemByCache cache 中的item会填充到items中
func (c *cachedCallerImpl) findItemByCache(items []*Item) error {
	cache := c.config.cache
	if cache == nil {
		return nil
	}
	itemCodec := c.config.itemCodec
	for i, item := range items {
		if !item.Empty() {
			continue
		}
		key, _, err := itemCodec.Encode(item)
		if err != nil {
			return fmt.Errorf("itemCodec Encode failed, err=%v", err)
		}
		value, err := cache.Get(key)
		if err != nil {
			continue
		}
		item, err = itemCodec.Decode(key, value)
		if err != nil {
			return fmt.Errorf("itemCodec Decode failed, err=%v", err)
		}
		items[i] = item
	}
	return nil
}

func (c *cachedCallerImpl) setItemsToCache(items []*Item, forceUpdateTs ...bool) error {
	cache := c.config.cache
	itemCodec := c.config.itemCodec
	flag := false
	if len(forceUpdateTs) > 0 {
		flag = true
	}
	for _, item := range items {
		if item.TS <= 0 || flag {
			item.TS = utils.NowTS()
		}
		key, value, err := itemCodec.Encode(item)
		if err != nil {
			return fmt.Errorf("itemCodec Encode failed, err=%v", err)
		}
		err = cache.Put(key, value)
		if err != nil {
			return fmt.Errorf("put cache failed, err=%v", err)
		}
	}
	return nil
}

func (c *cachedCallerImpl) Init(caller Caller, opts ...Option) error {
	c.caller = caller
	c.config = DefaultConfig()
	for _, opt := range opts {
		err := opt(c.config)
		if err != nil {
			return err
		}
	}
	if c.config.backgroundUpdater != nil {
		go c.backgroundUpdate()
	}
	return nil
}

func (c *cachedCallerImpl) backgroundUpdate() {
	duration := c.config.backgroundUpdateDuration
	cache := c.config.cache
	log := c.config.logger
	iterCache, ok := cache.(IterableCache)
	mon := c.config.monitor
	batchRequestNum := c.config.backgroundUpdateBatchNum
	if !ok {
		mon.Inc("cacheNotIterable")
		log.Errorf("cache not iterable can not backgroundUpdate")
		return
	}
	for range time.Tick(duration) {
		iter := iterCache.GetIter()
		items := c.getNeedUpdateItems(iter)
		for k := 0; k < (len(items)/batchRequestNum + 1); k++ {
			end := (k + 1) * batchRequestNum
			if end > len(items) {
				end = len(items)
			}
			c.updateCache(items[k*batchRequestNum : end])
		}
	}
}

func (c *cachedCallerImpl) updateCache(items []*Item) {
	reqCodec := c.config.reqCodec
	rspCodec := c.config.rspCodec
	timeout := c.config.backgroundUpdateTimeout
	req, err := reqCodec.Decode(items)
	mon := c.config.monitor
	log := c.config.logger
	if err != nil {
		mon.Inc("reqCodecDecodeFail")
		log.Errorf("reqCodec Decode failed, err=%v", err)
		return
	}
	rsp, delta, err := c.realCall(context.Background(), timeout, req)
	c.reportTransCtrl(delta, timeout, err)
	if err != nil {
		mon.Inc("bgRealCallFail")
		log.Errorf("realCall failed, err=%v", err)
		err = c.setItemsToCache(items, true)
		if err != nil {
			mon.Inc("bgSetCacheFail")
			log.Errorf("setItemsToCache failed, err=%v", err)
		}
		return
	}
	rspItems, err := rspCodec.Encode(rsp)
	if err != nil {
		mon.Inc("rspCodecEncodeFail")
		log.Errorf("rspCodec Encode failed, err=%v", err)
		err = c.setItemsToCache(items, true)
		if err != nil {
			mon.Inc("bgSetCacheFail")
			log.Errorf("setItemsToCache failed, err=%v", err)
		}
		return
	}
	err = c.setItemsToCache(rspItems)
	if err != nil {
		mon.Inc("bgSetCacheFail")
		log.Errorf("setItemsToCache failed, err=%v", err)
	}
}

func (c *cachedCallerImpl) getNeedUpdateItems(iter CacheIter) []*Item {
	log := c.config.logger
	itemCodec := c.config.itemCodec
	updateCheck := c.config.backgroundUpdater
	items := make([]*Item, 0, 1000)
	mon := c.config.monitor
	keyNum := 0
	for {
		key, value, err := iter.Next()
		keyNum++
		if err == errors.ErrCacheIterStop {
			break
		}
		if err != nil {
			mon.Inc("iterNextFail")
			log.Errorf("iter next failed, err=%v", err)
			break
		}
		item, err := itemCodec.Decode(key, value)
		if err != nil {
			mon.Inc("itemDecodeFail")
			log.Errorf("itemCodec Decode, err=%v", err)
			break
		}
		if updateCheck != nil && !updateCheck.NeedUpdate(item) {
			continue
		}
		items = append(items, item)
	}
	mon.Inc("cacheKeyNum", keyNum)
	mon.Inc("needUpdateKey", len(items))
	log.Debugf("getNeedUpdateItems cacheKeyNum=%v, needUpdateKey=%v", keyNum, len(items))
	return items
}

func NewCachedCaller() CachedCaller {
	return &cachedCallerImpl{}
}
