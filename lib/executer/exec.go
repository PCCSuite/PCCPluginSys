package executer

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

type ExecCmd struct {
	requestId int
	logPath   string
	command   *exec.Cmd
}

func Exec(cmddata common.ExecuterExecData) {
	cmd := exec.Command("cmd.exe", append([]string{"/C"}, cmddata.Args...)...)
	cmd.Dir = cmddata.WorkDir
	cmd.Env = append(GetSystemEnv(), cmddata.Env...)
	execcmd := ExecCmd{
		requestId: cmddata.RequestId,
		logPath:   cmddata.LogFile,
		command:   cmd,
	}
	cmds[cmddata.RequestId] = &execcmd
	go execcmd.run()
}

func (c *ExecCmd) run() {
	logFile, err := os.OpenFile(c.logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Print("Failed to open log file")
		send(common.NewExecuterResult(-1, c.requestId))
	}
	defer logFile.Close()
	c.command.Stdout = logFile
	c.command.Stderr = logFile
	if c.command.Dir == "" {
		executable, err := exec.LookPath(c.command.Args[0])
		if err != nil {
			c.command.Stderr.Write([]byte("Failed to find exec file dir"))
			send(common.NewExecuterResult(-1, c.requestId))
			return
		}
		c.command.Dir = filepath.Dir(executable)
	}
	c.command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	err = c.command.Start()
	c.command.Stderr.Write([]byte("EXEC: running " + strings.Join(c.command.Args, " , ") + "\n"))
	if err != nil {
		c.command.Stderr.Write([]byte("Failed to start process"))
		send(common.NewExecuterResult(-1, c.requestId))
		return
	} else {
		c.command.Wait()
		send(common.NewExecuterResult(c.command.ProcessState.ExitCode(), c.requestId))
	}
}

func (c *ExecCmd) stop() {
	err := c.command.Process.Kill()
	if err != nil {
		log.Println("failed to kill process: ", err)
		return
	}
}
