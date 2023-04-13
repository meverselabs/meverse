import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers, network } from "hardhat";
import { BigNumber, Wallet, utils } from "ethers";
import { expect } from "chai";
import { jestSnapshotPlugin } from "mocha-chai-jest-snapshot";

import { UserReward } from "../types/UserReward";
import { isMeverseFailTx } from "./util";

import { PktTokenTest } from "../typechain-types/contracts/test/PktTokenTest";
import { RewardPool } from "../typechain-types/contracts/RewardPool";

const chai = require("chai");
chai.use(jestSnapshotPlugin());

describe("RewardPool", function () {
  const initSupply = utils.parseEther("100000.0");
  const pktMevlocal = "0xE08FBAd440dfF3267f5A42061D64FC3b953C1Ef1";
  const blockGasLimit = 500000; // mevlocal blockGasLimit = 5000000  / 10
  let wallets: Wallet[];
  let alice: Wallet, bob: Wallet, charlie: Wallet;
  let owner: Wallet;
  let pkt: PktTokenTest;
  let pool: RewardPool;

  before("get wallets", async () => {
    wallets = (await (ethers as any).getSigners()) as Wallet[];
    [alice, bob, charlie] = wallets;
    owner = alice;
  });

  //fixture
  async function fixture() {
    let pkt: PktTokenTest;
    if (network.name !== "mevlocal") {
      const testPktTokenFactory = await ethers.getContractFactory(
        "PktTokenTest"
      );
      pkt = await testPktTokenFactory.deploy(initSupply);
    } else {
      pkt = (await ethers.getContractAt("IERC20", pktMevlocal)) as PktTokenTest;
    }

    const rewardPoolFactory = await ethers.getContractFactory("RewardPool");
    const pool = await rewardPoolFactory.deploy(pkt.address);

    return { pkt, pool };
  }

  // randomAddress : randomly return bytes[20] address
  const hexNumbers = [
    "0",
    "1",
    "2",
    "3",
    "4",
    "5",
    "6",
    "7",
    "8",
    "9",
    "a",
    "b",
    "c",
    "d",
    "e",
    "f",
  ];
  const HexLength = hexNumbers.length;
  const AddressLength = 40;
  function randomAddress() {
    let result = "0x";

    for (let i = 0; i < AddressLength; i++) {
      const randomIndex = Math.floor(Math.random() * HexLength);
      result += hexNumbers[randomIndex];
    }
    return result;
  }

  // makeRewards : make userRewards structs of length size for using pool.addRewards
  const MaxAmount = 1000000000;
  function makeRewards(rewardsMap: Map<string, BigNumber>, size: number) {
    let total: BigNumber = BigNumber.from(0);
    const userRewards: UserReward[] = [];
    for (let i = 0; i < size; i++) {
      const user = randomAddress();
      const amount = BigNumber.from(
        new Number(Math.floor(Math.random() * MaxAmount) + 1).toString()
      );

      total = total.add(amount);
      const u: UserReward = { user, amount };
      userRewards.push(u);
      addMap(rewardsMap, u);
    }
    return { total, userRewards };
  }

  function addMap(rewardsMap: Map<string, BigNumber>, user: UserReward) {
    const existent = rewardsMap.get(user.user);
    if (existent) {
      rewardsMap.set(user.user, existent.add(user.amount));
    } else {
      rewardsMap.set(user.user, user.amount);
    }
  }

  function makeRewardsWithWallet(
    rewardsMap: Map<string, BigNumber>,
    size: number
  ) {
    let total: BigNumber = BigNumber.from(0);
    const userRewards: UserReward[] = [];
    for (let i = 0; i < size; i++) {
      const idx = Math.floor(Math.random() * wallets.length);
      const user = wallets[idx].address;
      const amount = BigNumber.from(
        new Number(Math.floor(Math.random() * MaxAmount) + 1).toString()
      );

      total = total.add(amount);
      const u: UserReward = { user, amount };
      userRewards.push(u);
      addMap(rewardsMap, u);
    }
    return { total, userRewards };
  }

  beforeEach("deploy fixture", async () => {
    if (network.name === "hardhat") {
      ({ pkt, pool } = await loadFixture(fixture));
    } else {
      ({ pkt, pool } = await fixture());
    }
  });

  describe.only("fallback", function () {
    it("fallback", async function () {
      try {
        await expect(
          alice.sendTransaction({
            to: pool.address,
            data: "0x12345678",
            value: utils.parseEther("1.0"),
            gasLimit: blockGasLimit,
          })
        ).to.be.reverted;
      } catch (e) {
        //meverse net
        if (!isMeverseFailTx(e)) {
          throw e;
        }
      }
    });
  });

  describe("receive", function () {
    it("revert", async function () {
      try {
        await expect(
          alice.sendTransaction({
            to: pool.address,
            value: utils.parseEther("1.0"),
            gasLimit: blockGasLimit,
          })
        ).to.be.reverted;
      } catch (e) {
        //meverse net
        if (!isMeverseFailTx(e)) {
          throw e;
        }
      }
    });
  });

  describe("token", function () {
    it("token same with pkt", async function () {
      expect(await pool.token()).to.equal(pkt.address);
    });
  });

  describe("owner", function () {
    it("ownerChange with Event", async function () {
      await expect(pool.transferOwnership(bob.address))
        .to.emit(pool, "OwnershipTransferred")
        .withArgs(owner.address, bob.address);
    });

    it("ownerChange with owner address change", async function () {
      await pool.transferOwnership(bob.address);

      expect(await pool.owner()).to.equal(bob.address);
    });

    it("fail if owner change to 0-address", async function () {
      await expect(
        pool.transferOwnership(ethers.constants.AddressZero)
      ).to.be.revertedWith("Ownable: new owner is the zero address");
    });

    it("only owner", async function () {
      await expect(
        pool.connect(bob).transferOwnership(charlie.address)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  describe("addReward", function () {
    it("addReward with Event", async function () {
      const rewardMap = new Map();
      const { total, userRewards } = makeRewards(rewardMap, 10);
      await pkt.approve(pool.address, total);

      if (network.name !== "mevlocal") {
        await expect(pool.addReward(total, userRewards))
          .to.emit(pool, "RewardAdded")
          .withArgs(total)
          .to.emit(pkt, "Transfer")
          .withArgs(owner.address, pool.address, total);
      } else {
        await expect(pool.addReward(total, userRewards))
          .to.emit(pool, "RewardAdded")
          .withArgs(total);
      }
    });

    it("gas Price with userRewards 100", async function () {
      const rewardMap = new Map();
      const { total, userRewards } = makeRewards(rewardMap, 100);
      await pkt.approve(pool.address, total);

      const tx = await pool.addReward(total, userRewards);
      const receipt = await tx.wait();

      expect(receipt.gasUsed.toString()).toMatchSnapshot();
    });

    it("fail if allowance is smaller than total", async function () {
      const rewardMap = new Map();
      const { total, userRewards } = makeRewards(rewardMap, 10);
      await pkt.approve(pool.address, total.sub(1));

      try {
        await expect(
          pool.addReward(total, userRewards, {
            gasLimit: blockGasLimit,
          })
        ).to.be.reverted;
      } catch (e) {
        //meverse net
        if (!isMeverseFailTx(e)) {
          throw e;
        }
      }
    });

    it("fail if total is smaller", async function () {
      const rewardMap = new Map();
      const { total, userRewards } = makeRewards(rewardMap, 10);
      await expect(
        pool.addReward(total.sub(1), userRewards)
      ).to.be.rejectedWith("NFTFarm: input total is different from sum");
    });

    it("fail if total is larger", async function () {
      const rewardMap = new Map();
      const { total, userRewards } = makeRewards(rewardMap, 10);
      await expect(
        pool.addReward(total.add(1), userRewards)
      ).to.be.rejectedWith("NFTFarm: input total is different from sum");
    });

    it("onlyOwner", async function () {
      const rewardMap = new Map();
      const { total, userRewards } = makeRewards(rewardMap, 10);
      await expect(
        pool.connect(bob).addReward(total, userRewards)
      ).to.be.rejectedWith("Ownable: caller is not the owner");
    });

    // 100 is safe when block gas limit = 5000000
    for (let i = 500; i < 600; i += 10) {
      it.skip("maximum length of userRewards", async function () {
        const { total, userRewards } = makeRewards(new Map(), i);
        await pkt.approve(pool.address, total);
        await pool.addReward(total, userRewards);
        expect(i).toMatchSnapshot();
      });
    }

    it("pkt.balanceOf, pool.totalReward, pool.rewards", async function () {
      const rewardMap = new Map();
      let totalReward: BigNumber = BigNumber.from(0);

      for (const userLength of [100, 100, 100]) {
        let adminBalanceBefore = await pkt.balanceOf(owner.address);
        let contractBalanceBefore = await pkt.balanceOf(pool.address);

        const { total, userRewards } = makeRewards(rewardMap, userLength);

        await pkt.approve(pool.address, total);
        await pool.addReward(total, userRewards);

        //balance
        expect(await pkt.balanceOf(owner.address)).to.equal(
          adminBalanceBefore.sub(total)
        );
        expect(await pkt.balanceOf(pool.address)).to.equal(
          contractBalanceBefore.add(total)
        );

        //totalReward
        totalReward = totalReward.add(total);
        expect(await pool.totalReward()).to.equal(totalReward);
      }

      //rewards
      rewardMap.forEach(async (v, k) => {
        expect(await pool.rewards(k)).to.equal(v);
      });
    });

    it("claim", async function () {
      const rewardMap = new Map();
      let totalReward: BigNumber = BigNumber.from(0);

      if (network.name === "gethlocal") {
        for (let i = 1; i < wallets.length; i++) {
          await alice.sendTransaction({
            to: wallets[i].address,
            value: utils.parseEther("1.0"),
          });
        }
      }

      for (const userLength of [100, 100, 100]) {
        const { total, userRewards } = makeRewardsWithWallet(
          rewardMap,
          userLength
        );
        await pkt.approve(pool.address, total);
        await pool.addReward(total, userRewards);
        totalReward = totalReward.add(total);
      }

      for (let i = 0; i < wallets.length; i++) {
        const wallet = wallets[i];
        expect(await pool.rewards(wallet.address)).to.equal(
          rewardMap.get(wallet.address)
        );
        const r = await pool.rewards(wallet.address);
        await pool.connect(wallet).claim();
        expect(await pool.rewards(wallet.address)).to.equal(0);
        totalReward = totalReward.sub(r);
        expect(await pool.totalReward()).to.equal(totalReward);
      }

      expect(totalReward).to.equal(0);
      expect(await pool.totalReward()).to.equal(0);
    });
  });

  describe("rewards", function () {
    let userRewards: UserReward[];
    const rewardAmount = BigNumber.from(123456);

    beforeEach("addReward", async () => {
      userRewards = [];
      userRewards.push({ user: bob.address, amount: rewardAmount });
      await pkt.approve(pool.address, rewardAmount);
      await pool.addReward(rewardAmount, userRewards);
    });

    it("reward > 0", async function () {
      expect(await pool.rewards(bob.address)).to.equal(rewardAmount);
    });

    it("reward = 0", async function () {
      expect(await pool.rewards(charlie.address)).to.equal(0);
    });
  });

  describe("claim", function () {
    let userRewards: UserReward[];
    const rewardAmount = BigNumber.from(123456);

    beforeEach("addReward", async () => {
      userRewards = [];
      userRewards.push({ user: bob.address, amount: rewardAmount });
      await pkt.approve(pool.address, rewardAmount);
      await pool.addReward(rewardAmount, userRewards);

      // ether send to bob :  in case of gethlocal
      await alice.sendTransaction({
        to: bob.address,
        value: utils.parseEther("1.0"),
      });
    });

    it("claim with Event", async function () {
      if (network.name !== "mevlocal") {
        await expect(pool.connect(bob).claim())
          .to.emit(pool, "Claim")
          .withArgs(bob.address, rewardAmount)
          .to.emit(pkt, "Transfer")
          .withArgs(pool.address, bob.address, rewardAmount);
      } else {
        await expect(pool.connect(bob).claim())
          .to.emit(pool, "Claim")
          .withArgs(bob.address, rewardAmount);
      }
    });

    it("gas Price", async function () {
      const tx = await pool.connect(bob).claim();
      const receipt = await tx.wait();

      expect(receipt.gasUsed.toString()).toMatchSnapshot();
    });

    it("pkt.balanceOf", async function () {
      const beforeAmount = await pkt.balanceOf(bob.address);
      await pool.connect(bob).claim();
      expect(await pkt.balanceOf(bob.address)).to.equal(
        beforeAmount.add(rewardAmount)
      );
    });

    it("fail if pool is unClaimable", async function () {
      await pool.setClaimable(false);
      await expect(pool.connect(bob).claim()).to.be.revertedWith(
        "RewardPool: pool is unclaimable"
      );
    });

    it("should claim after claimable", async function () {
      await pool.setClaimable(false);
      await pool.setClaimable(true);
      await pool.connect(bob).claim();
    });

    it("fail if reward is 0", async function () {
      await expect(pool.connect(charlie).claim()).to.be.revertedWith(
        "RewardPool: claim amount is zero"
      );
    });
  });

  describe("revokeTotalReward", function () {
    let userRewards: UserReward[];
    const rewardAmount = BigNumber.from(123456);

    beforeEach("addReward", async () => {
      userRewards = [];
      userRewards.push({ user: bob.address, amount: rewardAmount });
      await pkt.approve(pool.address, rewardAmount);
      await pool.addReward(rewardAmount, userRewards);
    });

    it("revokeTotalReward with Event", async function () {
      const totalReward = await pool.totalReward();
      if (network.name !== "mevlocal") {
        await expect(pool.revokeTotalReward())
          .to.emit(pool, "Revoke")
          .withArgs(owner.address, totalReward)
          .to.emit(pkt, "Transfer")
          .withArgs(pool.address, owner.address, totalReward);
      } else {
        await expect(pool.revokeTotalReward())
          .to.emit(pool, "Revoke")
          .withArgs(owner.address, totalReward);
      }
    });

    it("gas Price", async function () {
      const tx = await pool.revokeTotalReward();
      const receipt = await tx.wait();

      expect(receipt.gasUsed.toString()).toMatchSnapshot();
    });

    it("balance change of the owner", async function () {
      const balanceBefore = await pkt.balanceOf(owner.address);
      const totalReward = await pool.totalReward();
      await pool.revokeTotalReward();
      expect(await pkt.balanceOf(owner.address)).to.equal(
        balanceBefore.add(totalReward)
      );
    });

    it("user balance unchange", async function () {
      const balanceBefore = await pkt.balanceOf(bob.address);
      await pool.revokeTotalReward();
      expect(await pkt.balanceOf(bob.address)).to.equal(balanceBefore);
    });

    it("totalReward unchange", async function () {
      const totalRewardBefore = await pool.totalReward();
      await pool.revokeTotalReward();
      expect(await pool.totalReward()).to.equal(totalRewardBefore);
    });

    it("unClaimable after revokeTotalReward", async function () {
      await pool.revokeTotalReward();
      expect(await pool.isClaimable()).to.be.false;
    });

    it("onlyOwner", async function () {
      await expect(pool.connect(bob).revokeTotalReward()).to.be.rejectedWith(
        "Ownable: caller is not the owner"
      );
    });
  });

  describe("setClaimable", function () {
    it("setClaimable with Event", async function () {
      await expect(pool.setClaimable(false))
        .to.emit(pool, "Claimable")
        .withArgs(false);
    });

    it("gas Price", async function () {
      const tx = await pool.setClaimable(false);
      const receipt = await tx.wait();

      expect(receipt.gasUsed.toString()).toMatchSnapshot();
    });

    it("fail setClaimable(true) if isClaimable is true", async function () {
      await expect(pool.setClaimable(true)).to.be.revertedWith(
        "RewardPool: isClaimable is already set"
      );
    });

    it("fail setClaimable(false) if isClaimable is false", async function () {
      await pool.setClaimable(false);
      await expect(pool.setClaimable(false)).to.be.revertedWith(
        "RewardPool: isClaimable is already set"
      );
    });

    it("onlyOwner", async function () {
      await expect(pool.connect(bob).setClaimable(false)).to.be.rejectedWith(
        "Ownable: caller is not the owner"
      );
    });
  });
});
