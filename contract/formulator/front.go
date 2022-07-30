package formulator

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *FormulatorContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *FormulatorContract
}

func (f *front) CreateGenesisAlpha(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	return f.cont.CreateGenesisAlpha(cc, owner)
}

func (f *front) CreateGenesisSigma(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	return f.cont.CreateGenesisSigma(cc, owner)
}

func (f *front) CreateGenesisOmega(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	return f.cont.CreateGenesisOmega(cc, owner)
}

func (f *front) AddGenesisStakingAmount(cc *types.ContractContext, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) error {
	return f.cont.AddGenesisStakingAmount(cc, HyperAddress, StakingAddress, StakingAmount)
}

func (f *front) CreateAlpha(cc *types.ContractContext) (common.Address, error) {
	return f.cont.CreateAlpha(cc)
}

func (f *front) CreateAlphaBatch(cc *types.ContractContext, count *big.Int) ([]common.Address, error) {
	return f.cont.CreateAlphaBatch(cc, count)
}

func (f *front) CreateSigma(cc *types.ContractContext, TokenIDs []common.Address) error {
	return f.cont.CreateSigma(cc, TokenIDs)
}

func (f *front) CreateOmega(cc *types.ContractContext, TokenIDs []common.Address) error {
	return f.cont.CreateOmega(cc, TokenIDs)
}

func (f *front) Revoke(cc *types.ContractContext, TokenID common.Address) error {
	return f.cont.Revoke(cc, TokenID)
}

func (f *front) RevokeBatch(cc *types.ContractContext, TokenIDs []common.Address) ([]common.Address, error) {
	return f.cont.RevokeBatch(cc, TokenIDs)
}

func (f *front) Stake(cc *types.ContractContext, HyperAddress common.Address, Amount *amount.Amount) error {
	return f.cont.Stake(cc, HyperAddress, Amount)
}

func (f *front) Unstake(cc *types.ContractContext, HyperAddress common.Address, Amount *amount.Amount) error {
	return f.cont.Unstake(cc, HyperAddress, Amount)
}

func (f *front) Approve(cc *types.ContractContext, To common.Address, TokenID common.Address) error {
	return f.cont.Approve(cc, To, TokenID)
}

func (f *front) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, TokenID common.Address) error {
	return f.cont.TransferFrom(cc, From, To, TokenID)
}

func (f *front) RegisterSales(cc *types.ContractContext, TokenID common.Address, Amount *amount.Amount) error {
	return f.cont.RegisterSales(cc, TokenID, Amount)
}

func (f *front) CancelSales(cc *types.ContractContext, TokenID common.Address) error {
	return f.cont.CancelSales(cc, TokenID)
}

func (f *front) BuyFormulator(cc *types.ContractContext, TokenID common.Address) error {
	return f.cont.BuyFormulator(cc, TokenID)
}

func (f *front) SetURI(cc *types.ContractContext, uri string) error {
	return f.cont.SetURI(cc, uri)
}

func (f *front) SetRewardPolicy(cc *types.ContractContext, bs []byte) error {
	return f.cont.SetRewardPolicy(cc, bs)
}

func (f *front) SetRewardPerBlock(cc *types.ContractContext, RewardPerBlock *amount.Amount) error {
	return f.cont.SetRewardPerBlock(cc, RewardPerBlock)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) Formulator(cc types.ContractLoader, _tokenID common.Address) (uint8, uint32, *amount.Amount, common.Address, common.Address, error) {
	formulator, err := f.cont._formulator(cc, _tokenID)
	if err != nil {
		return 0, 0, nil, common.Address{}, common.Address{}, err
	}
	return formulator.Type, formulator.Height, formulator.Amount, formulator.Owner, formulator.TokenID, nil
}

/// @notice Enumerate valid NFTs
/// @dev Throws if `_index` >= `totalSupply()`.
/// @param _index A counter less than `totalSupply()`
/// @return The token identifier for the `_index`th NFT,
///  (sort order not specified)
func (f *front) TokenByIndex(cc types.ContractLoader, _id uint32) (*big.Int, error) {
	return f.cont.TokenByIndex(cc, _id)
}
func (f *front) TokenByRange(cc types.ContractLoader, from, to uint32) ([]*big.Int, error) {
	return f.cont.TokenByRange(cc, from, to)
}

func (f *front) StakingAmount(cc types.ContractLoader, HyperAddress common.Address, StakingAddress common.Address) *amount.Amount {
	return f.cont.StakingAmount(cc, HyperAddress, StakingAddress)
}

func (f *front) StakingAmountMap(cc types.ContractLoader, HyperAddress common.Address, addr common.Address) *amount.Amount {
	return f.StakingAmount(cc, HyperAddress, addr)
}

func (f *front) FormulatorMap(cc types.ContractLoader) (map[common.Address]*Formulator, error) {
	return f.cont.FormulatorMap(cc)
}

func (f *front) BalanceOf(cc types.ContractLoader, _owner common.Address) uint32 {
	return f.cont.BalanceOf(cc, _owner)
}

func (f *front) OwnerOf(cc types.ContractLoader, _tokenID common.Address) (common.Address, error) {
	return f.cont.OwnerOf(cc, _tokenID)
}

func (f *front) GetApproved(cc types.ContractLoader, TokenID common.Address) common.Address {
	return f.cont.GetApproved(cc, TokenID)
}

func (f *front) TotalSupply(cc types.ContractLoader) uint32 {
	return f.cont.TotalSupply(cc)
}

func (f *front) Name(cc types.ContractLoader) string {
	return f.cont.Name()
}

func (f *front) Decimals(cc types.ContractLoader) *big.Int {
	return big.NewInt(0)
}

func (f *front) Uri(cc types.ContractLoader, _id *big.Int) string {
	return f.cont.tokenURI(cc, _id)
}

/// @notice A distinct Uniform Resource Identifier (URI) for a given asset.
/// @dev Throws if `_tokenId` is not a valid NFT. URIs are defined in RFC
///  3986. The URI may point to a JSON file that conforms to the "ERC721
///  Metadata JSON Schema".
func (f *front) TokenURI(cc types.ContractLoader, _id *big.Int) string {
	return f.cont.tokenURI(cc, _id)
}

func (f *front) BaseURI(cc types.ContractLoader) string {
	return f.cont.URI(cc)
}

func (f *front) SupportsInterface(cc types.ContractLoader, interfaceID []byte) bool {
	return f.cont.SupportsInterface(cc, interfaceID)
}

/// @notice Enumerate NFTs assigned to an owner
/// @dev Throws if `_index` >= `balanceOf(_owner)` or if
///  `_owner` is the zero address, representing invalid NFTs.
/// @param _owner An address where we are interested in NFTs owned by them
/// @param _index A counter less than `balanceOf(_owner)`
/// @return The token identifier for the `_index`th NFT assigned to `_owner`,
///   (sort order not specified)
func (f *front) TokenOfOwnerByIndex(cc types.ContractLoader, _owner common.Address, _index uint32) (*big.Int, error) {
	return f.cont.TokenOfOwnerByIndex(cc, _owner, _index)
}
func (f *front) TokenOfOwnerByRange(cc types.ContractLoader, _owner common.Address, from, to uint32) ([]*big.Int, error) {
	return f.cont.TokenOfOwnerByRange(cc, _owner, from, to)
}

func (f *front) SetApprovalForAll(cc *types.ContractContext, _operator common.Address, _approved bool) {
	f.cont.SetApprovalForAll(cc, _operator, _approved)
}
func (f *front) IsApprovedForAll(cc types.ContractLoader, _owner common.Address, _operator common.Address) bool {
	return f.cont.IsApprovedForAll(cc, _owner, _operator)
}
