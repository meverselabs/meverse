// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import './helpers/Context.sol';
import './helpers/Ownable.sol';
import './helpers/Address.sol';
import './helpers/ReentrancyGuard.sol';

/// @title MevSample for mev transfer sample to and from the comtract
/// @notice
contract MevTransferSample is Ownable, ReentrancyGuard {
  uint256 public data;

  constructor() {}

  function setData(uint256 _data) external payable {
    data = _data;
  }

  function transfer(address payable to) external onlyOwner nonReentrant {
    uint256 amount = address(this).balance;
    Address.sendValue(to, amount);
  }

  function getBalance() external view returns (uint256) {
    return address(this).balance;
  }
}
