package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/xiaoxuxiansheng/goredis/persist"
)

type Config struct {
	Bind                    string `cfg:"bind"`                        // ip 地址
	Port                    int    `cfg:"port"`                        // 启动端口号
	AppendOnly_             bool   `cfg:"appendonly"`                  // 是否启用 aof
	AppendFileName_         string `cfg:"appendfilename"`              // aof 文件名称
	AppendFsync_            string `cfg:"appendfsync"`                 // aof 级别
	AutoAofRewriteAfterCmd_ int    `cfg:"auto-aof-rewrite-after-cmds"` // 每执行多少次 aof 操作后，进行一次重写
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Bind, c.Port)
}

func (c *Config) AppendOnly() bool {
	return c.AppendOnly_
}

func (c *Config) AppendFileName() string {
	return c.AppendFileName_
}

func (c *Config) AppendFsync() string {
	return c.AppendFsync_
}

func (c *Config) AutoAofRewriteAfterCmd() int {
	return c.AutoAofRewriteAfterCmd_
}

/*
通过提供全局配置对象，获取配置信息。
全局配置对象先从配置文件加载，加载失败，使用默认配置对象。
*/
var (
	confOnce   sync.Once
	globalConf *Config
)

func PersistThinker() persist.Thinker {
	return SetUpConfig()
}

func SetUpConfig() *Config {
	confOnce.Do(func() {
		defer func() {
			if globalConf == nil {
				globalConf = defaultConf()
			}
		}()

		file, err := os.Open("./redis.conf")
		if err != nil {
			return
		}
		defer file.Close()
		globalConf = setUpConfig(file)
	})

	return globalConf
}

func setUpConfig(src io.Reader) *Config {
	tmpkv := make(map[string]string)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		// 注释行，跳过
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && trimmed[0] == '#' {
			continue
		}

		// 寻找合法的空格分隔符位置
		pivot := strings.Index(trimmed, " ")
		if pivot <= 0 || pivot >= len(trimmed)-1 {
			continue
		}

		key := trimmed[:pivot]
		value := trimmed[pivot+1:]
		tmpkv[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	conf := &Config{}
	// 通过反射设置 conf 属性值
	// 1. 生成对象实例，获取指针
	// 2. 针对结构体所有字段遍历
	// 3.1 对第i个字段，获取field对象（代表字段，可获取字段每个字段类型、字段tag值）和fieldVal对象（代表字段值，可设置某个字段的值）
	// 3.2 通过field对象，获取字段的tag对应的key（key和配置文件中的key名相同）
	// 3.3 从对应map中获取对应key的配置值
	// 3.4 通过fieldVal对象，设置字段值
	t := reflect.TypeOf(conf)
	v := reflect.ValueOf(conf)
	for i := 0; i < t.Elem().NumField(); i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok || strings.TrimSpace(key) == "" {
			key = field.Name
		}
		value, ok := tmpkv[key]
		if !ok {
			continue
		}
		switch field.Type.Kind() {
		case reflect.String:
			fieldVal.SetString(value)
		case reflect.Int:
			intv, _ := strconv.ParseInt(value, 10, 64)
			fieldVal.SetInt(intv)
		case reflect.Bool:
			fieldVal.SetBool(value == "yes")
		}
	}

	return conf
}

func defaultConf() *Config {
	return &Config{
		Bind:        "0.0.0.0",
		Port:        6379,
		AppendOnly_: false, // 默认不启用 aof
	}
}
