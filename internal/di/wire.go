// +build wireinject
// The build tag makes sure the stub is not built in the final build.

package di

import (
	"git.bell.ai/Technology/go-delayer/internal/dao"
	"git.bell.ai/Technology/go-delayer/internal/server/grpc"
	"git.bell.ai/Technology/go-delayer/internal/service"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(dao.Provider, service.Provider, grpc.New, NewApp))
}
