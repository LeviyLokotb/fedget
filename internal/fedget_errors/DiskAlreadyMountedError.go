package fedgeterrors

import (
	"fmt"

	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"
)

type DiskAlreadyMountedError struct {
	Disk osprovider.DiskInfo
}

func (e DiskAlreadyMountedError) Error() string {
	return fmt.Sprintf("disk %s already mounted at %s", e.Disk.Name, e.Disk.MountPointPath)
}
