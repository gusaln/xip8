const url = "localhost:9999";

document.addEventListener("alpine:init", () => {
  Alpine.data("dashboard", () => ({
    source: null,

    screenWidth: 64,
    screenHeight: 32,

    // const cpu = {
    instruction: "0",
    instruction_op: "",
    instruction_x: "",
    instruction_y: "",
    instruction_nnn: "",
    instruction_kk: "",
    instruction_n: "",
    pc: 0,
    registers: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    i: 0,
    stackPointer: 0,
    stack: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    delay: 0,
    timer: 0,
    // };

    init() {
      const component = this;

      const debugEvent = new WebSocket("ws://" + url + "/debugger");
      debugEvent.binaryType = "arraybuffer";
      debugEvent.addEventListener("open", function (event) {
        console.log("new debugger connection", event);
      });
      debugEvent.addEventListener("close", function (event) {
        console.log("debugger connection closed", event);
      });
      debugEvent.addEventListener("message", function (event) {
        /** @type {ArrayBuffer} msg */
        const msg = event.data;

        const view = new DataView(msg);

        const instruction = view.getUint16(0);

        component.instruction = instruction.toString(16).padStart(4, "0");
        component.instruction_op = (instruction & 0xf000)
          .toString(16)
          .padStart(4, "0");
        component.instruction_x = (instruction & 0x0f00) >> 8;
        component.instruction_y = (instruction & 0x00f0) >> 4;
        component.instruction_nnn = (instruction & 0x0fff) >> 0;
        component.instruction_kk = (instruction & 0x00ff) >> 0;
        component.instruction_n = (instruction & 0x000f) >> 0;

        component.pc = view.getUint16(2);
        component.registers = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
        component.i = view.getUint16(20);
        component.stackPointer = view.getUint8(22);
        component.stack = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
        component.delay = view.getUint8(54);
        component.timer = view.getUint8(55);
        // Why is it jumping one byte? IDK
        component.screenWidth = view.getUint8(57);
        component.screenHeight = view.getUint8(58);

        for (let i = 0; i < 16; i++) {
          component.registers[i] = view.getUint8(4 + i).toString(16);
          component.stack[i] = view.getInt16(23 + i * 2);
        }

        // console.log(view.getUint8(56), view.getUint8(57), view.getUint8(58));
      });

      /** @type {HTMLCanvasElement} canvasEl */
      const canvasEl = document.getElementById("screen");
      const canvasCtx = canvasEl.getContext("2d");

      // buffer canvas
      const canvasBuffer = document.createElement("canvas");
      canvasBuffer.width = canvasEl.width;
      canvasBuffer.height = canvasEl.height;
      const bufferCtx = canvasBuffer.getContext("2d");

      const M = { x: 0, y: 0 };
      const SCALE = 10;
      const ON_COLOR = "white";
      const OFF_COLOR = "black";
      canvasCtx.fillStyle = "gray";
      canvasCtx.fillRect(0, 0, canvasEl.width, canvasEl.height);
      // canvasCtx.fillStyle = "red";
      // canvasCtx.fillRect(M.x * SCALE + 0, M.y * SCALE + 0, SCALE, SCALE);
      // canvasCtx.fillRect(M.x * SCALE + 63, M.y * SCALE + 0, SCALE, SCALE);
      // canvasCtx.fillRect(M.x * SCALE + 0, M.y * SCALE + 31, SCALE, SCALE);
      // canvasCtx.fillRect(M.x * SCALE + 63, M.y * SCALE + 31, SCALE, SCALE);

      const displayWs = new WebSocket("ws://" + url + "/display");
      displayWs.binaryType = "arraybuffer";
      displayWs.addEventListener("open", function (event) {
        console.log("new display connection", event);
      });
      displayWs.addEventListener("close", function (event) {
        console.log("display connection closed", event);
      });

      // const screen = Array.from(Array(20 * 10));
      // const msPerFrame = 1000 / 2;
      // let lastFrameAt = 0;

      // function draw() {
      //   for (y = 0; y < component.screenHeight; y++) {
      //     for (x = 0; x < component.screenWidth; x++) {
      //       screenT = y * component.screenHeight + x;

      //       bufferCtx.fillStyle = screen[screenT] ? ON_COLOR : OFF_COLOR;
      //       bufferCtx.fillRect(M.x + x * SCALE, M.y + y * SCALE, SCALE, SCALE);
      //     }
      //   }

      //   canvasCtx.drawImage(canvasBuffer, 0, 0);
      // }

      // function render(ts) {
      //   draw();
      //   setTimeout(
      //     () => requestAnimationFrame(render),
      //     Math.min(msPerFrame, ts - lastFrameAt)
      //   );

      //   lastFrameAt = ts;
      // }
      // requestAnimationFrame(render);

      let screenT = 0,
        screenSize = 0,
        x = 0,
        y = 0;
      displayWs.addEventListener("message", function (event) {
        /** @type {ArrayBuffer} msg */
        const msg = event.data;
        const view = new DataView(msg);
        // canvasCtx.fillStyle = OFF_COLOR;
        // canvasCtx.fillRect(
        //   0,
        //   0,
        //   component.screenWidth * SCALE,
        //   component.screenHeight * SCALE
        // );

        screenSize = component.screenWidth * component.screenHeight;
        // console.log(
        //   "frame :: byteLength=%d screenSize=%d byteSizeOfScreen=%f",
        //   view.byteLength,
        //   screenSize,
        //   screenSize / 8
        // );

        screenT = x = y = 0;
        for (
          let index = 0;
          index < view.byteLength && screenT < screenSize;
          index++
        ) {
          const byte = view.getUint8(index);

          // Works
          for (let bit = 7; bit >= 0 && screenT < screenSize; bit--) {
            y = Math.floor(screenT / component.screenWidth);
            x = screenT % component.screenWidth;
            bufferCtx.fillStyle = byte & (1 << bit) ? ON_COLOR : OFF_COLOR;
            bufferCtx.fillRect(M.x + x * SCALE, M.y + y * SCALE, SCALE, SCALE);
            screenT++;
          }
        }
        // draw();
        canvasCtx.drawImage(canvasBuffer, 0, 0);
      });

      this.source = debugEvent;
    },
  }));
});

document.getElementById("start").addEventListener("submit", (event) => {
  event.preventDefault();

  fetch("http://" + url + "/start", {
    method: "post",
  }).then((res) => console.log(res));
});

document.getElementById("stop").addEventListener("submit", (event) => {
  event.preventDefault();

  fetch("http://" + url + "/stop", {
    method: "post",
  }).then((res) => console.log(res));
});

document.getElementById("step").addEventListener("submit", (event) => {
  event.preventDefault();

  fetch("http://" + url + "/step", {
    method: "post",
  }).then((res) => console.log(res));
});

document.getElementById("reset").addEventListener("submit", (event) => {
  event.preventDefault();

  fetch("http://" + url + "/reset", {
    method: "post",
  }).then((res) => console.log(res));
});
