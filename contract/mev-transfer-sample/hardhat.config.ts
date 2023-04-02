import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

const config: HardhatUserConfig = {
  networks: {
    hardhat: {
      allowUnlimitedContractSize: false,
    },
    mevlocal: {
      url: "http://127.0.0.1:8541",
      chainId: 65535,
      accounts: [
        "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", //0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
        "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d", //0x70997970c51812dc3a010c7d01b50e0d17dc79c8
        "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a", //0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc
      ],
      timeout: 1000000,
    },
    mevtestnet: {
      url: "https://rpc.meversetestnet.io",
      chainId: 0x1297,
      accounts: [
        "0xdf6a85a142a59e783ce9a12e3639a586afe4a8d548038c0cb9c054efc287d6f1", //0x04EfbDABCBF3F68AE80899f7798E63Bc3E02E8b7
      ],
      timeout: 1000000,
    },
    mevmainnet: {
      url: "https://rpc.meversemainnet.io",
      chainId: 0x1d5e,
      accounts: [
        "0xdf6a85a142a59e783ce9a12e3639a586afe4a8d548038c0cb9c054efc287d6f1", //0x04EfbDABCBF3F68AE80899f7798E63Bc3E02E8b7
      ],
      timeout: 1000000,
    },
  },
  mocha: {
    timeout: 3600000000000,
  },
  solidity: {
    version: "0.8.17",
  },
};

export default config;
