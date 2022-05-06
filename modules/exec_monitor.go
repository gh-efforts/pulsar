package modules

import (
	"github.com/bitrainforest/pulsar/libs/exec_monitor"

	"github.com/filecoin-project/lotus/chain/store"
)

func NewBufferedExecMonitor(cs *store.ChainStore) *exec_monitor.BufferedExecMonitor {
	return exec_monitor.NewBufferedExecMonitor(cs)
}
