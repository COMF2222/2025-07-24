package repository

import (
	"errors"
	"sync"
)

type TaskStatus string

const (
	TaskCreated    TaskStatus = "created"
	TaskProcessing TaskStatus = "processing"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
)

type Task struct {
	ID          string
	Status      TaskStatus
	Links       []string
	ArchivePath string
	Errors      []string
}

type InMemoryRepository struct {
	tasks map[string]Task
	mu    sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{tasks: make(map[string]Task)}
}

func (r *InMemoryRepository) CreateTask(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[id]; exists {
		return errors.New("Задача уже существует")
	}
	r.tasks[id] = Task{
		ID:     id,
		Status: TaskCreated,
		Links:  []string{},
	}
	return nil
}

func (r *InMemoryRepository) AddLink(id, link string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, exists := r.tasks[id]
	if !exists {
		return errors.New("задача не найдена")
	}
	task.Links = append(task.Links, link)
	r.tasks[id] = task
	return nil
}

func (r *InMemoryRepository) GetTask(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, exists := r.tasks[id]
	return task, exists
}

func (r *InMemoryRepository) UpdateTask(task Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryRepository) ActiveTaskCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, task := range r.tasks {
		if task.Status == TaskProcessing {
			count++
		}
	}
	return count
}
