package api

import (
	"context"
	"github.com/go-kratos/kratos/pkg/naming/etcd"
	"github.com/go-kratos/kratos/pkg/net/rpc/warden"
	"github.com/go-kratos/kratos/pkg/net/rpc/warden/resolver"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "go_delayer"

func init() {
	resolver.Register(etcd.Builder(nil))
}

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (GoDelayerClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), "etcd://default/"+AppID)
	if err != nil {
		return nil, err
	}
	return NewGoDelayerClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc api.proto
