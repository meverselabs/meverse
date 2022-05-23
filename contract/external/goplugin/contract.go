package goplugin

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"plugin"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/external/goplugin/goplugincontext"
	"github.com/meverselabs/meverse/core/types"
)

type PluginContract struct {
	addr   common.Address
	master common.Address
}

func (cont *PluginContract) Name() string {
	return "plugin"
}

func (cont *PluginContract) Address() common.Address {
	return cont.addr
}

func (cont *PluginContract) Master() common.Address {
	return cont.master
}

func (cont *PluginContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

var pluginCache map[common.Address]*plugin.Plugin

func (cont *PluginContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &PluginContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	if pluginCache == nil {
		pluginCache = make(map[common.Address]*plugin.Plugin)
	}
	if _, has := pluginCache[cont.addr]; has {
		return nil
	}
	if err := os.MkdirAll("./plugin/", 0755); err != nil {
		return err
	}
	path := fmt.Sprintf("./plugin/%v.so", cont.addr.String())

	if err := ioutil.WriteFile(path, data.Bin, 0644); err != nil {
		return err
	}
	if pg, err := plugin.Open(path); err != nil {
		return err
	} else {
		pluginCache[cont.addr] = pg
	}

	return nil
}

func (cont *PluginContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////
func (cont *PluginContract) loadPlugin(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

func (cont *PluginContract) ContractInvoke(cc *types.ContractContext, method string, params []interface{}) ([]interface{}, error) {
	pg, has := pluginCache[cont.addr]
	if !has {
		return nil, errors.New("plugin not found")
	}

	mt, err := pg.Lookup("Contract")
	if err != nil {
		return nil, err
	}
	var pc IPluginContract
	pc, ok := mt.(IPluginContract)
	if !ok {
		return nil, errors.New("invalid plugin")
	}

	return pc.Invoke(getIcc(cont, cc), mt, method, params...)
}

func getIcc(cont types.Contract, cc *types.ContractContext) ContractContext {
	return goplugincontext.NewPluginContextContract(cont, cc)
}
