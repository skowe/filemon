package filemon

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Transaltions of original fsnotify events to uint32, implemented for ease of use
const (
	CREATE uint32 = 1 << iota
	WRITE
	REMOVE
	RENAME
	CHMOD
)

type Event struct {
	Name      string
	Op        fsnotify.Op
	Event     fsnotify.Event
	Timestamp time.Time
	Err       error
}

func (e *Event) Has(op uint32) bool {
	return e.Op.Has(fsnotify.Op(op))
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
