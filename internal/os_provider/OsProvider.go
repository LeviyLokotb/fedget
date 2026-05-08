package osprovider

type OsProvider interface {
	GetAllDiscsInfo() ([]DiskInfo, error)
	MountDisk(disk *DiskInfo) error
	UnmountDisk(disk *DiskInfo) error
}
