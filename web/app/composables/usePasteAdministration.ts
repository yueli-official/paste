export function usePasteAdministration() {
 const { user } = useAuth();
 const api = usePasteApi();
 const { data, refresh } = useAsyncData("paste-administration", async () => {
  if (!user.value) return false;
  try { return (await api.getAdministrationSession()).allowed; } catch { return false; }
 }, {watch:[user],default:()=>false});
 return {isAdmin:computed(()=>Boolean(user.value && data.value)),refresh};
}
