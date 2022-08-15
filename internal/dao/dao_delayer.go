package dao

import (
	"context"
	"errors"
	pb "git.bell.ai/Technology/go-delayer/api"
	"github.com/garyburd/redigo/redis"
	"github.com/go-kratos/kratos/pkg/log"
	"reflect"
	"strconv"
	"strings"

	"time"
)

const (
	KeyJobPool       = "delayer:job_pool:"
	PrefixJobBucket  = "delayer:job_bucket"
	PrefixReadyQueue = "delayer:ready_queue:"

	Topic = "topic"
	Body  = "body"
	TTR   = "ttr"
)

func (d *dao) PushJob(ctx context.Context, job *pb.Job) (err error) {
	p := d.redis.Pipeline()
	var pushJobCmd = []struct {
		args     []interface{}
		expected interface{}
	}{
		{
			[]interface{}{"MULTI"},
			"OK",
		},
		{
			[]interface{}{"HMSET", KeyJobPool + job.ID, Topic, job.Topic, Body, job.Body, TTR, job.TTR},
			"QUEUED",
		},
		{
			[]interface{}{"ZADD", PrefixJobBucket, time.Now().Unix() + job.Delay, job.ID},
			"QUEUED",
		},
		{
			[]interface{}{"EXEC"},
			[]interface{}{"OK", int64(1)},
		},
	}
	for _, cmd := range pushJobCmd {
		p.Send(cmd.args[0].(string), cmd.args[1:]...)
	}

	replies, err := p.Exec(ctx)
	if err != nil {
		log.Error("Redis Pipeline err %v", err)
		return
	}
	i := 0
	for replies.Next() {
		cmd := pushJobCmd[i]
		actual, err1 := replies.Scan()
		if err1 != nil {
			log.Error("Receive(%v) returned error %v", cmd.args, err1)
			err = err1
			break
		}
		if !reflect.DeepEqual(actual, cmd.expected) {
			log.Error("Receive(%v) = %v, want %v", cmd.args, actual, cmd.expected)
			err = errors.New("Reids Client Receive Unwanted!")
			break
		}
		i++
	}
	return
}

func (d *dao) PopJob(ctx context.Context, t *pb.Topic) (job *pb.Job, err error) {
	id, err := redis.String(d.redis.Do(ctx, "RPOP", PrefixReadyQueue+t.Topic))
	if err != nil {
		return nil, err
	}
	result, err := redis.StringMap(d.redis.Do(ctx, "HGETALL", KeyJobPool+id))
	if err != nil {
		return nil, err
	}
	if result["topic"] == "" || result["body"] == "" {
		return nil, errors.New("Job has expired or is incomplete")
	}
	_, err = d.redis.Do(ctx, "DEL", KeyJobPool+id)
	ttr, _ := strconv.ParseInt(result["ttr"], 10, 64)
	job = &pb.Job{
		ID:   id,
		Body: result["body"],
		TTR:  ttr,
	}
	return
}

func (d *dao) getExpireJobs(ctx context.Context) ([]string, error) {
	return redis.Strings(d.redis.Do(ctx, "ZRANGEBYSCORE", PrefixJobBucket, 0, time.Now().Unix()))
}

func (d *dao) getJobTopic(ctx context.Context, jobID string, ch chan []string) {
	topic, err := redis.Strings(d.redis.Do(ctx, "HMGET", KeyJobPool+jobID, Topic))
	if err != nil {
		return
	}
	arr := []string{jobID, topic[0]}
	ch <- arr
}

// 移动任务至ReadyQueue
func (d *dao) moveJobToReadyQueue(ctx context.Context, jobIDs []string, topic string) {
	p := d.redis.Pipeline()
	jobIDsStr := strings.Join(jobIDs, ",")
	p.Send("MULTI")
	// 移除JobBucket
	args1 := make([]interface{}, len(jobIDs)+1)
	args1[0] = PrefixJobBucket
	for k, v := range jobIDs {
		args1[k+1] = v
	}
	p.Send("ZREM", args1[0:]...)
	// 插入ReadyQueue
	args2 := make([]interface{}, len(jobIDs)+1)
	args2[0] = PrefixReadyQueue + topic
	for k, v := range jobIDs {
		args2[k+1] = v
	}
	p.Send("LPUSH", args2[0:]...)
	p.Send("EXEC")
	_, err := p.Exec(ctx)
	if err != nil {
		log.Error("Redis Pipeline err %v", err)
		return
	}
	// 打印日志
	log.Info("Job is ready, Topic: %s, IDs: [%s]", topic, jobIDsStr)
}

func (d *dao) RemoveJob(ctx context.Context, jobId *pb.JobID) (err error) {
	p := d.redis.Pipeline()
	var removeJobCmd = []struct {
		args      []interface{}
		expected  interface{}
		expected2 interface{}
	}{
		{
			[]interface{}{"MULTI"},
			"OK", nil,
		},
		{
			[]interface{}{"ZREM", PrefixJobBucket, jobId.ID},
			"QUEUED", nil,
		},
		{
			[]interface{}{"DEL", KeyJobPool + jobId.ID},
			"QUEUED", nil,
		},
		{
			[]interface{}{"EXEC"},
			[]interface{}{int64(1), int64(1)}, []interface{}{int64(0), int64(0)},
		},
	}
	for _, cmd := range removeJobCmd {
		p.Send(cmd.args[0].(string), cmd.args[1:]...)
	}

	replies, err := p.Exec(ctx)
	if err != nil {
		log.Error("Redis Pipeline err %v", err)
		return
	}
	i := 0
	for replies.Next() {
		cmd := removeJobCmd[i]
		actual, err1 := replies.Scan()
		if err1 != nil {
			log.Error("Receive(%v) returned error %v", cmd.args, err1)
			err = err1
			break
		}
		if !reflect.DeepEqual(actual, cmd.expected) && !reflect.DeepEqual(actual, cmd.expected2) {
			log.Error("Receive(%v) = %v, want %v", cmd.args, actual, cmd.expected)
			err = errors.New("Reids Client Receive Unwanted!")
			break
		}
		i++
	}
	return
}
