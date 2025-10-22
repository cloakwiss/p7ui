class DragResizer {
  constructor(dragElement, options = {}) {
    this.dragElement = dragElement;
    this.containerElement = options.container;

    this.startX = undefined;
    this.initWidth = undefined;

    this.init();
  }

  init() {
    this.dragElement.addEventListener("mousedown", (e) => this.onMouseDown(e));
  }

  onMouseDown(e) {
    e.preventDefault();
    this.startX = e.clientX;

    // Get current width from CSS variable
    const style = window.getComputedStyle(this.containerElement);
    const width = style.getPropertyValue("--col1-width");
    console.log(width);
    this.initWidth = parseFloat(width);

    this.dragElement.classList.add("dragging");

    document.addEventListener("mousemove", this.onMouseMove);
    document.addEventListener("mouseup", this.onMouseUp);
  }

  onMouseMove = (e) => {
    const offset = e.clientX - this.startX;
    let newWidth = this.initWidth + offset;

    this.containerElement.style.setProperty("--col1-width", `${newWidth}px`);
  };

  onMouseUp = (e) => {
    this.dragElement.classList.remove("dragging");
    document.removeEventListener("mousemove", this.onMouseMove);
    document.removeEventListener("mouseup", this.onMouseUp);
  };

  destroy() {
    this.dragElement.removeEventListener("mousedown", this.onMouseDown);
    document.removeEventListener("mousemove", this.onMouseMove);
    document.removeEventListener("mouseup", this.onMouseUp);
  }
}

const drag = document.getElementById("drag");
const target = document.getElementById("panes-container");

const resize = new DragResizer(drag, {
  container: target,
});

const toggleDetails = document.getElementById("toggle-detail");
toggleDetails.addEventListener("click", (_) => {
  switch (target.dataset.toShowCol1) {
    case "true":
      target.dataset.toShowCol1 = "false";
      break;
    case "false":
      target.dataset.toShowCol1 = "true";
      break;
    default:
      console.error("Contaimenated value");
      target.dataset.toShowCol1 = "false";
      break;
  }
});
