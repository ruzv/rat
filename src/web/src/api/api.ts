export function ratAPIBaseURL() {
  if (process.env.NODE_ENV === "production") {
    // enables use of relative path when app is embedded in rat server.
    return apiAuthority;
  }

  return process.env.REACT_APP_RAT_API_BASE_URL;
}

const apiAuthority = await (async () => {
  if (process.env.NODE_ENV !== "production") {
    return "";
  }

  const response = await fetch(`/api-authority`);
  if (!response.ok) {
    throw new Error(`Failed to fetch api authority: ${response.status}`);
  }

  return (await response.json()).authority;
})();
