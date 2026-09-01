// MFA type identifiers as the Casdoor backend uses them.
export const EmailMfaType = "email";
export const SmsMfaType = "sms";
export const TotpMfaType = "app";
export const RadiusMfaType = "radius";
export const PushMfaType = "push";
export const RecoveryMfaType = "recovery";

// Answers the login API can return instead of a session.
export const NextMfa = "NextMfa";
export const RequiredMfa = "RequiredMfa";

// `method` values of the MFA forms, mirroring web/src/auth/mfa/MfaVerifyForm.js.
export const mfaAuth = "mfaAuth";
export const mfaSetup = "mfaSetup";
