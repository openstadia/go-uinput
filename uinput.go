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

func ValidateDevicePath(path string) error {
	if path == "" {
		return errors.New("device path must not be empty")
	}
	_, err := os.Stat(path)
	return err
}

func ValidateUinputName(name []byte) error {
	if name == nil || len(name) == 0 {
		return errors.New("device name may not be empty")
	}
	if len(name) > uinputMaxNameSize {
		return fmt.Errorf("device name %s is too long (maximum of %d characters allowed)", name, uinputMaxNameSize)
	}
	return nil
}

func ToUinputName(name []byte) (uinputName [uinputMaxNameSize]byte) {
	var fixedSizeName [uinputMaxNameSize]byte
	copy(fixedSizeName[:], name)
	return fixedSizeName
}

func CreateDeviceFile(path string) (fd *os.File, err error) {
	deviceFile, err := os.OpenFile(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, errors.New("could not open device file")
	}
	return deviceFile, err
}

func RegisterDevice(deviceFile *os.File, evType uintptr) error {
	err := Ioctl(deviceFile, uiSetEvBit, evType)
	if err != nil {
		defer deviceFile.Close()
		err = releaseDevice(deviceFile)
		if err != nil {
			return fmt.Errorf("failed to close device: %v", err)
		}
		return fmt.Errorf("invalid file handle returned from ioctl: %v", err)
	}
	return nil
}

func CreateUsbDevice(deviceFile *os.File, dev UinputUserDev) (fd *os.File, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, dev)
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("failed to write user device buffer: %v", err)
	}
	_, err = deviceFile.Write(buf.Bytes())
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("failed to write uidev struct to device file: %v", err)
	}

	err = Ioctl(deviceFile, uiDevCreate, uintptr(0))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("failed to create device: %v", err)
	}

	time.Sleep(time.Millisecond * 200)

	return deviceFile, err
}

func CloseDevice(deviceFile *os.File) (err error) {
	err = releaseDevice(deviceFile)
	if err != nil {
		return fmt.Errorf("failed to close device: %v", err)
	}
	return deviceFile.Close()
}

func releaseDevice(deviceFile *os.File) (err error) {
	return Ioctl(deviceFile, uiDevDestroy, uintptr(0))
}

func fetchSyspath(deviceFile *os.File) (string, error) {
	sysInputDir := "/sys/devices/virtual/input/"
	// 64 for name + 1 for null byte
	path := make([]byte, 65)
	err := Ioctl(deviceFile, uiGetSysname, uintptr(unsafe.Pointer(&path[0])))

	sysInputDir = sysInputDir + string(path)
	return sysInputDir, err
}

func SendBtnEvent(deviceFile *os.File, keys []int, btnState int) (err error) {
	for _, key := range keys {
		buf, err := InputEventToBuffer(InputEvent{
			Time:  syscall.Timeval{Sec: 0, Usec: 0},
			Type:  EvKey,
			Code:  uint16(key),
			Value: int32(btnState)})
		if err != nil {
			return fmt.Errorf("key event could not be set: %v", err)
		}
		_, err = deviceFile.Write(buf)
		if err != nil {
			return fmt.Errorf("writing btnEvent structure to the device file failed: %v", err)
		}
	}
	return nil
}

func SyncEvents(deviceFile *os.File) (err error) {
	buf, err := InputEventToBuffer(InputEvent{
		Time:  syscall.Timeval{Sec: 0, Usec: 0},
		Type:  evSyn,
		Code:  uint16(synReport),
		Value: 0})
	if err != nil {
		return fmt.Errorf("writing sync event failed: %v", err)
	}
	_, err = deviceFile.Write(buf)
	return err
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
