import "./commands";

// Google One Tap (GSI) fails in a headless browser with no Google session and
// rejects an unhandled promise. That is not a web2 failure, so it must not fail
// the run; anything else still does.
Cypress.on("uncaught:exception", (err) => {
  if (/GSI_LOGGER|FedCM|google\.accounts/.test(err.message)) {
    return false;
  }
  return true;
});
