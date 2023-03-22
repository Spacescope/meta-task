package utils

import (
	"syscall"

	"github.com/jinzhu/configor"
)

var CNF TomlConfig

type TomlConfig struct {
	Title    string
	MetaTask MetaTask `toml:"meta-task"`
}

type MetaTask struct {
	Addr         string `toml:"listen"`
	DB           string `toml:"db"`
	TaskName     string `toml:"task_name"`
	Lotus        string `toml:"lotus"`
	MQ           string `toml:"mq"`
	ReportServer string `toml:"report_server"`
}

func InitConfFile(file string, cf *TomlConfig) error {
	err := syscall.Access(file, syscall.O_RDONLY)
	if err != nil {
		return err
	}
	err = configor.Load(cf, file)
	if err != nil {
		return err
	}

	return nil
}
