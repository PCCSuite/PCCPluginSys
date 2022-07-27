package subp

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

func StartExecuters() {

}

func startUserExecuter() {
	executable, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.Create(filepath.Join(filepath.Dir(executable), "executer-user.log"))
	if err != nil {
		log.Fatal(err)
	}

	// start executer-user
	proc := exec.Command(executable, "executer-user")
	proc.Stdout = logFile
	proc.Stderr = proc.Stdout
	proc.Run()
}

func startAdminExecuter() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := "executer-admin"

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		log.Fatal(err)
	}
}
