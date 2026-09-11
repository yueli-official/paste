export default defineNuxtRouteMiddleware(async () => {
  const setup = await usePasteApi().getSetup();
  if (!setup.claimed) return navigateTo("/setup");
  const { user, refresh } = useAuth();
  if (!user.value) await refresh();
  if (!user.value) return navigateTo("/login");
  try {
    const session = await usePasteApi().getAdministrationSession();
    if (!session.allowed) return navigateTo("/mine");
  } catch {
    return navigateTo("/mine");
  }
});
