package filemon

type Worker interface {
	Work() bool
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
}

func (s *SignalReciever) Update(e *Event) {
	if !s.isTrigger {
		worker := s.spawner().Open(e)
		worker.Work()
	} else {
		_, ok := s.waitingWorkers[e.Name]
		if !ok {
			s.waitingWorkers[e.Name] = s.spawner()
		} else {
			s.waitingWorkers[e.Name].Open(e)
			canFree := s.waitingWorkers[e.Name].Work() && s.freeOnCompletion

			if canFree {
				delete(s.waitingWorkers, e.Name)
			}

		}
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
func NewReciever(path string, workersWait bool, freeOnCompletion bool, spawner func() Worker) *SignalReciever {
	s := new(SignalReciever).WithSpawner(spawner).Watching(path).WorkersWork()
	if workersWait {
		s = s.WorkersWait()
	}
	s.freeOnCompletion = freeOnCompletion
	return s
}
