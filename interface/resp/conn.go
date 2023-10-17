package resp

/*
*
代表一个客户端的连接
*/
type Connection interface {
	//给客户端去回复消息
	Write([]byte) error

	//查询当前客户端用的哪个DB
	GetDBIndex() int

	// 选择DB
	SelectDB(int)
}
