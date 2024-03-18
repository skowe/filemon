package filemon

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Name string
	Op fsnotify.Op
	Event     fsnotify.Event
	Timestamp time.Time
	Err       error
}

func (e *Event) String() string {
	var mes string
	format := "%v:%v"
	if e.Err != nil {
		mes = e.Err.Error()
	} else {
		v := e.Event.Name + ":" + e.Event.Op.String()
		mes = fmt.Sprintf(format, e.Timestamp, v)
	}
	return mes
}

func (e *Event) IsError() bool {
	return e.Err != nil
}
