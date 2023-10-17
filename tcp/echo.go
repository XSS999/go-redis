package tcp

import (
	"bufio"
	"context"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

/*
*
你给我发什么，我给你回复过去
*/

// 客户端信息
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// 关闭客户端
func (e *EchoClient) Close() error {
	//等10s，让客户端把一些没做完的东西做完
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

// 接收客户端
type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

// 处理一个新来的客户端
func (handler *EchoHandler) Handler(ctx context.Context, conn net.Conn) {
	if handler.closing.Get() {
		_ = conn.Close()
	}

	// 把客户端包装成内部的EchoClient
	client := &EchoClient{
		Conn: conn,
	}

	// 把新来的客户端存到map里面去
	handler.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	// 不断的收业务传过来的报文
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			// 手动关闭
			if err == io.EOF {
				logger.Info("Conneting close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 我在处理写的业务
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("handler shutting down")
	handler.closing.Set(true)

	// 关掉所有的客户端
	handler.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		//是否继续遍历下一个
		return true
	})
	return nil
}
