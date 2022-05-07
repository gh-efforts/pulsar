package modules

import (
	"github.com/filecoin-project/lotus/chain/beacon"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/vm"
	"go.uber.org/fx"
)

func StateManager(lc fx.Lifecycle, cs *store.ChainStore, exec stmgr.Executor, sys vm.SyscallBuilder, us stmgr.UpgradeSchedule, bs beacon.Schedule, em stmgr.ExecMonitor) (*stmgr.StateManager, error) {
	sm, err := stmgr.NewStateManagerWithUpgradeScheduleAndMonitor(cs, exec, sys, us, bs, em)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: sm.Start,
		OnStop:  sm.Stop,
	})
	return sm, nil
}
