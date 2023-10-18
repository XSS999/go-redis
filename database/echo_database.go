package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

type EchoDatabase struct {
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) Close() {

}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {

}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}
