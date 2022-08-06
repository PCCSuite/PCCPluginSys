package executer

import (
	"bytes"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

type ExecCmd struct {
	requestId int
	command   *exec.Cmd
}

func Exec(cmddata data.ExecuterExecData) {
	cmd := exec.Command(cmddata.Args[0], cmddata.Args[1:]...)
	cmd.Dir = cmddata.WorkDir
	cmd.Env = append(cmd.Env, cmddata.Env...)
	execcmd := ExecCmd{
		requestId: cmddata.RequestId,
		command:   cmd,
	}
	cmds[cmddata.RequestId] = &execcmd
	go execcmd.run()
}

func (c *ExecCmd) run() {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.command.Stdout = &stdout
	c.command.Stderr = &stderr
	if c.command.Dir == "" {
		executable, err := exec.LookPath(c.command.Args[0])
		if err != nil {
			send(data.NewExecuterResult(-1, "", "Failed to find exec file dir", c.requestId))
		}
		c.command.Dir = filepath.Dir(executable)
	}
	err := c.command.Start()
	if err != nil {
		send(data.NewExecuterResult(-1, "", "Failed to start process", c.requestId))
	} else {
		c.command.Wait()
		send(data.NewExecuterResult(c.command.ProcessState.ExitCode(), stdout.String(), stderr.String(), c.requestId))
	}
}

func (c *ExecCmd) stop() {
	err := c.command.Process.Kill()
	if err != nil {
		log.Println("failed to kill process: ", err)
		return
	}
}
