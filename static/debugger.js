document.addEventListener("alpine:init", () => {
  Alpine.data("dashboard", () => ({
    source: null,

    // const cpu = {
    instruction: "",
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

      this.source = new EventSource("http://localhost:9999/events");
      this.source.addEventListener("message", function (event) {
        /** @type {string} msg */
        const msg = event.data;

        component.instruction = msg.charCodeAt(0).toString(16);
        component.instruction_op = msg.charCodeAt(0) & 0xf000;
        component.instruction_x = (msg.charCodeAt(0) & 0x0f00) >> 8;
        component.instruction_y = (msg.charCodeAt(0) & 0x00f0) >> 4;
        component.instruction_nnn = (msg.charCodeAt(0) & 0x0fff) >> 0;
        component.instruction_kk = (msg.charCodeAt(0) & 0x00ff) >> 0;
        component.instruction_n = (msg.charCodeAt(0) & 0x000f) >> 0;

        component.pc = msg.charCodeAt(1);
        component.registers = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
        component.i = msg.charCodeAt(18);
        component.stackPointer = msg.charCodeAt(19);
        component.stack = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
        component.delay = msg.charCodeAt(36);
        component.timer = msg.charCodeAt(37);

        for (let i = 0; i < 16; i++) {
          component.registers[i] = msg.charCodeAt(2 + i).toString(16);
          component.stack[i] = msg.charCodeAt(20 + i);
        }

        console.log(msg.length, cpu);
      });
    },
  }));
});

const container = document.getElementById("sse-data");
const eventSource = new EventSource("http://localhost:9999/events");
eventSource.addEventListener("message", function (event) {
  /** @type {string} msg */
  const msg = event.data;

  const cpu = {
    instruction: msg.charCodeAt(0).toString(16),
    pc: msg.charCodeAt(1),
    registers: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    i: msg.charCodeAt(18),
    stackPointer: msg.charCodeAt(19),
    stack: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    delay: msg.charCodeAt(36),
    timer: msg.charCodeAt(37),
  };

  for (let i = 0; i < 16; i++) {
    cpu.registers[i] = msg.charCodeAt(2 + i).toString(16);
    cpu.stack[i] = msg.charCodeAt(20 + i);
  }

  console.log(msg.length, cpu);

  container.innerText = JSON.stringify(cpu, null, 2);
});

document.getElementById("start").addEventListener("submit", (event) => {
  event.preventDefault();

  fetch("http://localhost:9999/start", {
    method: "post",
  }).then((res) => console.log(res));
});

document.getElementById("stop").addEventListener("submit", (event) => {
  event.preventDefault();

  fetch("http://localhost:9999/stop", {
    method: "post",
  }).then((res) => console.log(res));
});
