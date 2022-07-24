package viewchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/meverselabs/meverse/service/txsearch"
)

type INode interface {
	AddTx(tx *types.Transaction, sig common.Signature) error
}

type viewchain struct {
	api     *apiserver.APIServer
	chainID *big.Int
	ts      txsearch.ITxSearch
	cn      *chain.Chain
	st      *chain.Store
	in      INode
}

func NewViewchain(api *apiserver.APIServer, ts txsearch.ITxSearch, cn *chain.Chain, st *chain.Store, in INode) {
	v := &viewchain{
		api:     api,
		chainID: cn.Provider().ChainID(),
		ts:      ts,
		cn:      cn,
		st:      st,
		in:      in,
	}

	s, err := v.api.JRPC("view")
	if err != nil {
		panic(err)
	}

	// chainID := "0xffff"
	s.Set("version", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return fmt.Sprintf("0x%x", v.chainID.Uint64()), nil
	})
	s.Set("chainId", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return fmt.Sprintf("0x%x", v.chainID.Uint64()), nil
	})
	s.Set("maintoken", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return v.cn.NewContext().MainToken().String(), nil
	})
	s.Set("blockNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return cn.Provider().Height(), nil
	})
	s.Set("getBlockByNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height, err := arg.Uint32(0)
		if err != nil {
			str, err2 := arg.String(0)
			if err2 != nil {
				return nil, err
			}
			cheight := cn.Provider().Height()
			if str == "latest" {
				height = cheight
			}
		}
		return v.cn.Provider().Block(uint32(height))
	})
	s.Set("getBlockByHash", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		bhash, _ := arg.String(0)
		cleaned := strings.Replace(bhash, "0x", "", -1)
		bs, err := hex.DecodeString(cleaned)
		if err != nil {
			return nil, err
		}
		var hs hash.Hash256
		if len(hs) != len(bs) {
			return nil, errors.New("invalid hash length")
		}
		copy(hs[:], bs[:])

		hei, err := ts.BlockHeight(hs)
		if err != nil {
			return nil, err
		}
		if uint32(hei) > cn.Provider().Height() {
			hei = 1
		}
		return v.cn.Provider().Block(uint32(hei))
	})
	s.Set("getBalance", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addrStr, _ := arg.String(0)
		mainaddr := cn.NewContext().MainToken()
		if mainaddr == nil {
			return "0x0", nil
		}

		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return nil, err
		}

		return v.getTokenBalanceOf(*mainaddr, addr)
	})
	s.Set("gasPrice", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		// return "0x3B9ACA00", nil
		return "0xE8D4A51000", nil
	})

	s.Set("estimateGas", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		// return "0xcf08", nil 79500
		// return "0x1DCD6500", nil
		return "0x07A120", nil
	})

	s.Set("getCode", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		// check is contract
		return "0x", nil
	})

	s.Set("seq", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addrStr, _ := arg.String(0)
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return nil, err
		}
		seq := v.cn.NewContext().AddrSeq(addr)
		return "0x" + strconv.FormatUint(seq, 16), nil
	})

	s.Set("getTxByHash", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		txhash, _ := arg.String(0)
		log.Println("eth_getTransactionReceipt", txhash)

		cleaned := strings.Replace(txhash, "0x", "", -1)

		bs, err := hex.DecodeString(cleaned)
		if err != nil {
			return nil, err
		}
		var hs hash.Hash256
		if len(hs) != len(bs) {
			return nil, errors.Errorf("invalid hash length want 32, got %v", len(bs))
		}
		copy(hs[:], bs[:])

		TxID, err := ts.TxIndex(hs)
		if err != nil {
			return nil, err
		}

		if TxID.Err != nil {
			return hs, TxID.Err
		}

		b, err := v.st.Block(TxID.Height)
		if err != nil {
			return nil, err
		}
		if int(TxID.Index) >= len(b.Body.Transactions) {
			return nil, errors.New("invalid txhash")
		}
		tx := b.Body.Transactions[TxID.Index]
		if !tx.IsEtherType {
			return nil, errors.New("invalid txhash")
		}

		return tx, nil
	})

	s.Set("isContract", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contract, err := arg.String(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		cont, err := common.ParseAddress(contract)
		if err != nil {
			return nil, err
		}

		ctx := v.cn.NewContext()
		return ctx.IsContract(cont), nil
	})

	s.Set("call", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contract, err := arg.String(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		method, err := arg.String(1)
		if err != nil {
			return nil, errors.New("method not allow")
		}
		param, err := arg.Array(2)
		if err != nil {
			return nil, errors.New("parameter not allow")
		}
		for _, i := range param {
			if i == nil {
				return nil, errors.New("nil params")
			}
		}
		from, _ := arg.String(3)
		return v.Call(contract, from, method, param)
	})

	s.Set("multi_call", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contract, err := arg.Array(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		method, err := arg.Array(1)
		if err != nil {
			return nil, errors.New("method not allow")
		}
		params, err := arg.Array(2)
		if err != nil {
			return nil, errors.New("parameter not allow")
		}
		for _, i := range params {
			if i == nil {
				return nil, errors.New("nil params")
			}
		}
		paramss := make([][]interface{}, len(params))
		for i, p := range params {
			if ps, ok := p.([]interface{}); ok {
				paramss[i] = ps
			} else {
				paramss[i] = []interface{}{ps}
			}
		}
		from, _ := arg.String(3)

		if len(contract) != len(method) {
			return nil, errors.New("not matched contract method pair")
		}
		contracts := []string{}
		methods := []string{}
		for i, icont := range contract {
			cont, ok := icont.(string)
			if !ok {
				return nil, errors.New("invalid contract param(not string)")
			}
			methodStr, ok := method[i].(string)
			if !ok {
				return nil, errors.New("invalid method param(not string)")
			}
			contracts = append(contracts, cont)
			methods = append(methods, methodStr)
		}

		return v.MultiCall(contracts, from, methods, paramss)
	})

	s.Set("searchMap", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		searchMap, err := arg.Map(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		startBlock, err := arg.Uint32(1)
		if err != nil {
			return nil, fmt.Errorf("startBlock(%v) not allow", err)
		}
		lastBlock, err := arg.Uint32(2)
		if err != nil {
			lastBlockStr, _ := arg.String(2)
			if lastBlockStr == "latest" {
				lastBlock = st.TargetHeight()
			} else {
				return nil, errors.New("method not allow")
			}
		}
		th := st.TargetHeight()
		if startBlock > th {
			startBlock = th
		}
		if lastBlock > th {
			lastBlock = th
		}
		if startBlock > lastBlock {
			startBlock, lastBlock = lastBlock, startBlock
		}
		if lastBlock-startBlock > 1024 {
			return nil, errors.New("search ragne less then 1024")
		}
		smap := map[common.Address]map[string]bool{}
		for cont, methods := range searchMap {
			var m map[string]bool
			var ok bool
			contAddr := common.HexToAddress(cont)
			if m, ok = smap[contAddr]; !ok {
				m = map[string]bool{}
				smap[contAddr] = m
			}
			if methodStr, ok := methods.(string); ok {
				m[strings.ToLower(methodStr)] = true
			} else if methodArr, ok := methods.([]interface{}); ok {
				for _, s := range methodArr {
					if methodStr, ok = s.(string); ok {
						m[strings.ToLower(methodStr)] = true
					}
				}
			} else if methodArr, ok := methods.([]string); ok {
				for _, methodStr := range methodArr {
					m[strings.ToLower(methodStr)] = true
				}
			}
		}

		return v.Search(smap, startBlock, lastBlock), nil
	})

	s.Set("search", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contract, err := arg.Array(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		method, err := arg.Array(1)
		if err != nil {
			return nil, errors.New("method not allow")
		}
		startBlock, err := arg.Uint32(2)
		if err != nil {
			return nil, errors.New("startBlock not allow")
		}
		lastBlock, err := arg.Uint32(3)
		if err != nil {
			lastBlockStr, _ := arg.String(3)
			if lastBlockStr == "latest" {
				lastBlock = st.TargetHeight()
			} else {
				return nil, errors.New("method not allow")
			}
		}
		if len(contract) != len(method) {
			return nil, errors.New("not matched contract method pair")
		}
		if startBlock > lastBlock {
			startBlock, lastBlock = lastBlock, startBlock
		}
		if lastBlock-startBlock > 1024 {
			return nil, errors.New("search ragne less then 1024")
		}
		th := st.TargetHeight()
		if startBlock > th || lastBlock > th {
			return nil, fmt.Errorf("current block height is %v. requested further heights", th)
		}
		smap := map[common.Address]map[string]bool{}
		// contracts := []common.Address{}
		// methods := []string{}
		for i, icont := range contract {
			cont, ok := icont.(string)
			if !ok {
				return nil, errors.New("invalid contract param(not string)")
			}
			methodStr, ok := method[i].(string)
			if !ok {
				return nil, errors.New("invalid method param(not string)")
			}
			contAddr := common.HexToAddress(cont)
			var m map[string]bool
			if m, ok = smap[contAddr]; !ok {
				m = map[string]bool{}
				smap[contAddr] = m
			}
			m[strings.ToLower(methodStr)] = true
			// contracts = append(contracts, common.HexToAddress(cont))
			// methods = append(methods, methodStr)
		}

		return v.Search(smap, startBlock, lastBlock), nil
	})

	s.Set("calcRewardPower", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contract, err := arg.String(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		cont, err := common.ParseAddress(contract)
		if err != nil {
			return nil, err
		}
		return v.CalcRewardPower(cont)
	})
	s.Set("rewardPolicy", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contract, err := arg.String(0)
		if err != nil {
			return nil, errors.New("need contract address")
		}
		cont, err := common.ParseAddress(contract)
		if err != nil {
			return nil, err
		}
		p, err := v.RewardPolicy(cont)
		if err != nil {
			return nil, err
		}
		return map[string]int{
			"Alpha": int(p.AlphaEfficiency1000),
			"Omega": int(p.OmegaEfficiency1000),
			"Sigma": int(p.SigmaEfficiency1000),
		}, nil
	})
	s.Set("formulatorCount", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		tokenStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		token := common.HexToAddress(tokenStr)
		return v.GetFormulatorCount(token), nil
	})
	s.Set("rtx", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		method, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		sContAddr, err := arg.String(1)
		if err != nil {
			return nil, err
		}
		contAddr := common.HexToAddress(sContAddr)
		tim := uint64(time.Now().UnixNano())
		tx := &types.Transaction{
			ChainID:   v.chainID,
			Timestamp: tim,
			To:        contAddr,
			Method:    method,
		}

		switch method {
		case "Burn":
			sBi, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			am, err := amount.ParseAmount(sBi)
			if err != nil {
				return nil, err
			}
			bs := bin.TypeWriteAll(am)
			bsw := bin.TypeWriteAll(tx.Method, tx.To, tim, bs)
			tx.Args = bs
			strResult := hex.EncodeToString(bsw)
			return []interface{}{tx.HashSig().String(), strResult}, nil

		case "Transfer":
			toStr, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			to := common.HexToAddress(toStr)
			sBi, err := arg.String(3)
			if err != nil {
				return nil, err
			}
			am, err := amount.ParseAmount(sBi)
			if err != nil {
				return nil, err
			}
			bs := bin.TypeWriteAll(to, am)
			bsw := bin.TypeWriteAll(tx.Method, tx.To, tim, bs)
			tx.Args = bs
			strResult := hex.EncodeToString(bsw)
			return []interface{}{tx.HashSig().String(), strResult}, nil
		case "TokenIndexIn":
			Platform, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			ercHash, err := arg.String(3)
			if err != nil {
				return nil, err
			}
			sTo, err := arg.String(4)
			if err != nil {
				return nil, err
			}
			to := common.HexToAddress(sTo)
			sBi, err := arg.String(5)
			if err != nil {
				return nil, err
			}
			bi, ok := big.NewInt(0).SetString(sBi, 10)
			if !ok {
				return nil, errors.New("invalid amount")
			}
			am := amount.NewAmountFromBytes(bi.Bytes())
			bs := bin.TypeWriteAll(Platform, ercHash, to, am)
			tx.Args = bs
			bsw := bin.TypeWriteAll(tx.Method, tx.To, tim, bs)

			return []interface{}{
				tx.HashSig().String(),
				hex.EncodeToString(bsw),
			}, nil
		case "TokenLeave":
			CoinTXID, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			ERC20TXID, err := arg.String(3)
			if err != nil {
				return nil, err
			}
			Platform, err := arg.String(4)
			if err != nil {
				return nil, err
			}
			bs := bin.TypeWriteAll(CoinTXID, ERC20TXID, Platform)
			tx.Args = bs
			bsw := bin.TypeWriteAll(tx.Method, tx.To, tim, bs)
			strResult := hex.EncodeToString(bsw)
			return []interface{}{tx.HashSig().String(), strResult}, nil
		}
		return nil, errors.New("not support tx")
	})
	s.Set("srtx", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		ssig, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		body, err := arg.String(1)
		if err != nil {
			return nil, err
		}
		bs, err := hex.DecodeString(ssig)
		if err != nil {
			return nil, err
		}
		sig := common.Signature(bs)

		bs, err = hex.DecodeString(body)
		if err != nil {
			return nil, err
		}
		is, err := bin.TypeReadAll(bs, -1)
		if err != nil {
			return nil, err
		}
		method := is[0].(string)
		contAddr := is[1].(common.Address)
		tim := is[2].(uint64)
		bs = is[3].([]byte)
		tx := &types.Transaction{
			ChainID:   v.chainID,
			Timestamp: tim,
			To:        contAddr,
			Method:    method,
			Args:      bs,
		}
		if len(is) > 4 {
			seq, ok := is[4].(uint64)
			if !ok {
				return nil, errors.New("invalid parameter")
			}
			tx.Seq = seq
			tx.UseSeq = true
		}

		return tx.HashSig().String(), in.AddTx(tx, sig)
	})
	s.Set("clientVersion", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return GetVersion(), nil
	})
}

func (v *viewchain) getTokenBalanceOf(conAddr common.Address, addr common.Address) (string, error) {
	ctx := v.cn.NewContext()
	con, err := ctx.Contract(conAddr)
	if err != nil {
		return "", err
	}
	if cont, ok := con.(*token.TokenContract); ok {
		cc := ctx.ContractLoader(cont.Address())
		if err != nil {
			return "0", err
		}
		am := cont.BalanceOf(cc, addr)
		return am.String(), nil
	}
	return "", errors.New("not match contract")
}

func (v *viewchain) MultiCall(contract []string, from string, methods []string, paramss [][]interface{}) (interface{}, error) {
	caller := NewViewCaller(v.cn)
	toAddrs := []common.Address{}
	for _, addrStr := range contract {
		toAddr, err := common.ParseAddress(addrStr)
		if err != nil {
			return nil, err
		}
		toAddrs = append(toAddrs, toAddr)
	}
	output, err := caller.MultiExecute(toAddrs, from, methods, paramss)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// map[common.Address]map[string]bool{}
func (v *viewchain) Search(searchMap map[common.Address]map[string]bool, start, end uint32) (txList []map[string]string) {
	txList = []map[string]string{}
	for i := start; i <= end; i++ {
		b, err := v.st.Block(i)
		if err != nil {
			continue
		}

		for _, e := range b.Body.Events {
			if e.Type != types.EventTagCallHistory {
				continue
			}
			bf := bytes.NewBuffer(e.Result)
			mc := &types.MethodCallEvent{}
			if _, err := mc.ReadFrom(bf); err != nil {
				continue
			}
			sm, ok := searchMap[mc.To]
			if !ok {
				continue
			}
			if _, ok := sm[strings.ToLower(mc.Method)]; !ok {
				continue
			}
			if len(b.Body.Transactions) <= int(e.Index) {
				continue
			}
			tx := b.Body.Transactions[e.Index]

			m := map[string]string{
				"Contract":  mc.To.String(),
				"From":      mc.From.String(),
				"TxFrom":    tx.From.String(),
				"TxTo":      tx.To.String(),
				"Method":    mc.Method,
				"Height":    fmt.Sprintf("%v", b.Header.Height),
				"Timestamp": fmt.Sprintf("%v", b.Header.Timestamp),
				"Index":     fmt.Sprintf("%v", e.Index),
				"Hash":      tx.Hash(b.Header.Height).String(),
				"TXID":      types.TransactionID(b.Header.Height, e.Index),
			}
			args, err := json.Marshal(mc.Args)
			if err == nil {
				m["Args"] = string(args)
			}
			result, err := json.Marshal(mc.Result)
			if err == nil {
				m["Result"] = string(result)
			}

			txList = append(txList, m)
		}
	}
	return
}

func (v *viewchain) Call(contract, from, method string, params []interface{}) (interface{}, error) {
	toAddr, err := common.ParseAddress(contract)
	if err != nil {
		return nil, err
	}
	caller := NewViewCaller(v.cn)
	output, err := caller.Execute(toAddr, from, method, params)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (v *viewchain) RewardPolicy(cont common.Address) (*formulator.RewardPolicy, error) {
	var tagRewardPolicy = byte(0x03)
	ctx := v.cn.NewContext()
	rewardPolicy := &formulator.RewardPolicy{}
	if _, err := rewardPolicy.ReadFrom(bytes.NewReader(ctx.Data(cont, common.Address{}, []byte{tagRewardPolicy}))); err != nil {
		return nil, err
	}

	return rewardPolicy, nil
}

func (v *viewchain) GetFormulatorCount(cont common.Address) uint32 {
	ctx := v.cn.NewContext()
	if bs := ctx.Data(cont, common.Address{}, []byte{tagFormulatorCount}); len(bs) > 0 {
		return bin.Uint32(bs)
	}
	return 0
}

var (
	tagFormulatorPolicy     = byte(0x02)
	tagRewardPolicy         = byte(0x03)
	tagFormulator           = byte(0x10)
	tagFormulatorReverse    = byte(0x12)
	tagFormulatorCount      = byte(0x13)
	tagStakingAmount        = byte(0x20)
	tagStakingAmountReverse = byte(0x22)
	tagStakingAmountCount   = byte(0x23)
	tagStackRewardMap       = byte(0x31)
)

func (v *viewchain) CalcRewardPower(cont common.Address) (interface{}, error) {
	ctx := v.cn.NewContext()

	toFormulatorReverseKey := func(Num uint32) []byte {
		bs := make([]byte, 6)
		bs[0] = tagFormulatorReverse
		bin.PutUint32(bs[1:], Num)
		return bs
	}
	toStakingAmountReverseKey := func(Num uint32) []byte {
		bs := make([]byte, 6)
		bs[0] = tagStakingAmountReverse
		bin.PutUint32(bs[1:], Num)
		return bs
	}
	toStakingAmountKey := func(addr common.Address) []byte {
		bs := make([]byte, 1+common.AddressLength)
		bs[0] = tagStakingAmount
		copy(bs[1:], addr[:])
		return bs
	}

	rewardPolicy := &formulator.RewardPolicy{}
	if _, err := rewardPolicy.ReadFrom(bytes.NewReader(ctx.Data(cont, common.Address{}, []byte{tagRewardPolicy}))); err != nil {
		return nil, err
	}
	formulatorPolicy := &formulator.FormulatorPolicy{}
	if _, err := formulatorPolicy.ReadFrom(bytes.NewReader(ctx.Data(cont, common.Address{}, []byte{tagFormulatorPolicy}))); err != nil {
		return nil, err
	}

	formulatorMap := map[common.Address]*formulator.Formulator{}
	if bs := ctx.Data(cont, common.Address{}, []byte{tagFormulatorCount}); len(bs) > 0 {
		Count := bin.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], ctx.Data(cont, common.Address{}, toFormulatorReverseKey(i)))
			fr := &formulator.Formulator{}
			bs := ctx.Data(cont, addr, []byte{tagFormulator})
			if len(bs) == 0 {
				return nil, errors.WithStack(formulator.ErrNotExistFormulator)
			}
			if _, err := fr.ReadFrom(bytes.NewReader(bs)); err != nil {
				return nil, err
			}
			formulatorMap[addr] = fr
		}
	}

	StackRewardMap := map[common.Address]*amount.Amount{}
	if bs := ctx.Data(cont, common.Address{}, []byte{tagStackRewardMap}); len(bs) > 0 {
		if err := types.UnmarshalAddressAmountMap(bs, StackRewardMap); err != nil {
			return nil, err
		}
	}

	RewardPowerSum := amount.NewAmount(0, 0)
	for _, fr := range formulatorMap {
		var effic uint32 = 0
		switch fr.Type {
		case formulator.AlphaFormulatorType:
			effic = rewardPolicy.AlphaEfficiency1000
		case formulator.SigmaFormulatorType:
			effic = rewardPolicy.SigmaEfficiency1000
		case formulator.OmegaFormulatorType:
			effic = rewardPolicy.OmegaEfficiency1000
		default:
			return nil, errors.WithStack(formulator.ErrUnknownFormulatorType)
		}
		am := fr.Amount.MulC(int64(effic)).DivC(1000)
		RewardPowerSum = RewardPowerSum.Add(am)
	}

	gr, err := v.st.Generators()
	if err != nil {
		return nil, err
	}
	for _, HyperAddress := range gr {
		am := formulatorPolicy.HyperAmount.MulC(int64(rewardPolicy.HyperEfficiency1000)).DivC(1000)
		RewardPowerSum = RewardPowerSum.Add(am)

		var Number uint32
		if bs := ctx.Data(cont, HyperAddress, []byte{tagStakingAmountCount}); len(bs) > 0 {
			Number = bin.Uint32(bs)
		}
		for i := uint32(0); i <= Number; i++ {
			if bs := ctx.Data(cont, HyperAddress, toStakingAmountReverseKey(i)); len(bs) > 0 {
				sAddr := common.BytesToAddress(bs)
				if bs := ctx.Data(cont, HyperAddress, toStakingAmountKey(sAddr)); len(bs) > 0 {
					RewardPowerSum = RewardPowerSum.Add(amount.NewAmountFromBytes(bs).MulC(int64(rewardPolicy.StakingEfficiency1000)).DivC(1000))
				}
			}
		}
	}

	TotalReward := rewardPolicy.RewardPerBlock.MulC(int64(172800))

	TotalReward = TotalReward.MulC(1000000000000000000)
	TotalReward.Int.Div(TotalReward.Int, RewardPowerSum.Int)

	Fee := TotalReward.MulC(int64(rewardPolicy.MiningFee1000)).DivC(1000)
	TotalReward.Int.Sub(TotalReward.Int, Fee.Int)

	return TotalReward, nil
}
