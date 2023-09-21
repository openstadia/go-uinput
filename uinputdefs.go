package uinput

import "syscall"

// types needed from uinput.h
const (
	UinputMaxNameSize = 80
	UiDevCreate       = 0x5501
	UiDevDestroy      = 0x5502
	// this is for 64 length buffer to store name
	// for another length generate using : (len << 16) | 0x8000552C
	UiGetSysname = 0x8041552c
	UiSetEvBit   = 0x40045564
	UiSetKeyBit  = 0x40045565

	UiSetRelBit = 0x40045566
	UiSetAbsBit = 0x40045567
	BusUsb      = 0x03
)

// input event codes as specified in input-event-codes.h
const (
	EvSyn     = 0x00
	EvKey     = 0x01
	EvRel     = 0x02
	EvAbs     = 0x03
	relX      = 0x0
	relY      = 0x1
	relHWheel = 0x6
	relWheel  = 0x8
	relDial   = 0x7

	AbsX     = 0x00
	AbsY     = 0x01
	AbsZ     = 0x02
	AbsRX    = 0x03
	AbsRY    = 0x04
	AbsRZ    = 0x05
	AbsHat0X = 0x10
	AbsHat0Y = 0x11

	SynReport        = 0
	evMouseBtnLeft   = 0x110
	evMouseBtnRight  = 0x111
	evMouseBtnMiddle = 0x112
	evBtnTouch       = 0x14a
)

const (
	BtnStateReleased = 0
	BtnStatePressed  = 1
	absSize          = 64
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
	Absmax       [absSize]int32
	Absmin       [absSize]int32
	Absfuzz      [absSize]int32
	Absflat      [absSize]int32
}

// translated to go from input.h
type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}
