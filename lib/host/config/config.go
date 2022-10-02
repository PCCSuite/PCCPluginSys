package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type ConfigFile struct {
	PluginsList  string            `json:"plugins_list"`
	Repositories map[string]string `json:"repositories"`
	TempDir      string            `json:"temp_dir"`
	DataDir      string            `json:"data_dir"`
}

var Config ConfigFile

func ReadConfig() {
	raw, err := os.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}
	Config = ConfigFile{}
	err = json.Unmarshal(raw, &Config)
	if err != nil {
		log.Fatal("Error unmarshaling config file: ", err)
	}
	Config.TempDir, err = filepath.Abs(Config.TempDir)
	if err != nil {
		log.Fatal("Error parsing tempdir config: ", err)
	}
	err = os.MkdirAll(Config.TempDir, os.ModeDir)
	if err != nil {
		log.Fatal("Error creating tempdir: ", err)
	}
	Config.DataDir, err = filepath.Abs(Config.DataDir)
	if err != nil {
		log.Fatal("Error parsing datadir config: ", err)
	}
	err = os.MkdirAll(Config.DataDir, os.ModeDir)
	if err != nil {
		log.Fatal("Error creating datadir: ", err)
	}
}

// func copy(src string, dest string) error {
// 	srcFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer srcFile.Close()
// 	destFile, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
// 	defer destFile.Close()
// 	_, err = io.Copy(destFile, srcFile)
// 	return err
// }
