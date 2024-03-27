package filemon

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

type testWorker struct {
	passOnOp uint32
	t        *testing.T
	canWork  bool
}

func (t *testWorker) Open(e *Event) Worker {
	t.canWork = e.Has(t.passOnOp)
	return Worker(t)
}
func (t *testWorker) Work() error {
	t.t.Log("In Worker")
	if !t.canWork {
		t.t.Log("Failing Worker")
		return nil
	}
	t.t.Log("Passing Worker")
	t.t.Log("Got ", fsnotify.Op(t.passOnOp).String())
	Pass <- 1
	return nil
}

var (
	ErrTestOutOfTime = fmt.Errorf("test out of time")

	// testWorker will send a pass signal if it ever comes to it
	Pass = make(chan int)
)

const toMonit = "to_monit"

func createTester(t *testing.T, passOn uint32) func() Worker {
	tw := &testWorker{
		passOnOp: passOn,
		t:        t,
		canWork:  false,
	}
	return func() Worker {
		return Worker(tw)
	}
}

func createEnv(t *testing.T, passOn uint32, workerWaits bool) *Tracker {
	os.Mkdir(toMonit, 0744)
	logger := log.New(os.Stderr, "ERROR", log.Ldate|log.Ltime)
	toTest, err := New(logger)

	if err != nil {
		t.Errorf("failed to init test: monitor creation fail %v", err)
		return nil
	}
	err = toTest.Add(toMonit)
	if err != nil {
		t.Errorf("failed to assign folder to monitor %v", err)
		return nil
	}
	spawner := createTester(t, passOn)
	freeOnCompletion := true

	to := NewReciever(toMonit, workerWaits, freeOnCompletion, logger, spawner)
	err = toTest.Register(to)
	if err != nil {
		t.Errorf("failed to register observer to the monitor: %v", err)
		return nil
	}
	return toTest
}

func createFile(t *testing.T) {
	f, err := os.CreateTemp(toMonit, "test.txt")
	if err != nil {
		t.Errorf("failed to create a file %v", err)
	}
	f.Close()
	os.Remove(f.Name())
}

func createFolder(t *testing.T) {
	n, err := os.MkdirTemp(toMonit, "test_dir")
	if err != nil {
		t.Errorf("failed to create a file %v", err)
	}

	os.Remove(n)
}

func writeToFile(t *testing.T) {
	data, err := GenerateString()
	if err != nil {
		t.Errorf("failed to generate random content")
		return
	}
	strRead := strings.NewReader(data)
	f, err := os.CreateTemp(toMonit, "testFile.txt")

	if err != nil {
		t.Errorf("failed to create a temporary file for writing")
		return
	}

	_, err = io.Copy(f, strRead)
	if err != nil {
		t.Errorf("failed to write to file")
		return
	}
	f.Close()
	os.Remove(f.Name())
}

func renameFile(t *testing.T) {
	f, err := os.CreateTemp(toMonit, "test.txt")
	if err != nil {
		t.Errorf("failed to create a file %v", err)
		return
	}
	f.Close()
	newName := path.Join(toMonit, "renamed.txt")
	err = os.Rename(f.Name(), newName)
	if err != nil {
		t.Log(err)
		t.Errorf("failed to rename")
		return
	}

	os.Remove(newName)

}
func testRunner(t *testing.T, toRun func(*testing.T), waitBeforeFail int) {

	go toRun(t)

	ticker := time.NewTicker(time.Duration(waitBeforeFail) * time.Second)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Errorf("%v", ErrTestOutOfTime)
	case <-Pass:
		ticker.Stop()
		// tocker is here to help sync with execution
		// call to log on t.Log in Work may sometimes finish after testRunner is done causing a panic
		tocker := time.NewTicker(300 * time.Millisecond)
		<-tocker.C
		return
	}
}
func TestCreateFile(t *testing.T) {
	toTest := createEnv(t, CREATE, false)

	go toTest.Run()
	testRunner(t, createFile, 5)
}

func TestRemoveFile(t *testing.T) {
	toTest := createEnv(t, REMOVE, false)

	go toTest.Run()
	testRunner(t, createFile, 5)
}
func TestCreateFolder(t *testing.T) {
	toTest := createEnv(t, CREATE, false)

	go toTest.Run()
	testRunner(t, createFolder, 5)
}

func TestWriteFile(t *testing.T) {
	toTest := createEnv(t, WRITE, false)

	go toTest.Run()
	testRunner(t, writeToFile, 5)
}

func TestRename(t *testing.T) {
	toTest := createEnv(t, RENAME, false)

	go toTest.Run()
	testRunner(t, renameFile, 5)
}

func TestWaitingWorker(t *testing.T) {
	toTest := createEnv(t, REMOVE, true)

	go toTest.Run()
	testRunner(t, writeToFile, 5)
	os.Remove(toMonit)
}
