package fedgeterrors

import (
	"fmt"

	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"
)

type DiskNotMountedError struct {
	Disk osprovider.DiskInfo
}

func (e DiskNotMountedError) Error() string {
	return fmt.Sprintf("disk %s is not mounted", e.Disk.Name)
}
