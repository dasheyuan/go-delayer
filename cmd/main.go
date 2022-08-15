package main

import (
	"context"
	"flag"
	"github.com/go-kratos/kratos/pkg/conf/env"
	"github.com/go-kratos/kratos/pkg/conf/paladin/apollo"
	"github.com/go-kratos/kratos/pkg/naming"
	"github.com/go-kratos/kratos/pkg/naming/etcd"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.bell.ai/Technology/go-delayer/internal/di"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
)

func etcdRegister() (cancel context.CancelFunc, err error) {
	log.Info("Service Register")
	cfg := new(struct {
		Url   string
		AppId string
	})

	if err := paladin.Get("etcd.txt").UnmarshalTOML(cfg); err != nil {
		panic(err)
	}

	hn, _ := os.Hostname()
	e, _ := etcd.New(nil)
	ins := &naming.Instance{
		Zone:     env.Zone,
		Env:      env.DeployEnv,
		AppID:    cfg.AppId,
		Hostname: hn,
		Addrs:    strings.Split(cfg.Url, ","),
	}

	cancel, err = e.Register(context.Background(), ins)
	if err != nil {
		panic(err)
	}
	return
}

func main() {
	flag.Parse()
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("git.bell.ai/Technology/go-delayer start")
	paladin.Init(apollo.PaladinDriverApollo)

	cancel, err := etcdRegister()
	if err != nil {
		panic(err)
	}
	_, closeFunc, err := di.InitApp()
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			cancel()
			log.Info("git.bell.ai/Technology/go-delayer exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
