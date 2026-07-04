const query = new URLSearchParams(location.search);
const returnTo = query.get("return_to");
if (returnTo && returnTo.startsWith("/") && !returnTo.startsWith("//")) {
  document.querySelector("#return-to").value = returnTo;
}
document.querySelector("#login-error").hidden = query.get("error") !== "invalid";
