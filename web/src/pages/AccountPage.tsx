import UserEditPage from "@/pages/UserEditPage";

/** "My account" — the same editor as /users/:org/:name, bound to the signed-in user. */
export default function AccountPage() {
  return <UserEditPage self />;
}
