package filemon

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Target int

const ()

var (
	ErrCantGenerateTag = fmt.Errorf("failed to generate the tag")
)

type Observer interface {
	Update(event *Event)
	Tag(id string)
	GetTag() string
}

// Implementing Tracker with a map for observers
type Tracker struct {
	watcher   *fsnotify.Watcher
	observers map[string]Observer
}

func New() (*Tracker, error) {
	res := new(Tracker)

	res.observers = make(map[string]Observer)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	res.addWatcher(w)
	return res, nil
}

func (t *Tracker) addWatcher(w *fsnotify.Watcher) {
	t.watcher = w
}

func (t *Tracker) Register(o Observer) error {
	failCounter := 0
	for {
		if failCounter == 5 {
			return ErrCantGenerateTag
		}
		tag, err := GenerateString()
		if err != nil {
			failCounter = failCounter + 1
			continue
		}

		_, ok := t.observers[tag]
		if ok {
			continue
		}

		t.observers[tag] = o
		break
	}

	return nil
}

func (t *Tracker) NotifyAll(event Event) {
	for _, o := range t.observers {
		o.Update(&event)
	}
}

func (t *Tracker) Add(path string) error {
	return t.watcher.Add(path)
}

func (t *Tracker) Run() {

	for {
		select {
		case ev := <-t.watcher.Events:
			e := Event{
				Name: ev.Name,
				Op: ev.Op,
				Err:       nil,
				Event:     ev,
				Timestamp: time.Now(),
			}
			t.NotifyAll(e)
		case err := <-t.watcher.Errors:
			e := Event{
				Err:       err,
				Timestamp: time.Now(),
			}
			t.NotifyAll(e)
		}
	}
}
