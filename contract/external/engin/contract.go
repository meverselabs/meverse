package engin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"plugin"
	"sync"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/external/deployer"
	"github.com/meverselabs/meverse/core/types"
)

var enginCache = map[string]types.IEngin{}

type EnginContract struct {
	addr   common.Address
	master common.Address
}

func (cont *EnginContract) Name() string {
	return "EnginContract"
}

func (cont *EnginContract) Address() common.Address {
	return cont.addr
}

func (cont *EnginContract) Master() common.Address {
	return cont.master
}

func (cont *EnginContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *EnginContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &EnginContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	return nil
}

func (cont *EnginContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *EnginContract) enginID(Name string, Version uint32) string {
	return fmt.Sprintf("%v_%v_%v", cont.addr.String(), Name, Version)
}

func (cont *EnginContract) NextEnginVersion(cc *types.ContractContext, name string) uint32 {
	version := cont.EnginVersion(cc, name)
	version++
	cc.SetContractData(makeEnginVersionKey(name), bin.Uint32Bytes(version))
	return version
}

func (cont *EnginContract) NextContractSeq(cc *types.ContractContext) uint32 {
	seqbs := cc.ContractData([]byte{tagContractSeq})
	var seq uint32
	if len(seqbs) != 0 {
		seq = bin.Uint32(seqbs)
	}
	seq++
	cc.SetContractData([]byte{tagContractSeq}, bin.Uint32Bytes(seq))
	return seq
}

var fileLock sync.Mutex

func (cont *EnginContract) saveEngin(Name string, Version uint32, EnginURL string) error {
	ID := cont.enginID(Name, Version)
	if _, has := enginCache[ID]; has {
		return nil
	}
	if err := os.MkdirAll("./engin/", 0744); err != nil {
		return err
	}

	fileLock.Lock()
	path := fmt.Sprintf("./engin/%v.so", ID)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		out, err := os.Create(path)
		if err != nil {
			fileLock.Unlock()
			return err
		}
		defer out.Close()
		resp, err := http.Get(EnginURL)
		if err != nil {
			fileLock.Unlock()
			return err
		}
		defer resp.Body.Close()
		n, err := io.Copy(out, resp.Body)
		if err != nil {
			fileLock.Unlock()
			return err
		}
		if n == 0 {
			fileLock.Unlock()
			return errors.New("engin is empty")
		}
	}
	fileLock.Unlock()

	// if err := ioutil.WriteFile(path, Engin, 0644); err != nil {
	// 	return err
	// }
	pg, err := plugin.Open(path)
	if err != nil {
		return err
	}

	mt, err := pg.Lookup("Engin")
	if err != nil {
		return err
	}
	var eg types.IEngin
	eg, ok := mt.(types.IEngin)
	if !ok {
		return errors.New("invalid engin")
	}

	enginCache[ID] = eg
	return nil
}

func (cont *EnginContract) loadEngin(cc *types.ContractContext, Name string, Version uint32) (types.IEngin, error) {
	_Version := cont.EnginVersion(cc, Name)
	if _Version < Version {
		if _Version == 0 {
			return nil, fmt.Errorf("not exist engin current version is %v, get %v", _Version, Version)
		}
		return nil, fmt.Errorf("not exist engin version last version is %v", _Version)
	}
	ID := cont.enginID(Name, Version)
	EnginURL := cc.ContractData(makeEnginURLKey(Name, Version))

	eg, has := enginCache[ID]
	if !has {
		if len(EnginURL) == 0 {
			return nil, errors.New("engin not exist")
		}

		err := cont.saveEngin(Name, Version, string(EnginURL))
		if err != nil {
			return nil, err
		}
		eg, has = enginCache[ID]
		if !has {
			return nil, errors.New("engin not loaded")
		}
	}

	return eg, nil
}

//////////////////////////////////////////////////
// Public admin only Writer Functions
//////////////////////////////////////////////////

func (cont *EnginContract) addEngin(cc *types.ContractContext, Name string, Description string, EnginURL string) error {
	if cc.From() != cont.master {
		return errors.New("not owner")
	}
	if len(EnginURL) == 0 {
		return errors.New("engin not provided")
	}
	Version := cont.NextEnginVersion(cc, Name)
	cc.SetContractData(makeDescriptionKey(Name, Version), []byte(Description))
	err := cont.saveEngin(Name, Version, EnginURL)
	if err != nil {
		return err
	}
	cc.SetContractData(makeEnginURLKey(Name, Version), []byte(EnginURL))

	return nil
}

func (cont *EnginContract) EnginDescription(cc *types.ContractContext, Name string, Version uint32) (string, error) {
	DescriptionBs := cc.ContractData(makeDescriptionKey(Name, Version))
	if len(DescriptionBs) == 0 {
		return "", errors.New("not exist")
	}
	return string(DescriptionBs), nil
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *EnginContract) deploryContract(cc *types.ContractContext, EnginName string, EnginVersion uint32, contract []byte, InitArgs []interface{}, updateable bool) (contAddr common.Address, err error) {
	_, ClassID := types.GetContractClassID(&deployer.DeployerContract{})
	base := make([]byte, 1+common.AddressLength+8+4)
	base[0] = 0xff
	copy(base[1:], cont.addr[:])
	copy(base[1+common.AddressLength:], bin.Uint64Bytes(ClassID))
	copy(base[1+common.AddressLength+8:], bin.Uint32Bytes(cont.NextContractSeq(cc)))

	height := cc.TargetHeight()
	if height > 0 {
		bs := bin.Uint32Bytes(height)
		base = append(base, bs...)
	}
	h := hash.Hash(base)
	contAddr = common.BytesToAddress(h[12:])

	deployerConstrunction := &deployer.DeployerContractConstruction{
		EnginAddress: cont.addr,
		EnginName:    EnginName,
		EnginVersion: EnginVersion,
		Binary:       contract,
		Owner:        cc.From(),
		Updateable:   updateable,
	}
	bs, _, err := bin.WriterToBytes(deployerConstrunction)
	if err != nil {
		return
	}

	externalCont, err := cc.DeployContractWithAddress(cc.From(), ClassID, contAddr, bs)
	if err != nil {
		return
	}

	externalAddr := externalCont.Address()
	if externalAddr != contAddr {
		err = errors.New("invalid contract address")
		return
	}
	_, err = cc.Exec(cc, externalAddr, "InitContract", []interface{}{contract, InitArgs})
	return
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *EnginContract) EnginVersion(cc *types.ContractContext, name string) uint32 {
	seqbs := cc.ContractData(makeEnginVersionKey(name))
	var seq uint32
	if len(seqbs) != 0 {
		seq = bin.Uint32(seqbs)
	}
	return seq
}
