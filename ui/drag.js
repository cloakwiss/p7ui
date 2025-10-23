///////////////////////////// Layout ///////////////////////////////////////////
// Drag Resizer
class DragResizer {
  constructor(dragElement, containerElement) {
    this.dragElement = dragElement;
    this.containerElement = containerElement;
    this.startX = null;
    this.initWidth = null;
    this.init();
  }

  init() {
    this.dragElement.addEventListener("mousedown", (e) => this.onMouseDown(e));
  }

  onMouseDown(e) {
    e.preventDefault();
    this.startX = e.clientX;

    const style = window.getComputedStyle(this.containerElement);
    const width = style.getPropertyValue("--console-width");
    this.initWidth = parseFloat(width) || 480;

    document.addEventListener("mousemove", this.onMouseMove);
    document.addEventListener("mouseup", this.onMouseUp);
  }

  onMouseMove = (e) => {
    const offset = e.clientX - this.startX;
    const newWidth = this.initWidth + offset;
    this.containerElement.style.setProperty("--console-width", `${newWidth}px`);
  };

  onMouseUp = () => {
    document.removeEventListener("mousemove", this.onMouseMove);
    document.removeEventListener("mouseup", this.onMouseUp);
  };
}

// Initialize
const contentArea = document.getElementById("content-area");
const resizeHandle = document.getElementById("resize-handle");
const toggleButton = document.getElementById("toggle-detail");

new DragResizer(resizeHandle, contentArea);

// toggleButton.addEventListener("click", () => {
//   const isShown = contentArea.dataset.showConsole === "true";
//   contentArea.dataset.showConsole = isShown ? "false" : "true";
//   toggleButton.textContent = isShown ? "Show Details" : "Hide Details";
// });
