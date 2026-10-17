// Opens and closes the "Don't have an account?" dialog; Escape and the Close button are native to <dialog>.
const dialog = document.getElementById("sf-account-help")
const opener = document.getElementById("sf-account-help-open")

opener.addEventListener("click", () => dialog.showModal())
// The dialog's own padding is zero, so a click whose target is the dialog itself hit the backdrop.
dialog.addEventListener("click", (event) => {
  if (event.target === dialog) dialog.close()
})
dialog.addEventListener("close", () => opener.focus())
