package uinput

import "syscall"

const (
	UinputMaxNameSize = 80
)

const (
	UiDevCreate  = 0x5501
	UiDevDestroy = 0x5502
)

const (
	UiSetEvBit  = 0x40045564
	UiSetKeyBit = 0x40045565
	UiSetRelBit = 0x40045566
	UiSetAbsBit = 0x40045567
)

const (
	UiGetSysname = 0x8041552c
)

const (
	BtnStateReleased = 0
	BtnStatePressed  = 1
	AbsSize          = 64
)

// translated to go from input.h
type InputId struct {
	Bustype uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

// translated to go from uinput.h
type UinputUserDev struct {
	Name         [UinputMaxNameSize]byte
	Id           InputId
	FfEffectsMax uint32
	Absmax       [AbsSize]int32
	Absmin       [AbsSize]int32
	Absfuzz      [AbsSize]int32
	Absflat      [AbsSize]int32
}

// translated to go from input.h
type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}
