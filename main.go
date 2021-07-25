package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/phuslu/log"
	"github.com/robfig/cron/v3"
	"github.com/valyala/fasthttp"
)

var (
	version = "r9999"
)

func main() {
	executable, err := os.Executable()
	if err != nil {
		println("cannot get executable path")
		os.Exit(1)
	}

	var validate bool
	flag.BoolVar(&validate, "validate", false, "parse the config toml and exit")
	flag.Parse()

	config, err := NewConfig(flag.Arg(0))
	if err != nil {
		log.Fatal().Err(err).Str("filename", flag.Arg(0)).Msg("read config error")
	}

	if validate {
		os.Exit(0)
	}

	if log.IsTerminal(os.Stderr.Fd()) {
		log.DefaultLogger = log.Logger{
			Level:      log.ParseLevel(config.Log.Level),
			Caller:     1,
			TimeFormat: "15:04:05",
			Writer: &log.ConsoleWriter{
				ColorOutput: true,
			},
		}
	} else {
		log.DefaultLogger = log.Logger{
			Level: log.ParseLevel(config.Log.Level),
			Writer: &log.FileWriter{
				Filename:   executable + ".log",
				MaxSize:    config.Log.Maxsize,
				MaxBackups: config.Log.Backups,
				LocalTime:  false,
			},
		}
	}

	address := config.Listen.Address + ":" + strconv.Itoa(config.Listen.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("listen_tcp", address).Msg("listen error")
	}

	server := &fasthttp.Server{
		Concurrency:        config.Fasthttp.Concurrency,
		ReadTimeout:        time.Duration(config.Fasthttp.ReadTimeout) * time.Second,
		MaxRequestsPerConn: config.Fasthttp.MaxRequestsPerConn,
		ReadBufferSize:     config.Fasthttp.ReadBufferSize,
		ReduceMemoryUsage:  config.Fasthttp.ReduceMemoryUsage,
		Logger:             &log.DefaultLogger,
		Handler:            fasthttp.CompressHandler(requestHandler),
	}

	log.Info().Str("version", version).Str("listen_tcp", ln.Addr().String()).Msg("listen and serve")
	go server.Serve(ln)

	runner := cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC), cron.WithLogger(cron.PrintfLogger(&log.DefaultLogger)))
	// log rotating daily
	if !log.IsTerminal(os.Stderr.Fd()) {
		runner.AddFunc("0 0 0 * * *", func() { log.DefaultLogger.Writer.(*log.FileWriter).Rotate() })
	}

	go runner.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGHUP)

	switch <-c {
	case syscall.SIGTERM, syscall.SIGINT:
		log.Info().Msg("exit")
		os.Exit(0)
	}

	log.Warn().Msg("start graceful shutdown...")

	var wg sync.WaitGroup
	go func(server *fasthttp.Server) {
		wg.Add(1)
		defer wg.Done()

		server.Shutdown()
	}(server)
	wg.Wait()

	log.Info().Msg("server shutdown")
}
