package osprovider

type DiskInfo struct {
	Name           string
	MountPointPath string
	FStype         string
	MemUsedGb      string
	MemTotalGb     string
}

func (d DiskInfo) IsMounted() bool {
	if d.MountPointPath == "" {
		return false
	}
	return true
}
