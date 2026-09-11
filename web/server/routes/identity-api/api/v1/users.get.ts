import { createBffHandler } from "@yueli/nuxt-runtime/server";

// Public profile reads use the shared Identity target; Paste owns no user directory.
export default createBffHandler({
  mountPath: "/identity-api/api/v1",
  resolveTarget({ event }) {
    return identityBffTarget(String(useRuntimeConfig(event).identityBase));
  },
});
