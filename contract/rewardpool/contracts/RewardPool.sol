// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import './helpers/Context.sol';
import './helpers/Ownable.sol';
import './helpers/ReentrancyGuard.sol';

import './interfaces/UserReward.sol';
import './interfaces/IRewardPool.sol';
import './interfaces/IMRC20.sol';

/// @title Reward pool for Pocket Battles NFT with token
/// @notice implements the IRewardPool interface
contract RewardPool is IRewardPool, Ownable, ReentrancyGuard {
  address public immutable override token;

  uint256 public override totalReward;

  mapping(address => uint256) public override rewards;

  bool public override isClaimable;

  constructor(
    address _token //IMRC20 PKT token
  ) {
    token = _token;
    isClaimable = true;
  }

  fallback() external payable {
    revert();
  }

  receive() external payable {
    revert();
  }

  function addReward(uint256 _total, UserReward[] calldata userRewards) external override onlyOwner {
    uint256 total;
    for (uint256 i = 0; i < userRewards.length; i++) {
      total += userRewards[i].amount;
    }
    require(total == _total, 'NFTFarm: input total is different from sum');

    IMRC20(token).transferFrom(_msgSender(), address(this), total);

    for (uint256 i = 0; i < userRewards.length; i++) {
      uint256 amount = userRewards[i].amount;
      if (amount == 0) continue;
      rewards[userRewards[i].user] += amount;
    }
    totalReward += total;

    emit RewardAdded(total);
  }

  function claim() external override nonReentrant {
    require(isClaimable, 'RewardPool: pool is unclaimable');
    address _msgSender = _msgSender();
    uint256 amount = rewards[_msgSender];
    require(amount > 0, 'RewardPool: claim amount is zero');
    require(totalReward >= amount, 'RewardPool: totalReward is lower');

    IMRC20(token).transfer(_msgSender, amount);
    unchecked {
      totalReward -= amount;
    }
    delete rewards[_msgSender];

    emit Claim(_msgSender, amount);
  }

  function revokeTotalReward() external override onlyOwner nonReentrant {
    setClaimable(false);

    address owner = owner();
    IMRC20(token).transfer(owner, totalReward);

    emit Revoke(owner, totalReward);
  }

  function setClaimable(bool _isClaimable) public override onlyOwner {
    require(isClaimable != _isClaimable, 'RewardPool: isClaimable is already set');
    isClaimable = _isClaimable;

    emit Claimable(isClaimable);
  }
}
