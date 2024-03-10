package geecache

/*只能读，不可修改，因为lru有缓存添加，删除;这只是一个可视的结构，返回的是备份*/
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v *ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}
func (v ByteView) String() string {
	return string(v.b)
}
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
