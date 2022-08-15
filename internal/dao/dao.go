package dao

import (
	"context"
	pb "git.bell.ai/Technology/go-delayer/api"
	"github.com/go-kratos/kratos/pkg/log"
	"time"

	"github.com/go-kratos/kratos/pkg/cache/memcache"
	"github.com/go-kratos/kratos/pkg/cache/redis"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/database/sql"
	"github.com/go-kratos/kratos/pkg/sync/pipeline/fanout"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB, NewRedis, NewMC)

//go:generate kratos tool genbts
// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)

	PushJob(ctx context.Context, job *pb.Job) (err error)
	PopJob(ctx context.Context, topic *pb.Topic) (job *pb.Job, err error)
	RemoveJob(ctx context.Context, jobId *pb.JobID) (err error)
}

// dao dao.
type dao struct {
	db         *sql.DB
	redis      *redis.Redis
	mc         *memcache.Memcache
	cache      *fanout.Fanout
	demoExpire int32
	ticker     *time.Ticker
}

// New new a dao and return.
func New(r *redis.Redis, mc *memcache.Memcache, db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(r, mc, db)
}

func newDao(r *redis.Redis, mc *memcache.Memcache, db *sql.DB) (d *dao, cf func(), err error) {
	var cfg struct {
		DemoExpire    xtime.Duration
		TimerInterval int64
	}
	if err = paladin.Get("application.txt").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:         db,
		redis:      r,
		mc:         mc,
		cache:      fanout.New("cache"),
		demoExpire: int32(time.Duration(cfg.DemoExpire) / time.Second),
		ticker:     time.NewTicker(time.Duration(cfg.TimerInterval) * time.Millisecond),
	}
	d.Start()
	cf = d.Close
	return
}

// Start
func (d *dao) Start() {
	go func() {
		for range d.ticker.C {
			d.Run()
		}
	}()
}

// Run
func (d *dao) Run() {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	jobs, err := d.getExpireJobs(ctx)
	if err != nil {
		log.Error("err %v", err)
	}
	topics := make(map[string][]string)
	ch := make(chan []string)
	for _, jobID := range jobs {
		go d.getJobTopic(ctx, jobID, ch)
	}
	// Topic分组
	for i := 0; i < len(jobs); i++ {
		arr := <-ch
		if arr[1] != "" {
			if _, ok := topics[arr[1]]; !ok {
				jobIDs := []string{arr[0]}
				topics[arr[1]] = jobIDs
			} else {
				topics[arr[1]] = append(topics[arr[1]], arr[0])
			}
		}
	}
	// 并行移动至Topic对应的ReadyQueue
	for topic, jobIDs := range topics {
		go d.moveJobToReadyQueue(ctx, jobIDs, topic)
	}
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
	d.ticker.Stop()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
