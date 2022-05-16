package cached_caller

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/davidhacking/cached_caller/utils"
	"github.com/stretchr/testify/assert"
)

const (
	fid1      = 1
	fid1Value = int64(1)
)

type FeatureCenterServer struct {
}

type ItemInfo struct {
	ItemID  string
	TS      int64
	Feature map[int32]*Feature
}

func (i *ItemInfo) Marshal() (data []byte, err error) {
	return json.Marshal(i)
}

func (i *ItemInfo) Unmarshal(data []byte) (err error) {
	return json.Unmarshal(data, i)
}

type Feature struct {
	IntVal int64
}

type FeatureRequest struct {
	userInfo  *Feature
	itemInfos []*ItemInfo
}

type FeatureResponse struct {
	userInfo  *Feature
	itemInfos []*ItemInfo
}

func (f *FeatureCenterServer) GetFeature(req *FeatureRequest) (rsp *FeatureResponse, err error) {
	rsp = &FeatureResponse{itemInfos: make([]*ItemInfo, 0, len(req.itemInfos))}
	for _, item := range req.itemInfos {
		item.Feature = map[int32]*Feature{
			fid1: {IntVal: fid1Value},
		}
		rsp.itemInfos = append(rsp.itemInfos, item)
	}
	return
}

type FeatureReqCodec struct {
}

func (v *FeatureReqCodec) Encode(req Req) (items []*Item, err error) {
	fReq := req.(*FeatureRequest)
	items = make([]*Item, 0, len(fReq.itemInfos))
	for _, i := range fReq.itemInfos {
		if err != nil {
			return nil, err
		}
		items = append(items, &Item{
			Key: i.ItemID,
		})
	}
	return items, nil
}

func (v *FeatureReqCodec) Decode(items []*Item) (req Req, err error) {
	fReq := &FeatureRequest{itemInfos: make([]*ItemInfo, 0, len(items))}
	for _, i := range items {
		fItem := &ItemInfo{}
		err = fItem.Unmarshal(i.Data)
		if err != nil {
			return nil, err
		}
		fItem.ItemID = i.Key
		fReq.itemInfos = append(fReq.itemInfos, fItem)
	}
	return fReq, nil
}

type FeatureRspCodec struct {
}

func (v *FeatureRspCodec) Encode(rsp Rsp) (items []*Item, err error) {
	fRsp := rsp.(*FeatureResponse)
	items = make([]*Item, 0, len(fRsp.itemInfos))
	for _, i := range fRsp.itemInfos {
		data, err := i.Marshal()
		if err != nil {
			return nil, err
		}
		items = append(items, &Item{
			Key:  i.ItemID,
			Data: data,
		})
	}
	return items, nil
}

func (v *FeatureRspCodec) Decode(items []*Item) (rsp Rsp, err error) {
	fRsp := &FeatureResponse{itemInfos: make([]*ItemInfo, 0, len(items))}
	for _, i := range items {
		fItem := &ItemInfo{}
		fItem.ItemID = i.Key
		fItem.TS = i.TS
		fRsp.itemInfos = append(fRsp.itemInfos, fItem)
		if len(i.Data) <= 0 {
			continue
		}
		err = fItem.Unmarshal(i.Data)
		if err != nil {
			return nil, err
		}
		fItem.ItemID = i.Key
		fItem.TS = i.TS
	}
	return fRsp, nil
}

type VideoFeatureCaller struct {
	client *FeatureCenterServer
}

func (v *VideoFeatureCaller) Call(ctx context.Context, timeout time.Duration, req Req) (rsp Rsp, err error) {
	return v.client.GetFeature(req.(*FeatureRequest))
}

type VideoBackgroundUpdater struct {
}

func (v *VideoBackgroundUpdater) NeedUpdate(item *Item) bool {
	if utils.NowTS()-item.TS > 5 {
		return true
	}
	return false
}

func TestCachedCaller(t *testing.T) {
	c := NewCachedCaller()
	opts := []Option{
		WithReqCodec(&FeatureReqCodec{}),
		WithRspCodec(&FeatureRspCodec{}),
		WithCache(NewBigCaching(bigcache.DefaultConfig(10 * time.Second))),
		WithBackgroundUpdater(&VideoBackgroundUpdater{}),
		WithBackgroundUpdateDuration(5 * time.Second),
	}
	err := c.Init(&VideoFeatureCaller{}, opts...)
	assert.Nil(t, err)
	id := "111"
	req := &FeatureRequest{
		itemInfos: []*ItemInfo{
			{
				ItemID: id,
			},
		},
	}
	rsp, err := c.Call(context.Background(), time.Second, req)
	assert.Nil(t, err)
	cRsp, ok := rsp.(*FeatureResponse)
	assert.True(t, ok)
	assert.Equal(t, cRsp.itemInfos[0].ItemID, id)
	assert.Equal(t, cRsp.itemInfos[0].Feature[fid1].IntVal, fid1Value)
}

func TestCallInCache(t *testing.T) {
	c := NewCachedCaller()
	opts := []Option{
		WithReqCodec(&FeatureReqCodec{}),
		WithRspCodec(&FeatureRspCodec{}),
		WithCache(NewBigCaching(bigcache.DefaultConfig(10 * time.Second))),
		WithBackgroundUpdater(&VideoBackgroundUpdater{}),
		WithBackgroundUpdateDuration(5 * time.Second),
	}
	err := c.Init(&VideoFeatureCaller{}, opts...)
	assert.Nil(t, err)
	id := "111"
	req := &FeatureRequest{
		itemInfos: []*ItemInfo{
			{
				ItemID: id,
			},
		},
	}
	rsp, err := c.Call(context.Background(), time.Second, req)
	assert.Nil(t, err)
	cRsp, ok := rsp.(*FeatureResponse)
	assert.True(t, ok)
	assert.Equal(t, cRsp.itemInfos[0].ItemID, id)
	assert.Equal(t, cRsp.itemInfos[0].Feature[fid1].IntVal, fid1Value)

	c2, ok := c.(FindInCacheCaller)
	assert.True(t, ok)
	rsp2, err2 := c2.CallInCache(&FeatureRequest{
		itemInfos: []*ItemInfo{
			{
				ItemID: id,
			},
		},
	})
	cRsp2, ok := rsp2.(*FeatureResponse)
	assert.True(t, ok)
	assert.Nil(t, err2)
	assert.Equal(t, cRsp2.itemInfos[0].ItemID, id)
	assert.Equal(t, cRsp2.itemInfos[0].Feature[fid1].IntVal, fid1Value)
}

func TestBackgroundUpdate(t *testing.T) {
	c := NewCachedCaller()
	opts := []Option{
		WithReqCodec(&FeatureReqCodec{}),
		WithRspCodec(&FeatureRspCodec{}),
		WithCache(NewBigCaching(bigcache.DefaultConfig(10 * time.Second))),
		WithBackgroundUpdater(&VideoBackgroundUpdater{}),
		WithBackgroundUpdateDuration(2 * time.Second),
	}
	err := c.Init(&VideoFeatureCaller{}, opts...)
	assert.Nil(t, err)
	id := "111"
	req := &FeatureRequest{
		itemInfos: []*ItemInfo{
			{
				ItemID: id,
			},
		},
	}
	rsp, err := c.Call(context.Background(), time.Second, req)
	assert.Nil(t, err)
	cRsp, ok := rsp.(*FeatureResponse)
	assert.True(t, ok)
	assert.Equal(t, cRsp.itemInfos[0].ItemID, id)
	assert.Equal(t, cRsp.itemInfos[0].Feature[fid1].IntVal, fid1Value)

	var getTS = func() int64 {
		c2, ok := c.(FindInCacheCaller)
		assert.True(t, ok)
		rsp2, err2 := c2.CallInCache(&FeatureRequest{
			itemInfos: []*ItemInfo{
				{
					ItemID: id,
				},
			},
		})
		cRsp2, ok := rsp2.(*FeatureResponse)
		assert.True(t, ok)
		assert.Nil(t, err2)
		assert.Equal(t, cRsp2.itemInfos[0].ItemID, id)
		assert.Equal(t, cRsp2.itemInfos[0].Feature[fid1].IntVal, fid1Value)
		return cRsp2.itemInfos[0].TS
	}
	ts1 := getTS()
	time.Sleep(10 * time.Second)
	ts2 := getTS()
	assert.True(t, ts2-ts1 > 5)
}

type VideoFeatureCaller2 struct {
	callCnt int
}

func (v *VideoFeatureCaller2) Call(ctx context.Context, timeout time.Duration, req Req) (rsp Rsp, err error) {
	time.Sleep(100 * time.Millisecond)
	v.callCnt++
	return nil, fmt.Errorf("fake error")
}

func TestTransCtrl(t *testing.T) {
	c := NewCachedCaller()
	opts := []Option{
		WithReqCodec(&FeatureReqCodec{}),
		WithRspCodec(&FeatureRspCodec{}),
		WithCache(NewBigCaching(bigcache.DefaultConfig(10 * time.Second))),
		WithBackgroundUpdater(&VideoBackgroundUpdater{}),
		WithBackgroundUpdateDuration(2 * time.Second),
	}
	vc2 := &VideoFeatureCaller2{}
	err := c.Init(vc2, opts...)
	assert.Nil(t, err)
	id := "111"
	req := &FeatureRequest{
		itemInfos: []*ItemInfo{
			{
				ItemID: id,
			},
		},
	}
	rsp, err := c.Call(context.Background(), 200*time.Millisecond, req)
	assert.NotNil(t, err)
	assert.Nil(t, rsp)
	rsp, err = c.Call(context.Background(), 200*time.Millisecond, req)
	assert.NotNil(t, rsp)
	assert.Nil(t, err)
}
