package app

import (
	"fmt"

	"github.com/xiaoxuxiansheng/goredis/database"
	"github.com/xiaoxuxiansheng/goredis/datastore"
	"github.com/xiaoxuxiansheng/goredis/handler"
	"github.com/xiaoxuxiansheng/goredis/log"
	"github.com/xiaoxuxiansheng/goredis/persist"
	"github.com/xiaoxuxiansheng/goredis/protocol"
	"github.com/xiaoxuxiansheng/goredis/server"

	"go.uber.org/dig"
)

var container = dig.New()

/*
*server.Server
├── server.Handler
│   ├── handler.DB
│   │   └── database.Executor
│   │       └── database.DataStore
│   │           └── handler.Persister
│   │               └── persist.Thinker
│   ├── handler.Persister
│   ├── handler.Parser
│   │   └── log.Logger
│   └── log.Logger
└── log.Logger
*/
func init() {
	/**
	   其它
	**/
	// 配置加载 conf
	_ = container.Provide(SetUpConfig)
	_ = container.Provide(PersistThinker)
	// 日志打印 logger
	_ = container.Provide(log.GetDefaultLogger)

	/**
	   存储引擎
	**/
	// 数据持久化
	_ = container.Provide(persist.NewPersister)
	// 存储介质
	_ = container.Provide(datastore.NewKVStore)
	// 执行器
	_ = container.Provide(database.NewDBExecutor)
	// 触发器
	_ = container.Provide(database.NewDBTrigger)

	/**
	   逻辑处理层
	**/
	// 协议解析
	_ = container.Provide(protocol.NewParser)
	// 指令处理
	_ = container.Provide(handler.NewHandler)

	/**
	   服务端
	**/
	_ = container.Provide(server.NewServer)
	fmt.Println("dig::", container.String())
}

func ConstructServer() (*server.Server, error) {
	//var h server.Handler
	//if err := container.Invoke(func(_h server.Handler) {
	//    h = _h
	//}); err != nil {
	//    return nil, err
	//}
	//
	//var l log.Logger
	//if err := container.Invoke(func(_l log.Logger) {
	//    l = _l
	//}); err != nil {
	//    return nil, err
	//}
	var s *server.Server
	if err := container.Invoke(func(_s *server.Server) {
		s = _s
	}); err != nil {
		return nil, err
	}
	return s, nil
	//return server.NewServer(h, l), nil
}
