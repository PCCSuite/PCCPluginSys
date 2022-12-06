package subp

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
	"golang.org/x/sys/windows"
)

var ExecutablePath string

func StartExecuters() {
	copyBinary()
	go startUserExecuter()
	startAdminExecuter()
}

func copyBinary() {
	executable, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable: ", err)
	}

	ExecutablePath = filepath.Join(config.Config.TempDir, filepath.Base(executable))

	dest, err := os.Create(ExecutablePath)
	if err != nil {
		log.Fatal("Failed to create dest exec file: ", err)
	}

	source, err := os.Open(executable)
	if err != nil {
		log.Fatal("Failed to open exec file: ", err)
	}

	_, err = io.Copy(dest, source)
	if err != nil {
		log.Fatal("Failed to copy exec file: ", err)
	}

	err = source.Close()
	if err != nil {
		log.Fatal("Failed close source exec file: ", err)
	}

	err = dest.Close()
	if err != nil {
		log.Fatal("Failed close dest exec file: ", err)
	}
}

func startUserExecuter() {
	logFile, err := os.Create(filepath.Join(config.Config.TempDir, "executer-user.log"))
	if err != nil {
		log.Fatal("Failed to open exec-user log file: ", err)
	}
	for {
		log.Print("Starting executer-user")
		// start executer-user
		proc := exec.Command(ExecutablePath, "executer-user")
		proc.Stdout = logFile
		proc.Stderr = proc.Stdout
		proc.Run()
		log.Print("Stopped executer-user")
		time.Sleep(2 * time.Second)
	}
}

func startAdminExecuter() {
	verb := "runas"
	args := "executer-admin"

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(ExecutablePath)
	cwdPtr, _ := syscall.UTF16PtrFromString(config.Config.TempDir)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		log.Fatal("Failed to start exec-admin: ", err)
	}
}
