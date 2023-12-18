export function ratAPIBaseURL() {
  if (process.env.NODE_ENV === "production") {
    // enables use of relative path when app is embedded in rat server.
    return "";
  }

  return process.env.REACT_APP_RAT_API_BASE_URL;
}

export function Spacer({
  width = 0,
  height = 0,
}: {
  width?: number;
  height?: number;
}) {
  let style = {};

  if (width !== 0) {
    style = {
      width: `${width}px`,
      height: `${width}px`,
      float: "left",
    };
  }

  if (height !== 0) {
    style = { marginBottom: `${height}px` };
  }

  return <div style={style}></div>;
}
