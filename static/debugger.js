const url = "localhost:9999";

document.addEventListener("alpine:init", () => {
  Alpine.data("dashboard", () => ({
    source: null,

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

      const source = new WebSocket("ws://" + url + "/debugger");
      source.binaryType = "arraybuffer";
      source.addEventListener("open", function (event) {
        console.log("new connection", event)
      })
      source.addEventListener("close", function (event) {
        console.log("connection closed", event)
      })
      source.addEventListener("message", function (event) {
        /** @type {ArrayBuffer} msg */
        const msg = event.data;

        const view = new DataView(msg);

        const instruction = view.getUint16(0);

        component.instruction = instruction.toString(16).padStart(4, "0");
        component.instruction_op = (instruction & 0xf000).toString(16).padStart(4, "0");
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

        for (let i = 0; i < 16; i++) {
          component.registers[i] = view.getUint8(4 + i).toString(16);
          component.stack[i] = view.getInt16(23 + i * 2);
        }

        // console.log(component);
      });

      this.source = source;
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
