const cookieSecure =
  process.env.NUXT_COOKIE_SECURE === undefined
    ? process.env.NODE_ENV === "production"
    : process.env.NUXT_COOKIE_SECURE === "true";

export default defineNuxtConfig({
  devtools: { enabled: process.env.NUXT_DEVTOOLS === "true" },
  extends: ["@yueli/identity-nuxt"],
  modules: ["@nuxt/ui", "@yueli/ui", "@yueli/nuxt-runtime"],
  yueliRuntime: {
    defaultTarget: "platform",
    targets: {
      identity: {
        path: "/identity-api",
        ssr: { cookies: [], headers: ["accept-language", "user-agent"] },
      },
      platform: {
        path: "/",
        ssr: {
          cookies: [],
          headers: ["accept-language", "user-agent"],
        },
      },
    },
  },
  icon: {
    serverBundle: { collections: ["tabler"] },
    clientBundle: {
      scan: {
        globInclude: [
          "app/**/*.{vue,ts}",
          "node_modules/@yueli/**/*.{vue,js,mjs,ts}",
        ],
        globExclude: ["test/**", "tests/**", ".*"],
      },
      sizeLimitKb: 256,
    },
  },
  css: ["~/assets/css/main.css"],
  app: {
    head: {
      htmlAttrs: { lang: "zh-CN" },
      titleTemplate: "%s",
      meta: [
        {
          name: "theme-color",
          content: "#f8f7f3",
          media: "(prefers-color-scheme: light)",
        },
        {
          name: "theme-color",
          content: "#111820",
          media: "(prefers-color-scheme: dark)",
        },
        { property: "og:site_name", content: "代码片段" },
        { property: "og:type", content: "website" },
        { name: "twitter:card", content: "summary" },
      ],
    },
  },
  fonts: {
    providers: {
      google: false,
      googleicons: false,
      bunny: false,
      fontshare: false,
      fontsource: false,
    },
  },
  nitro: {
    esbuild: {
      options: {
        target: "es2022",
        exclude: /node_modules(?!.*(?:@yueli\+|@yueli[\\/]))/,
      },
    },
  },
  buildDir: process.env.NUXT_BUILD_DIR || ".nuxt",
  devServer: {
    host: "127.0.0.1",
    port: Number(process.env.NUXT_DEV_PORT || "3010"),
  },
  runtimeConfig: {
    downstreamBase:
      process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8091",
    identityBase: process.env.NUXT_IDENTITY_BASE || "http://127.0.0.1:8081",
    cookieSecure,
    authCookieSecure: cookieSecure,
    sealSecret:
      process.env.NUXT_SEAL_SECRET ||
      "dev-paste-seal-secret-change-me-0123456789abcdef",
    public: {
      oidcIssuer: process.env.NUXT_PUBLIC_OIDC_ISSUER || "http://localhost:8081",
      oidcClientId: process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "paste-yueli-web",
      oidcRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_REDIRECT_URI ||
        "http://localhost:3010/auth/callback",
      oidcPostLogoutRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_POST_LOGOUT_REDIRECT_URI ||
        "http://localhost:3010/",
      oidcScopes:
        process.env.NUXT_PUBLIC_OIDC_SCOPES ||
        "openid profile email roles offline_access",
      accountUrl:
        process.env.NUXT_PUBLIC_ACCOUNT_URL || "http://localhost:3000",
      operatorSubs: process.env.NUXT_PUBLIC_OPERATOR_SUBS || "",
      publicBase:
        process.env.NUXT_PUBLIC_PASTE_BASE || "http://localhost:3010",
      siteBrand: "月离 Paste",
    },
  },
});
