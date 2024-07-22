package xip8

type Hook func(cpu *Cpu)

// AddBeforeHook adds a hook that will before every cicle of the CPU
func (cpu *Cpu) AddBeforeHook(h Hook) int {
	cpu.BeforeHooks = append(cpu.BeforeHooks, h)

	return len(cpu.BeforeHooks)
}

// AddAfterHook adds a hook that will after every cicle of the CPU
func (cpu *Cpu) AddAfterHook(h Hook) int {
	cpu.AfterHooks = append(cpu.AfterHooks, h)

	return len(cpu.AfterHooks)
}

// RunBeforeHooks runs all the hooks
func (cpu *Cpu) RunBeforeHooks() {
	cpu.runHooks(cpu.BeforeHooks)
}

// RunAfterHooks runs all the hooks
func (cpu *Cpu) RunAfterHooks() {
	cpu.runHooks(cpu.AfterHooks)
}

// RunAfterHooks runs all the hooks
func (cpu *Cpu) runHooks(hooks []Hook) {
	for _, h := range hooks {
		h(cpu)
	}
}
