package tcp

import (
	"context"
	"fmt"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

func ListenAndServerWithSignal(cfg *Config, handler tcp.Handler) error {

	closeChan := make(chan struct{})
	// sigChan用来监听系统的关闭信号
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	// 如果监听到了系统的关闭信号，就给closeChan发送信号，关闭Listener和Handler
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)

	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)

	return nil
}

func ListenAndServe(listener net.Listener,
	handler tcp.Handler,
	closeChan <-chan struct{}) {

	// 通过chan来关闭
	go func() {
		<-closeChan //不需要返回值，因为是一个空的结构体，也就相当于是一个信号的作用
		logger.Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()

	//在退出时关闭Listener和Handler
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	// 等待所有的客户端handler完
	var waitDone sync.WaitGroup
	ctx := context.Background()
	for true {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("accept waring", err)
			break
		}
		logger.Info("accepted link")
		waitDone.Add(1)
		go func() {
			//这里用defer是因为如果在执行handler.Handler(ctx,conn)时如果出现了panic，也能正常执行
			defer func() {
				waitDone.Done()
			}()
			logger.Info("start client")
			handler.Handler(ctx, conn)
		}()
	}

	waitDone.Wait()
}
