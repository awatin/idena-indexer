package common

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-go/core/mempool"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-indexer/import/words"
	"github.com/idena-network/idena-indexer/incoming"
)

type TestWordsLoader struct {
}

func (l *TestWordsLoader) LoadWords() ([]words.Word, error) {
	return nil, nil
}

type TestListener struct {
	bus         eventbus.Bus
	collector   collector.StatsCollector
	handleBlock func(block *types.Block)
	appState    *appstate.AppState
	nodeCtx     *node.NodeCtx
	keysPool    *mempool.KeysPool
}

func NewTestListener(
	bus eventbus.Bus,
	collector collector.StatsCollector,
	appState *appstate.AppState,
	nodeCtx *node.NodeCtx,
) incoming.Listener {
	return &TestListener{
		bus:       bus,
		collector: collector,
		appState:  appState,
		nodeCtx:   nodeCtx,
		keysPool:  &mempool.KeysPool{},
	}
}

func (l *TestListener) Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64) {
	l.handleBlock = handleBlock
	l.bus.Subscribe(events.AddBlockEventID,
		func(e eventbus.Event) {
			newBlockEvent := e.(*events.NewBlockEvent)
			l.handleBlock(newBlockEvent.Block)
		})
}

func (l *TestListener) AppStateReadonly(height uint64) (*appstate.AppState, error) {
	return l.appState, nil
}

func (l *TestListener) AppState() *appstate.AppState {
	return nil
}

func (l *TestListener) NodeCtx() *node.NodeCtx {
	return l.nodeCtx
}

func (l *TestListener) StatsCollector() collector.StatsCollector {
	return l.collector
}

func (l *TestListener) Blockchain() *blockchain.Blockchain {
	return nil
}

func (l *TestListener) Flipper() *flip.Flipper {
	return nil
}

func (l *TestListener) Config() *config.Config {
	return nil
}

func (l *TestListener) KeysPool() *mempool.KeysPool {
	return l.keysPool
}

func (l *TestListener) OfflineDetector() *blockchain.OfflineDetector {
	return nil
}

func (l *TestListener) NodeEventBus() eventbus.Bus {
	return l.bus
}

func (l *TestListener) Destroy() {

}

func (l *TestListener) WaitForStop() {

}

type TestFlipLoader struct {
}

func (l *TestFlipLoader) SubmitToLoad(cidBytes []byte, txHash common.Hash) {
}
