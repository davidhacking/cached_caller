package cached_caller

type Item struct {
	TS   int64
	Key  string
	Data []byte
}

func (i *Item) Empty() bool {
	return len(i.Data) == 0
}
