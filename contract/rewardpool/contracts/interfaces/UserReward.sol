// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity >=0.5.0;

/// @title UserReward struct
/// @notice user(address), reward Struct
struct UserReward {
  // address
  address user;
  // amount
  uint256 amount;
}
