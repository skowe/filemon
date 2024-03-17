package filemon

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	event     fsnotify.Event
	timestamp time.Time
	err       error
}

func (e *Event) String() string {
	var mes string
	format := "%v:%v"
	if e.err != nil {
		mes = e.err.Error()
	} else {
		v := e.event.Name + ":" + e.event.Op.String()
		mes = fmt.Sprintf(format, e.timestamp, v)
	}
	return mes
}

func (e *Event) IsError() bool {
	return e.err != nil
}
