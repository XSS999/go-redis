package cluster

import (
	pool "github.com/jolestar/go-commons-pool/v2"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistenthash"
)

type ClusterDatabase struct {
	self string

	node       []string
	peerPicker *consistenthash.NodeMap
	// 保存多个连接池
	peerConnection map[string]*pool.ObjectPool
	db             database.Database
}

func MakeClusterDatabase() *ClusterDatabase {

}
func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	//TODO implement me
	panic("implement me")
}

func (c *ClusterDatabase) Close() {
	//TODO implement me
	panic("implement me")
}

func (c *ClusterDatabase) AfterClientClose(conn resp.Connection) {
	//TODO implement me
	panic("implement me")
}
