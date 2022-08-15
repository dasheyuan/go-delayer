package service

import (
	"context"
	pb "git.bell.ai/Technology/go-delayer/api"
	"git.bell.ai/Technology/go-delayer/internal/dao"
	"github.com/garyburd/redigo/redis"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.GoDelayerServer), new(*Service)))

// Service service.
type Service struct {
	ac  *paladin.Map
	dao dao.Dao
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	cf = s.Close
	err = paladin.Watch("application.txt", s.ac)

	return
}

func (s *Service) PushJob(ctx context.Context, req *pb.Job) (reply *empty.Empty, err error) {
	reply = new(empty.Empty)
	err = s.dao.PushJob(ctx, req)
	if err != nil {
		log.Error("PushJob err(%v) req(%v)", err, req)
		err = pb.PushJobFail
	}
	return
}

func (s *Service) PopJob(ctx context.Context, topic *pb.Topic) (job *pb.Job, err error) {
	job, err = s.dao.PopJob(ctx, topic)
	if err == redis.ErrNil {
		return &pb.Job{}, nil
	}
	if err != nil {
		log.Error("PopJob Err:%v", err)
		return nil, err
	}
	return
}

func (s *Service) RemoveJob(ctx context.Context, jobId *pb.JobID) (resp *pb.RemoveJobResp, err error) {
	resp = &pb.RemoveJobResp{
		Removed: true,
	}
	err = s.dao.RemoveJob(ctx, jobId)
	if err != nil {
		log.Error("RemoveJob Err:%v", err)
		resp.Removed = false
		return
	}
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
