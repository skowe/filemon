package filemon

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

type testObserver struct {
	passOnOp fsnotify.Op
	tag      string
	test     *testing.T
	Pass     chan int
}

func (t *testObserver) Update(event *Event) {

	if event.IsError() {
		t.test.Errorf("failed test: %v", event.err)
		return
	}
	if !event.event.Has(t.passOnOp) {
		return
	}
	t.test.Log("passed")
	t.Pass <- 1
}

func (t *testObserver) Tag(s string) {
	t.tag = s
}

func (t *testObserver) GetTag() string {
	return t.tag
}

var (
	ErrTestOutOfTime = fmt.Errorf("test out of time")
)

const toMonit = "to_monit"

func createFile(t *testing.T) {
	f, err := os.CreateTemp(toMonit, "test.txt")
	if err != nil {
		t.Errorf("failed to create a file %v", err)
	}
	f.Close()
	err = os.Remove(f.Name())
	t.Log(err)
}

func createEnv(t *testing.T, passOn fsnotify.Op) (*Tracker, *testObserver) {
	toTest, err := New()
	if err != nil {
		t.Errorf("failed to init test: monitor creation fail")
		return nil, nil
	}
	err = toTest.Add(toMonit)
	if err != nil {
		t.Errorf("failed to assign folder to monitor")
		return nil, nil
	}
	to := &testObserver{
		Pass:     make(chan int),
		test:     t,
		passOnOp: passOn,
	}

	err = toTest.Register(to)
	if err != nil {
		t.Error("failed to register observer to the monitor")
		return nil, nil
	}
	return toTest, to
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
func testRunner(o *testObserver, t *testing.T, toRun func(*testing.T), waitBeforeFail int) {

	go toRun(t)

	ticker := time.NewTicker(time.Duration(waitBeforeFail) * time.Second)
	select {
	case <-ticker.C:
		ticker.Stop()
		t.Errorf("%v", ErrTestOutOfTime)
	case <-o.Pass:
		ticker.Stop()
		t.Log("Got to pass")
		return
	}
}
func TestCreateFile(t *testing.T) {
	toTest, to := createEnv(t, fsnotify.Create)
	go toTest.Run()
	testRunner(to, t, createFile, 5)
}

func TestRemoveFile(t *testing.T) {
	toTest, to := createEnv(t, fsnotify.Remove)
	go toTest.Run()
	testRunner(to, t, createFile, 5)
}
func TestCreateFolder(t *testing.T) {
	toTest, to := createEnv(t, fsnotify.Create)
	go toTest.Run()
	testRunner(to, t, createFolder, 5)
}

func TestWriteFile(t *testing.T) {
	toTest, to := createEnv(t, fsnotify.Write)
	go toTest.Run()
	testRunner(to, t, writeToFile, 5)
}

func TestRename(t *testing.T) {
	toTest, to := createEnv(t, fsnotify.Rename)
	go toTest.Run()
	testRunner(to, t, renameFile, 5)
}
