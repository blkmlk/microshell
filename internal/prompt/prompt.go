package prompt

func newPrompt() Prompt {
	return &prompt{
		username:  "none",
		startChar: ">",
	}
}

type prompt struct {
	username  string
	hostname  string
	startChar string
}

func (p *prompt) StartChar() string {
	return p.startChar
}

func (p *prompt) Hostname() string {
	return p.hostname
}

func (p *prompt) SetHostname(hostname string) {
	p.hostname = hostname
}

func (p *prompt) Username() string {
	return p.username
}

func (p *prompt) SetUsername(username string) {
	p.username = username
}
