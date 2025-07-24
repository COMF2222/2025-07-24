package service

import (
	"2025-07-24/internal/config"
	"2025-07-24/internal/repository"
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Service struct {
	repo       *repository.InMemoryRepository
	cfg        *config.Config
	workerPool chan struct{}
	wg         sync.WaitGroup
}

func NewService(repo *repository.InMemoryRepository, cfg *config.Config) *Service {
	return &Service{
		repo:       repo,
		cfg:        cfg,
		workerPool: make(chan struct{}, cfg.MaxTasks),
	}
}

func (s *Service) createArchive(id string, links []string) (string, []string, error) {
	archiveDir := "archives"
	os.MkdirAll(archiveDir, os.ModePerm)
	archivePath := filepath.Join(archiveDir, id+".zip")
	f, err := os.Create(archivePath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	var errs []string

	for _, link := range links {
		resp, err := http.Get(link)
		if err != nil {
			errs = append(errs, fmt.Sprintf("ошибка скачивания %s: %v", link, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errs = append(errs, fmt.Sprintf("некорректный статус от %s: %d", link, resp.StatusCode))
			continue
		}

		filename := filepath.Base(link)
		w, err := zw.Create(filename)
		if err != nil {
			errs = append(errs, fmt.Sprintf("ошибка архивации %s: %v", link, err))
			continue
		}
		if _, err := io.Copy(w, resp.Body); err != nil {
			errs = append(errs, fmt.Sprintf("ошибка копирования %s: %v", link, err))
			continue
		}
	}

	return archivePath, errs, nil
}

func (s *Service) processTask(id string) {
	defer s.wg.Done()
	defer func() { <-s.workerPool }()

	task, exists := s.repo.GetTask(id)
	if !exists {
		return
	}
	task.Status = repository.TaskProcessing
	s.repo.UpdateTask(task)

	_, errs, err := s.createArchive(id, task.Links)
	if err != nil {
		task.Status = repository.TaskFailed
		task.Errors = append(task.Errors, err.Error())
		s.repo.UpdateTask(task)
		return
	}

	task.Status = repository.TaskCompleted
	task.ArchivePath = fmt.Sprintf("http://localhost:%s/archives/%s", s.cfg.Port, id)
	task.Errors = errs
	s.repo.UpdateTask(task)
}

func (s *Service) isValidExtension(link string) bool {
	ext := strings.ToLower(filepath.Ext(link))
	for _, allowed := range s.cfg.AllowedExt {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (s *Service) CreateTask() (string, error) {
	if s.repo.ActiveTaskCount() >= s.cfg.MaxTasks {
		return "", errors.New("сервер занят, достигнуто максимальное количество задач")
	}

	id := uuid.New().String()
	return id, s.repo.CreateTask(id)
}

func (s *Service) AddLink(id, link string) error {
	task, exists := s.repo.GetTask(id)
	if !exists {
		return errors.New("задача не найдена")
	}
	if len(task.Links) >= s.cfg.MaxFilesPerTask {
		return errors.New("достигнуто максимальное количество файлов на задачу")
	}
	if !s.isValidExtension(link) {
		return errors.New("недопустимое расширение файла")
	}
	if err := s.repo.AddLink(id, link); err != nil {
		return err
	}

	task, _ = s.repo.GetTask(id)
	if task.Status == repository.TaskCreated && len(task.Links) == s.cfg.MaxFilesPerTask {
		s.wg.Add(1)
		s.workerPool <- struct{}{}
		go s.processTask(id)
	}
	return nil
}

func (s *Service) GetStatus(id string) (repository.Task, error) {
	task, exists := s.repo.GetTask(id)
	if !exists {
		return repository.Task{}, errors.New("задача не найдена")
	}
	return task, nil
}

func (s *Service) GetArchive(id string) (*os.File, error) {
	task, exists := s.repo.GetTask(id)
	if !exists || task.Status != repository.TaskCompleted {
		return nil, errors.New("архив не готов или не найден")
	}
	path := filepath.Join("archives", task.ID+".zip")
	return os.Open(path)
}
