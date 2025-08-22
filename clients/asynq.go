package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type AsynqTask interface {
	Name() string
	Handler(ctx context.Context, task *asynq.Task) error
}

type SchedulerClient interface {
	RegisterTask(name string, task AsynqTask)
	Enqueue(task *asynq.Task, at time.Time) error
	Start() error
}

type AsynqClient struct {
	asyncClient *asynq.Client
	mux         *asynq.ServeMux
	srv         *asynq.Server
}

func NewAsynqClient(redisURL string) *AsynqClient {
	// Create a new asynq client
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisURL})
	mux := asynq.NewServeMux()
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisURL},
		asynq.Config{
			Concurrency: 3,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	return &AsynqClient{
		asyncClient: client,
		mux:         mux,
		srv:         srv,
	}
}

func (c *AsynqClient) RegisterTask(name string, task AsynqTask) {
	c.mux.HandleFunc(name, task.Handler)
}

func (c *AsynqClient) Enqueue(task *asynq.Task, at time.Time) error {
	_, err := c.asyncClient.Enqueue(task, asynq.ProcessAt(at))
	return err
}

func (c *AsynqClient) Start() error {
	// Create error channel to catch any server errors
	errChan := make(chan error, 1)

	// Start the server in a separate goroutine
	go func() {
		if err := c.srv.Run(c.mux); err != nil {
			fmt.Println("Error starting asynq server")
			errChan <- err
		}
	}()

	// Return any immediate error, otherwise return nil
	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Second): // Give server a moment to start
		return nil
	}
}
