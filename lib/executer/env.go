package executer

import (
	"log"
	"os/exec"
	"strings"
)

func GetSystemEnv() []string {
	machine := getSystemEnvs("Machine")
	user := getSystemEnvs("User")
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

func getSystemEnvs(target string) map[string]string {
	cmd := exec.Command("cmd", "/C", "start", "/B", "/WAIT", "powershell.exe", "-WindowStyle", "Hidden", "-NoProfile", "-NonInteractive", "[System.Environment]::GetEnvironmentVariables('"+target+"').GetEnumerator() | Select-Object -Property Key,Value | ConvertTo-CSV -NoTypeInformation")
	output, err := cmd.Output()
	if err != nil {
		log.Panic("Failed to get Env from Powershell: ", err)
	}
	res := map[string]string{}
	for i, row := range strings.Split(string(output), "\n") {
		if i == 0 {
			continue
		}
		split := strings.SplitN(row, ",", 2)
		if len(split) != 2 {
			break
		}
		key := strings.Trim(split[0], "\"")
		value := strings.Trim(split[1], "\"")
		res[key] = value
	}
	return res
}
