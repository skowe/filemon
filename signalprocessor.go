package filemon

/*
*	Implementing a simple observer that can be used for simple tasks
*	It is up to the user to define how the tasks are to be done by implementing a worker interface
* and giving the SignalReciever object a spawner function that returns a valid worker object
 */

import (
	"errors"
	"log"
	"path/filepath"
	"sync"
)

var (
	ErrorLoggerCantbeNil = errors.New("SignalReciever object must be passed a logger object")
	ErrorCantWork        = errors.New("Worker.Work: can't work waiting for signal")
)

// The implementation of the worker interface is left to the user
// Signal reciever can only work with one type of worker.
type Worker interface {
	Work() error
	Open(e *Event) Worker
}

// SignalReciever object is a simple implementation of an observer.
// It spawns workers as needed by calling the designated spawner function
// The spawner function for a worker takes no arguments and returns a Worker type object!
// There are 2 modes of execution.
// First is the simple linear execution  Create Worker -> Load Event by calling Worker.Open(event) -> Call Worker.Work() -> Log errors
// The Work method is always called so in case a worker doesn't handle all types, it is up to user to exclude unwanted events
// 
// The second method is defined by an initiator event and 1 or more trigger events
// Upon recieveing an event the Update function looks into it's map of workers and in no worker is tied to the filename in the event
// it calls a spawner and creates a worker
// the created worker then opens the event object and checks if the event is the defined initiator event
// if the Open method recieves any event that is not considered an initiator it should return nil after it completes it's objective otherwise returns the same worker object
// said worker object is then stored in the map with the filename it's tied to as the key, the following events  
type SignalReciever struct {
	tag              string
	spawner          func() Worker
	tagSet           bool
	watching         string
	isTrigger        bool
	waitingWorkers   map[string]Worker
	freeOnCompletion bool
	logger           *log.Logger
	mut sync.Mutex
}
// 
func (s *SignalReciever) Update(e *Event) {
	// log.Println(e.Name, s.watching)
	if s.watching != filepath.Dir(e.Name) {
		return
	}
	if !s.isTrigger {
		worker := s.spawner()
		worker.Open(e)
		err := worker.Work()
		if err != nil {
			s.logger.Printf("SignalReciever.Update: Failed to execute worker for %s, %s", e.Name, err)
		}
		return
	}
	worker, ok := s.waitingWorkers[e.Name]
	if !ok {
		worker := s.spawner()
		worker = worker.Open(e)
		if worker != nil {
			s.waitingWorkers[e.Name] = worker
		}
		return
	}
	s.mut.Lock()
	worker.Open(e)
	err := worker.Work()

	if err != nil && !errors.Is(err, ErrorCantWork) {
		log.Println(err)

		s.logger.Printf("SignalReciever.Update: Failed to execute worker for %s, %s", e.Name, err)
	}

	if s.freeOnCompletion && err == nil{
		delete(s.waitingWorkers, e.Name)
	}
	s.mut.Unlock()
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
func (s *SignalReciever) FreeWorkersOnCompletion() *SignalReciever{
	s.freeOnCompletion = !s.freeOnCompletion
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
