import i18next from "i18next";
import {getAuthUrl} from "@/auth/Provider";
import * as Setting from "@/lib/setting";

/**
 * MetaMask sign-in, ported from web/src/auth/Web3Auth.js. The EIP-712 payload,
 * the localStorage key and the `web3AuthTokenKey` query parameter are unchanged,
 * so the backend verifies exactly what it verified before.
 *
 * Web3Onboard is not ported: it needs the whole @web3-onboard wallet-module set.
 */

interface Web3AuthToken {
  address: string;
  createAt: number;
  typedData: string;
  signature: string;
}

function getEthereum(): any {
  return (window as any).ethereum;
}

export function getWeb3AuthTokenKey(address: string) {
  return `Web3AuthToken_${address}`;
}

export function setWeb3AuthToken(token: Web3AuthToken) {
  localStorage.setItem(getWeb3AuthTokenKey(token.address), JSON.stringify(token));
}

export function getWeb3AuthToken(address: string): Web3AuthToken | null {
  const raw = localStorage.getItem(getWeb3AuthTokenKey(address));
  try {
    return raw === null ? null : JSON.parse(raw);
  } catch {
    return null;
  }
}

export function delWeb3AuthToken(address: string) {
  localStorage.removeItem(getWeb3AuthTokenKey(address));
}

export function clearWeb3AuthToken() {
  Object.keys(localStorage)
    .filter((key) => key.startsWith("Web3AuthToken_"))
    .forEach((key) => localStorage.removeItem(key));
}

export function detectMetaMaskPlugin() {
  const ethereum = getEthereum();
  return !!ethereum && !!ethereum.isMetaMask;
}

export function requestEthereumAccount(): Promise<string> {
  return getEthereum().request({method: "eth_requestAccounts"}).then((accounts: string[]) => accounts[0]);
}

export function signEthereumTypedData(from: string, nonce: string): Promise<Web3AuthToken> {
  // https://docs.metamask.io/wallet/how-to/sign-data/
  const date = new Date();
  const typedData = JSON.stringify({
    domain: {
      chainId: getEthereum().chainId,
      name: "Casdoor",
      version: "1",
    },
    message: {
      prompt: "In order to authenticate to this website, sign this request and your public address will be sent to the server in a verifiable way.",
      nonce: nonce,
      createAt: `${date.toLocaleString()}`,
    },
    primaryType: "AuthRequest",
    types: {
      EIP712Domain: [
        {name: "name", type: "string"},
        {name: "version", type: "string"},
        {name: "chainId", type: "uint256"},
      ],
      AuthRequest: [
        {name: "prompt", type: "string"},
        {name: "nonce", type: "string"},
        {name: "createAt", type: "string"},
      ],
    },
  });

  return getEthereum()
    .request({method: "eth_signTypedData_v4", params: [from, typedData]})
    .then((signature: string) => ({
      address: from,
      createAt: Math.floor(date.getTime() / 1000),
      typedData,
      signature,
    }));
}

/**
 * Whether a cached token can be reused. The antd frontend also recovers the
 * signer locally (@metamask/eth-sig-util); here the check is structural and the
 * signature is verified server-side, as it always was.
 */
export function checkEthereumSignedTypedData(token: Web3AuthToken | null, address: string) {
  return !!token && token.address === address && !!token.typedData && !!token.signature;
}

export async function authViaMetaMask(application: any, provider: any, method: string) {
  if (!detectMetaMaskPlugin()) {
    Setting.showMessage("error", i18next.t("login:MetaMask plugin not detected"));
    return;
  }
  try {
    const account = await requestEthereumAccount();
    let token = getWeb3AuthToken(account);
    if (!checkEthereumSignedTypedData(token, account)) {
      token = await signEthereumTypedData(account, crypto.randomUUID());
      setWeb3AuthToken(token);
    }
    Setting.goToLink(`${getAuthUrl(application, provider, method)}&web3AuthTokenKey=${getWeb3AuthTokenKey(account)}`);
  } catch (err: any) {
    Setting.showMessage("error", `${i18next.t("login:Failed to obtain MetaMask authorization")}: ${err.message}`);
  }
}
