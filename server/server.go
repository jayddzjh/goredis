package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xiaoxuxiansheng/goredis/lib/pool"
	"github.com/xiaoxuxiansheng/goredis/log"
)

// 先分析类的依赖关系
// 大致分析每个类的作用
// 从main函数，找启动方法
// 分析依赖了类中的什么方法
// 函数调用时机，总结常见场景处理流程
// 如果自己实现，如何实现

// 处理器
type Handler interface {
	Start() error // 启动 handler
	// 处理到来的每一笔 tcp 连接
	Handle(ctx context.Context, conn net.Conn)
	// 关闭处理器
	Close()
}

type Server struct {
	runOnce  sync.Once
	stopOnce sync.Once
	handler  Handler
	logger   log.Logger
	stopc    chan struct{}
}

func NewServer(handler Handler, logger log.Logger) *Server {
	return &Server{
		handler: handler,
		logger:  logger,
		stopc:   make(chan struct{}),
	}
}

func (s *Server) Serve(address string) error {
	if err := s.handler.Start(); err != nil { // 启动处理器，阻塞处理任务
		return err
	}
	// 场景提取：接收来自系统的终止信号
	var _err error
	s.runOnce.Do(func() {
		// 监听进程信号
		exitWords := []os.Signal{syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT}

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, exitWords...) // 注册监听信号和通信管道
		closec := make(chan struct{}, 4)  //

		pool.Submit(func() {
			for {
				select {
				case signal := <-sigc:
					switch signal {
					case exitWords[0], exitWords[1], exitWords[2], exitWords[3]:
						closec <- struct{}{}
						return
					default:
					}
				case <-s.stopc:
					closec <- struct{}{}
					return
				}
			}
		})
		// 监听地址
		listener, err := net.Listen("tcp", address)
		if err != nil {
			_err = err
			return
		}

		s.listenAndServe(listener, closec)
	})

	return _err
}

func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopc)
	})
}

func (s *Server) listenAndServe(listener net.Listener, closec chan struct{}) {
	errc := make(chan error, 1)
	defer close(errc)

	// 遇到意外错误，则终止流程
	ctx, cancel := context.WithCancel(context.Background())
	// 异步任务功能：
	// 1. 监听正常退出信号
	// 2. 发生错误退出
	pool.Submit(
		func() {
			select {
			case <-closec:
				s.logger.Errorf("[server]server closing...")
			case err := <-errc:
				s.logger.Errorf("[server]server err: %s", err.Error())
			}
			cancel()
			// cancel方法功能如下：
			/*
				func main() {
					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						time.Sleep(2 * time.Second)
						cancel()  // 在2秒后取消ctx
					}()

					select {
					case <-ctx.Done():
						fmt.Println("Canceled")
					case <-time.After(3 * time.Second):
						fmt.Println("Timeout")
					}
				}
			*/
			s.logger.Warnf("[server]server closeing...")
			s.handler.Close() // 终止hander
			if err := listener.Close(); err != nil {
				s.logger.Errorf("[server]server close listener err: %s", err.Error())
			}
		})

	s.logger.Warnf("[server]server starting...")
	var wg sync.WaitGroup
	// io 多路复用模型，goroutine for per conn
	for {
		conn, err := listener.Accept()
		if err != nil {
			// 超时类错误，忽略
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				time.Sleep(5 * time.Millisecond)
				continue
			}

			// 意外错误，则停止运行
			errc <- err
			break
		}

		// 为每个到来的 conn 分配一个 goroutine 处理
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			s.handler.Handle(ctx, conn)
		})
	}

	// 通过 waitGroup 保证优雅退出
	wg.Wait()
}
