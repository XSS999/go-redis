package aof

import (
	"go-redis/config"
	databaseface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
	"sync"
)

type CmdLine = [][]byte

const aofQueueSize = 1 << 16 //避免魔法值 65535

type payload struct {
	cmdLine CmdLine //指令本身
	dbIndex int     //写入那个DB
}

type AofHandler struct {
	db          databaseface.Database //Redis核心
	aofChan     chan *payload         //写文件的一个缓冲区，文件要落入到硬盘中，速度较慢，需要加Chan
	aofFile     *os.File              //后期读取appendonly.aof文件
	aofFilename string
	aofFinished chan struct{}
	pausingAof  sync.RWMutex
	currentDB   int //记录指令保存到那个DB
}

func NewAofHandler(database databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename //找到配置文件的文件名
	handler.db = database
	handler.loadAof()

	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600) //入参依次是，文件名，flag（只读，只写，追加），文件模式
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	// channel
	handler.aofChan = make(chan *payload, aofQueueSize)

	go func() {
		handler.handleAof()
	}()

	return handler, err
}

// Add payload(set k v) --> aofChan
func (handler *AofHandler) AddAof(dbIndex int, cmd CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmd,
			dbIndex: dbIndex,
		}
	}
}

// handleAof payload(set k v) <- aofChan
func (handler *AofHandler) handleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {

			data := reply.MakeMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Error(err)
				continue
			}
			handler.currentDB = p.dbIndex

		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Error(err)
		}
	}

}

// loadAof
func (handler *AofHandler) loadAof() {
	file, err := os.Open(handler.aofFilename)
	if err != nil {
		logger.Error(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)
	fackConn := &connection.Connection{}

	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error(p.Err)
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("need multi mulk")
			continue
		}
		rep := handler.db.Exec(fackConn, r.Args)
		if reply.IsErrorReply(rep) {
			logger.Error(rep)
		}
	}
}
