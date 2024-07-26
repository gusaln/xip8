package xip8

type Hook func(cpu *Cpu)

func (cpu *Cpu) Start() {
	cpu.isPaused = false
}

func (cpu *Cpu) Stop() {
	cpu.isPaused = true
}

// AddBeforeFrameHook adds a hook that will before every cicle of the CPU
func (cpu *Cpu) AddBeforeFrameHook(h Hook) int {
	cpu.beforeFrameHooks = append(cpu.beforeFrameHooks, h)

	return len(cpu.beforeFrameHooks)
}

// AddBeforeCycleHook adds a hook that will before every cicle of the CPU
func (cpu *Cpu) AddBeforeCycleHook(h Hook) int {
	cpu.beforeCycleHooks = append(cpu.beforeCycleHooks, h)

	return len(cpu.beforeCycleHooks)
}

// AddAfterCycleHook adds a hook that will after every cicle of the CPU
func (cpu *Cpu) AddAfterCycleHook(h Hook) int {
	cpu.afterCycleHooks = append(cpu.afterCycleHooks, h)

	return len(cpu.afterCycleHooks)
}

// AddAfterFrameHook adds a hook that will after every cicle of the CPU
func (cpu *Cpu) AddAfterFrameHook(h Hook) int {
	cpu.afterFrameHooks = append(cpu.afterFrameHooks, h)

	return len(cpu.afterFrameHooks)
}

// AddErrorHook adds a hook that will after every cicle of the CPU
func (cpu *Cpu) AddErrorHook(h Hook) int {
	cpu.errorHooks = append(cpu.errorHooks, h)

	return len(cpu.errorHooks)
}

// runBeforeFrameHooks
func (cpu *Cpu) runBeforeFrameHooks() {
	cpu.runHooks(cpu.beforeFrameHooks)
}

// runBeforeCycleHooks
func (cpu *Cpu) runBeforeCycleHooks() {
	cpu.runHooks(cpu.beforeCycleHooks)
}

// runAfterCycleHooks
func (cpu *Cpu) runAfterCycleHooks() {
	cpu.runHooks(cpu.afterCycleHooks)
}

// runAfterFrameHooks
func (cpu *Cpu) runAfterFrameHooks() {
	cpu.runHooks(cpu.afterFrameHooks)
}

// runErrorHooks
func (cpu *Cpu) runErrorHooks() {
	cpu.runHooks(cpu.errorHooks)
}

// runHooks executes the given set of hooks
func (cpu *Cpu) runHooks(hooks []Hook) {
	for _, h := range hooks {
		h(cpu)
	}
}
