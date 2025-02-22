window.EXCALIDRAW_ASSET_PATH = window.origin;

function App() {
  return React.createElement("div", {
    style: { height: "calc(100% - 44px)" },
  }, React.createElement(ExcalidrawLib.Excalidraw, {}));
}

let excalidrawWrapper
if (excalidrawWrapper = document.getElementById("app")) {
  const root = ReactDOM.createRoot(excalidrawWrapper);
  root.render(React.createElement(App));
}