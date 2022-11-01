package config

import (
	"syscall"

	"github.com/jinzhu/configor"
)

type CFG struct {
	Storage     *Storage     `toml:"storage"`
	Lotus       *Lotus       `toml:"lotus_cluster"`
	ChainNotify *ChainNotify `toml:"chain_notify"`
	Task        *Task        `toml:"task"`
}

type Task struct {
	Name string `toml:"name"`
}

type MQ struct {
	Name   string                 `toml:"name"`
	Params map[string]interface{} `toml:"params"`
}

type ChainNotify struct {
	Host string `toml:"host"`
	MQ   *MQ    `toml:"mq"`
}

type Lotus struct {
	Addr string `toml:"addr"`
}

type Storage struct {
	Name   string                 `toml:"name"`
	Params map[string]interface{} `toml:"params"`
}

// InitConfFile init config file to CFG
func InitConfFile(file string) (*CFG, error) {
	var cf CFG
	err := syscall.Access(file, syscall.O_RDONLY)
	if err != nil {
		return nil, err
	}
	err = configor.Load(&cf, file)
	if err != nil {
		return nil, err
	}

	return &cf, nil
}
