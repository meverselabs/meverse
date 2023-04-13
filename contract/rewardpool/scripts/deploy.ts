import { ethers, network } from "hardhat";

async function main() {
  let pktAddress: string;

  switch (network.name) {
    case "mevlocal":
      pktAddress = "0xE08FBAd440dfF3267f5A42061D64FC3b953C1Ef1";
      break;
    case "mevtesnet":
      pktAddress = "";
    case "mevmainnet":
      pktAddress = "";
      break;
    default:
      console.log("pkt address of the network is not specified");
      return;
  }

  const rewardPoolFactory = await ethers.getContractFactory("RewardPool");
  const pool = await rewardPoolFactory.deploy(pktAddress);

  await pool.deployed();

  console.log(`RewardPool with PKT ${pktAddress} deployed to ${pool.address}`);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
