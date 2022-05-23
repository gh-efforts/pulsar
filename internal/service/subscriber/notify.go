package subscriber

import (
	"github.com/bitrainforest/filmeta-hic/core/log"

	"github.com/bitrainforest/filmeta-hic/core/threading"

	"github.com/bitrainforest/filmeta-hic/model"

	"github.com/nats-io/nats.go"

	"sync"
)

type Notify interface {
	Notify(wgDone *sync.WaitGroup, appIds []string, msg *model.Message) error
	Close()
}

type notify struct {
	connect *nats.Conn
}

func NewNotify(natsUri string) (Notify, error) {
	connect, err := nats.Connect(natsUri)
	if err != nil {
		return nil, err
	}
	return &notify{connect: connect}, nil
}

func (n *notify) Notify(wgDone *sync.WaitGroup, appIds []string, msg *model.Message) error {
	defer wgDone.Done()
	var wg sync.WaitGroup
	for _, appId := range appIds {
		wg.Add(1)
		to := appId
		threading.GoSafe(func() {
			defer wg.Done()
			msgByte, err := msg.Marshal()
			if err != nil {
				log.Errorf("[core.processing] marshal msg:%+v err: %s", msg, err)
				return
			}
			if err := n.connect.Publish(to, msgByte); err != nil {
				log.Errorf("[core.processing] publish appId:%v msg:%+v err: %s", to, msg, err)
			}
		})
	}
	wg.Wait()
	return nil
}

func (n *notify) Close() {
	n.connect.Close()
}