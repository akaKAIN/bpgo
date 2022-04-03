package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ErrorWrongPath = "invalid path argument"
const ErrorEmptyPath = "empty path argument"
const defaultMaxDeep = 2

type FileInfo interface {
	os.FileInfo
	Path() string
}

type fileInfo struct {
	os.FileInfo
	path string
}

func (fi fileInfo) Path() string {
	return fi.path
}

type Searcher interface {
	Start() error
	AddFile(info os.FileInfo, path string) error
	HasMatch(file os.FileInfo, substr string) bool
	List() []FileInfo
	Find(substr string) []FileInfo
	IncreaseDeep()
}

type Search struct {
	maxDeep  int
	basePath string
	list     []FileInfo
}

func (s *Search) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.readDir(ctx, s.basePath); err != nil {
		return err
	}

	return nil
}

func (s *Search) readDir(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		if s.canSearchDeeper(path) {
			dirEntry, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			for _, file := range dirEntry {
				if file.IsDir() {
					err := s.readDir(ctx, filepath.Join(path, file.Name()))
					if err != nil {
						log.Println(err)
					}
				} else {
					fi, err := file.Info()
					if err != nil {
						return err
					}

					if err = s.AddFile(fi, path); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

func (s *Search) canSearchDeeper(path string) bool {
	paths := strings.Split(path, s.basePath)
	return len(paths) == 2 && (len(paths[1]) < 2 || len(strings.Split(paths[1], "\\")) <= s.maxDeep)
}

// AddFile Добавление файла в массив результатов
func (s *Search) AddFile(file os.FileInfo, path string) error {
	s.list = append(s.list, FileInfo(fileInfo{
		FileInfo: file,
		path:     filepath.Join(path, file.Name()),
	}))
	return nil
}

// HasMatch Проверка, содержит ли имя файла заданную подстроку
func (s *Search) HasMatch(file os.FileInfo, substr string) bool {
	return strings.Contains(file.Name(), substr)
}

// Find получаем список FileInfo, имена которых соответствуют HasMatch
func (s *Search) Find(substr string) []FileInfo {
	matchList := make([]FileInfo, 0)
	for _, fi := range s.List() {
		if s.HasMatch(fi, substr) {
			log.Println("find: ", fi.Name())
			matchList = append(matchList, fi)
		}
	}

	return matchList
}

// List Возвращает результаты поиска
func (s *Search) List() []FileInfo {
	return s.list
}

func (s *Search) IncreaseDeep() {
	s.maxDeep += 2
	fmt.Println("max search deep was increased by 2")
}

// InitSearch Создание Search объекта
func InitSearch(path string) (Searcher, error) {
	if path == "" {
		return nil, errors.New(ErrorEmptyPath)
	}

	dir, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !dir.IsDir() {
		return nil, errors.New(ErrorWrongPath)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	search := new(Search)
	search.list = make([]FileInfo, 0)
	search.basePath = absPath
	search.maxDeep = defaultMaxDeep

	return search, nil
}
