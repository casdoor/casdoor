import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import * as TokenBackend from "@/backend/TokenBackend";

export default function TokenEditPage() {
  const {tokenName = ""} = useParams();

  const fields: EditField[] = [
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "application", labelKey: "general:Application"},
    {type: "text", name: "organization", labelKey: "general:Organization"},
    {type: "text", name: "user", labelKey: "general:User"},
    {type: "textarea", name: "code", labelKey: "token:Authorization code", rows: 3},
    {type: "textarea", name: "accessToken", labelKey: "token:Access token", rows: 6},
    {type: "textarea", name: "refreshToken", labelKey: "token:Refresh token", rows: 6},
    {type: "number", name: "expiresIn", labelKey: "token:Expires in"},
    {type: "text", name: "scope", labelKey: "provider:Scope"},
    {type: "text", name: "tokenType", labelKey: "token:Token type"},
  ];

  return (
    <SimpleEditPage
      titleKey="token:Edit Token"
      backTo="/tokens"
      deps={[tokenName]}
      fields={fields}
      fetch={() => TokenBackend.getToken("admin", tokenName)}
      add={(record) => TokenBackend.addToken(record)}
      update={(record) => TokenBackend.updateToken("admin", tokenName, record)}
      editUrl={(record) => `/tokens/${record.name}`}
    />
  );
}
