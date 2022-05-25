package subscriber

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nats-io/nats.go"

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

func Test_Notify(t *testing.T) {
	notify, err := NewNotify(nats.DefaultURL)
	assert.Nil(t, err)
	sub, err := NewSub([]string{"test1", "test2"}, notify)
	assert.Nil(t, err)
	testAppId := "notify_test_all_001"
	// sub address
	sub.AppendAppId(testAppId)

	ch := make(chan struct{})

	var recvCount int64

	go func() {
		nc, err := nats.Connect(nats.DefaultURL)
		assert.Nil(t, err)
		_, err = nc.QueueSubscribe(testAppId, "test", func(m *nats.Msg) {
			atomic.AddInt64(&recvCount, 1)
		})
		assert.Nil(t, err)
		ch <- struct{}{}
		select {}
	}()
	// wait nats server start
	<-ch
	var wg sync.WaitGroup

	nums := rand.Intn(100) + 100
	for i := 0; i < nums; i++ {
		wg.Add(1)
		go func() {
			err := notify.Notify(&wg, []string{testAppId}, &model.Message{})
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
	time.Sleep(2 * time.Second)
	assert.Equal(t, int64(nums), atomic.LoadInt64(&recvCount))
}
