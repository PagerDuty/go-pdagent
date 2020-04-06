package persistentqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventqueue"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"go.uber.org/zap"
	"os"
	"path"
	"testing"
)

// TODO Consider replacing with `ioutil.TempDir` and/or `ioutil.TempFile`.
const tmpDir = "../../tmp"

var tmpDbFile = path.Join(tmpDir, "test.db")

type MockEventQueue struct {
	logger *zap.SugaredLogger
}

func NewMockEventQueue() *MockEventQueue {
	return &MockEventQueue{
		logger: common.Logger.Named("MockEventQueue"),
	}
}

func (q *MockEventQueue) Shutdown() {
	q.logger.Debug("Shutdown called.")
}

func (q *MockEventQueue) Enqueue(_ eventsapi.Event, c chan<- eventqueue.Response) error {
	q.logger.Debug("Enqueue called.")
	go func() {
		q.logger.Debug("Response sent called.")
		c <- eventqueue.Response{}
	}()

	q.logger.Debug("Enqueue returning.")
	return nil
}

// Clean up any existing tmp directory contents and create if necessary.
func setup(t *testing.T) {
	if err := os.RemoveAll(tmpDbFile); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(tmpDir, 0777); err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
}

// Clean up any leftover tmp files.
//
// Useful to comment out when troubleshooting DB issues.
func teardown(t *testing.T) {
	if err := os.RemoveAll(tmpDbFile); err != nil {
		t.Fatal(err)
	}
}
