package core

import (
	"github.com/LeviyLokotb/fedget/internal/config"
	fedgeterrors "github.com/LeviyLokotb/fedget/internal/fedget_errors"
	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"
)

type App struct {
	config  config.Config
	devices []*osprovider.DiskInfo
	os      osprovider.OsProvider
}

func NewApp(conf config.Config, os osprovider.OsProvider) *App {
	return &App{
		config:  conf,
		devices: nil,
		os:      os,
	}
}

func (app *App) MountDisk(d *osprovider.DiskInfo) error {
	if d.IsMounted() {
		return fedgeterrors.DiskAlreadyMountedError{
			Disk: *d,
		}
	}

	// d.MountPointPath = filepath.Join(app.config.MountPath, d.Name)

	err := app.os.MountDisk(d)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) UnmountDisk(d *osprovider.DiskInfo) error {
	if !d.IsMounted() {
		return fedgeterrors.DiskNotMountedError{
			Disk: *d,
		}
	}

	err := app.os.UnmountDisk(d)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) MountAll() []error {
	errs := []error{}
	for _, d := range app.devices {
		err := app.MountDisk(d)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (app *App) GetDevices() ([]*osprovider.DiskInfo, error) {
	devices, err := app.GetValueDevices()
	if err != nil {
		return nil, err
	}

	devicesPtrs := make([]*osprovider.DiskInfo, len(devices))
	for i := range devices {
		devicesPtrs[i] = &devices[i]
	}

	app.devices = devicesPtrs
	return devicesPtrs, nil
}

func (app App) GetValueDevices() ([]osprovider.DiskInfo, error) {
	devices, err := app.os.GetAllDiscsInfo()
	if err != nil {
		return nil, err
	}
	return devices, nil
}
