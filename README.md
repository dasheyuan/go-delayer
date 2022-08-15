参考 [有赞延迟队列设计](http://tech.youzan.com/queuing_delay) 中的部分设计，优化后实现。

## 应用场景

- 订单超过30分钟未支付，自动关闭订单。
- 订单完成后, 如果用户一直未评价, 5天后自动好评。
- 会员到期前3天，短信通知续费。
- 其他针对某个任务，延迟执行功能的需求。

## 实现原理

- 客户端：push 任务时，任务数据存入 hash 中，jobID 存入 zset 中，pop 时从指定的 list 中取准备好的数据。
- 服务器端：定时使用连接池并行将 zset 中到期的 jobID 放入对应的 list 中，供客户端 pop 取出。

# go-delayer 队列延时中间件
基于 Redis 的延迟消息队列中间件，采用 Golang 开发，支持 Kratos GRPC调用。

## 使用样例
1.导入服务（provider/consumer）
```
import delayer "git.bell.ai/Technology/go-delayer/api"

// Service service.
type Service struct {
	ac      *paladin.Map
	dao     dao.Dao
	delayer delayer.GoDelayerClient
}

cfg := &warden.ClientConfig{
		Timeout:                xtime.Duration(20 * time.Second),
		KeepAliveWithoutStream: true,
		NonBlock:               true,
	}
delayerCli, cerr := delayer.NewClient(cfg)
if cerr != nil {
    panic(cerr)
}

s = &Service{
    ac:      &paladin.TOML{},
    dao:     d,
    delayer: delayerCli,
}

```
2. 任务提交 provider
```
func (s *Service) MockPushJob(ctx context.Context, req *empty.Empty) (resp *empty.Empty, err error){
	_, err = s.delayer.PushJob(context.TODO(), &delayer.Job{
		Topic: "topic1", //Job类型。可以理解成具体的业务名称。
		ID:    "uuid1234",  //Job的唯一标识。用来检索和删除指定的Job信息。
		Delay: 60,      //Job需要延迟的时间。单位：秒。（服务端会将其转换为绝对时间）
		Body:  "{}",    //Job的内容，供消费者做具体的业务处理，以json格式存储。
		TTR:   0,       //Job执行超时时间。单位：秒。(TODO)
	})
	if err != nil {
		log.Error("er(%v)", err)
	}
	return
}
```
3. 任务执行 consumer
```
func (s *Service) MockPopJob(ctx context.Context, req *empty.Empty) (resp *empty.Empty, err error) {
	job1, err := s.delayer.PopJob(ctx, &delayer.Topic{
		Topic: "topic1",
	})
	if err != nil {
		log.Error("err(%v)", err)
	}
	log.Info("job(%v)", job1)
	return
}

```