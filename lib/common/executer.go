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
}

type ExecuterExecData struct {
	DataType  DataType        `json:"data_type"`
	Command   ExecuterCommand `json:"command"`
	Args      []string        `json:"args"`
	WorkDir   string          `json:"work_dir"`
	LogFile   string          `json:"log_file"`
	Env       []string        `json:"env"`
	RequestId int             `json:"request_id"`
}

func NewExecuterExec(args []string, workDir, logFile string, env []string, requestId int) ExecuterExecData {
	return ExecuterExecData{
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

type ExecuterEnvData struct {
	DataType  DataType          `json:"data_type"`
	Command   ExecuterCommand   `json:"command"`
	Target    ExecuterEnvTarget `json:"target"`
	Mode      ExecuterEnvMode   `json:"mode"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	RequestId int               `json:"request_id"`
}

func NewExecuterEnv(target ExecuterEnvTarget, mode ExecuterEnvMode, key, value string, requestId int) ExecuterEnvData {
	return ExecuterEnvData{
		DataType:  DataTypeExecuterCommand,
		Command:   ExecuterCommandEnv,
		Target:    target,
		Mode:      mode,
		Key:       key,
		Value:     value,
		RequestId: requestId,
	}
}

type ExecuterStopData struct {
	DataType DataType        `json:"data_type"`
	Command  ExecuterCommand `json:"command"`
	StopId   int             `json:"stop_id"`
}

func NewExecuterStop(stopId int) ExecuterStopData {
	return ExecuterStopData{
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
