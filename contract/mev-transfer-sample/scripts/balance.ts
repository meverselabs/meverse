import { ethers, network } from "hardhat";
import { getMevAddress, getMevTransferSampleAddress } from "../test/util";

async function main() {
  const mevAddress = getMevAddress();
  const mev = await ethers.getContractAt("IERC20", mevAddress);

  const mevTransferSampleAddress = getMevTransferSampleAddress();
  const mts = await ethers.getContractAt(
    "MevTransferSample",
    mevTransferSampleAddress
  );

  const mevBalance = await mev.balanceOf(mts.address);
  console.log(`mev.balanceOf(MevTransferSamples) = ${mevBalance}`);

  const balance = await mts.getBalance();
  console.log(`mts.getBalance() = ${balance}`);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
