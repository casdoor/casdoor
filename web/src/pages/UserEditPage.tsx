import * as React from "react";
import i18next from "i18next";
import {Link, useLocation, useNavigate, useParams} from "react-router-dom";
import {ShieldCheck, Users} from "lucide-react";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Popover, PopoverContent, PopoverTrigger} from "@/components/ui/popover";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {CountryCodeSelect} from "@/components/common/CountryCodeSelect";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {RegionSelect} from "@/components/common/RegionSelect";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow, formGridClass} from "@/components/crud/FormRow";
import {AffiliationAddressSelect, AffiliationField, useAffiliation} from "@/components/user/AffiliationSelect";
import {CartTable} from "@/components/user/CartTable";
import {CasdoorAppQrCode, CasdoorAppUrl} from "@/components/user/CasdoorAppConnector";
import {ConsentTable} from "@/components/user/ConsentTable";
import {FaceIdTable} from "@/components/user/FaceIdTable";
import {CropperDivModal, UserImageField} from "@/components/user/CropperDivModal";
import {ThirdPartyLogins} from "@/components/user/OAuthWidget";
import {AccountItemRow, AccountItemsProvider} from "@/components/user/AccountItemRow";
import {isAccountItemVisible} from "@/lib/account-items";
import {PasswordModal} from "@/components/user/PasswordModal";
import {ResetModal} from "@/components/user/ResetModal";
import {TransactionTable} from "@/components/user/TransactionTable";
import {WebauthnCredentialTable} from "@/components/user/WebauthnCredentialTable";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {dropExtraPhysicalGroups, useApplicationOptions, useGroupList, useGroupOptions, useOrganizationOptions} from "@/hooks/use-options";
import {getModeTitleKey, submitEdit} from "@/lib/crud";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as MfaBackend from "@/backend/MfaBackend";
import * as TransactionBackend from "@/backend/TransactionBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/** the value a Radix tab needs for the account item whose `tab` is the empty string */
const DEFAULT_TAB = "__default__";

const GENDERS = ["Male", "Female", "Other"];
/** the three ID card pictures, stored as user properties and uploaded under the same tag */
const ID_CARD_IMAGES = [
  {field: "idCardFront", uploadKey: "user:Upload ID card front picture", setKey: "user:ID card front"},
  {field: "idCardBack", uploadKey: "user:Upload ID card back picture", setKey: "user:ID card back"},
  {field: "idCardWithPerson", uploadKey: "user:Upload ID card with person picture", setKey: "user:ID card with person"},
];
const ID_CARD_TYPES = ["ID card", "Passport", "Driver license"];
/** the rows of `user.addresses` are object.Address: tag/line1/line2/city/state/zipCode/region */
const ADDRESS_TAGS = [
  {value: "Home", labelKey: "general:Home"},
  {value: "Work", labelKey: "user:Work"},
  {value: "Other", labelKey: "user:Other"},
];

export default function UserEditPage({self}: {self?: boolean} = {}) {
  const params = useParams();
  const navigate = useNavigate();
  const {account, accessToken, reload: reloadAccount} = useAccount();
  const returnUrl = new URLSearchParams(useLocation().search).get("returnUrl") ?? "";
  // "/account" edits the signed-in user, every other route takes the name from the URL
  const organizationName = self ? account?.owner ?? "" : params.organizationName ?? "";
  const userName = self ? account?.name ?? "" : params.userName ?? "";
  const [saving, setSaving] = React.useState(false);
  // the antd page keeps the open tab in the URL hash; "" is the unnamed tab
  const [activeTab, setActiveTab] = React.useState(() => window.location.hash.replace(/^#/, "") || DEFAULT_TAB);
  const selectTab = (key: string) => {
    setActiveTab(key);
    window.location.hash = key === DEFAULT_TAB ? "" : key;
  };
  const [organization, setOrganization] = React.useState<any>({});
  const [mfaItems, setMfaItems] = React.useState<any[]>([]);
  const [removingMfa, setRemovingMfa] = React.useState(false);
  // the signup application drives the 3rd-party login rows and the reset-code flow
  const [application, setApplication] = React.useState<any>(null);
  const [transactions, setTransactions] = React.useState<any[]>([]);

  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const groups = useGroupOptions(organizationName);
  const groupList = useGroupList(organizationName);

  const {record: user, updateField, updateFields, loading, denied, mode, setMode, reload} = useEditRecord<any>({
    fetch: () => UserBackend.getUser(organizationName, userName),
    deps: [organizationName, userName],
  });
  const affiliation = useAffiliation(application, user);

  React.useEffect(() => {
    setMfaItems(user?.multiFactorAuths ?? []);
  }, [user]);

  React.useEffect(() => {
    OrganizationBackend.getOrganization("admin", organizationName).then((res: any) => {
      if (res.status === "ok") {
        setOrganization(res.data ?? {});
      }
    });
  }, [organizationName]);

  React.useEffect(() => {
    // in "add" mode the user is not persisted yet, so both requests would fail
    if (mode === "add" || !organizationName || !userName) {
      return;
    }
    ApplicationBackend.getUserApplication(organizationName, userName).then((res: any) => {
      if (res.status === "ok") {
        setApplication(res.data ?? null);
      }
    });
    TransactionBackend.getTransactions(organizationName, "", "", "user", userName).then((res: any) => {
      if (res.status === "ok") {
        setTransactions(res.data ?? []);
      }
    });
  }, [mode, organizationName, userName]);

  React.useEffect(() => {
    // in "add" mode the user has no application yet, so the accountItems layout is
    // resolved from the signup application the new user will belong to
    if (mode !== "add" || !user?.signupApplication) {
      return;
    }
    ApplicationBackend.getApplication("admin", user.signupApplication).then((res: any) => {
      if (res.status === "ok") {
        setApplication(res.data ?? null);
      }
    });
  }, [mode, user?.signupApplication]);

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || user === null || (self && !account)) {
    return <Loading />;
  }

  const isAdmin = Setting.isLocalAdminUser(account);
  const isSelf = account?.owner === user.owner && account?.name === user.name;

  const save = async(exitAfterSave: boolean) => {
    const isAdd = mode === "add";
    setSaving(true);
    await submitEdit({
      mode,
      record: Setting.deepCopy(user),
      add: (record) => UserBackend.addUser(record),
      update: (record) => UserBackend.updateUser(organizationName, userName, record),
      onSaved: () => {
        setMode("edit");
        if (isSelf) {
          reloadAccount();
        }
        if (exitAfterSave) {
          // Casdoor can send a user here to complete their profile and hand them
          // back afterwards; without that the user list is admin-only, so a plain
          // user goes home instead.
          if (returnUrl) {
            Setting.goToLink(returnUrl);
          } else if (self || !isAdmin) {
            navigate("/");
          } else {
            navigate("/users");
          }
        } else if (user.name !== userName) {
          navigate(`/users/${user.owner}/${user.name}`, {replace: true});
        } else if (isAdd) {
          // the user exists now: its ID is generated by the server, and its avatar,
          // MFA and consents can be loaded
          reload();
        }
      },
      onFailed: () => {
        // a rejected rename would otherwise leave the form pointing at a user that
        // does not exist, and the next save would hit the wrong URL
        if (!isAdd) {
          updateFields({owner: organizationName, name: userName});
        }
      },
    });
    setSaving(false);
  };

  const setPreferredMfa = (mfaType: string) => {
    MfaBackend.SetPreferredMfa({owner: user.owner, name: user.name, mfaType}).then((res: any) => {
      if (res.status === "ok") {
        setMfaItems(res.data ?? []);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const deleteMfa = () => {
    setRemovingMfa(true);
    return MfaBackend.DeleteMfa({owner: user.owner, name: user.name})
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          setMfaItems(res.data ?? []);
        } else {
          Setting.showMessage("error", i18next.t("general:Failed to delete"));
        }
      })
      .finally(() => setRemovingMfa(false));
  };

  const verifyIdentification = () => {
    if (!user.idCard || !user.idCardType) {
      Setting.showMessage("error", i18next.t("user:Please fill in ID card information first"));
      return;
    }
    if (!user.realName) {
      Setting.showMessage("error", i18next.t("user:Please fill in your real name first"));
      return;
    }
    // the backend picks the provider and the logged-in user itself
    UserBackend.verifyIdentification(user.owner, user.name, "").then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("user:Identity verification successful"));
        reload();
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  // The form layout is a per-organization policy, and `get-user-application` is the
  // only way a non-admin can read their own organization, so prefer it over the
  // admin-only `get-organization` this page also fetches.
  const userOrganization = application?.organizationObj ?? organization;
  const isSelfOrAdmin = isSelf || isAdmin;

  const avatarUrl = Setting.getEffectiveAvatarUrl(user);
  // `properties` is a plain map on the wire, the table edits it as rows
  const propertyRows = Object.entries(user.properties ?? {}).map(([key, value]) => ({key, value}));

  const rows: Record<string, React.ReactNode> = {
    "Organization": (
      <AccountItemRow name="Organization" labelKey="general:Organization">
        <SearchableSelect
          value={user.owner}
          disabled={!Setting.isAdminUser(account)}
          onChange={(v) => updateField("owner", v)}
          options={organizations}
        />
      </AccountItemRow>
    ),
    "ID": (
      <AccountItemRow name="ID" labelKey="general:ID">
        <Input value={user.id ?? ""} onChange={(e) => updateField("id", e.target.value)} />
      </AccountItemRow>
    ),
    "Name": (
      <AccountItemRow name="Name" labelKey="general:Name">
        <Input value={user.name ?? ""} disabled={!isAdmin} onChange={(e) => updateField("name", e.target.value)} />
      </AccountItemRow>
    ),
    "Display name": (
      <AccountItemRow name="Display name" labelKey="general:Display name">
        <Input value={user.displayName ?? ""} onChange={(e) => updateField("displayName", e.target.value)} />
      </AccountItemRow>
    ),
    "First name": (
      <AccountItemRow name="First name" labelKey="general:First name">
        <Input value={user.firstName ?? ""} onChange={(e) => updateField("firstName", e.target.value)} />
      </AccountItemRow>
    ),
    "Last name": (
      <AccountItemRow name="Last name" labelKey="general:Last name">
        <Input value={user.lastName ?? ""} onChange={(e) => updateField("lastName", e.target.value)} />
      </AccountItemRow>
    ),
    "Avatar": (
      <AccountItemRow name="Avatar" labelKey="general:Avatar">
        <div className="flex items-center gap-3">
          <Avatar className="h-16 w-16">
            {avatarUrl ? <AvatarImage src={avatarUrl} alt={user.name} /> : null}
            <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(user.name ?? "?"), color: "#fff"}}>
              {(user.name || "?").charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <Input value={user.avatar ?? ""} onChange={(e) => updateField("avatar", e.target.value)} />
          {/* the upload creates a resource owned by the user, so it needs the user to exist */}
          {mode === "add" ? null : (
            <CropperDivModal
              tag="avatar"
              title={i18next.t("user:Upload a photo")}
              setTitle={i18next.t("user:Set new profile picture")}
              buttonText={`${i18next.t("user:Upload a photo")}...`}
              user={user}
              organization={userOrganization}
              onUploaded={reload}
            />
          )}
        </div>
      </AccountItemRow>
    ),
    "Email": (
      <AccountItemRow name="Email" labelKey="general:Email">
        <div className="flex gap-2">
          <Input type="email" value={user.email ?? ""} disabled={!isAdmin} onChange={(e) => updateField("email", e.target.value)} />
          {/* the backend resolves the current user itself, so only the user can reset their own */}
          {isSelf ? (
            <ResetModal application={application} destType="email" buttonText={i18next.t("user:Reset Email...")} />
          ) : null}
        </div>
      </AccountItemRow>
    ),
    "Phone": (
      <AccountItemRow name="Phone" labelKey="general:Phone">
        <div className="flex gap-2">
          <div className="w-32 shrink-0">
            <CountryCodeSelect
              value={user.countryCode ?? ""}
              onChange={(v) => updateField("countryCode", v)}
              countryCodes={userOrganization.countryCodes}
            />
          </div>
          <Input value={user.phone ?? ""} disabled={!isAdmin} onChange={(e) => updateField("phone", e.target.value)} />
          {isSelf ? (
            <ResetModal
              application={application}
              destType="phone"
              countryCode={user.countryCode ?? ""}
              buttonText={i18next.t("user:Reset Phone...")}
            />
          ) : null}
        </div>
      </AccountItemRow>
    ),
    "User type": (
      <AccountItemRow name="User type" labelKey="general:User type">
        <SearchableSelect
          value={user.type || "normal-user"}
          onChange={(v) => updateField("type", v)}
          options={(userOrganization.userTypes?.length > 0
            ? userOrganization.userTypes
            : ["normal-user", "paid-user"]
          ).map((item: string) => ({value: item, label: item}))}
        />
      </AccountItemRow>
    ),
    "Tag": (
      <AccountItemRow name="Tag" labelKey="general:Tag">
        {/* an organization that lists its tags picks from them; the second
              token of "<name>|<zh name>" is the label shown in Chinese */}
        {(userOrganization?.tags?.length ?? 0) > 0 ? (
          <SearchableSelect
            value={user.tag ?? ""}
            onChange={(v) => updateField("tag", v)}
            options={(userOrganization.tags as string[]).map((tag) => {
              const tokens = tag.split("|");
              return {value: tokens[0], label: Setting.getLanguage() !== "zh" ? tokens[0] : tokens[1]};
            })}
          />
        ) : (
          <Input value={user.tag ?? ""} onChange={(e) => updateField("tag", e.target.value)} />
        )}
      </AccountItemRow>
    ),
    "Signup application": (
      <AccountItemRow name="Signup application" labelKey="general:Signup application">
        <SearchableSelect
          value={user.signupApplication ?? ""}
          onChange={(v) => updateField("signupApplication", v)}
          options={applications}
        />
      </AccountItemRow>
    ),
    "Groups": (
      <AccountItemRow name="Groups" labelKey="general:Groups">
        <MultiSelect
          value={user.groups ?? []}
          onChange={(v) => {
            // a user belongs to at most one Physical group
            if (dropExtraPhysicalGroups(v, groupList)) {
              Setting.showMessage("error", i18next.t("general:You can only select one physical group"));
              return;
            }
            updateField("groups", v);
          }}
          options={groups}
        />
      </AccountItemRow>

    ),
    "Country/Region": (
      <AccountItemRow name="Country/Region" labelKey="user:Country/Region">
        <Input value={user.region ?? ""} onChange={(e) => updateField("region", e.target.value)} />
      </AccountItemRow>
    ),
    "Location": (
      <AccountItemRow name="Location" labelKey="user:Location">
        <Input value={user.location ?? ""} onChange={(e) => updateField("location", e.target.value)} />
      </AccountItemRow>
    ),
    "Address": (
      <AccountItemRow name="Address" labelKey="user:Address">
        {affiliation.enabled ? (
          <AffiliationAddressSelect
            value={user.address}
            options={affiliation.addressOptions}
            onChange={(value) => {
              // a new address invalidates the affiliation picked under the old one
              updateFields({address: value, affiliation: "", score: 0});
              affiliation.loadAffiliationOptions(value);
            }}
          />
        ) : (
        // antd edits the two lines of `address` as their own inputs
          <div className="space-y-2">
            {[0, 1].map((line) => (
              <div key={line} className="flex items-center gap-2">
                <span className="w-28 shrink-0 text-sm text-muted-foreground">
                  {`${i18next.t("user:Address line")} ${line + 1}`}
                </span>
                <Input
                  value={user.address?.[line] ?? ""}
                  onChange={(e) => {
                    const next = [...(user.address ?? [])];
                    next[line] = e.target.value;
                    updateField("address", next);
                  }}
                />
              </div>
            ))}
          </div>
        )}
      </AccountItemRow>
    ),
    "Addresses": (
      <AccountItemRow name="Addresses" labelKey="user:Addresses" block>
        <EditableTable
          rows={user.addresses ?? []}
          onChange={(rows) => updateField("addresses", rows)}
          newRow={() => ({tag: "", line1: "", line2: "", city: "", state: "", zipCode: "", region: ""})}
          columns={[
            {
              key: "tag",
              title: i18next.t("general:Tag"),
              width: 130,
              render: (row: any, _i, patch) => (
                <SearchableSelect
                  value={row.tag ?? ""}
                  onChange={(v) => patch({tag: v})}
                  options={ADDRESS_TAGS.map((item) => ({value: item.value, label: i18next.t(item.labelKey)}))}
                />
              ),
            },
            {
              key: "line1",
              title: i18next.t("user:Line 1"),
              width: 160,
              render: (row: any, _i, patch) => (
                <Input value={row.line1 ?? ""} onChange={(e) => patch({line1: e.target.value})} />
              ),
            },
            {
              key: "line2",
              title: i18next.t("user:Line 2"),
              width: 160,
              render: (row: any, _i, patch) => (
                <Input value={row.line2 ?? ""} onChange={(e) => patch({line2: e.target.value})} />
              ),
            },
            {
              key: "city",
              title: i18next.t("user:City"),
              width: 130,
              render: (row: any, _i, patch) => (
                <Input value={row.city ?? ""} onChange={(e) => patch({city: e.target.value})} />
              ),
            },
            {
              key: "state",
              title: i18next.t("general:State"),
              width: 120,
              render: (row: any, _i, patch) => (
                <Input value={row.state ?? ""} onChange={(e) => patch({state: e.target.value})} />
              ),
            },
            {
              key: "zipCode",
              title: i18next.t("user:Zip code"),
              width: 120,
              render: (row: any, _i, patch) => (
                <Input value={row.zipCode ?? ""} onChange={(e) => patch({zipCode: e.target.value})} />
              ),
            },
            {
              key: "region",
              title: i18next.t("provider:Region"),
              width: 170,
              render: (row: any, _i, patch) => (
                <RegionSelect value={row.region ?? ""} onChange={(v) => patch({region: v})} />
              ),
            },
          ]}
        />
      </AccountItemRow>
    ),
    "Affiliation": (
      <AccountItemRow name="Affiliation" labelKey="user:Affiliation">
        <AffiliationField
          enabled={affiliation.enabled}
          value={user.affiliation}
          options={affiliation.affiliationOptions}
          onChange={(name, score) =>
            score === undefined ? updateField("affiliation", name) : updateFields({affiliation: name, score})
          }
        />
      </AccountItemRow>
    ),
    "Title": (
      <AccountItemRow name="Title" labelKey="general:Title">
        <Input value={user.title ?? ""} onChange={(e) => updateField("title", e.target.value)} />
      </AccountItemRow>
    ),
    "Homepage": (
      <AccountItemRow name="Homepage" labelKey="user:Homepage">
        <Input value={user.homepage ?? ""} onChange={(e) => updateField("homepage", e.target.value)} />
      </AccountItemRow>
    ),
    "Bio": (
      <AccountItemRow name="Bio" labelKey="user:Bio">
        <Input value={user.bio ?? ""} onChange={(e) => updateField("bio", e.target.value)} />
      </AccountItemRow>
    ),
    "Gender": (
      <AccountItemRow name="Gender" labelKey="user:Gender">
        <SearchableSelect
          value={user.gender ?? ""}
          onChange={(v) => updateField("gender", v)}
          options={GENDERS.map((item) => ({value: item, label: item}))}
        />
      </AccountItemRow>
    ),
    "Birthday": (
      <AccountItemRow name="Birthday" labelKey="user:Birthday">
        <Input value={user.birthday ?? ""} onChange={(e) => updateField("birthday", e.target.value)} />
      </AccountItemRow>
    ),
    "Education": (
      <AccountItemRow name="Education" labelKey="user:Education">
        <Input value={user.education ?? ""} onChange={(e) => updateField("education", e.target.value)} />
      </AccountItemRow>
    ),
    "ID card type": (
      <AccountItemRow name="ID card type" labelKey="user:ID card type">
        <SearchableSelect
          value={user.idCardType ?? ""}
          onChange={(v) => updateField("idCardType", v)}
          options={ID_CARD_TYPES.map((item) => ({value: item, label: item}))}
        />
      </AccountItemRow>
    ),
    "ID card": (
      <AccountItemRow name="ID card" labelKey="user:ID card">
        <Input value={user.idCard ?? ""} onChange={(e) => updateField("idCard", e.target.value)} />
      </AccountItemRow>
    ),
    "Real name": (
      <AccountItemRow name="Real name" labelKey="application:Real name">
        <Input
          value={user.realName ?? ""}
          placeholder={i18next.t("user:Please enter your real name")}
          onChange={(e) => updateField("realName", e.target.value)}
        />
      </AccountItemRow>
    ),
    "ID card info": (
      <AccountItemRow name="ID card info" labelKey="user:ID card info" block>
        <div className="flex flex-wrap gap-4">
          {ID_CARD_IMAGES.map((entry) => (
            <UserImageField
              key={entry.field}
              imageUrl={user.properties?.[entry.field] ?? ""}
              title={i18next.t(entry.uploadKey)}
              setTitle={i18next.t(entry.setKey)}
              tag={entry.field}
              user={user}
              organization={userOrganization}
              canUpload={mode !== "add"}
              onUploaded={reload}
            />
          ))}
        </div>
      </AccountItemRow>
    ),
    "ID verification": (
      <AccountItemRow name="ID verification" labelKey="user:ID verification">
        <div className="flex items-center gap-2">
          <Button
            // the verification result is written back to the saved user, so it needs the user to exist
            disabled={!!user.isVerified || mode === "add"}
            onClick={verifyIdentification}
          >
            {user.isVerified ? i18next.t("user:Verified") : i18next.t("user:Verify Identity")}
          </Button>
          {user.isVerified ? (
            <Badge variant="success">{i18next.t("user:Identity verified")}</Badge>
          ) : null}
        </div>
      </AccountItemRow>
    ),
    "Language": (
      <AccountItemRow name="Language" labelKey="user:Language">
        <Input value={user.language ?? ""} onChange={(e) => updateField("language", e.target.value)} />
      </AccountItemRow>
    ),
    "Score": (
      <AccountItemRow name="Score" labelKey="user:Score">
        <Input
          type="number"
          value={user.score ?? 0}
          onChange={(e) => updateField("score", Setting.myParseInt(e.target.value))}
        />
      </AccountItemRow>
    ),
    "UID number": (
      <AccountItemRow name="UID number" labelKey="general:UID number">
        <Input value={user.uidNumber ?? ""} onChange={(e) => updateField("uidNumber", e.target.value)} />
      </AccountItemRow>
    ),
    "Ranking": (
      <AccountItemRow name="Ranking" labelKey="user:Ranking">
        <Input
          type="number"
          value={user.ranking ?? 0}
          onChange={(e) => updateField("ranking", Setting.myParseInt(e.target.value))}
        />
      </AccountItemRow>
    ),
    "Karma": (
      <AccountItemRow name="Karma" labelKey="user:Karma">
        <Input
          type="number"
          value={user.karma ?? 0}
          onChange={(e) => updateField("karma", Setting.myParseInt(e.target.value))}
        />
      </AccountItemRow>
    ),
    "Balance credit": (
      <AccountItemRow name="Balance credit" labelKey="organization:Balance credit">
        <Input
          type="number"
          step="0.01"
          value={user.balanceCredit ?? 0}
          onChange={(e) => updateField("balanceCredit", Number(e.target.value))}
        />
      </AccountItemRow>
    ),
    "Properties": (
      <AccountItemRow name="Properties" labelKey="user:Properties" block>
        <EditableTable
          rows={propertyRows}
          onChange={(rows) =>
            updateField(
              "properties",
              Object.fromEntries(rows.filter((row: any) => row.key).map((row: any) => [row.key, row.value])),
            )
          }
          newRow={() => ({key: "", value: ""})}
          reorderable={false}
          columns={[
            {
              key: "key",
              title: i18next.t("general:Name"),
              width: 240,
              render: (row: any, _i, patch) => (
                <Input value={row.key ?? ""} onChange={(e) => patch({key: e.target.value})} />
              ),
            },
            {
              key: "value",
              title: i18next.t("webhook:Value"),
              render: (row: any, _i, patch) => (
                <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
              ),
            },
          ]}
        />
      </AccountItemRow>
    ),
    "Balance": (
      <AccountItemRow name="Balance" labelKey="user:Balance">
        <Input
          type="number"
          step="0.01"
          value={user.balance ?? 0}
          onChange={(e) => updateField("balance", Number(e.target.value))}
        />
      </AccountItemRow>
    ),
    "Balance currency": (
      <AccountItemRow name="Balance currency" labelKey="organization:Balance currency">
        <SearchableSelect
          value={user.balanceCurrency || "USD"}
          onChange={(v) => updateField("balanceCurrency", v)}
          options={(Setting.CurrencyOptions as any[]).map((item) => ({
            value: item.id,
            label: Setting.getCurrencyWithFlag(item.id),
          }))}
        />
      </AccountItemRow>
    ),
    "Cart": (
      <AccountItemRow name="Cart" labelKey="general:Cart" block>
        <CartTable cart={user.cart ?? []} />
      </AccountItemRow>
    ),
    "Transactions": (
      <AccountItemRow name="Transactions" labelKey="general:Transactions" block>
        <TransactionTable transactions={transactions} />
      </AccountItemRow>

    ),
    "Password": (
      <AccountItemRow name="Password" labelKey="general:Password">
        {/* set-password needs an existing user, so in "add" mode the initial
            password is edited directly on the user to be created */}
        {mode === "add" ? (
          <Input
            type="password"
            autoComplete="new-password"
            value={user.password ?? ""}
            placeholder={i18next.t("user:New Password")}
            onChange={(e) => updateField("password", e.target.value)}
          />
        ) : user.name !== userName ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-block">
                <PasswordModal
                  user={user}
                  userName={userName}
                  organization={userOrganization}
                  account={account}
                  disabled={true}
                />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              {i18next.t("user:You have changed the username, please save your change first before modifying the password")}
            </TooltipContent>
          </Tooltip>
        ) : (
          <PasswordModal
            user={user}
            userName={userName}
            organization={userOrganization}
            account={account}
            disabled={!isAdmin && !isSelf}
            onPasswordUpdated={() => updateField("needUpdatePassword", false)}
          />
        )}
      </AccountItemRow>
    ),
    "Need update password": (
      <AccountItemRow name="Need update password" labelKey="user:Need update password">
        <Switch
          checked={!!user.needUpdatePassword}
          onCheckedChange={(v) => updateField("needUpdatePassword", v)}
        />
      </AccountItemRow>
    ),
    "Is admin": (
      <AccountItemRow name="Is admin" labelKey="user:Is admin">
        <Switch
          checked={!!user.isAdmin}
          disabled={!isAdmin}
          onCheckedChange={(v) => updateField("isAdmin", v)}
        />
      </AccountItemRow>
    ),
    "Is forbidden": (
      <AccountItemRow name="Is forbidden" labelKey="user:Is forbidden">
        <Switch
          checked={!!user.isForbidden}
          disabled={!isAdmin}
          onCheckedChange={(v) => updateField("isForbidden", v)}
        />
      </AccountItemRow>
    ),
    "Is deleted": (
      <AccountItemRow name="Is deleted" labelKey="user:Is deleted">
        <Switch
          checked={!!user.isDeleted}
          disabled={!isAdmin}
          onCheckedChange={(v) => updateField("isDeleted", v)}
        />
      </AccountItemRow>
    ),
    "IP whitelist": (
      <AccountItemRow name="IP whitelist" labelKey="general:IP whitelist">
        <Input
          value={user.ipWhitelist ?? ""}
          onChange={(e) => updateField("ipWhitelist", e.target.value)}
        />
      </AccountItemRow>
    ),
    "MFA items": (
      <AccountItemRow name="MFA items" labelKey="general:MFA items" block>
        <EditableTable
          rows={user.mfaItems ?? []}
          onChange={(rows) => updateField("mfaItems", rows)}
          newRow={() => ({name: "Email", rule: "Optional"})}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 220,
              render: (row: any, _i, patch) => (
                <SearchableSelect
                  value={row.name}
                  onChange={(v) => patch({name: v})}
                  options={["Email", "SMS", "TOTP"].map((item) => ({value: item, label: item}))}
                />
              ),
            },
            {
              key: "rule",
              title: i18next.t("application:Rule"),
              width: 220,
              render: (row: any, _i, patch) => (
                <SearchableSelect
                  value={row.rule}
                  onChange={(v) => patch({rule: v})}
                  options={["Optional", "Prompted", "Required"].map((item) => ({
                    value: item,
                    label: i18next.t(`general:${item}`),
                  }))}
                />
              ),
            },
          ]}
        />
      </AccountItemRow>
    ),
    "Managed accounts": (
      <AccountItemRow name="Managed accounts" labelKey="user:Managed accounts" block>
        <EditableTable
          rows={user.managedAccounts ?? []}
          onChange={(rows) => updateField("managedAccounts", rows)}
          newRow={() => ({application: "", username: "", password: "", signinUrl: ""})}
          reorderable={false}
          columns={[
            {
              key: "application",
              title: i18next.t("general:Application"),
              width: 200,
              render: (row: any, _i, patch) => (
                <SearchableSelect
                  value={row.application}
                  onChange={(v) => patch({application: v})}
                  options={applications}
                />
              ),
            },
            {
              key: "username",
              title: i18next.t("signup:Username"),
              width: 180,
              render: (row: any, _i, patch) => (
                <Input value={row.username ?? ""} onChange={(e) => patch({username: e.target.value})} />
              ),
            },
            {
              key: "password",
              title: i18next.t("general:Password"),
              width: 180,
              render: (row: any, _i, patch) => (
                <Input
                  type="password"
                  value={row.password ?? ""}
                  onChange={(e) => patch({password: e.target.value})}
                />
              ),
            },
            {
              key: "signinUrl",
              title: i18next.t("general:Signin URL"),
              render: (row: any, _i, patch) => (
                <Input value={row.signinUrl ?? ""} onChange={(e) => patch({signinUrl: e.target.value})} />
              ),
            },
          ]}
        />
      </AccountItemRow>
    ),
    "MFA accounts": (
      <AccountItemRow name="MFA accounts" labelKey="user:MFA accounts" block>
        <div className="space-y-2">
          {/* the Casdoor Authenticator app takes these over by scanning the QR / opening the link */}
          <div className="flex flex-wrap gap-2">
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="outline" size="sm">{i18next.t("general:QR Code")}</Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto">
                <CasdoorAppQrCode accessToken={accessToken ?? undefined} icon={user.avatar} />
              </PopoverContent>
            </Popover>
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="outline" size="sm">{i18next.t("general:URL")}</Button>
              </PopoverTrigger>
              <PopoverContent className="w-96">
                <CasdoorAppUrl accessToken={accessToken ?? undefined} />
              </PopoverContent>
            </Popover>
          </div>
          <EditableTable
            rows={user.mfaAccounts ?? []}
            onChange={(rows) => updateField("mfaAccounts", rows)}
            newRow={() => ({accountName: "", issuer: "", secretKey: "", origin: ""})}
            columns={[
              {
                key: "accountName",
                title: i18next.t("cert:Account"),
                width: 180,
                render: (row: any, _i, patch) => (
                  <Input value={row.accountName ?? ""} onChange={(e) => patch({accountName: e.target.value})} />
                ),
              },
              {
                key: "issuer",
                title: "Issuer",
                width: 180,
                render: (row: any, _i, patch) => (
                  <Input value={row.issuer ?? ""} onChange={(e) => patch({issuer: e.target.value})} />
                ),
              },
              {
                key: "origin",
                title: i18next.t("general:URL"),
                width: 180,
                render: (row: any, _i, patch) => (
                  <Input value={row.origin ?? ""} onChange={(e) => patch({origin: e.target.value})} />
                ),
              },
              {
                key: "secretKey",
                title: i18next.t("provider:Secret key"),
                render: (row: any, _i, patch) => (
                  <Input
                    type="password"
                    value={row.secretKey ?? ""}
                    onChange={(e) => patch({secretKey: e.target.value})}
                  />
                ),
              },
              {
                key: "logo",
                title: i18next.t("general:Logo"),
                width: 70,
                render: (row: any) => <IssuerLogo issuer={row.issuer} />,
              },
            ]}
          />
        </div>
      </AccountItemRow>
    ),
    "WebAuthn credentials": (
      <AccountItemRow name="WebAuthn credentials" labelKey="user:WebAuthn credentials" block>
        <WebauthnCredentialTable
          table={user.webauthnCredentials ?? []}
          isSelf={isSelf}
          onUpdateTable={(rows) => updateField("webauthnCredentials", rows)}
          refresh={reload}
        />
      </AccountItemRow>
    ),
    "Face ID": (
      <AccountItemRow name="Face ID" labelKey="user:Face IDs" block>
        <FaceIdTable
          table={user.faceIds ?? []}
          account={account}
          onUpdateTable={(rows) => updateField("faceIds", rows)}
        />
      </AccountItemRow>
    ),
    "Last change password time": (
      <AccountItemRow name="Last change password time" labelKey="user:Last change password time">
        <Input value={Setting.getFormattedDate(user.lastChangePasswordTime) ?? ""} disabled />
      </AccountItemRow>
    ),
    "Multi-factor authentication": (
      <AccountItemRow name="Multi-factor authentication" labelKey="mfa:Multi-factor authentication" block>
        <div className="space-y-2">
          <div className="divide-y rounded-lg border">
            {mfaItems.length === 0 ? (
              <div className="p-4 text-center text-sm text-muted-foreground">
                {i18next.t("general:No data")}
              </div>
            ) : (
              mfaItems.map((item: any) => (
                <div key={item.mfaType} className="flex flex-wrap items-center gap-2 p-3">
                  <span className="text-sm font-medium">{item.mfaType}</span>
                  {item.secret ? (
                    <span className="truncate text-xs text-muted-foreground">{item.secret}</span>
                  ) : null}
                  <span className="flex-1" />
                  {item.enabled ? (
                    <>
                      <Badge variant="success">{i18next.t("general:Enabled")}</Badge>
                      {item.isPreferred ? (
                        <Badge>{i18next.t("mfa:preferred")}</Badge>
                      ) : (
                        <Button size="sm" onClick={() => setPreferredMfa(item.mfaType)}>
                          {i18next.t("mfa:Set preferred")}
                        </Button>
                      )}
                    </>
                  ) : (
                    <Badge variant="secondary">{i18next.t("general:Disabled")}</Badge>
                  )}
                  {isSelf ? (
                    <Button variant="outline" size="sm" asChild>
                      <Link to={`/mfa/setup?mfaType=${item.mfaType}`}>{i18next.t("general:Edit")}</Link>
                    </Button>
                  ) : null}
                </div>
              ))
            )}
          </div>
          <div className="flex gap-2">
            {isSelf ? (
              <Button variant="outline" size="sm" asChild>
                <Link to="/mfa/setup">{i18next.t("general:Enable")}</Link>
              </Button>
            ) : null}
            {mfaItems.some((item: any) => item.enabled) ? (
              <ConfirmButton
                variant="destructive"
                size="sm"
                loading={removingMfa}
                description={i18next.t("mfa:Multi-factor authentication")}
                onConfirm={deleteMfa}
              >
                {i18next.t("general:Delete")}
              </ConfirmButton>
            ) : null}
          </div>
        </div>
      </AccountItemRow>

    ),
    "Roles": (
      <AccountItemRow name="Roles" labelKey="general:Roles">
        <div className="flex flex-wrap gap-1">
          {(user.roles ?? []).length === 0 ? (
            <span className="text-sm text-muted-foreground">{i18next.t("general:No data")}</span>
          ) : (
            (user.roles ?? []).map((role: any) => {
              const sourceGroups: string[] = role.sourceGroups ?? [];
              const key = `${role.owner}/${role.name}`;
              if (sourceGroups.length === 0) {
                return (
                  <Badge key={key} variant="secondary">
                    {role.name}
                  </Badge>
                );
              }
              return (
                <Tooltip key={key}>
                  <TooltipTrigger asChild>
                    <Badge variant="outline" className="cursor-default">
                      <Users className="h-3 w-3" />
                      {role.name}
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>{`${i18next.t("general:Groups")}: ${sourceGroups.join(", ")}`}</TooltipContent>
                </Tooltip>
              );
            })
          )}
        </div>
      </AccountItemRow>
    ),
    "Permissions": (
      <AccountItemRow name="Permissions" labelKey="general:Permissions">
        <div className="flex flex-wrap gap-1">
          {(user.permissions ?? []).length === 0 ? (
            <span className="text-sm text-muted-foreground">{i18next.t("general:No data")}</span>
          ) : (
            (user.permissions ?? []).map((permission: any) => {
              const sourceGroups: string[] = permission.sourceGroups ?? [];
              const sourceRoles: string[] = permission.sourceRoles ?? [];
              const key = `${permission.owner}/${permission.name}`;
              if (sourceGroups.length === 0 && sourceRoles.length === 0) {
                return (
                  <Badge key={key} variant="secondary">
                    {permission.name}
                  </Badge>
                );
              }
              const sourceTexts: string[] = [];
              if (sourceRoles.length !== 0) {
                sourceTexts.push(`${i18next.t("general:Roles")}: ${sourceRoles.join(", ")}`);
              }
              if (sourceGroups.length !== 0) {
                sourceTexts.push(`${i18next.t("general:Groups")}: ${sourceGroups.join(", ")}`);
              }
              return (
                <Tooltip key={key}>
                  <TooltipTrigger asChild>
                    <Badge variant="outline" className="cursor-default">
                      {sourceRoles.length !== 0 ? <ShieldCheck className="h-3 w-3" /> : <Users className="h-3 w-3" />}
                      {permission.name}
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>{sourceTexts.join(" | ")}</TooltipContent>
                </Tooltip>
              );
            })
          )}
        </div>
      </AccountItemRow>
    ),
    "Register type": (
      <AccountItemRow name="Register type" labelKey="user:Register type">
        <Input value={user.registerType ?? ""} disabled />
      </AccountItemRow>
    ),
    "Register source": (
      <AccountItemRow name="Register source" labelKey="user:Register source">
        <Input value={user.registerSource ?? ""} disabled />
      </AccountItemRow>
    ),
    // linking and unlinking go through the saved user, so they need the user to exist
    "3rd-party logins": mode === "add" || application === null ? null : (
      <AccountItemRow name="3rd-party logins" labelKey="user:3rd-party logins" block>
        <ThirdPartyLogins
          user={user}
          application={application}
          account={account}
          onUnlinked={reload}
        />
      </AccountItemRow>
    ),
    "Consents": (
      <AccountItemRow name="Consents" labelKey="consent:Consents" block>
        <ConsentTable table={user.applicationScopes ?? []} onUpdateTable={reload} />
      </AccountItemRow>
    ),
  };

  // rows this frontend adds on top of the antd list; no account item governs them
  const extraRows = (
    <>
      <FormRow block labelKey="user:Is verified">
        <Switch checked={!!user.isVerified} onCheckedChange={(v) => updateField("isVerified", v)} />
      </FormRow>
      <FormRow block labelKey="user:Deleted time">
        <Input value={Setting.getFormattedDate(user.deletedTime) ?? ""} disabled />
      </FormRow>
      <FormRow block labelKey="user:Last signin time">
        <Input value={Setting.getFormattedDate(user.lastSigninTime) ?? ""} disabled />
      </FormRow>
      <FormRow block labelKey="user:Last signin IP">
        <Input value={user.lastSigninIp ?? ""} disabled />
      </FormRow>
    </>
  );

  // `organization.accountItems` decides which rows appear, in which order and
  // under which tab, exactly as getUniqueTabs()/getAccountItemsByTab() do in
  // web/src/UserEditPage.js.
  const visibleAccountItems = ((userOrganization?.accountItems ?? []) as any[]).filter((item) =>
    isAccountItemVisible(item, {isAdmin, isSelfOrAdmin, user}),
  );
  const tabNames = Array.from(new Set(visibleAccountItems.map((item: any) => item.tab || ""))).sort((a, b) => {
    // the unnamed tab always comes first
    if (a === "") {
      return -1;
    }
    if (b === "") {
      return 1;
    }
    return a.localeCompare(b);
  });
  const rowsOfTab = (tab: string) =>
    visibleAccountItems
      .filter((item: any) => (item.tab || "") === tab)
      .map((item: any) => <React.Fragment key={item.name}>{rows[item.name]}</React.Fragment>);
  const tabValues = tabNames.map((tab) => tab || DEFAULT_TAB);
  // a stale hash would otherwise leave every tab closed
  const currentTab = tabValues.includes(activeTab) ? activeTab : tabValues[0];

  return (
    <EditPageShell
      // the antd page calls it "My Account" when you are looking at yourself
      title={`${i18next.t(
        mode === "add" ? "user:New User" : isSelf ? "account:My Account" : getModeTitleKey("user:Edit User", mode),
      )} - ${user.displayName || user.name}`}
      mode={mode}
      backTo={self ? "/" : "/users"}
      onSave={save}
      saving={saving}
    >
      {isSelf && user.needUpdatePassword ? (
        <Alert variant="warning" className="mb-4">
          <AlertTitle>{i18next.t("user:You need to update your password")}</AlertTitle>
          <AlertDescription>
            {i18next.t(
              "user:Your password must be updated before you can continue to use your account, please click the \"Modify password...\" button below",
            )}
          </AlertDescription>
        </Alert>
      ) : null}
      <AccountItemsProvider
        organization={userOrganization}
        isAdmin={isAdmin}
        isSelfOrAdmin={isSelfOrAdmin}
        user={user}
      >
        {visibleAccountItems.length === 0 ? (
          // the organization configures no account items (or has not loaded yet);
          // showing every row beats showing an empty form
          <div className={formGridClass}>
            {Object.entries(rows).map(([name, row]) => <React.Fragment key={name}>{row}</React.Fragment>)}
            {extraRows}
          </div>
        ) : tabNames.length <= 1 ? (
          <div className={formGridClass}>
            {rowsOfTab(tabNames[0] ?? "")}
            {extraRows}
          </div>
        ) : (
          <Tabs value={currentTab} onValueChange={selectTab}>
            <TabsList className="mb-2 flex-wrap">
              {tabNames.map((tab) => (
                <TabsTrigger key={tab} value={tab || DEFAULT_TAB}>
                  {tab === "" ? i18next.t("general:Default") : tab}
                </TabsTrigger>
              ))}
            </TabsList>
            {tabNames.map((tab) => (
              <TabsContent key={tab} value={tab || DEFAULT_TAB} className={formGridClass}>
                {rowsOfTab(tab)}
                {tab === tabNames[0] ? extraRows : null}
              </TabsContent>
            ))}
          </Tabs>
        )}
      </AccountItemsProvider>
    </EditPageShell>
  );
}

/** the well-known logo of an MFA issuer, falling back to a generic one */
function IssuerLogo({issuer}: {issuer?: string}) {
  const src = issuer
    ? `${Setting.StaticBaseUrl}/img/social_${issuer.toLowerCase()}.png`
    : `${Setting.StaticBaseUrl}/img/social_default.png`;
  return (
    <img
      className="h-9 w-9 rounded"
      src={src}
      alt={issuer ?? "default"}
      onError={(e) => {
        (e.currentTarget as HTMLImageElement).src = `${Setting.StaticBaseUrl}/img/social_default.png`;
      }}
    />
  );
}
