  document.getElementById("hookdll_pick").addEventListener("click", async () => {
    // call the Go backend
    const path = await window.PickHookdll();
    if (path) {
      HookdllPathBox.value = path;
    }
  });
  
  document.addEventListener("DOMContentLoaded", () => {
  const TargetPathBox = document.getElementById("target_path");
  document.getElementById("target_pick").addEventListener("click", async () => {
    // call the Go backend
    const path = await window.PickTarget();
    if (path) {
      TargetPathBox.value = path;
    }
  });