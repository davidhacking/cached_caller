package cached_caller

import (
	"bytes"
	"encoding/binary"
)

type defaultItemCodec struct {
}

func (i *defaultItemCodec) Encode(item *Item) (key string, value []byte, err error) {
	length := 8 + 4 + len(item.Data)
	value = make([]byte, length, length)
	buf := bytes.NewBuffer(value[:0])
	var uint32Buff = make([]byte, 4)
	var uint64Buff = make([]byte, 8)
	binary.BigEndian.PutUint64(uint64Buff, uint64(item.TS))
	buf.Write(uint64Buff)
	binary.BigEndian.PutUint32(uint32Buff, uint32(len(item.Data)))
	buf.Write(uint32Buff)
	buf.Write(item.Data)
	return item.Key, value, nil
}

func (i *defaultItemCodec) Decode(key string, value []byte) (item *Item, err error) {
	item = &Item{}
	buf := bytes.NewBuffer(value)
	item.Key = key
	item.TS = int64(binary.BigEndian.Uint64(buf.Next(8)))
	item.Data = buf.Next(int(binary.BigEndian.Uint32(buf.Next(4))))
	return item, nil
}
