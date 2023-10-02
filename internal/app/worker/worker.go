package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
	log "github.com/sirupsen/logrus"
)

type DeleteURLQueue struct {
	ch     chan *store.Task
	store  store.Database
	logger *log.Logger
	tasks  []store.Task
}

func NewDeleteURLQueue(storage store.Database, logger *log.Logger, maxWorker int) *DeleteURLQueue {
	return &DeleteURLQueue{
		store:  storage,
		logger: logger,
		ch:     make(chan *store.Task, maxWorker),
		tasks:  make([]store.Task, 0, 500),
	}
}

// Start добавляет задачи в массив и запускаети удаление задач через каждые 10 секунд.
func (q *DeleteURLQueue) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)

	go func() {
		for {
			select {
			case task := <-q.ch:
				q.tasks = append(q.tasks, *task)
			case <-ctx.Done():
				if err := q.doDeleteTasks(); err != nil {
					q.logger.Info(err.Error())
				}
			case <-ticker.C:
				if err := q.doDeleteTasks(); err != nil {
					q.logger.Info(err.Error())
				}
			}
		}
	}()
}

// Push отправляет url в канал для дальнейшего удаления.
func (q *DeleteURLQueue) Push(task *store.Task) {
	q.ch <- task
}

// doDeleteTasks отвечает за удаления url.
func (q *DeleteURLQueue) doDeleteTasks() error {
	if len(q.tasks) == 0 {
		return nil
	}

	if err := q.store.DeleteURL(q.tasks); err != nil {
		fmt.Println(err)
		return err
	}

	q.logger.Info(fmt.Sprintf("Successfully did %d delete url tasks", len(q.tasks)))
	q.tasks = q.tasks[:0]
	return nil
}
