import { network } from "hardhat";
import { ProviderError } from "hardhat/internal/core/providers/errors";

export function isMeverseFailTx(e: any) {
  if (
    network.name === "mevlocal" &&
    e instanceof ProviderError &&
    e.message == "failtx"
  )
    return true;
  return false;
}
