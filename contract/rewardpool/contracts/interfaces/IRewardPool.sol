// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity >=0.5.0;

import './UserReward.sol';

/// @title The Interface for the NftFarm for NftPocketBattles
/// @notice The NftFarm facilitates the farming of Nft staking
interface IRewardPool {
  /// @notice The ERC20 contract for rewarding
  /// @return The contract address
  function token() external view returns (address);

  /// @notice The total reward of this farm
  /// @return The amount of total reward
  function totalReward() external view returns (uint256);

  /// @notice The user's credit amount
  /// @return The pkt amount of the user
  function rewards(address user) external view returns (uint256);

  /// @notice whether claiming is possible or not
  /// @return true when claiming is possible
  function isClaimable() external view returns (bool);

  /// @notice Emitted when the owner adds the reward
  /// @param total The total amount of added rewards
  event RewardAdded(uint256 indexed total);

  /// @notice Emitted when a user claims his/her reward
  /// @param user The address of the user who claims
  /// @param amount The PKT amount
  event Claim(address indexed user, uint256 indexed amount);

  /// @notice Emitted when the owner of this contract claims the total reward
  /// @param owner The address of the user who claims
  /// @param totalReward The total reward amount
  event Revoke(address indexed owner, uint256 indexed totalReward);

  /// @notice Emitted when the owner sets whether the claiming is possible or not
  /// @param isClaimable The true means that the claming is possible
  event Claimable(bool indexed isClaimable);

  /// @notice Add rewards for multi users
  /// @dev emit RewardAdded event
  /// @param _total The total amount of rewards
  /// @param userRewards The array of UserReward struct which has the address and the reward amount
  function addReward(uint256 _total, UserReward[] calldata userRewards) external;

  /// @notice Claim his reward by owner
  /// @dev emit Claim event
  function claim() external;

  /// @notice Revoke total reward to the owner and makes the contract unClaimable
  /// @dev emit Revoke event
  function revokeTotalReward() external;

  /// @notice Makes the contract claimable or unClaimable by the owner
  /// @dev emit Claimable event
  function setClaimable(bool _isClaimable) external;
}
