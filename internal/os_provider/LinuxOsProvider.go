package osprovider

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type LinuxOsProvider struct{}

func (sp LinuxOsProvider) GetAllDiscsInfo() ([]DiskInfo, error) {
	// lsblk -J -o NAME,MOUNTPOINT,FSTYPE,SIZE
	out, err := exec.Command("lsblk", "-J", "-o", "NAME,MOUNTPOINT,FSTYPE,SIZE").Output()
	if err != nil {
		return nil, err
	}

	ls := listFromLSBLK{}
	err = json.Unmarshal(out, &ls)
	if err != nil {
		return nil, err
	}

	disks := []diskFromLSBLK{}
	for _, dev := range ls.Blockdevices {
		disks = append(disks, dev.Children...)
	}

	result := make([]DiskInfo, len(disks))
	for i, d := range disks {
		if d.FStype == "" || strings.HasPrefix(d.Name, "loop") {
			continue
		}
		result[i] = d.toDiskInfo()
	}
	return result, nil
}

type listFromLSBLK struct {
	Blockdevices []struct {
		Children []diskFromLSBLK `json:"children"`
	} `json:"blockdevices"`
}

type diskFromLSBLK struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountpoint"`
	FStype     string `json:"fstype"`
	Size       string `json:"size"`
}

func (d diskFromLSBLK) toDiskInfo() DiskInfo {
	return DiskInfo{
		Name:           d.Name,
		MountPointPath: d.MountPoint, // "" if null
		FStype:         d.FStype,
		MemUsedGb:      "",
		MemTotalGb:     d.Size,
	}
}

func (sp LinuxOsProvider) MountDisk(disk *DiskInfo) error {
	diskPath := filepath.Join("/dev", disk.Name)

	// udisksctl стандартно монтирует в /run/media/$USER/имя_диска
	cmd := exec.Command("udisksctl", "mount", "-b", diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("udisksctl mount failed: %v, output: %s", err, string(output))
	}

	// Парсим вывод чтобы узнать точку монтирования
	// "Mounted /dev/sda1 at /run/media/lev/disk."
	outputStr := string(output)
	if idx := strings.Index(outputStr, "at "); idx != -1 {
		mountPoint := strings.TrimSpace(outputStr[idx+3:])
		mountPoint = strings.TrimSuffix(mountPoint, ".")
		disk.MountPointPath = mountPoint
	}

	return nil
}

func (sp LinuxOsProvider) UnmountDisk(disk *DiskInfo) error {
	if !disk.IsMounted() {
		return fmt.Errorf("disk %s is not mounted", disk.Name)
	}

	diskPath := filepath.Join("/dev", disk.Name)

	// Размонтируем через udisksctl
	cmd := exec.Command("udisksctl", "unmount", "-b", diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("udisksctl unmount failed: %v, output: %s", err, string(output))
	}

	// Очищаем путь монтирования в структуре
	disk.MountPointPath = ""

	return nil
}
