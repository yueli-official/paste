// Identity owns guest credentials and the claim exchange. Paste only supplies
// its resource endpoint, resolved from the same runtime binding as its BFF.
export default defineEventHandler((event) => {
  if (getRequestURL(event).pathname !== "/auth/session") return;
  const config = useRuntimeConfig(event);
  config.guestClaimTargets = [{
    audience: String(config.pasteAudience),
    base: String(config.downstreamBase),
    path: "/api/v1/pastes/guest-claims",
  }];
});
