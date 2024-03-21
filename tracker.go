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
	stop      chan int
	Events    chan Event
}

func New() (*Tracker, error) {
	res := new(Tracker)

	res.observers = make(map[string]Observer)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("New failed to create a tracker %w", err)
	}
	res.addWatcher(w)
	res.stopper()
	res.Events = make(chan Event)
	return res, nil
}

func (t *Tracker) addWatcher(w *fsnotify.Watcher) {
	t.watcher = w
}
func (t *Tracker) stopper() {
	t.stop = make(chan int)
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
	t.Events <- event
}

func (t *Tracker) Add(path string) error {
	err := t.watcher.Add(path)
	if err != nil {
		return fmt.Errorf("Tracker.Add failed to add path to tracker: %w", err)
	}
	return nil
}
func (t *Tracker) Stop() {
	close(t.stop)
}
func (t *Tracker) Run() {

	for {
		select {
		case ev := <-t.watcher.Events:
			e := Event{
				Name:      ev.Name,
				Op:        ev.Op,
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
		case <-t.stop:
			err := t.watcher.Close()
			if err != nil {
				e := Event{
					Err:       err,
					Timestamp: time.Now(),
				}
				t.NotifyAll(e)
			}
			return
		}
	}
}
