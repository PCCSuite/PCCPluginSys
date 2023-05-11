package executer

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"golang.org/x/sys/windows/registry"
)

type ExecCmd struct {
	requestId int
	logPath   string
	command   *exec.Cmd
}

func Exec(cmddata common.ExecuterCommandData) {
	cmd := exec.Command("cmd.exe", append([]string{"/C"}, cmddata.Args...)...)
	cmd.Dir = cmddata.WorkDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, GetSystemEnv()...)
	cmd.Env = append(cmd.Env, cmddata.Env...)
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
		log.Print("Failed to open log file: ", err)
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

func GetSystemEnv() []string {
	machine := getSystemEnvs(true)
	user := getSystemEnvs(false)
	for k := range user {
		if k == "Path" {
			machine[k] = machine[k] + ";" + user[k]
		} else {
			machine[k] = user[k]
		}
	}
	var res []string = make([]string, 0)
	for k, v := range machine {
		res = append(res, k+"="+v)
	}
	return res
}

func getSystemEnvs(machine bool) map[string]string {
	res := map[string]string{}
	var key registry.Key
	var err error
	if machine {
		key, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.READ)
	} else {
		key, err = registry.OpenKey(registry.CURRENT_USER, "Environment", registry.READ)
	}
	if err != nil {
		log.Panicln("Failed to open registry key: ", err)
	}
	values, err := key.ReadValueNames(0)
	if err != nil {
		log.Panicln("Failed to read registry values list: ", err)
	}
	for _, v := range values {
		data, valType, err := key.GetStringValue(v)
		if err != nil {
			log.Panicln("Failed to get value: ", err)
		}
		if valType == registry.EXPAND_SZ {
			data, err = registry.ExpandString(data)
			if err != nil {
				log.Panicln("Failed to expand value: ", err)
			}
		}
		res[v] = data
	}
	return res
}
