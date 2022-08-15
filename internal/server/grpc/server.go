package grpc

import (
	pb "git.bell.ai/Technology/go-delayer/api"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/net/rpc/warden"
)

// New new a grpc server.
func New(svc pb.GoDelayerServer) (ws *warden.Server, err error) {
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("config.txt").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("GRPC").UnmarshalTOML(&cfg); err != nil {
		return
	}
	ws = warden.NewServer(&cfg)
	pb.RegisterGoDelayerServer(ws.Server(), svc)
	ws, err = ws.Start()
	return
}
