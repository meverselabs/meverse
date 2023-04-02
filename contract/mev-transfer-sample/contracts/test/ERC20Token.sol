// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import '@openzeppelin/contracts/token/ERC20/ERC20.sol';

contract ERC20Token is ERC20 {
  constructor(uint256 totalSupply) ERC20('Test Token', 'MEV') {
    _mint(msg.sender, totalSupply);
  }
}
