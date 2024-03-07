package power

type PowerOnInterface interface {
	PowerOn() (result string, err error)
}

type PowerOffInterface interface {
	PowerOff() (result string, err error)
}

type PowerResetInterface interface {
	PowerReset() (result string, err error)
}

type PowerSoftInterface interface {
	PowerSoft() (result string, err error)
}

type PowerCycleInterface interface {
	PowerCycle() (result string, err error)
}

type PowerStatusInterface interface {
	PowerStatus() (result string, err error)
}
