window.EXCALIDRAW_ASSET_PATH = window.origin;

const App = () => {
  return React.createElement("div", {
    style: { height: "100vh" },
  }, React.createElement(ExcalidrawLib.Excalidraw, {}));
};

const excalidrawWrapper = document.getElementById("app");
if (excalidrawWrapper) {
  const root = ReactDOM.createRoot(excalidrawWrapper);
  root.render(React.createElement(App));
}