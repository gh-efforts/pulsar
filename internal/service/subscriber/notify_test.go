package subscriber

import (
	"sync"
	"sync/atomic"

	"github.com/bitrainforest/filmeta-hic/model"
)

var _ Notify = (*MockNotify)(nil)

type MockNotify struct {
	count int64
}

func (m *MockNotify) Notify(wgDone *sync.WaitGroup, appIds []string, msg *model.Message) error {
	//time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
	atomic.AddInt64(&m.count, int64(len(appIds)))
	wgDone.Done()
	return nil
}

func (m *MockNotify) Close() {
}
