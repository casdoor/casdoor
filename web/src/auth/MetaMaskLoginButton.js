// Copyright 2024 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {getAuthUrl} from "./Provider";
import {getProviderLogoURL, goToLink, showMessage} from "../Setting";
import i18next from "i18next";
import {
  generateNonce,
  getWeb3AuthTokenKey,
  setWeb3AuthToken
} from "./Web3Auth";
import {useSDK} from "@metamask/sdk-react";
import React, {useEffect} from "react";

export function MetaMaskLoginButton(props) {
  const {application, web3Provider, method, width, margin} = props;
  const {sdk, chainId, account} = useSDK();
  const [typedData, setTypedData] = React.useState("");
  const [nonce, setNonce] = React.useState("");
  const [signature, setSignature] = React.useState();

  useEffect(() => {
    if (account && signature) {
      const date = new Date();

      const token = {
        address: account,
        nonce: nonce,
        createAt: Math.floor(date.getTime() / 1000),
        typedData: typedData,
        signature: signature,
      };
      setWeb3AuthToken(token);

      const redirectUri = `${getAuthUrl(application, web3Provider, method)}&web3AuthTokenKey=${getWeb3AuthTokenKey(account)}`;
      goToLink(redirectUri);
    }
  }, [account, signature]);

  const handleConnectAndSign = async() => {
    try {
      terminate();

      const date = new Date();

      const nonce = generateNonce();
      setNonce(nonce);

      const prompt = web3Provider?.metadata === "" ? "Casdoor: In order to authenticate to this website, sign this request and your public address will be sent to the server in a verifiable way." : web3Provider.metadata;
      const typedData = JSON.stringify({
        domain: {
          chainId: chainId,
          name: "Casdoor",
          version: "1",
        },
        message: {
          prompt: `${prompt}`,
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
      setTypedData(typedData);

      const sig = await sdk.connectAndSign({msg: typedData});
      setSignature(sig);
    } catch (err) {
      showMessage("error", `${i18next.t("login:Failed to obtain MetaMask authorization")}: ${err.message}`);
    }
  };

  const terminate = () => {
    sdk?.terminate();
  };

  return (
    <a key={web3Provider.displayName} onClick={handleConnectAndSign}>
      <img width={width} height={width} src={getProviderLogoURL(web3Provider)} alt={web3Provider.displayName}
        className="provider-img" style={{margin: margin}} />
    </a>
  );
}

export default MetaMaskLoginButton;
