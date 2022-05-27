package actoraddress

import (
	"context"
	"fmt"

	"github.com/bitrainforest/filmeta-hic/core/log"

	"github.com/bitrainforest/pulsar/library/httpclient"
	"github.com/go-resty/resty/v2"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
)

var (
	proxyUri = "https://rpc.%s.filmeta.pro/rpc/v0"
)

const (
	CalibHost = "calibnet"
	MainHost  = "mainnet"
)

type (
	Param struct {
		Jsonrpc string        `json:"jsonrpc"`
		Id      int           `json:"id"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}
	Resp struct {
		Jsonrpc string `json:"jsonrpc"`
		Id      int    `json:"id"`
		Error   struct {
			Code int64  `json:"code"`
			Msg  string `json:"message"`
		} `json:"error"`
		Result string `json:"result"`
	}
)

type (
	ProxyActorAddress struct {
		client  *resty.Client
		uri     string
		netWork address.Network
	}
)

func NewProxyActorAddress() *ProxyActorAddress {
	p := &ProxyActorAddress{
		client:  httpclient.NewDefaultHttpClient(),
		uri:     fmt.Sprintf(proxyUri, MainHost),
		netWork: address.Mainnet,
	}
	netWork := address.CurrentNetwork
	if netWork == address.Testnet {
		p.netWork = netWork
		p.uri = fmt.Sprintf(proxyUri, CalibHost)
	}
	return p
}

func (p *ProxyActorAddress) GetActorAddress(ctx context.Context, next *types.TipSet,
	a address.Address) (address.Address, error) {
	var (
		req  Param
		resp Resp
	)
	req.Method = "Filecoin.StateLookupID"
	req.Jsonrpc = "2.0"
	req.Id = 0
	req.Params = append(req.Params, a.String())
	req.Params = append(req.Params, []interface{}{})
	_, err := p.client.R().SetHeader("Content-Type", "application/json").
		SetBody(req).SetResult(&resp).Post(p.uri)
	if err != nil {
		return a, err
	}
	if resp.Error.Code != 0 {
		log.Errorf("get actor address error: %+v", resp.Error)
		return a, fmt.Errorf("address:%v err:%v", a.String(), resp.Error.Msg)
	}
	if resp.Result != "" {
		return address.NewFromString(resp.Result)
	}
	// return the original address
	return a, nil
}
