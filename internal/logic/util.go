package logic

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gitlab.com/david_mbuvi/go_asterisks"
)

type options struct {
	AllowEmpty   bool
	ForceMinimum bool
	MinimumValue int
	ForceMaximum bool
	MaximumValue int
}
type Option func(*options)

func WithAllowEmpty() Option {
	return func(o *options) {
		o.AllowEmpty = true
	}
}

func WithMinimumValue(min int) Option {
	return func(o *options) {
		o.MinimumValue = min
		o.ForceMinimum = true
	}
}

func WithMaximumValue(max int) Option {
	return func(o *options) {
		o.MaximumValue = max
		o.ForceMaximum = true
	}
}

func GetValueFromPrompt(prompt string, opts ...Option) (string, error) {
	var (
		reader = bufio.NewReader(os.Stdin)
		val    string
		err    error
	)

	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	// read user input from terminal until the user input is not empty
	for val == "" {
		//nolint:forbidigo
		fmt.Print(prompt + ": ")

		val, err = reader.ReadString('\n')
		if err != nil {
			return "", errors.Wrap(err, "reading from prompt")
		}

		val = strings.TrimSpace(val)

		if val == "" {
			if o.AllowEmpty {
				break
			}

			log.Errorf(prompt + ": cannot be empty. Please try again.")
		}
	}

	if o.ForceMinimum || o.ForceMaximum {
		iVal, err := strconv.Atoi(val)
		if err != nil {
			return "", errors.Wrap(err, "converting input to integer")
		}

		if o.ForceMinimum && iVal < o.MinimumValue {
			return "", errors.Errorf("value must be greater than or equal to %d", o.MinimumValue)
		}

		if o.ForceMaximum && iVal > o.MaximumValue {
			return "", errors.Errorf("value must be less than or equal to %d", o.MaximumValue)
		}
	}

	return val, nil
}

func GetPasswordFromPrompt(prompt string) (string, error) {
	var password string

	// read user input from terminal until the user input is not empty
	for password == "" {
		//nolint:forbidigo
		fmt.Print(prompt + ": ")

		bytePassword, err := go_asterisks.GetUsersPassword("", true, os.Stdin, os.Stdout)
		if err != nil {
			return "", errors.Wrap(err, "reading from prompt")
		}

		if bytePassword == nil {
			//nolint:forbidigo
			fmt.Println()
			log.Error(prompt + " cannot be empty. Please try again.")
		}

		//nolint:forbidigo
		fmt.Println()

		password = string(bytePassword)
	}

	return password, nil
}

type tableWriterOptions struct {
	WidthMax int
}

type TableWriterOption func(*tableWriterOptions)

func WithTableWidthMax(width int) TableWriterOption {
	return func(o *tableWriterOptions) {
		o.WidthMax = width
	}
}

type TableWriter struct {
	table.Writer
}

func NewTableWriter(opts ...TableWriterOption) TableWriter {
	o := tableWriterOptions{
		WidthMax: 24,
	}

	for _, opt := range opts {
		opt(&o)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleDefault)
	t.SetAllowedRowLength(180)

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:           2,
			WidthMax:         o.WidthMax,
			WidthMaxEnforcer: text.WrapSoft,
		},
		{
			Number:           3,
			WidthMax:         o.WidthMax / 2,
			WidthMaxEnforcer: text.WrapSoft,
		},
		{
			Number:           4,
			WidthMax:         o.WidthMax / 2,
			WidthMaxEnforcer: text.WrapSoft,
		},
		{
			Number:           5,
			WidthMax:         o.WidthMax / 2,
			WidthMaxEnforcer: text.WrapSoft,
		},
		{
			Number:           6,
			WidthMax:         o.WidthMax / 2,
			WidthMaxEnforcer: text.WrapSoft,
		},
		{
			Number:           7,
			WidthMax:         o.WidthMax / 2,
			WidthMaxEnforcer: text.WrapSoft,
		},
	})

	return TableWriter{t}
}

func (t TableWriter) Render() {
	t.Writer.Render()
}

func GetIDFromPath(path string) string {
	pathParts := strings.Split(path, "/")

	return pathParts[len(pathParts)-1]
}
