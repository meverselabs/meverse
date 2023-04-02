import { network } from "hardhat";

export function getMevAddress(): string {
  switch (network.name) {
    case "mevlocal":
      return "0xa1f093A1d8D4Ed5a7cC8fE29586266C5609a23e8";
    case "mevtestnet":
      return "0xa1f093A1d8D4Ed5a7cC8fE29586266C5609a23e8";
    case "mevmainnet":
      return "0x04EfbDABCBF3F68AE80899f7798E63Bc3E02E8b7";
    default:
      throw new Error("network doesn't exist");
  }
}

export function getMevTransferSampleAddress(): string {
  switch (network.name) {
    case "mevtestnet":
      return "0xdA2Fc6c5D30e0eDf453bD201Ec4a941D1fa72392";
    case "mevmainnet":
      return "0xdA2Fc6c5D30e0eDf453bD201Ec4a941D1fa72392";
    default:
      throw new Error("network doesn't exist");
  }
}
