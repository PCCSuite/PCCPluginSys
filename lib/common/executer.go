package common

const (
	DataTypeExecuterCommand DataType = "executer_command"
	DataTypeExecuterResult  DataType = "executer_result"
)

type ExecuterCommand string

const (
	ExecuterCommandExec ExecuterCommand = "EXEC"
	ExecuterCommandEnv  ExecuterCommand = "ENV"
	ExecuterCommandStop ExecuterCommand = "STOP"
)

type ExecuterCommandData struct {
	DataType DataType        `json:"data_type"`
	Command  ExecuterCommand `json:"command"`

	// ExecCommand, EnvCommand
	RequestId int `json:"request_id"`

	// ExecCommand
	Args []string `json:"args"`
	// ExecCommand
	WorkDir string `json:"work_dir"`
	// ExecCommand
	LogFile string `json:"log_file"`
	// ExecCommand
	Env []string `json:"env"`

	// EnvCommand
	Target ExecuterEnvTarget `json:"target"`
	// EnvCommand
	Mode ExecuterEnvMode `json:"mode"`
	// EnvCommand
	Key string `json:"key"`
	// EnvCommand
	Value string `json:"value"`

	// StopCommand
	StopId int `json:"stop_id"`
}

func NewExecuterExec(args []string, workDir, logFile string, env []string, requestId int) ExecuterCommandData {
	return ExecuterCommandData{
		DataType:  DataTypeExecuterCommand,
		Command:   ExecuterCommandExec,
		Args:      args,
		WorkDir:   workDir,
		LogFile:   logFile,
		Env:       env,
		RequestId: requestId,
	}
}

type ExecuterEnvTarget string

const (
	ExecuterEnvTargetMachine ExecuterEnvTarget = "MACHINE"
	ExecuterEnvTargetUser    ExecuterEnvTarget = "USER"
)

type ExecuterEnvMode string

const (
	ExecuterEnvModeSet       ExecuterEnvMode = "SET"
	ExecuterEnvModeAdd       ExecuterEnvMode = "ADD"
	ExecuterEnvModeAddPrefix ExecuterEnvMode = "ADD_PREFIX"
)

func NewExecuterEnv(target ExecuterEnvTarget, mode ExecuterEnvMode, key, value string, requestId int) ExecuterCommandData {
	return ExecuterCommandData{
		DataType:  DataTypeExecuterCommand,
		Command:   ExecuterCommandEnv,
		Target:    target,
		Mode:      mode,
		Key:       key,
		Value:     value,
		RequestId: requestId,
	}
}

func NewExecuterStop(stopId int) ExecuterCommandData {
	return ExecuterCommandData{
		DataType: DataTypeExecuterCommand,
		Command:  ExecuterCommandStop,
		StopId:   stopId,
	}
}

type ExecuterResultData struct {
	Data_type  DataType `json:"data_type"`
	Code       int      `json:"code"`
	Request_id int      `json:"request_id"`
}

func NewExecuterResult(statuscode int, request_id int) ExecuterResultData {
	return ExecuterResultData{
		Data_type:  DataTypeExecuterResult,
		Code:       statuscode,
		Request_id: request_id,
	}
}
