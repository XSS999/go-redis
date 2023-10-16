package tcp

import (
	"context"
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

	ListenAndServe(listener, handler, closeChan)

	return nil
}

func ListenAndServe(listener net.Listener,
	handler tcp.Handler,
	closeChan <-chan struct{}) {

	// 通过chan来关闭
	go func() {
		<-closeChan
		logger.Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()

	//在退出时关闭Listener和Handler
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	var waitDone sync.WaitGroup
	ctx := context.Background()
	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		waitDone.Add(1)
		go func() {
			//这里用defer是因为如果在执行handler.Handler(ctx,conn)时如果出现了panic，也能正常执行
			defer func() {
				waitDone.Done()
			}()
			handler.Handler(ctx, conn)
		}()
	}

	waitDone.Wait()
}
