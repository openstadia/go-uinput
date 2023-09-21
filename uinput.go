package uinput

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type KeyEvent uint16
type AbsEvent uint16

type DeviceInfo struct {
	Name    string
	Vendor  uint16
	Product uint16
	Version uint16
}

type UinputDevice struct {
	file *os.File
}

func CreateUinputDevice(
	path string,
	info DeviceInfo,
	keyEvents []KeyEvent,
	absEvents []AbsEvent,
) (*UinputDevice, error) {
	err := ValidateDevicePath(path)
	if err != nil {
		return nil, err
	}

	err = ValidateUinputName(info.Name)
	if err != nil {
		return nil, err
	}

	deviceFile, err := CreateDeviceFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual gamepad device: %v", err)
	}

	err = EnableDevice(deviceFile, uintptr(EvKey))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("failed to register virtual gamepad device: %v", err)
	}

	for _, code := range keyEvents {
		err = Ioctl(deviceFile, UiSetKeyBit, uintptr(code))
		if err != nil {
			_ = deviceFile.Close()
			return nil, fmt.Errorf("failed to register key number %d: %v", code, err)
		}
	}

	// register absolute events
	err = EnableDevice(deviceFile, uintptr(EvAbs))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("failed to register absolute event input device: %v", err)
	}

	for _, event := range absEvents {
		err = Ioctl(deviceFile, UiSetAbsBit, uintptr(event))
		if err != nil {
			_ = deviceFile.Close()
			return nil, fmt.Errorf("failed to register absolute event %v: %v", event, err)
		}
	}

	err = CreateUsbDevice(deviceFile,
		UinputUserDev{
			Name: ToUinputName(info.Name),
			Id: InputId{
				Bustype: BusUsb,
				Vendor:  info.Vendor,
				Product: info.Product,
				Version: info.Vendor}})
	if err != nil {
		return nil, err
	}

	return &UinputDevice{
		file: deviceFile,
	}, nil
}

func ToUinputName(name string) [UinputMaxNameSize]byte {
	var fixedSizeName [UinputMaxNameSize]byte
	copy(fixedSizeName[:], name)
	return fixedSizeName
}

func CreateDeviceFile(path string) (*os.File, error) {
	deviceFile, err := os.OpenFile(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, errors.New("could not open device file")
	}
	return deviceFile, err
}

func EnableDevice(deviceFile *os.File, evType uintptr) error {
	err := Ioctl(deviceFile, UiSetEvBit, evType)
	if err != nil {
		defer deviceFile.Close()
		err = ReleaseDevice(deviceFile)
		if err != nil {
			return fmt.Errorf("failed to close device: %v", err)
		}
		return fmt.Errorf("invalid file handle returned from ioctl: %v", err)
	}
	return nil
}

func CreateUsbDevice(deviceFile *os.File, dev UinputUserDev) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, dev)
	if err != nil {
		_ = deviceFile.Close()
		return fmt.Errorf("failed to write user device buffer: %v", err)
	}
	_, err = deviceFile.Write(buf.Bytes())
	if err != nil {
		_ = deviceFile.Close()
		return fmt.Errorf("failed to write uidev struct to device file: %v", err)
	}

	err = Ioctl(deviceFile, UiDevCreate, uintptr(0))
	if err != nil {
		_ = deviceFile.Close()
		return fmt.Errorf("failed to create device: %v", err)
	}

	time.Sleep(time.Millisecond * 200)

	return err
}

func (u *UinputDevice) CloseDevice() error {
	err := ReleaseDevice(u.file)
	if err != nil {
		return fmt.Errorf("failed to close device: %v", err)
	}
	return u.file.Close()
}

func ReleaseDevice(deviceFile *os.File) (err error) {
	return Ioctl(deviceFile, UiDevDestroy, uintptr(0))
}

func fetchSyspath(deviceFile *os.File) (string, error) {
	sysInputDir := "/sys/devices/virtual/input/"
	// 64 for name + 1 for null byte
	path := make([]byte, 65)
	err := Ioctl(deviceFile, UiGetSysname, uintptr(unsafe.Pointer(&path[0])))

	sysInputDir = sysInputDir + string(path)
	return sysInputDir, err
}

func (u *UinputDevice) SendKeyEvent(keyCode uint16, value int32) error {
	return u.SendEvent(EvKey, keyCode, value)
}

func (u *UinputDevice) SendAbsEvent(absCode uint16, value int32) error {
	return u.SendEvent(EvAbs, absCode, value)
}

func (u *UinputDevice) SendRelEvent(relCode uint16, value int32) error {
	return u.SendEvent(EvRel, relCode, value)
}

func (u *UinputDevice) SendSyncEvent() error {
	return u.SendEvent(EvSyn, SynReport, 0)
}

func (u *UinputDevice) SendEvent(eventType uint16, code uint16, value int32) error {
	buf, err := InputEventToBuffer(InputEvent{
		Time:  syscall.Timeval{Sec: 0, Usec: 0},
		Type:  eventType,
		Code:  code,
		Value: value})
	if err != nil {
		return fmt.Errorf("event could not be set: %v", err)
	}
	_, err = u.file.Write(buf)
	if err != nil {
		return fmt.Errorf("writing event to the device file failed: %v", err)
	}
	return nil
}

func InputEventToBuffer(iev InputEvent) (buffer []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, 24))
	err = binary.Write(buf, binary.LittleEndian, iev)
	if err != nil {
		return nil, fmt.Errorf("failed to write input event to buffer: %v", err)
	}
	return buf.Bytes(), nil
}

func Ioctl(deviceFile *os.File, cmd uintptr, ptr uintptr) error {
	_, _, errorCode := syscall.Syscall(syscall.SYS_IOCTL, deviceFile.Fd(), cmd, ptr)
	if errorCode != 0 {
		return errorCode
	}
	return nil
}
