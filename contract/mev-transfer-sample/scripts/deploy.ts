import { ethers, network } from "hardhat";

async function main() {
  const mevTransferSampleFactory = await ethers.getContractFactory(
    "MevTransferSample"
  );
  const mts = await mevTransferSampleFactory.deploy();

  await mts.deployed();

  console.log(
    `mevTransferSample deployed to ${mts.address} int network ${network.name}`
  );
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
