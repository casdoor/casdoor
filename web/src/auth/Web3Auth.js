// // Copyright 2023 The Casdoor Authors. All Rights Reserved.
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //      http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.

import {goToLink, showMessage} from "../Setting";
import i18next from "i18next";
import {v4 as uuidv4} from "uuid";
import {SignTypedDataVersion, recoverTypedSignature} from "@metamask/eth-sig-util";
import {getAuthUrl} from "./Provider";
import {Buffer} from "buffer";
import Onboard from "@web3-onboard/core";
import injectedModule from "@web3-onboard/injected-wallets";
import infinityWalletModule from "@web3-onboard/infinity-wallet";
import sequenceModule from "@web3-onboard/sequence";
import trustModule from "@web3-onboard/trust";
import frontierModule from "@web3-onboard/frontier";
import tahoModule from "@web3-onboard/taho";
import coinbaseModule from "@web3-onboard/coinbase";
import gnosisModule from "@web3-onboard/gnosis";
// import keystoneModule from "@web3-onboard/keystone";
// import keepkeyModule from "@web3-onboard/keepkey";
// import dcentModule from "@web3-onboard/dcent";
// import ledgerModule from "@web3-onboard/ledger";
// import trezorModule from "@web3-onboard/trezor";
// import walletConnectModule from "@web3-onboard/walletconnect";
// import fortmaticModule from "@web3-onboard/fortmatic";
// import portisModule from "@web3-onboard/portis";
// import magicModule from "@web3-onboard/magic";

global.Buffer = Buffer;

export function generateNonce() {
  const nonce = uuidv4();
  return nonce;
}

export function getWeb3AuthTokenKey(address) {
  return `Web3AuthToken_${address}`;
}

export function setWeb3AuthToken(token) {
  const key = getWeb3AuthTokenKey(token.address);
  localStorage.setItem(key, JSON.stringify(token));
}

export function getWeb3AuthToken(address) {
  const key = getWeb3AuthTokenKey(address);
  return JSON.parse(localStorage.getItem(key));
}

export function delWeb3AuthToken(address) {
  const key = getWeb3AuthTokenKey(address);
  localStorage.removeItem(key);
}

export function clearWeb3AuthToken() {
  const keys = Object.keys(localStorage);
  keys.forEach(key => {
    if (key.startsWith("Web3AuthToken_")) {
      localStorage.removeItem(key);
    }
  });
}

export function detectMetaMaskPlugin() {
  // check if ethereum extension MetaMask is installed
  return window.ethereum && window.ethereum.isMetaMask;
}

export function requestEthereumAccount() {
  const method = "eth_requestAccounts";
  const selectedAccount = window.ethereum.request({method})
    .then((accounts) => {
      return accounts[0];
    });
  return selectedAccount;
}

export function signEthereumTypedData(from, nonce) {
  // https://docs.metamask.io/wallet/how-to/sign-data/
  const date = new Date();
  const typedData = JSON.stringify({
    domain: {
      chainId: window.ethereum.chainId,
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

  const method = "eth_signTypedData_v4";
  const params = [from, typedData];

  return window.ethereum.request({method, params})
    .then((sign) => {
      return {
        address: from,
        createAt: Math.floor(date.getTime() / 1000),
        typedData: typedData,
        signature: sign,
      };
    });
}

export function checkEthereumSignedTypedData(token) {
  if (token === undefined || token === null) {
    return false;
  }
  if (token.address && token.typedData && token.signature) {
    const recoveredAddr = recoverTypedSignature({
      data: JSON.parse(token.typedData),
      signature: token.signature,
      version: SignTypedDataVersion.V4,
    });
    // const recoveredAddr = token.address;
    return recoveredAddr === token.address;
    // return toChecksumAddress(recoveredAddr) === toChecksumAddress(token.address);
  }
  return false;
}

export async function authViaMetaMask(application, provider, method) {
  if (!detectMetaMaskPlugin()) {
    showMessage("error", `${i18next.t("login:MetaMask plugin not detected")}`);
    return;
  }
  try {
    const account = await requestEthereumAccount();
    let token = getWeb3AuthToken(account);
    if (!checkEthereumSignedTypedData(token)) {
      const nonce = generateNonce();
      token = await signEthereumTypedData(account, nonce);
      setWeb3AuthToken(token);
    }
    const redirectUri = `${getAuthUrl(application, provider, method)}&web3AuthTokenKey=${getWeb3AuthTokenKey(account)}`;
    goToLink(redirectUri);
  } catch (err) {
    showMessage("error", `${i18next.t("login:Failed to obtain MetaMask authorization")}: ${err.message}`);
  }
}

const web3Wallets = {
  // injected wallets
  injected: {
    label: "Injected",
    wallet: injectedModule(),
  },
  // sdk wallets
  coinbase: {
    label: "Coinbase",
    wallet: coinbaseModule(),
  },
  trust: {
    label: "Trust",
    wallet: trustModule(),
  },
  gnosis: {
    label: "Gnosis",
    wallet: gnosisModule(),
  },
  sequence: {
    label: "Sequence",
    wallet: sequenceModule(),
  },
  taho: {
    label: "Taho",
    wallet: tahoModule(),
  },
  frontier: {
    label: "Frontier",
    wallet: frontierModule(),
  },
  infinityWallet: {
    label: "Infinity Wallet",
    wallet: infinityWalletModule(),
  },
  // hardware wallets
  // keystone: {
  //   label: "Keystone",
  //   wallet: keystoneModule(),
  // },
  // keepkey: {
  //   label: "KeepKey",
  //   wallet: keepkeyModule(),
  // },
  // dcent: {
  //   label: "D'CENT",
  //   wallet: dcentModule(),
  // },

  // some wallet need custome `apiKey` or `projectId` configure item
  // const magic = magicModule({
  //   apiKey: "magicApiKey",
  // });
  // const fortmatic = fortmaticModule({
  //   apiKey: "fortmaticApiKey",
  // });
  // const portis = portisModule({
  //   apiKey: "portisApiKey",
  // });
  // const ledger = ledgerModule({
  //   projectId: "ledgerProjectId"
  // });
  // const walletConnect = walletConnectModule({
  //   projectId: "walletConnectProjectId",
  // });
};

export function getWeb3OnboardWalletsOptions() {
  return Object.entries(web3Wallets).map(([key, value]) => ({
    label: value.label,
    value: key,
  }));
}

function getWeb3OnboardWallets(options) {
  if (options === null || options === undefined || !Array.isArray(options)) {
    return [];
  }
  return options.map(walletType => {
    if (walletType && web3Wallets[walletType]?.wallet) {
      return web3Wallets[walletType]?.wallet;
    }
  });
}

export function initWeb3Onboard(application, provider) {
  // init wallet
  // options = ["injected","coinbase",...]
  const options = JSON.parse(provider.metadata);
  const wallets = getWeb3OnboardWallets(options);

  // init chain
  // const InfuraKey = "2fa45cbe531e4e65be4fcbf408e651a8";
  const chains = [
    // {
    //   id: "0x1",
    //   token: "ETH",
    //   label: "Ethereum Mainnet",
    //   rpcUrl: `https://mainnet.infura.io/v3/${InfuraKey}`,
    // },
    // {
    //   id: "0x5",
    //   token: "ETH",
    //   label: "Goerli",
    //   rpcUrl: `https://goerli.infura.io/v3/${InfuraKey}`,
    // },
    {
      id: "0x13881",
      token: "MATIC",
      label: "Polygon - Mumbai",
      rpcUrl: "https://matic-mumbai.chainstacklabs.com",
    },
    {
      id: "0x38",
      token: "BNB",
      label: "Binance",
      rpcUrl: "https://bsc-dataseed.binance.org/",
    },
    {
      id: "0xA",
      token: "OETH",
      label: "Optimism",
      rpcUrl: "https://mainnet.optimism.io",
    },
    {
      id: "0xA4B1",
      token: "ARB-ETH",
      label: "Arbitrum",
      rpcUrl: "https://rpc.ankr.com/arbitrum",
    },
  ];

  const appMetadata = {
    name: "Casdoor",
    description: "Connect a wallet using Casdoor",
    recommendedInjectedWallets: [
      {name: "MetaMask", url: "https://metamask.io"},
      {name: "Coinbase", url: "https://www.coinbase.com/wallet"},
    ],
  };

  const web3Onboard = Onboard({
    wallets,
    chains,
    appMetadata,
  });
  return web3Onboard;
}

export async function authViaWeb3Onboard(application, provider, method) {
  try {
    const onboard = initWeb3Onboard(application, provider);
    const connectedWallets = await onboard.connectWallet();
    if (connectedWallets.length > 0) {
      const wallet = connectedWallets[0];
      const account = wallet.accounts[0];
      const address = account.address;
      const token = {
        address: address, // e.g."0xbd5444d31fe4139ee36bea29e43d4ac67ae276de"
        walletType: wallet.label, // e.g."MetaMask"
        createAt: Math.floor(new Date().getTime() / 1000),
      };
      setWeb3AuthToken(token);
      const redirectUri = `${getAuthUrl(application, provider, method)}&web3AuthTokenKey=${getWeb3AuthTokenKey(address)}`;
      goToLink(redirectUri);
    }
  } catch (err) {
    showMessage("error", `${i18next.t("login:Failed to obtain Web3-Onboard authorization")}: ${err}`);
  }
}
