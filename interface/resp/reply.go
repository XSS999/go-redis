package resp

// 代表回复
type Reply interface {
	// 转成字节，因为TCP是通过字节流去传输的
	ToBytes() []byte
}
