package formulator

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
)

type FormulatorContract struct {
	addr   common.Address
	master common.Address
}

func (cont *FormulatorContract) Name() string {
	return "FormulatorContract"
}

func (cont *FormulatorContract) Address() common.Address {
	return cont.addr
}

func (cont *FormulatorContract) Master() common.Address {
	return cont.master
}

func (cont *FormulatorContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *FormulatorContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &FormulatorContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	cc.SetContractData([]byte{tagTokenContractAddress}, data.TokenAddress[:])
	if bs, _, err := bin.WriterToBytes(&data.FormulatorPolicy); err != nil {
		return err
	} else {
		cc.SetContractData([]byte{tagFormulatorPolicy}, bs)
	}
	if bs, _, err := bin.WriterToBytes(&data.RewardPolicy); err != nil {
		return err
	} else {
		cc.SetContractData([]byte{tagRewardPolicy}, bs)
	}
	return nil
}

func (cont *FormulatorContract) OnReward(cc *types.ContractContext, b *types.Block, GenCountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	taddr := cont.TokenAddress(cc)

	rewardPolicy := &RewardPolicy{}
	if _, err := rewardPolicy.ReadFrom(bytes.NewReader(cc.ContractData([]byte{tagRewardPolicy}))); err != nil {
		return nil, err
	}
	formulatorPolicy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return nil, err
	}
	formulatorMap, err := cont.FormulatorMap(cc)
	if err != nil {
		return nil, err
	}

	CountMap := map[common.Address]uint32{}
	for addr, GenCount := range GenCountMap {
		CountMap[addr] = GenCount
	}
	CountSum := uint32(0)
	for _, GenCount := range GenCountMap {
		CountSum += GenCount
	}
	CountPerFormulator := CountSum / uint32(len(GenCountMap))
	for addr := range formulatorMap {
		CountMap[addr] = CountPerFormulator
	}

	StackRewardMap := map[common.Address]*amount.Amount{}
	if bs := cc.ContractData([]byte{tagStackRewardMap}); len(bs) > 0 {
		if err := types.UnmarshalAddressAmountMap(bs, StackRewardMap); err != nil {
			return nil, err
		}
	}

	RewardPowerSum := amount.NewAmount(0, 0)
	RewardPowerMap := map[common.Address]*amount.Amount{}
	StakingRewardPowerMap := map[common.Address]*amount.Amount{}
	Hypers := []common.Address{}
	for GenAddress, GenCount := range CountMap {
		if fr, has := formulatorMap[GenAddress]; has {
			var effic uint32 = 0
			switch fr.Type {
			case AlphaFormulatorType:
				effic = rewardPolicy.AlphaEfficiency1000
			case SigmaFormulatorType:
				effic = rewardPolicy.SigmaEfficiency1000
			case OmegaFormulatorType:
				effic = rewardPolicy.OmegaEfficiency1000
			default:
				return nil, errors.WithStack(ErrUnknownFormulatorType)
			}
			am := fr.Amount.MulC(int64(GenCount)).MulC(int64(effic)).DivC(1000)
			RewardPowerSum = RewardPowerSum.Add(am)
			RewardPowerMap[GenAddress] = am
		} else if !cc.IsGenerator(GenAddress) {
			delete(StackRewardMap, GenAddress)
		} else {
			Hypers = append(Hypers, GenAddress)

			am := formulatorPolicy.HyperAmount.MulC(int64(GenCount)).MulC(int64(rewardPolicy.HyperEfficiency1000)).DivC(1000)
			RewardPowerSum = RewardPowerSum.Add(am)
			RewardPowerMap[GenAddress] = am

			PrevAmountMap := map[common.Address]*amount.Amount{}
			if bs := cc.AccountData(GenAddress, []byte{tagStakingAmountMap}); len(bs) > 0 {
				if err := types.UnmarshalAddressAmountMap(bs, PrevAmountMap); err != nil {
					return nil, err
				}
			}
			AmountMap, err := cont.StakingAmountMap(cc, GenAddress)
			if err != nil {
				return nil, err
			}
			CurrentAmountMap := map[common.Address]*amount.Amount{}
			CrossAmountMap := map[common.Address]*amount.Amount{}
			for StakingAddress, StakingAmount := range AmountMap {
				CurrentAmountMap[StakingAddress] = StakingAmount
				if PrevStakingAmount, has := PrevAmountMap[StakingAddress]; has {
					if !PrevStakingAmount.IsZero() && !StakingAmount.IsZero() {
						if StakingAmount.Less(PrevStakingAmount) {
							CrossAmountMap[StakingAddress] = StakingAmount
						} else {
							CrossAmountMap[StakingAddress] = PrevStakingAmount
						}
					}
				}
			}

			if bs, err := types.MarshalAddressAmountMap(CurrentAmountMap); err != nil {
				return nil, err
			} else {
				cc.SetAccountData(GenAddress, []byte{tagStakingAmountMap}, bs)
			}

			StakingRewardPower := amount.NewAmount(0, 0)
			StakingPowerMap := map[common.Address]*amount.Amount{}
			if bs := cc.AccountData(GenAddress, []byte{tagStakingPowerMap}); len(bs) > 0 {
				if err := types.UnmarshalAddressAmountMap(bs, StakingPowerMap); err != nil {
					return nil, err
				}
			}
			for StakingAddress, StakingAmount := range CrossAmountMap {
				if sm, has := StakingPowerMap[StakingAddress]; has {
					StakingPowerMap[StakingAddress] = sm.Add(StakingAmount)
				} else {
					StakingPowerMap[StakingAddress] = StakingAmount
				}
				StakingRewardPower = StakingRewardPower.Add(StakingAmount.MulC(int64(GenCount)).MulC(int64(rewardPolicy.StakingEfficiency1000)).DivC(1000))
			}

			StackReward, has := StackRewardMap[GenAddress]
			if has {
				StakingPowerSum := amount.NewAmount(0, 0)
				for _, StakingPower := range StakingPowerMap {
					StakingPowerSum = StakingPowerSum.Add(StakingPower)
				}
				if !StakingPowerSum.IsZero() {
					Ratio := StackReward.Div(StakingPowerSum)
					for StakingAddress, StakingPower := range StakingPowerMap {
						StackStakingAmount := StakingPower.Mul(Ratio)
						StakingPowerMap[StakingAddress] = StakingPower.Add(StackStakingAmount)
						StakingRewardPower = StakingRewardPower.Add(StackStakingAmount.MulC(int64(GenCount)).MulC(int64(rewardPolicy.StakingEfficiency1000)).DivC(1000))
					}
				}
			}

			if bs, err := types.MarshalAddressAmountMap(StakingPowerMap); err != nil {
				return nil, err
			} else {
				cc.SetAccountData(GenAddress, []byte{tagStakingPowerMap}, bs)
			}
			StakingRewardPowerMap[GenAddress] = StakingRewardPower
			RewardPowerSum = RewardPowerSum.Add(StakingRewardPower)
		}
	}

	rewardEvent := map[common.Address]*amount.Amount{}
	var TotalForCmp *amount.Amount
	if !RewardPowerSum.IsZero() {
		TotalReward := rewardPolicy.RewardPerBlock.MulC(int64(prefix.RewardIntervalBlocks))

		is, err := cc.Exec(cc, taddr, "CollectedFee", nil)
		if err != nil {
			return nil, err
		}
		TotalFee := is[0].(*amount.Amount)
		if !TotalFee.IsZero() {
			_, err := cc.Exec(cc, taddr, "SubCollectedFee", []interface{}{TotalFee})
			if err != nil {
				return nil, err
			}
		}
		TotalReward = TotalReward.Add(TotalFee)
		TotalForCmp = amount.NewAmountFromBytes(TotalReward.Int.Bytes())

		Ratio := TotalReward.Div(RewardPowerSum)
		for RewardAddress, RewardPower := range RewardPowerMap {
			RewardAmount := RewardPower.Mul(Ratio)
			if !RewardAmount.IsZero() {
				if cc.IsGenerator(RewardAddress) {
					if err := cont.sendReward(cc, taddr, rewardEvent, RewardAddress, RewardAmount); err != nil {
						// if _, err := cc.Exec(cc, taddr, "Mint", []interface{}{RewardAddress, RewardAmount}); err != nil {
						return nil, err
					}
				} else {
					if fr, has := formulatorMap[RewardAddress]; has {
						RewardAddress = fr.Owner
					}
					fee := RewardAmount.MulC(int64(rewardPolicy.MiningFee1000)).DivC(1000)
					if !fee.IsZero() {
						if err := cont.sendReward(cc, taddr, rewardEvent, rewardPolicy.MiningFeeAddress, fee); err != nil {
							// if _, err := cc.Exec(cc, taddr, "Mint", []interface{}{rewardPolicy.MiningFeeAddress, fee}); err != nil {
							return nil, err
						}
					}
					processedAmount := RewardAmount.Sub(fee)
					if err := cont.sendReward(cc, taddr, rewardEvent, RewardAddress, processedAmount); err != nil {
						// if _, err := cc.Exec(cc, taddr, "Mint", []interface{}{RewardAddress, processedAmount}); err != nil {
						return nil, err
					}
				}
			}
		}
		for GenAddress, StakingRewardPower := range StakingRewardPowerMap {
			RewardAmount := StakingRewardPower.Mul(Ratio)
			if sm, has := StackRewardMap[GenAddress]; has {
				StackRewardMap[GenAddress] = sm.Add(RewardAmount)
			} else {
				StackRewardMap[GenAddress] = RewardAmount
			}
		}
	}
	for _, GenAddress := range Hypers {
		if StackReward, has := StackRewardMap[GenAddress]; has {
			StakingPowerMap := map[common.Address]*amount.Amount{}
			if bs := cc.AccountData(GenAddress, []byte{tagStakingPowerMap}); len(bs) > 0 {
				if err := types.UnmarshalAddressAmountMap(bs, StakingPowerMap); err != nil {
					return nil, err
				}
			}

			StakingPowerSum := amount.NewAmount(0, 0)
			for _, StakingPower := range StakingPowerMap {
				StakingPowerSum = StakingPowerSum.Add(StakingPower)
			}
			if !StakingPowerSum.IsZero() {
				CommissionSum := amount.NewAmount(0, 0)
				Ratio := StackReward.Div(StakingPowerSum)
				for StakingAddress, StakingPower := range StakingPowerMap {
					RewardAmount := StakingPower.Mul(Ratio)
					if rewardPolicy.CommissionRatio1000 > 0 {
						Commission := RewardAmount.MulC(int64(rewardPolicy.CommissionRatio1000)).DivC(1000)
						CommissionSum = CommissionSum.Add(Commission)
						RewardAmount = RewardAmount.Sub(Commission)
					}
					if !RewardAmount.IsZero() {
						// cont.addRewardMap(rewardEvent, StakingAddress, RewardAmount)
						// cont.addStakingAmount(cc, GenAddress, StakingAddress, RewardAmount)
						if err := cont.sendReward(cc, taddr, rewardEvent, StakingAddress, RewardAmount); err != nil {
							// if _, err := cc.Exec(cc, taddr, "Mint", []interface{}{GenAddress, CommissionSum}); err != nil {
							return nil, err
						}
					}
				}

				if !CommissionSum.IsZero() {
					if err := cont.sendReward(cc, taddr, rewardEvent, GenAddress, CommissionSum); err != nil {
						// if _, err := cc.Exec(cc, taddr, "Mint", []interface{}{GenAddress, CommissionSum}); err != nil {
						return nil, err
					}
				}
			}
			cc.SetAccountData(GenAddress, []byte{tagStakingPowerMap}, nil)

			delete(StackRewardMap, GenAddress)
		}
	}
	if bs, err := types.MarshalAddressAmountMap(StackRewardMap); err != nil {
		return nil, err
	} else {
		cc.SetContractData([]byte{tagStackRewardMap}, bs)
	}

	sum := amount.NewAmount(0, 0)
	for _, am := range rewardEvent {
		sum.Int.Add(sum.Int, am.Int)
	}

	if TotalForCmp.Cmp(sum.Int) < 0 {
		panic(errors.Errorf("%v %v %v", TotalForCmp.String(), sum.String(), TotalForCmp.Cmp(sum.Int)))
	}

	//log.Println("Paid at", cc.TargetHeight())
	return rewardEvent, nil
}

func (cont *FormulatorContract) sendReward(cc *types.ContractContext, taddr common.Address, rewardMap map[common.Address]*amount.Amount, RewardAddress common.Address, RewardAmount *amount.Amount) error {
	cont.addRewardMap(rewardMap, RewardAddress, RewardAmount)
	_, ExecErr := cc.Exec(cc, taddr, "Mint", []interface{}{RewardAddress, RewardAmount})
	return ExecErr
}

func (cont *FormulatorContract) addRewardMap(rewardMap map[common.Address]*amount.Amount, RewardAddress common.Address, RewardAmount *amount.Amount) {
	am, has := rewardMap[RewardAddress]
	if !has {
		am = amount.NewAmount(0, 0)
		rewardMap[RewardAddress] = am
	}
	am.Int.Add(am.Int, RewardAmount.Int)
}

func (cont *FormulatorContract) SetRewardPolicy(cc *types.ContractContext, bs []byte) error {
	if cont.master != cc.From() {
		return errors.New("is not master")
	}

	rewardPolicy := &RewardPolicy{}
	if _, err := rewardPolicy.ReadFrom(bytes.NewReader(bs)); err != nil {
		return err
	}

	if bs, _, err := bin.WriterToBytes(rewardPolicy); err != nil {
		return err
	} else {
		cc.SetContractData([]byte{tagRewardPolicy}, bs)
	}
	return nil
}

func (cont *FormulatorContract) SetRewardPerBlock(cc *types.ContractContext, RewardPerBlock *amount.Amount) error {
	if cont.master != cc.From() {
		return errors.New("is not master")
	}

	rewardPolicy := &RewardPolicy{}
	if _, err := rewardPolicy.ReadFrom(bytes.NewReader(cc.ContractData([]byte{tagRewardPolicy}))); err != nil {
		return err
	}

	rewardPolicy.RewardPerBlock = RewardPerBlock

	if bs, _, err := bin.WriterToBytes(rewardPolicy); err != nil {
		return err
	} else {
		cc.SetContractData([]byte{tagRewardPolicy}, bs)
	}

	return nil
}

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *FormulatorContract) TokenAddress(cc *types.ContractContext) common.Address {
	return common.BytesToAddress(cc.ContractData([]byte{tagTokenContractAddress}))
}

func (cont *FormulatorContract) formulatorPolicy(cc *types.ContractContext) (*FormulatorPolicy, error) {
	policy := &FormulatorPolicy{}
	if _, err := policy.ReadFrom(bytes.NewReader(cc.ContractData([]byte{tagFormulatorPolicy}))); err != nil {
		return nil, err
	}
	return policy, nil
}

func (cont *FormulatorContract) addStakingAmount(cc *types.ContractContext, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) {
	if ns := cc.AccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress)); len(ns) == 0 {
		var Count uint32
		if bs := cc.AccountData(HyperAddress, []byte{tagStakingAmountCount}); len(bs) > 0 {
			Count = bin.Uint32(bs)
		}
		cc.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), bin.Uint32Bytes(Count))
		cc.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count), StakingAddress[:])
		Count++
		cc.SetAccountData(HyperAddress, []byte{tagStakingAmountCount}, bin.Uint32Bytes(Count))
	}
	cc.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), cont.StakingAmount(cc, HyperAddress, StakingAddress).Add(StakingAmount).Bytes())
}

func (cont *FormulatorContract) subStakingAmount(cc *types.ContractContext, HyperAddress common.Address, StakingAddress common.Address, am *amount.Amount) error {
	total := cont.StakingAmount(cc, HyperAddress, StakingAddress)
	if total.Less(am) {
		return errors.WithStack(ErrInvalidStakeAmount)
	}

	total = total.Sub(am)
	if total.IsZero() {
		if ns := cc.AccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress)); len(ns) > 0 {
			var Count uint32
			if bs := cc.AccountData(HyperAddress, []byte{tagStakingAmountCount}); len(bs) > 0 {
				Count = bin.Uint32(bs)
			}
			Number := bin.Uint32(ns)
			if Number != Count-1 {
				var swapAddr common.Address
				copy(swapAddr[:], cc.AccountData(HyperAddress, toStakingAmountReverseKey(Count-1)))
				cc.SetAccountData(HyperAddress, toStakingAmountReverseKey(Number), swapAddr[:])
				cc.SetAccountData(HyperAddress, toStakingAmountNumberKey(swapAddr), bin.Uint32Bytes(Number))
			}
			cc.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), nil)
			cc.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count-1), nil)
			Count--
			if Count == 0 {
				cc.SetAccountData(HyperAddress, []byte{tagStakingAmountCount}, nil)
			} else {
				cc.SetAccountData(HyperAddress, []byte{tagStakingAmountCount}, bin.Uint32Bytes(Count))
			}
		}
		cc.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), nil)
	} else {
		cc.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), total.Bytes())
	}
	return nil
}

func (cont *FormulatorContract) addFormulator(cc *types.ContractContext, fr *Formulator) error {
	if ns := cc.ContractData(toFormulatorNumberKey(fr.TokenID)); len(ns) == 0 {
		var Count uint32
		if bs := cc.ContractData([]byte{tagFormulatorCount}); len(bs) > 0 {
			Count = bin.Uint32(bs)
		}
		cc.SetContractData(toFormulatorNumberKey(fr.TokenID), bin.Uint32Bytes(Count))
		cc.SetContractData(toFormulatorReverseKey(Count), fr.TokenID[:])
		Count++
		cc.SetContractData([]byte{tagFormulatorCount}, bin.Uint32Bytes(Count))
	}
	cont.increaseFormulatorCount(cc, fr.Owner)
	cont.updateFormulator(cc, fr)

	return nil
}

func (cont *FormulatorContract) updateFormulator(cc *types.ContractContext, fr *Formulator) error {
	if bs, _, err := bin.WriterToBytes(fr); err != nil {
		return err
	} else {
		cc.SetAccountData(fr.TokenID, []byte{tagFormulator}, bs)
	}
	return nil
}

func (cont *FormulatorContract) removeFormulator(cc *types.ContractContext, fr *Formulator) error {
	if ns := cc.ContractData(toFormulatorNumberKey(fr.TokenID)); len(ns) > 0 {
		var Count uint32
		if bs := cc.ContractData([]byte{tagFormulatorCount}); len(bs) > 0 {
			Count = bin.Uint32(bs)
		}
		Number := bin.Uint32(ns)
		if Number != Count-1 {
			var swapAddr common.Address
			copy(swapAddr[:], cc.ContractData(toFormulatorReverseKey(Count-1)))
			cc.SetContractData(toFormulatorReverseKey(Number), swapAddr[:])
			cc.SetContractData(toFormulatorNumberKey(swapAddr), bin.Uint32Bytes(Number))
		}
		cc.SetContractData(toFormulatorNumberKey(fr.TokenID), nil)
		cc.SetContractData(toFormulatorReverseKey(Count-1), nil)
		Count--
		if Count == 0 {
			cc.SetContractData([]byte{tagFormulatorCount}, nil)
		} else {
			cc.SetContractData([]byte{tagFormulatorCount}, bin.Uint32Bytes(Count))
		}
	}

	cont.decreaseFormulatorCount(cc, fr.Owner)
	cc.SetAccountData(fr.TokenID, []byte{tagFormulator}, nil)
	return nil
}

func (cont *FormulatorContract) increaseFormulatorCount(cc *types.ContractContext, _owner common.Address) {
	var formulatorCount uint32
	if count := cc.AccountData(_owner, []byte{tagFormulatorCount}); len(count) > 0 {
		formulatorCount = bin.Uint32(count)
	}
	formulatorCount++
	cc.SetAccountData(_owner, []byte{tagFormulatorCount}, bin.Uint32Bytes(formulatorCount))
}

func (cont *FormulatorContract) decreaseFormulatorCount(cc *types.ContractContext, _owner common.Address) {
	var formulatorCount uint32
	if count := cc.AccountData(_owner, []byte{tagFormulatorCount}); len(count) > 0 {
		formulatorCount = bin.Uint32(count)
	}

	if formulatorCount > 0 {
		formulatorCount--
		cc.SetAccountData(_owner, []byte{tagFormulatorCount}, bin.Uint32Bytes(formulatorCount))
	} else {
		cc.SetAccountData(_owner, []byte{tagFormulatorCount}, bin.Uint32Bytes(uint32(0)))
	}
}

func formulatorTokenID(baseKey byte, fr *Formulator, cc *types.ContractContext) common.Address {
	base := make([]byte, 5+common.AddressLength+4)
	base[0] = baseKey
	copy(base[1:], bin.Uint32Bytes(fr.Height))
	copy(base[5:], fr.Owner[:])
	copy(base[5+common.AddressLength:], bin.Uint32Bytes(cc.NextSeq()))
	h := hash.Hash(base)
	return common.BytesToAddress(h[12:])
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *FormulatorContract) CreateGenesisAlpha(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	if cc.TargetHeight() != 0 {
		return common.Address{}, errors.WithStack(ErrNotGenesis)
	}

	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return common.Address{}, err
	}
	fr := &Formulator{
		Type:   AlphaFormulatorType,
		Height: cc.TargetHeight(),
		Amount: policy.AlphaAmount,
		Owner:  owner,
	}
	fr.TokenID = formulatorTokenID(0xff, fr, cc)
	if err := cont.addFormulator(cc, fr); err != nil {
		return common.Address{}, err
	}

	taddr := cont.TokenAddress(cc)
	_, ExecErr := cc.Exec(cc, taddr, "Mint", []interface{}{cont.addr, policy.AlphaAmount})
	if ExecErr != nil {
		return common.Address{}, ExecErr
	}

	return fr.TokenID, nil
}

func (cont *FormulatorContract) CreateGenesisSigma(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	if cc.TargetHeight() != 0 {
		return common.Address{}, errors.WithStack(ErrNotGenesis)
	}

	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return common.Address{}, err
	}
	fr := &Formulator{
		Type:   SigmaFormulatorType,
		Height: cc.TargetHeight(),
		Amount: policy.AlphaAmount.MulC(int64(policy.SigmaCount)),
		Owner:  owner,
	}
	fr.TokenID = formulatorTokenID(0xff, fr, cc)
	if err := cont.addFormulator(cc, fr); err != nil {
		return common.Address{}, err
	}

	taddr := cont.TokenAddress(cc)
	_, ExecErr := cc.Exec(cc, taddr, "Mint", []interface{}{cont.addr, policy.AlphaAmount.MulC(int64(policy.SigmaCount))})
	if ExecErr != nil {
		return common.Address{}, ExecErr
	}

	return fr.TokenID, nil
}

func (cont *FormulatorContract) CreateGenesisOmega(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	if cc.TargetHeight() != 0 {
		return common.Address{}, errors.WithStack(ErrNotGenesis)
	}

	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return common.Address{}, err
	}
	fr := &Formulator{
		Type:   OmegaFormulatorType,
		Height: cc.TargetHeight(),
		Amount: policy.AlphaAmount.MulC(int64(policy.SigmaCount)).MulC(int64(policy.OmegaCount)),
		Owner:  owner,
	}
	fr.TokenID = formulatorTokenID(0xff, fr, cc)
	if err := cont.addFormulator(cc, fr); err != nil {
		return common.Address{}, err
	}

	taddr := cont.TokenAddress(cc)
	_, ExecErr := cc.Exec(cc, taddr, "Mint", []interface{}{cont.addr, policy.AlphaAmount.MulC(int64(policy.SigmaCount)).MulC(int64(policy.OmegaCount))})
	if ExecErr != nil {
		return common.Address{}, ExecErr
	}

	return fr.TokenID, nil
}

func (cont *FormulatorContract) AddGenesisStakingAmount(cc *types.ContractContext, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) error {
	if cc.TargetHeight() != 0 {
		return errors.WithStack(ErrNotGenesis)
	}

	if ns := cc.AccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress)); len(ns) == 0 {
		var Count uint32
		if bs := cc.AccountData(HyperAddress, []byte{tagStakingAmountCount}); len(bs) > 0 {
			Count = bin.Uint32(bs)
		}
		cc.SetAccountData(HyperAddress, toStakingAmountNumberKey(StakingAddress), bin.Uint32Bytes(Count))
		cc.SetAccountData(HyperAddress, toStakingAmountReverseKey(Count), StakingAddress[:])
		Count++
		cc.SetAccountData(HyperAddress, []byte{tagStakingAmountCount}, bin.Uint32Bytes(Count))
	}
	cc.SetAccountData(HyperAddress, toStakingAmountKey(StakingAddress), cont.StakingAmount(cc, HyperAddress, StakingAddress).Add(StakingAmount).Bytes())

	taddr := cont.TokenAddress(cc)
	_, ExecErr := cc.Exec(cc, taddr, "Mint", []interface{}{cont.addr, StakingAmount})
	if ExecErr != nil {
		return ExecErr
	}

	return nil
}

func (cont *FormulatorContract) CreateAlpha(cc *types.ContractContext) (common.Address, error) {
	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return common.Address{}, err
	}

	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "TransferFrom", []interface{}{cc.From(), cont.Address(), policy.AlphaAmount}); err != nil {
		return common.Address{}, err
	}

	fr := &Formulator{
		Type:   AlphaFormulatorType,
		Height: cc.TargetHeight(),
		Amount: policy.AlphaAmount,
		Owner:  cc.From(),
	}
	fr.TokenID = formulatorTokenID(0xff, fr, cc)
	if err := cont.addFormulator(cc, fr); err != nil {
		return common.Address{}, err
	}

	return fr.TokenID, nil
}

func (cont *FormulatorContract) CreateAlphaBatch(cc *types.ContractContext, count *big.Int) ([]common.Address, error) {
	one := big.NewInt(1)
	flist := []common.Address{}
	for count.Cmp(amount.ZeroCoin.Int) > 0 {
		count.Sub(count, one)
		if addr, err := cont.CreateAlpha(cc); err != nil {
			return nil, err
		} else {
			flist = append(flist, addr)
		}
	}
	return flist, nil
}

func (cont *FormulatorContract) CreateSigma(cc *types.ContractContext, TokenIDs []common.Address) error {
	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return err
	}
	if len(TokenIDs) != int(policy.SigmaCount) {
		return errors.WithStack(ErrInvalidSigmaCreationCount)
	}
	uniqueCheck := map[common.Address]interface{}{}
	frs := []*Formulator{}
	sum := amount.NewAmount(0, 0)
	for _, addr := range TokenIDs {
		if _, ok := uniqueCheck[addr]; ok {
			return errors.New("duplicate formulator")
		}
		uniqueCheck[addr] = addr
		fr, err := cont._formulator(cc, addr)
		if err != nil {
			return err
		}
		if cc.From() != fr.Owner {
			return errors.WithStack(ErrNotFormulatorOwner)
		}
		if fr.Type != AlphaFormulatorType {
			return errors.WithStack(ErrInvalidSigmaFormulatorType)
		}
		if cc.TargetHeight()-fr.Height < policy.SigmaBlocks {
			return errors.WithStack(ErrInvalidSigmaCreationBlocks)
		}
		frs = append(frs, fr)

		sum = sum.Add(fr.Amount)
	}
	for _, fr := range frs[1:] {
		if err := cont.removeFormulator(cc, fr); err != nil {
			return err
		}
	}

	fr := frs[0]
	fr.Type = SigmaFormulatorType
	fr.Height = cc.TargetHeight()
	fr.Amount = sum
	cont._Approve(cc, fr.TokenID, common.Address{})
	return cont.updateFormulator(cc, fr)
}

func (cont *FormulatorContract) CreateOmega(cc *types.ContractContext, TokenIDs []common.Address) error {
	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return err
	}
	if len(TokenIDs) != int(policy.OmegaCount) {
		return errors.WithStack(ErrInvalidOmegaCreationCount)
	}
	uniqueCheck := map[common.Address]interface{}{}
	frs := []*Formulator{}
	sum := amount.NewAmount(0, 0)
	for _, addr := range TokenIDs {
		if _, ok := uniqueCheck[addr]; ok {
			return errors.New("duplicate formulator")
		}
		uniqueCheck[addr] = addr
		fr, err := cont._formulator(cc, addr)
		if err != nil {
			return err
		}
		if cc.From() != fr.Owner {
			return errors.WithStack(ErrNotFormulatorOwner)
		}
		if fr.Type != SigmaFormulatorType {
			return errors.WithStack(ErrInvalidOmegaFormulatorType)
		}
		if cc.TargetHeight()-fr.Height < policy.OmegaBlocks {
			return errors.WithStack(ErrInvalidOmegaCreationBlocks)
		}
		frs = append(frs, fr)

		sum = sum.Add(fr.Amount)
	}
	for _, fr := range frs[1:] {
		if err := cont.removeFormulator(cc, fr); err != nil {
			return err
		}
	}

	fr := frs[0]
	fr.Type, fr.Height, fr.Amount = OmegaFormulatorType, cc.TargetHeight(), sum
	cont._Approve(cc, fr.TokenID, common.Address{})
	return cont.updateFormulator(cc, fr)
}

func (cont *FormulatorContract) Revoke(cc *types.ContractContext, TokenID common.Address) error {
	fr, err := cont._formulator(cc, TokenID)
	if err != nil {
		return err
	}
	if cc.From() != fr.Owner {
		return errors.WithStack(ErrNotFormulatorOwner)
	}
	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "Transfer", []interface{}{fr.Owner, fr.Amount}); err != nil {
		return err
	}
	if err := cont.removeFormulator(cc, fr); err != nil {
		return err
	}
	return nil
}

func (cont *FormulatorContract) RevokeBatch(cc *types.ContractContext, TokenIDs []common.Address) ([]common.Address, error) {
	flist := []common.Address{}
	for _, tokenID := range TokenIDs {
		if err := cont.Revoke(cc, tokenID); err != nil {
			return nil, err
		} else {
			flist = append(flist, tokenID)
		}
	}

	return flist, nil
}

func (cont *FormulatorContract) Stake(cc *types.ContractContext, HyperAddress common.Address, Amount *amount.Amount) error {
	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return err
	}
	if Amount.Less(policy.MinStakeAmount) {
		return errors.WithStack(ErrInvalidStakeAmount)
	}
	if !cc.IsGenerator(HyperAddress) {
		return errors.WithStack(ErrInvalidStakeGenerator)
	}

	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "TransferFrom", []interface{}{cc.From(), cont.Address(), Amount}); err != nil {
		return err
	}
	cont.addStakingAmount(cc, HyperAddress, cc.From(), Amount)
	return nil
}

func (cont *FormulatorContract) Unstake(cc *types.ContractContext, HyperAddress common.Address, Amount *amount.Amount) error {
	policy, err := cont.formulatorPolicy(cc)
	if err != nil {
		return err
	}
	if Amount.Less(policy.MinStakeAmount) {
		return errors.WithStack(ErrInvalidStakeAmount)
	}
	if !cc.IsGenerator(HyperAddress) {
		return errors.WithStack(ErrInvalidStakeGenerator)
	}
	if err := cont.subStakingAmount(cc, HyperAddress, cc.From(), Amount); err != nil {
		return err
	}
	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "Transfer", []interface{}{cc.From(), Amount}); err != nil {
		return err
	}
	return nil
}

func (cont *FormulatorContract) Approve(cc *types.ContractContext, To common.Address, TokenID common.Address) error {
	formulator, err := cont._formulator(cc, TokenID)
	if err != nil {
		return err
	}
	if formulator.Owner == To {
		return ErrApprovalToCurrentOwner
	}
	if formulator.Owner != cc.From() {
		return ErrNotFormulatorOwner
	}

	cont._Approve(cc, TokenID, To)
	return nil
}

func (cont *FormulatorContract) _Approve(cc *types.ContractContext, TokenID common.Address, To common.Address) {
	cc.SetAccountData(common.Address{}, makeApproveKey(TokenID), To[:])
}

func (cont *FormulatorContract) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, TokenID common.Address) error {
	formulator, err := cont._formulator(cc, TokenID)
	if err != nil {
		return err
	}

	Approved := cont.GetApproved(cc, TokenID)
	owner, err := cont.OwnerOf(cc, TokenID)
	if err != nil {
		return err
	}
	isApproved := cont.IsApprovedForAll(cc, owner, cc.From())
	if cc.From() != Approved && !isApproved && owner != cc.From() {
		return errors.WithStack(ErrNotFormulatorApproved)
	}
	if (To == common.Address{}) {
		return errors.WithStack(ErrTransferToZeroAddress)
	}

	cont._transferFrom(cc, formulator, From, To, TokenID)
	return nil
}

func (cont *FormulatorContract) _transferFrom(cc *types.ContractContext, formulator *Formulator, From common.Address, To common.Address, TokenID common.Address) {
	formulator.Owner = To
	cont.updateFormulator(cc, formulator)

	cont._Approve(cc, TokenID, common.Address{})
	cont.increaseFormulatorCount(cc, To)
	cont.decreaseFormulatorCount(cc, From)
}

func (cont *FormulatorContract) RegisterSales(cc *types.ContractContext, TokenID common.Address, Amount *amount.Amount) error {
	formulator, err := cont._formulator(cc, TokenID)
	if err != nil {
		return err
	}

	bs := cc.AccountData(formulator.Owner, makeSaleAmountKey(TokenID))
	if len(bs) > 0 {
		return ErrAlreadyRegisteredSalesFormulator
	}

	approvedTo := cont.GetApproved(cc, TokenID)
	if approvedTo == cont.Address() {
		return ErrApprovalToCurrentOwner
	}
	if formulator.Owner != cc.From() {
		return ErrNotFormulatorOwner
	}

	cont._Approve(cc, TokenID, cont.Address())
	cc.SetAccountData(formulator.Owner, makeSaleAmountKey(TokenID), Amount.Bytes())

	return nil
}

func (cont *FormulatorContract) CancelSales(cc *types.ContractContext, TokenID common.Address) error {
	formulator, err := cont._formulator(cc, TokenID)
	if err != nil {
		return err
	}

	bs := cc.AccountData(formulator.Owner, makeSaleAmountKey(TokenID))
	if len(bs) == 0 {
		return ErrNotRegisteredSalesFormulator
	}

	approvedTo := cont.GetApproved(cc, TokenID)
	if (approvedTo == common.Address{}) {
		return ErrApprovalToCurrentOwner
	}

	if formulator.Owner != cc.From() {
		return ErrNotFormulatorOwner
	}

	cont._Approve(cc, TokenID, common.Address{})
	cc.SetAccountData(formulator.Owner, makeSaleAmountKey(TokenID), nil)

	return nil
}

func (cont *FormulatorContract) BuyFormulator(cc *types.ContractContext, TokenID common.Address) error {
	// formulator transfer
	formulator, err := cont._formulator(cc, TokenID)
	if err != nil {
		return err
	}

	// token transfer
	bs := cc.AccountData(formulator.Owner, makeSaleAmountKey(TokenID))
	if len(bs) == 0 {
		return errors.New("not registerd sales")
	}
	salePrice := amount.NewAmountFromBytes(bs)

	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "TransferFrom", []interface{}{cc.From(), formulator.Owner, salePrice}); err != nil {
		return err
	}

	cont._transferFrom(cc, formulator, formulator.Owner, cc.From(), TokenID)
	cc.SetAccountData(formulator.Owner, makeSaleAmountKey(TokenID), nil)

	return nil
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////

func (cont *FormulatorContract) SetURI(cc *types.ContractContext, uri string) error {
	if cont.master != cc.From() {
		return errors.New("is not master")
	}
	cc.SetContractData([]byte{tagUri}, []byte(uri))
	return nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *FormulatorContract) _formulator(cc types.ContractLoader, _tokenID common.Address) (*Formulator, error) {
	fr := &Formulator{}
	bs := cc.AccountData(_tokenID, []byte{tagFormulator})
	if len(bs) == 0 {
		return nil, errors.WithStack(ErrNotExistFormulator)
	}
	if _, err := fr.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, err
	}
	return fr, nil
}

func (cont *FormulatorContract) StakingAmount(cc types.ContractLoader, HyperAddress common.Address, StakingAddress common.Address) *amount.Amount {
	if bs := cc.AccountData(HyperAddress, toStakingAmountKey(StakingAddress)); len(bs) > 0 {
		return amount.NewAmountFromBytes(bs)
	} else {
		return amount.NewAmount(0, 0)
	}
}

func (cont *FormulatorContract) StakingAmountMap(cc types.ContractLoader, HyperAddress common.Address) (map[common.Address]*amount.Amount, error) {
	PowerMap := map[common.Address]*amount.Amount{}
	if bs := cc.AccountData(HyperAddress, []byte{tagStakingAmountCount}); len(bs) > 0 {
		Count := bin.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var StakingAddress common.Address
			copy(StakingAddress[:], cc.AccountData(HyperAddress, toStakingAmountReverseKey(i)))
			PowerMap[StakingAddress] = cont.StakingAmount(cc, HyperAddress, StakingAddress)
		}
	}
	return PowerMap, nil
}

func (cont *FormulatorContract) FormulatorMap(cc types.ContractLoader) (map[common.Address]*Formulator, error) {
	FormulatorMap := map[common.Address]*Formulator{}
	if bs := cc.ContractData([]byte{tagFormulatorCount}); len(bs) > 0 {
		Count := bin.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], cc.ContractData(toFormulatorReverseKey(i)))
			fr, err := cont._formulator(cc, addr)
			if err != nil {
				return nil, err
			}
			FormulatorMap[addr] = fr
		}
	}
	return FormulatorMap, nil
}

func (cont *FormulatorContract) BalanceOf(cc types.ContractLoader, _owner common.Address) uint32 {
	formulatorCount := cc.AccountData(_owner, []byte{tagFormulatorCount})
	if len(formulatorCount) == 0 {
		return 0
	}
	return bin.Uint32(formulatorCount)
}

func (cont *FormulatorContract) TotalSupply(cc types.ContractLoader) uint32 {
	var Count uint32
	if bs := cc.ContractData([]byte{tagFormulatorCount}); len(bs) > 0 {
		Count = bin.Uint32(bs)
	}
	return Count
}

func (cont *FormulatorContract) OwnerOf(cc types.ContractLoader, _tokenID common.Address) (common.Address, error) {
	formulator, err := cont._formulator(cc, _tokenID)
	if err != nil {
		return common.Address{}, err
	}
	return formulator.Owner, nil
}

func (cont *FormulatorContract) GetApproved(cc types.ContractLoader, TokenID common.Address) common.Address {
	bs := cc.AccountData(common.Address{}, makeApproveKey(TokenID))
	var approvedTo common.Address
	copy(approvedTo[:], bs)
	return approvedTo
}

func (cont *FormulatorContract) tokenURI(cc types.ContractLoader, _id *big.Int) string {
	uri := cont.URI(cc)
	strID := hex.EncodeToString(_id.Bytes())
	idstr := fmt.Sprintf("%064v", strID)

	return strings.Replace(uri, "{id}", idstr, -1)
}

func (cont *FormulatorContract) URI(cc types.ContractLoader) string {
	bs := cc.ContractData([]byte{tagUri})
	return string(bs)
}

func (cont *FormulatorContract) SupportsInterface(cc types.ContractLoader, interfaceID []byte) bool {
	// 01ffc9a7 eip-165
	// d9b67a26 Multi Token Standard
	// 80ac58cd NFT
	// 0e89341c ERC1155Metadata_URI
	switch hex.EncodeToString(interfaceID) {
	case "80ac58cd", "d9b67a26", "01ffc9a7", "0e89341c":
		return true
	}
	return false
}

func (cont *FormulatorContract) TokenByIndex(cc types.ContractLoader, _id uint32) (*big.Int, error) {
	var addr common.Address
	copy(addr[:], cc.ContractData(toFormulatorReverseKey(_id)))
	fr, err := cont._formulator(cc, addr)
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetBytes(fr.TokenID[:]), nil
}

func (cont *FormulatorContract) TokenByRange(cc types.ContractLoader, from, to uint32) ([]*big.Int, error) {
	ts := cont.TotalSupply(cc)
	if from >= ts {
		from = ts - 1
	}
	if to >= ts {
		to = ts - 1
	}
	if from > to {
		return nil, errors.New("from less than to")
	}

	bis := []*big.Int{}
	for _id := from; _id <= to; _id++ {
		var addr common.Address
		copy(addr[:], cc.ContractData(toFormulatorReverseKey(_id)))
		fr, err := cont._formulator(cc, addr)
		if err != nil {
			return nil, err
		}
		bis = append(bis, big.NewInt(0).SetBytes(fr.TokenID[:]))
	}
	return bis, nil
}

func (cont *FormulatorContract) TokenOfOwnerByIndex(cc types.ContractLoader, _owner common.Address, _index uint32) (*big.Int, error) {
	var userCount uint32 = 0
	if bs := cc.ContractData([]byte{tagFormulatorCount}); len(bs) > 0 {
		Count := bin.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], cc.ContractData(toFormulatorReverseKey(i)))
			fr, err := cont._formulator(cc, addr)
			if err != nil {
				return nil, err
			}
			if fr.Owner == _owner {
				if _index == userCount {
					return big.NewInt(0).SetBytes(fr.TokenID[:]), nil
				}
				userCount++
			}
		}
	}
	return nil, errors.New("not exist")
}

func (cont *FormulatorContract) TokenOfOwnerByRange(cc types.ContractLoader, _owner common.Address, from, to uint32) ([]*big.Int, error) {
	ts := cont.BalanceOf(cc, _owner)
	if from > to {
		return nil, errors.New("from less than to")
	}
	if from >= ts {
		from = ts
	}
	if to >= ts {
		to = ts
	}
	if from > to {
		return nil, errors.New("from less than to")
	}

	bis := []*big.Int{}
	for _id := from; _id <= to; _id++ {
	}

	var userCount uint32 = 0
	if bs := cc.ContractData([]byte{tagFormulatorCount}); len(bs) > 0 {
		Count := bin.Uint32(bs)
		for i := uint32(0); i < Count; i++ {
			var addr common.Address
			copy(addr[:], cc.ContractData(toFormulatorReverseKey(i)))
			fr, err := cont._formulator(cc, addr)
			if err != nil {
				return nil, err
			}
			if fr.Owner == _owner {
				if from <= userCount {
					bid := big.NewInt(0).SetBytes(fr.TokenID[:])
					bis = append(bis, bid)
				}
				userCount++
				if to <= userCount {
					break
				}
				// if _index == userCount {
				// }
			}
		}
	}
	return bis, nil
}

func (cont *FormulatorContract) SetApprovalForAll(cc *types.ContractContext, _operator common.Address, _approved bool) {
	if _approved {
		cc.SetAccountData(cc.From(), makeApproveAllKey(_operator), []byte{1})
	} else {
		cc.SetAccountData(cc.From(), makeApproveAllKey(_operator), nil)
	}
}
func (cont *FormulatorContract) IsApprovedForAll(cc types.ContractLoader, _owner common.Address, _operator common.Address) bool {
	bs := cc.AccountData(_owner, makeApproveAllKey(_operator))
	if len(bs) == 0 || bs[0] != 1 {
		return false
	}
	return true
}
