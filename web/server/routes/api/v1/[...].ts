import { createBffHandler } from "@yueli/nuxt-runtime/server";

export default createBffHandler({
  mountPath: "/api/v1",
  resolveTarget({ event }) {
    return identityBffTarget(oidcConfig(event).downstreamBase);
  },
  credential: {
    async resolve({ event }) {
      const personalToken = identityPersonalTokenCredential(getHeader(event, "authorization"));
      if (personalToken) return personalToken;
      let headers = await sessionAuthHeaders(event);
      if (!headers.authorization) {
        // Guest access tokens target the Paste API, not the OIDC web client.
        const audience = String(useRuntimeConfig(event).pasteAudience);
        headers = await guestSessionAuthHeaders(event, audience,
          !["GET", "HEAD", "OPTIONS"].includes(event.method));
      }
      return identityBffCredential(headers);
    },
  },
});
