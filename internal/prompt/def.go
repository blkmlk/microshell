package prompt

import "github.com/sarulabs/di/v2"

const DefinitionName = "prompt"

var (
	Definition = di.Def{
		Name: DefinitionName,
		Build: func(ctn di.Container) (interface{}, error) {
			return newPrompt(), nil
		},
	}
)

type Prompt interface {
	StartChar() string
	Hostname() string
	SetHostname(hostname string)
	Username() string
	SetUsername(username string)
}
