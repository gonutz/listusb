package listusb

import (
	"strings"
	"unsafe"

	"github.com/gonutz/w32"
)

func Directories() []string {
	var dirs []string
	driveBits := w32.GetLogicalDrives()
	for i := 0; i < 26; i++ {
		if driveBits&(1<<uint(i)) != 0 {
			drive := string('A'+i) + ":"
			if getBusType(drive) == w32.BusTypeUsb {
				dirs = append(dirs, drive)
			}
		}
	}
	return dirs
}

func getBusType(drive string) uint32 {
	const IOCTL_STORAGE_QUERY_PROPERTY = 0x002D1400

	oldMode := w32.SetErrorMode(w32.SEM_FAILCRITICALERRORS)
	defer w32.SetErrorMode(oldMode)

	h := w32.CreateFile(
		`\\.\`+strings.ToLower(drive),
		0,
		w32.FILE_SHARE_READ|w32.FILE_SHARE_WRITE,
		nil,
		w32.OPEN_EXISTING,
		0,
		0,
	)
	if h != w32.INVALID_HANDLE_VALUE {
		defer w32.CloseHandle(h)
		var buffer [1024]byte
		var device *w32.STORAGE_DEVICE_DESCRIPTOR
		device = (*w32.STORAGE_DEVICE_DESCRIPTOR)(unsafe.Pointer(&buffer[0]))
		device.Size = uint32(len(buffer))
		query := w32.STORAGE_PROPERTY_QUERY{
			PropertyId: w32.StorageDeviceProperty,
			QueryType:  w32.PropertyStandardQuery,
		}
		if _, ok := w32.DeviceIoControl(
			h,
			IOCTL_STORAGE_QUERY_PROPERTY,
			unsafe.Pointer(&query),
			uint32(unsafe.Sizeof(query)),
			unsafe.Pointer(&buffer[0]),
			uint32(len(buffer)),
			nil,
		); ok {
			return device.BusType
		}
	}

	return 0
}
