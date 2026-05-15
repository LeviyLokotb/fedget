package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LeviyLokotb/fedget/internal/core"
	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type GUI struct {
	app         *core.App
	window      *gtk.Window
	grid        *gtk.Grid
	scrollWin   *gtk.ScrolledWindow
	statusBar   *gtk.Label
	devices     []*osprovider.DiskInfo
	checkboxes  map[int]*gtk.CheckButton
	openButtons map[int]*gtk.Button
	timer       glib.SourceHandle
	cssProvider *gtk.CssProvider
}

func NewGUI(app *core.App) (*GUI, error) {
	dev, err := app.GetDevices()
	if err != nil {
		return nil, err
	}
	gui := &GUI{
		app:         app,
		devices:     dev,
		checkboxes:  make(map[int]*gtk.CheckButton),
		openButtons: make(map[int]*gtk.Button),
	}
	err = gui.loadCSS()

	return gui, err
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

	// Создание скролла и грида
	g.scrollWin, _ = gtk.ScrolledWindowNew(nil, nil)
	g.scrollWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

	g.grid, _ = gtk.GridNew()
	g.grid.SetColumnSpacing(10)
	g.grid.SetRowSpacing(5)
	g.grid.SetColumnHomogeneous(false)

	g.scrollWin.Add(g.grid)
	mainBox.PackStart(g.scrollWin, true, true, 0)

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
	g.refreshDiskGrid()

	g.timer = glib.TimeoutAdd(300, g.autoRefresh)
	win.Connect("destroy", func() {
		if g.timer != 0 {
			glib.SourceRemove(g.timer)
		}
		gtk.MainQuit()
	})

	win.ShowAll()
}

func (g *GUI) loadCSS() error {
	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		return fmt.Errorf("failed to create CSS provider: %v", err)
	}
	g.cssProvider = cssProvider

	// Проверяем разные пути
	pathsToTry := []string{
		"style.css",
		"./style.css",
		filepath.Join(filepath.Dir(os.Args[0]), "style.css"),
		"/etc/fedget/style.css",
		filepath.Join(os.Getenv("HOME"), ".config/fedget/style.css"),
	}

	var loadedCSS bool
	for _, path := range pathsToTry {
		if _, err := os.Stat(path); err == nil {
			err = cssProvider.LoadFromPath(path)
			if err == nil {
				log.Printf("Loaded CSS from: %s", path)
				loadedCSS = true
				break
			}
		}
	}

	// Если файл не найден, загружаем дефолтный CSS
	if !loadedCSS {
		defaultCSS := ""
		err = cssProvider.LoadFromData(defaultCSS)
		if err != nil {
			return fmt.Errorf("failed to load default CSS: %v", err)
		}
		log.Println("Using default CSS theme")
	}

	// Применяем CSS ко всему приложению
	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return fmt.Errorf("failed to get screen: %v", err)
	}
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	return nil
}

func (g *GUI) refreshDiskGrid() {
	// Очистка грида
	g.grid.GetChildren().Foreach(func(item interface{}) {
		g.grid.Remove(item.(gtk.IWidget))
	})

	g.checkboxes = make(map[int]*gtk.CheckButton)
	g.openButtons = make(map[int]*gtk.Button)

	// Заголовки колонок с фиксированной шириной
	headers := []struct {
		title string
		width int
	}{
		{"Select", 60},
		{"Device", 100},
		{"Size", 80},
		{"FS Type", 80},
		{"Status", 100},
		{"Mount Point", 180},
		{"Actions", 80},
	}

	for col, h := range headers {
		label, _ := gtk.LabelNew(h.title)
		label.SetMarkup(fmt.Sprintf("<b>%s</b>", h.title))
		label.SetHAlign(gtk.ALIGN_START)
		label.SetWidthChars(h.width / 10)
		label.SetMarginStart(5)
		label.SetMarginEnd(5)
		g.grid.Attach(label, col, 0, 1, 1)
	}

	// Разделительная линия
	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	g.grid.Attach(separator, 0, 1, len(headers), 1)

	// Строки с дисками
	for i, device := range g.devices {
		row := i + 2 // после заголовка и разделителя

		// Чекбокс (без галочки по умолчанию)
		check, _ := gtk.CheckButtonNew()
		check.SetActive(false) // Всегда снята по умолчанию
		check.SetSensitive(true)
		check.SetHAlign(gtk.ALIGN_CENTER)
		g.grid.Attach(check, 0, row, 1, 1)
		g.checkboxes[i] = check

		// Имя устройства
		nameLabel, _ := gtk.LabelNew(device.Name)
		nameLabel.SetHAlign(gtk.ALIGN_START)
		nameLabel.SetMarginStart(5)
		g.grid.Attach(nameLabel, 1, row, 1, 1)

		// Размер
		sizeLabel, _ := gtk.LabelNew(device.MemTotalGb)
		sizeLabel.SetHAlign(gtk.ALIGN_START)
		sizeLabel.SetMarginStart(5)
		g.grid.Attach(sizeLabel, 2, row, 1, 1)

		// Файловая система
		fsLabel, _ := gtk.LabelNew(device.FStype)
		fsLabel.SetHAlign(gtk.ALIGN_START)
		fsLabel.SetMarginStart(5)
		g.grid.Attach(fsLabel, 3, row, 1, 1)

		// Статус
		statusText := "Unmounted"
		if device.IsMounted() {
			statusText = "Mounted"
		}
		statusLabel, _ := gtk.LabelNew(statusText)
		statusLabel.SetHAlign(gtk.ALIGN_START)
		statusLabel.SetMarginStart(5)
		g.grid.Attach(statusLabel, 4, row, 1, 1)

		// Путь монтирования
		mountPath := "-"
		if device.MountPointPath != "" {
			mountPath = device.MountPointPath
		}
		mountLabel, _ := gtk.LabelNew(mountPath)
		mountLabel.SetHAlign(gtk.ALIGN_START)
		mountLabel.SetMarginStart(5)
		mountLabel.SetEllipsize(2) // Обрезание длинного текста
		g.grid.Attach(mountLabel, 5, row, 1, 1)

		// Кнопка открытия файлового менеджера
		openBtn, _ := gtk.ButtonNewWithLabel("Open")
		openBtn.SetHAlign(gtk.ALIGN_CENTER)
		if device.IsMounted() {
			openBtn.SetSensitive(true)
			deviceCopy := device
			openBtn.Connect("clicked", func() {
				g.openFileManager(deviceCopy)
			})
		} else {
			openBtn.SetSensitive(false)
		}
		g.grid.Attach(openBtn, 6, row, 1, 1)
		g.openButtons[i] = openBtn
	}

	g.grid.ShowAll()
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
	selected := false

	for i, device := range g.devices {
		if check, ok := g.checkboxes[i]; ok && check.GetActive() {
			selected = true
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

	if !selected {
		g.statusBar.SetText("No disks selected")
	} else if mounted == 0 && failed == 0 {
		g.statusBar.SetText("Selected disks are already mounted")
	} else {
		g.statusBar.SetText(fmt.Sprintf("Mounted: %d, Failed: %d", mounted, failed))
	}
	g.refreshDiskGrid()
}

func (g *GUI) mountAll() {
	g.statusBar.SetText("Mounting all disks...")

	errs := g.app.MountAll()

	if len(errs) == 0 {
		g.statusBar.SetText("All disks mounted successfully")
	} else {
		g.statusBar.SetText(fmt.Sprintf("Errors mounting: %d disk(s)", len(errs)))
	}

	g.refreshDiskGrid()
}

func (g *GUI) unmountSelected() {
	unmounted := 0
	failed := 0
	selected := false

	for i, device := range g.devices {
		if check, ok := g.checkboxes[i]; ok && check.GetActive() {
			selected = true
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

	if !selected {
		g.statusBar.SetText("No disks selected")
	} else if unmounted == 0 && failed == 0 {
		g.statusBar.SetText("Selected disks are not mounted")
	} else {
		g.statusBar.SetText(fmt.Sprintf("Unmounted: %d, Failed: %d", unmounted, failed))
	}
	g.refreshDiskGrid()
}

func (g *GUI) refresh() {
	g.statusBar.SetText("Refreshing...")

	devices, err := g.app.GetDevices()
	if err != nil {
		g.statusBar.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	g.devices = devices
	g.refreshDiskGrid()
	g.statusBar.SetText(fmt.Sprintf("Found %d devices", len(devices)))
}

func (g *GUI) autoRefresh() bool {
	devices, err := g.app.GetDevices()
	if err != nil {
		log.Printf("Auto-refresh error: %v", err)
		return true
	}
	if len(g.devices) == len(devices) {
		return true
	}
	for _, d := range devices {
		if strings.TrimSpace(d.Name) == "" {
			return true
		}
	}

	g.devices = devices
	g.refreshDiskGrid()

	return true
}
