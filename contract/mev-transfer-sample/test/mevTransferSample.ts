import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers, network } from "hardhat";
import { Wallet, utils } from "ethers";
import { expect } from "chai";
import { getMevAddress } from "./util";

import { MevTransferSample } from "../typechain-types";
import { ERC20Token } from "../typechain-types/contracts/test/ERC20Token";

describe("mevTransferSample", function () {
  let alice: Wallet, bob: Wallet, charlie: Wallet;
  let owner: Wallet;
  let mts: MevTransferSample;
  let mev: ERC20Token;
  const initSupply = utils.parseEther("100000.0");
  let mevAddress = getMevAddress();

  before("get wallets", async () => {
    [alice, bob, charlie] = (await (ethers as any).getSigners()) as Wallet[];
    owner = alice;
  });

  //fixture
  async function fixture() {
    let mev: ERC20Token;
    if (network.name === "hardhat") {
      const testPktTokenFactory = await ethers.getContractFactory("ERC20Token");
      mev = await testPktTokenFactory.deploy(initSupply);
    } else {
      mev = (await ethers.getContractAt("IERC20", mevAddress)) as ERC20Token;
    }

    const mevTransferSampleFactory = await ethers.getContractFactory(
      "MevTransferSample"
    );
    const mts = await mevTransferSampleFactory.deploy();

    return { mev, mts };
  }

  beforeEach("deploy fixture", async () => {
    if (network.name === "hardhat") {
      ({ mev, mts } = await loadFixture(fixture));
    } else {
      ({ mev, mts } = await fixture());
    }
  });

  describe("owner", function () {
    it("ownerChange with Event", async function () {
      await expect(mts.transferOwnership(bob.address))
        .to.emit(mts, "OwnershipTransferred")
        .withArgs(owner.address, bob.address);
    });

    it("ownerChange with owner address change", async function () {
      await mts.transferOwnership(bob.address);

      expect(await mts.owner()).to.equal(bob.address);
    });

    it("fail if owner change to 0-address", async function () {
      await expect(
        mts.transferOwnership(ethers.constants.AddressZero)
      ).to.be.revertedWith("Ownable: new owner is the zero address");
    });

    it("only owner", async function () {
      await expect(
        mts.connect(bob).transferOwnership(charlie.address)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  describe("setData", function () {
    it("setData", async function () {
      await mts.setData(1234);
      expect(await mts.data()).to.equal(1234);
    });

    it("setData with value", async function () {
      const amount = utils.parseEther("5678.0");
      await mts.setData(1234, { value: amount });
      expect(await mts.data()).to.equal(1234);
      expect(await mts.getBalance()).to.equal(amount);
      if (network.name !== "hardhat") {
        expect(await mev.balanceOf(mts.address)).to.equal(amount);
      }
    });
  });

  describe("transfer", function () {
    // 1. send : alice -> contract
    // 2. transfer : contract -> bot
    // alice : alice.balance -> alice.balance - transferAmount - gas1 - gas2
    // contract : 0 -> transferAmount -> 0
    // bob :  0 -> tranaferAmount
    it("transfer", async function () {
      const amount = utils.parseEther("5678.0");

      const beforeTransferAlice = await alice.getBalance();

      const tx1 = await mts.setData(1234, { value: amount });
      const receipt1 = await tx1.wait();
      const gas1 = receipt1.gasUsed.mul(receipt1.effectiveGasPrice);

      const beforeTransferBob = await bob.getBalance();

      const tx2 = await mts.transfer(bob.address);
      const receipt2 = await tx2.wait();
      const gas2 = receipt2.gasUsed.mul(receipt2.effectiveGasPrice);

      expect(await alice.getBalance()).to.equal(
        beforeTransferAlice.sub(amount).sub(gas1).sub(gas2)
      );
      expect(await bob.getBalance()).to.equal(beforeTransferBob.add(amount));
      expect(await mts.getBalance()).to.equal(0);
      if (network.name !== "hardhat")
        expect(await mev.balanceOf(mts.address)).to.equal(0);
    });

    it("onlyOwner", async function () {
      const amount = utils.parseEther("5678.0");
      await mts.setData(1234, { value: amount });

      await expect(
        mts.connect(bob).transfer(charlie.address)
      ).to.be.rejectedWith("Ownable: caller is not the owner");
    });
  });
});
