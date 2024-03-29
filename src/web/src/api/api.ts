export function ratAPIBaseURL() {
  if (process.env.NODE_ENV === "production") {
    // enables use of relative path when app is embedded in rat server.
    return "";
  }

  return process.env.REACT_APP_RAT_API_BASE_URL;
}
