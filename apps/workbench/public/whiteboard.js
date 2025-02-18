window.EXCALIDRAW_ASSET_PATH = window.origin;

function App() {
  return React.createElement("div", {
    style: { height: "100%" },
  }, React.createElement(ExcalidrawLib.Excalidraw, {}));
}

const excalidrawWrapper = document.getElementById("app");
if (excalidrawWrapper) {
  const root = ReactDOM.createRoot(excalidrawWrapper);
  root.render(React.createElement(App));
}