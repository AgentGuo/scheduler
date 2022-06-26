package schedule

import (
	"fmt"
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/task"
	"github.com/go-redis/redis"
	"log"
)

type Scheduler struct {
	redisCli *redis.Client
}

func NewScheduler(config *config.SchedulerMainConfig) (*Scheduler, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	_, err := redisCli.Ping().Result()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &Scheduler{
		redisCli: redisCli,
	}, nil
}

func (s *Scheduler) Schedule(t *task.Task) (nodeName string, err error) {
	priorityList, err := s.score()
	if err != nil {
		return "", err
	}
	if len(priorityList) == 0 {
		return "", fmt.Errorf("no node for scheduling")
	}
	return priorityList[0].NodeName, nil
}
