package main

import "github.com/xiaoxuxiansheng/goredis/app"

/*
app
|__server.Server
|__app.Config

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
func main() {
	server, err := app.ConstructServer()
	if err != nil {
		panic(err)
	}

	app := app.NewApplication(server, app.SetUpConfig())
	defer app.Stop()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
