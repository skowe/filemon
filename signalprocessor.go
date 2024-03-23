package filemon

import (
	"errors"
	"log"
)

var (
	ErrorLoggerCantbeNil = errors.New("SignalReciever object must be passed a logger object")
)

type Worker interface {
	Work() error
	Open(e *Event) Worker
}

type SignalReciever struct {
	tag              string
	spawner          func() Worker
	tagSet           bool
	watching         string
	isTrigger        bool
	waitingWorkers   map[string]Worker
	freeOnCompletion bool
	logger           *log.Logger
}

func (s *SignalReciever) Update(e *Event) {
	if !s.isTrigger {
		worker := s.spawner().Open(e)
		worker.Work()
		return
	}
	_, ok := s.waitingWorkers[e.Name]
	if !ok {
		s.waitingWorkers[e.Name] = s.spawner()
		return
	}

	s.waitingWorkers[e.Name].Open(e)
	err := s.waitingWorkers[e.Name].Work()
	if err != nil {
		log.Printf("SignalReciever.Update: Failed to execute worker for %s, %s", e.Name, err)
	}
	if s.freeOnCompletion {
		delete(s.waitingWorkers, e.Name)
	}

}

func (s *SignalReciever) Tag(tag string) {
	if s.tagSet {
		return
	}
	s.tag = tag
}

func (s *SignalReciever) IsTagSet() bool {
	return s.tagSet
}

func (s *SignalReciever) GetTag() string {
	if s.tagSet {
		return s.tag
	}
	return ""
}

func (s *SignalReciever) WithSpawner(spawner func() Worker) *SignalReciever {
	s.spawner = spawner
	return s
}

func (s *SignalReciever) Watching(path string) *SignalReciever {
	s.watching = path
	return s
}

func (s *SignalReciever) WorkersWait() *SignalReciever {
	s.isTrigger = true
	s.waitingWorkers = make(map[string]Worker)
	return s
}

func (s *SignalReciever) WorkersWork() *SignalReciever {
	s.isTrigger = false
	s.waitingWorkers = nil
	return s
}

func (s *SignalReciever) WithLogger(logger *log.Logger) *SignalReciever {
	if logger == nil {
		panic(ErrorLoggerCantbeNil)
	}
	s.logger = logger
	return s
}

func NewReciever(path string, workersWait bool, freeOnCompletion bool, logger *log.Logger, spawner func() Worker) *SignalReciever {
	s := new(SignalReciever).WithSpawner(spawner).Watching(path).WorkersWork()
	if workersWait {
		s = s.WorkersWait()
	}
	s.freeOnCompletion = freeOnCompletion
	return s
}
