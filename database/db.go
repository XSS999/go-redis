package database

import (
	"go-redis/datastruct/dict"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"strings"
)

type DB struct {
	index  int
	data   dict.Dict
	addAof func(line CmdLine)
}

type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

func MakeDB() *DB {
	db := &DB{
		data:   dict.MakeSyncDict(),
		addAof: func(line CmdLine) {},
	}

	return db
}

func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	// PING SET
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command" + cmdName)
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.exector
	// SET k V --> K V
	return fun(db, cmdLine[1:])
}

// 校验参数个数
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity := raw.(*database.DataEntity)
	return entity, true
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove 指定的key清除
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// Removes 根据key，清除数据库
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Flush() {
	db.data.Clear()
}
