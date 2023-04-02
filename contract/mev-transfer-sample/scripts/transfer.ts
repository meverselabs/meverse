import { ethers } from "hardhat";
import { utils } from "ethers";
import { getMevTransferSampleAddress } from "../test/util";

async function main() {
  const mevTransferSampleAddress = getMevTransferSampleAddress();
  const mts = await ethers.getContractAt(
    "MevTransferSample",
    mevTransferSampleAddress
  );

  let recipient = "0x04EfbDABCBF3F68AE80899f7798E63Bc3E02E8b7";
  const tx = await mts.transfer(recipient);
  const receipt = await tx.wait();
  console.log("receipt", receipt);
}
// receipt {
//   to: '0xdA2Fc6c5D30e0eDf453bD201Ec4a941D1fa72392',
//   from: '0x04EfbDABCBF3F68AE80899f7798E63Bc3E02E8b7',
//   contractAddress: null,
//   transactionIndex: 0,
//   gasUsed: BigNumber { value: "29739" },
//   logsBloom: '0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000',
//   blockHash: '0x208681504632b844622e4155a458fe54076fc4028222a3f688f65f92c5bfe7d9',
//   transactionHash: '0x06e82703d207cd2de32a8faf798fd48c500bd6aa9e72229e05202a438d4121a3',
//   logs: [],
//   blockNumber: 44728600,
//   confirmations: 6,
//   cumulativeGasUsed: BigNumber { value: "29739" },
//   effectiveGasPrice: BigNumber { value: "38931713774040" },
//   status: 1,
//   type: 2,
//   byzantium: true,
//   events: []
// }
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
