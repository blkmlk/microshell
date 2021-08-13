package logger

import "github.com/sarulabs/di/v2"

const DefinitionName = "logger"

var (
	Definition = di.Def{
		Name: DefinitionName,
		Build: func(ctn di.Container) (interface{}, error) {
			return newLogger(), nil
		},
	}
)

type Logger interface {
	WriteMessages(msg ...interface{})
}
