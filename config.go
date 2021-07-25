package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/naoina/toml"
)

type Config struct {
	Listen struct {
		Address string
		Port    int
	}
	Log struct {
		Level   string
		Backups int
		Maxsize int64
	}
	Fasthttp struct {
		Concurrency        int
		ReadBufferSize     int
		ReadTimeout        int
		MaxRequestsPerConn int
		ReduceMemoryUsage  bool
	}
}

func NewConfig(filename string) (*Config, error) {
	if filename == "" {
		var env = "development"
		for _, name := range []string{"GOLANG_ENV", "ENV"} {
			if s := os.Getenv(name); s != "" {
				env = s
				break
			}
		}
		filename = env + ".toml"
	}

	tomlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := new(Config)
	if err = toml.Unmarshal(tomlData, c); err != nil {
		return nil, fmt.Errorf("toml.Decode(%#v) error: %+v", filename, err)
	}

	if filename == "development.toml" {
		fmt.Fprintf(os.Stderr, "%s WAN config.go:131 > running in the development mode.\n", time.Now().Format("15:04:05"))
	}

	return c, nil
}
