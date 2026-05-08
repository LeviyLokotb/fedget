package ui

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/LeviyLokotb/fedget/internal/core"
	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"

	"github.com/gotk3/gotk3/gtk"
)

type GUI struct {
	app         *core.App
	window      *gtk.Window
	listBox     *gtk.ListBox
	statusBar   *gtk.Label
	devices     []*osprovider.DiskInfo
	checkboxes  map[int]*gtk.CheckButton
	openButtons map[int]*gtk.Button
}

func NewGUI(app *core.App) (*GUI, error) {
	dev, err := app.GetDevices()
	if err != nil {
		return nil, err
	}
	return &GUI{
		app:         app,
		devices:     dev,
		checkboxes:  make(map[int]*gtk.CheckButton),
		openButtons: make(map[int]*gtk.Button),
	}, nil
}

func (g *GUI) Show() {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("FedGet - Disk Mounter")
	win.SetDefaultSize(900, 600)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	g.window = win

	mainBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
	mainBox.SetMarginTop(10)
	mainBox.SetMarginBottom(10)
	mainBox.SetMarginStart(10)
	mainBox.SetMarginEnd(10)

	// Заголовок
	titleLabel, _ := gtk.LabelNew("Available Disks")
	titleLabel.SetMarkup("<b><big>Available Disks</big></b>")
	titleLabel.SetHAlign(gtk.ALIGN_START)
	titleLabel.SetMarginBottom(10)
	mainBox.PackStart(titleLabel, false, false, 0)

	// Создание списка дисков
	scrollWin, _ := gtk.ScrolledWindowNew(nil, nil)
	scrollWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

	g.listBox, _ = gtk.ListBoxNew()
	g.listBox.SetSelectionMode(gtk.SELECTION_NONE)
	scrollWin.Add(g.listBox)
	mainBox.PackStart(scrollWin, true, true, 0)

	// Кнопки
	buttonBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	buttonBox.SetMarginTop(10)

	mountSelectedBtn, _ := gtk.ButtonNewWithLabel("Mount Selected")
	mountSelectedBtn.Connect("clicked", g.mountSelected)
	buttonBox.PackStart(mountSelectedBtn, false, false, 0)

	mountAllBtn, _ := gtk.ButtonNewWithLabel("Mount All")
	mountAllBtn.Connect("clicked", g.mountAll)
	buttonBox.PackStart(mountAllBtn, false, false, 0)

	unmountSelectedBtn, _ := gtk.ButtonNewWithLabel("Unmount Selected")
	unmountSelectedBtn.Connect("clicked", g.unmountSelected)
	buttonBox.PackStart(unmountSelectedBtn, false, false, 0)

	refreshBtn, _ := gtk.ButtonNewWithLabel("Refresh")
	refreshBtn.Connect("clicked", g.refresh)
	buttonBox.PackStart(refreshBtn, false, false, 0)

	mainBox.PackStart(buttonBox, false, false, 0)

	// Статус-бар
	g.statusBar, _ = gtk.LabelNew("Ready")
	g.statusBar.SetHAlign(gtk.ALIGN_START)
	g.statusBar.SetMarginTop(5)
	mainBox.PackStart(g.statusBar, false, false, 0)

	win.Add(mainBox)
	g.refreshDiskList()
	win.ShowAll()
}

func (g *GUI) refreshDiskList() {
	// Очистка списка
	for {
		row := g.listBox.GetRowAtIndex(0)
		if row == nil {
			break
		}
		g.listBox.Remove(row)
	}

	g.checkboxes = make(map[int]*gtk.CheckButton)
	g.openButtons = make(map[int]*gtk.Button)

	// Создание заголовка списка
	headerBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	headers := []string{"Select", "Device", "Size", "FS Type", "Status", "Mount Point", "Actions"}
	for _, h := range headers {
		label, _ := gtk.LabelNew(h)
		label.SetMarkup(fmt.Sprintf("<b>%s</b>", h))
		label.SetHAlign(gtk.ALIGN_START)
		label.SetWidthChars(15)
		headerBox.PackStart(label, false, false, 0)
	}

	headerRow, _ := gtk.ListBoxRowNew()
	headerRow.Add(headerBox)
	headerRow.SetSensitive(false)
	g.listBox.Add(headerRow)

	// Добавление дисков
	for i, device := range g.devices {
		rowBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

		// Чекбокс
		check, _ := gtk.CheckButtonNew()
		check.SetActive(!device.IsMounted())
		check.SetSensitive(true)
		rowBox.PackStart(check, false, false, 0)
		g.checkboxes[i] = check

		// Имя устройства
		nameLabel, _ := gtk.LabelNew(device.Name)
		nameLabel.SetHAlign(gtk.ALIGN_START)
		nameLabel.SetWidthChars(15)
		rowBox.PackStart(nameLabel, false, false, 0)

		// Размер
		sizeLabel, _ := gtk.LabelNew(device.MemTotalGb)
		sizeLabel.SetHAlign(gtk.ALIGN_START)
		sizeLabel.SetWidthChars(15)
		rowBox.PackStart(sizeLabel, false, false, 0)

		// Файловая система
		fsLabel, _ := gtk.LabelNew(device.FStype)
		fsLabel.SetHAlign(gtk.ALIGN_START)
		fsLabel.SetWidthChars(15)
		rowBox.PackStart(fsLabel, false, false, 0)

		// Статус
		statusText := "unmounted"
		if device.IsMounted() {
			statusText = "mounted"
		}
		statusLabel, _ := gtk.LabelNew(statusText)
		statusLabel.SetHAlign(gtk.ALIGN_START)
		statusLabel.SetWidthChars(15)
		rowBox.PackStart(statusLabel, false, false, 0)

		// Путь монтирования
		mountPath := "-"
		if device.MountPointPath != "" {
			mountPath = device.MountPointPath
		}
		mountLabel, _ := gtk.LabelNew(mountPath)
		mountLabel.SetHAlign(gtk.ALIGN_START)
		mountLabel.SetWidthChars(20)
		rowBox.PackStart(mountLabel, false, false, 0)

		// Кнопка открытия файлового менеджера
		openBtn, _ := gtk.ButtonNewWithLabel("Open")
		if device.IsMounted() {
			openBtn.SetSensitive(true)
			deviceCopy := device // захват переменной для замыкания
			openBtn.Connect("clicked", func() {
				g.openFileManager(deviceCopy)
			})
		} else {
			openBtn.SetSensitive(false)
		}
		rowBox.PackStart(openBtn, false, false, 0)
		g.openButtons[i] = openBtn

		row, _ := gtk.ListBoxRowNew()
		row.Add(rowBox)
		g.listBox.Add(row)
	}

	g.listBox.ShowAll()
}

// Открытие файлового менеджера в указанной директории
func (g *GUI) openFileManager(disk *osprovider.DiskInfo) {
	if disk.MountPointPath == "" {
		g.statusBar.SetText("Disk is not mounted")
		return
	}

	// Пытаемся открыть разными файловыми менеджерами
	managers := []string{
		"xdg-open", // системный обработчик по умолчанию
		"nautilus", // GNOME Files
		"nemo",     // Cinnamon/Nemo
		"thunar",   // XFCE
		"dolphin",  // KDE
		"pcmanfm",  // LXDE
		"caja",     // MATE
		"xfe",      // X File Explorer
	}

	var cmd *exec.Cmd
	for _, manager := range managers {
		if _, err := exec.LookPath(manager); err == nil {
			cmd = exec.Command(manager, disk.MountPointPath)
			break
		}
	}

	if cmd == nil {
		g.statusBar.SetText("No file manager found")
		return
	}

	err := cmd.Start()
	if err != nil {
		g.statusBar.SetText(fmt.Sprintf("Error opening file manager: %v", err))
		log.Printf("Failed to open file manager: %v", err)
	} else {
		g.statusBar.SetText(fmt.Sprintf("Opened %s in file manager", disk.MountPointPath))
	}
}

func (g *GUI) mountSelected() {
	mounted := 0
	failed := 0

	for i, device := range g.devices {
		if check, ok := g.checkboxes[i]; ok && check.GetActive() {
			if !device.IsMounted() {
				err := g.app.MountDisk(device)
				if err != nil {
					failed++
					log.Printf("Failed to mount %s: %v", device.Name, err)
				} else {
					mounted++
				}
			}
		}
	}

	g.statusBar.SetText(fmt.Sprintf("Mounted: %d, Failed: %d", mounted, failed))
	g.refreshDiskList()
}

func (g *GUI) mountAll() {
	g.statusBar.SetText("Mounting all disks...")

	errs := g.app.MountAll()

	if len(errs) == 0 {
		g.statusBar.SetText("All disks mounted successfully")
	} else {
		g.statusBar.SetText(fmt.Sprintf("Errors mounting: %d disk(s)", len(errs)))
	}

	g.refreshDiskList()
}

func (g *GUI) unmountSelected() {
	unmounted := 0
	failed := 0

	for i, device := range g.devices {
		if check, ok := g.checkboxes[i]; ok && check.GetActive() {
			if device.IsMounted() {
				err := g.app.UnmountDisk(device)
				if err != nil {
					failed++
					log.Printf("Failed to unmount %s: %v", device.Name, err)
				} else {
					unmounted++
				}
			}
		}
	}

	g.statusBar.SetText(fmt.Sprintf("Unmounted: %d, Failed: %d", unmounted, failed))
	g.refreshDiskList()
}

func (g *GUI) refresh() {
	g.statusBar.SetText("Refreshing...")

	devices, err := g.app.GetDevices()
	if err != nil {
		g.statusBar.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	g.devices = devices
	g.refreshDiskList()
	g.statusBar.SetText(fmt.Sprintf("Found %d devices", len(devices)))
}
