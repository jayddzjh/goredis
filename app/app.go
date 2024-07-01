package app

import "github.com/xiaoxuxiansheng/goredis/server"

/*
1. app代表应用，包括server和config2个对象
2. 之所以不直接用server，是因为还有一个config对象
*/
type Application struct {
	server *server.Server
	conf   *Config
}

func NewApplication(server *server.Server, conf *Config) *Application {
	return &Application{
		server: server,
		conf:   conf,
	}
}

func (a *Application) Run() error {
	return a.server.Serve(a.conf.Address())
}

func (a *Application) Stop() {
	a.server.Stop()
}
