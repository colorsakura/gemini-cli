package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/reugn/gemini-cli/gemini"
)

// ChatOpts represents Chat configuration options.
type ChatOpts struct {
	Model      string
	Format     bool
	Style      string
	Multiline  bool
	Terminator string
}

// Chat controls the chat flow.
type Chat struct {
	model  *gemini.ChatSession
	prompt *prompt
	reader *readline.Instance
	opts   *ChatOpts
}

// NewChat returns a new Chat.
func NewChat(user string, model *gemini.ChatSession, opts *ChatOpts) (*Chat, error) {
	reader, err := readline.NewEx(&readline.Config{})
	if err != nil {
		return nil, err
	}
	prompt := newPrompt(user)
	reader.SetPrompt(prompt.user)
	if opts.Multiline {
		reader.HistoryDisable()
	}
	return &Chat{
		model:  model,
		prompt: prompt,
		reader: reader,
		opts:   opts,
	}, nil
}

// StartChat starts the chat loop.
func (c *Chat) StartChat() {
	for {
		message, ok := c.read()
		if !ok {
			continue
		}
		command := c.parseCommand(message)
		if quit := command.run(message); quit {
			break
		}
	}
}

func (c *Chat) read() (string, bool) {
	if c.opts.Multiline {
		return c.readMultiLine()
	}
	return c.readLine()
}

func (c *Chat) readLine() (string, bool) {
	input, err := c.reader.Readline()
	if err != nil {
		return c.handleReadError(len(input), err)
	}
	return validateInput(input)
}

func (c *Chat) readMultiLine() (string, bool) {
	var builder strings.Builder
	term := c.opts.Terminator
	for {
		input, err := c.reader.Readline()
		if err != nil {
			c.reader.SetPrompt(c.prompt.user)
			return c.handleReadError(builder.Len()+len(input), err)
		}
		if strings.HasSuffix(input, term) {
			builder.WriteString(strings.TrimSuffix(input, term))
			break
		}
		if builder.Len() == 0 {
			c.reader.SetPrompt(c.prompt.userNext)
		}
		builder.WriteString(input + "\n")
	}
	c.reader.SetPrompt(c.prompt.user)
	return validateInput(builder.String())
}

func (c *Chat) parseCommand(message string) command {
	if strings.HasPrefix(message, systemCmdPrefix) {
		return newSystemCommand(c)
	}
	return newGeminiCommand(c)
}

func (c *Chat) handleReadError(inputLen int, err error) (string, bool) {
	if errors.Is(err, readline.ErrInterrupt) {
		if inputLen == 0 {
			return systemCmdQuit, true
		}
		return "", false
	}
	fmt.Printf("%s%s\n", c.prompt.cli, err)
	return "", false
}

func validateInput(input string) (string, bool) {
	input = strings.TrimSpace(input)
	return input, input != ""
}
