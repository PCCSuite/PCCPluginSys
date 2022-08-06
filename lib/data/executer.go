package data

const (
	DataTypeExecuterCommand DataType = "executer_command"
	DataTypeExecuterResult  DataType = "executer_result"
)

type ExecuterCommand string

const (
	ExecuterCommandExec ExecuterCommand = "EXEC"
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
	Env       []string        `json:"env"`
	RequestId int             `json:"request_id"`
}

func NewExecuterExec(args []string, workDir string, env []string, requestId int) ExecuterExecData {
	return ExecuterExecData{
		DataType:  DataTypeExecuterCommand,
		Command:   ExecuterCommandExec,
		Args:      args,
		WorkDir:   workDir,
		Env:       env,
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
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	Request_id int      `json:"request_id"`
}

func NewExecuterResult(statuscode int, stdout string, stderr string, request_id int) ExecuterResultData {
	return ExecuterResultData{
		Data_type:  DataTypeExecuterResult,
		Code:       statuscode,
		Stdout:     stdout,
		Stderr:     stderr,
		Request_id: request_id,
	}
}
