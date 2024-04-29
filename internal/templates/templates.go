package templates

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward-cloud-cli/assets"
)

type Client struct {
	fs embed.FS
}

func New() *Client {
	return &Client{
		fs: assets.Assets,
	}
}

// Cwd returns the current working directory.
func (c *Client) Cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panicln(err)
	}

	return cwd
}

func (c *Client) LoadTemplate(name string) (*template.Template, error) {
	funcMap := funcMap()

	t, err := template.New("Dockerfile").Funcs(funcMap).ParseFS(c.fs, name)
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}

	// Add the "include" template function to make it possible to bring in another templates and pass the results to
	// other template functions.
	// {{ include "mytpl" . | upper | quote }}
	//
	// Similarly as it works in Helm.
	// https://helm.sh/docs/howto/charts_tips_and_tricks/#know-your-template-functions
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)

		err := t.ExecuteTemplate(buf, name, data)
		if err != nil {
			return "", errors.Wrap(err, "cannot execute template")
		}

		return buf.String(), nil
	}

	t, err = template.New("Dockerfile").Funcs(funcMap).ParseFS(c.fs, name)
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}

	return t, nil
}

// ExecuteTemplate executes the templates, appending some specific template functions to the execution.
func (c *Client) ExecuteTemplate(t *template.Template, name string, data any, buffer io.Writer) error {
	if name == "" {
		name = t.Name()
	}

	funcMap := funcMap()
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)

		err := t.ExecuteTemplate(buf, name, data)
		if err != nil {
			return "", errors.Wrap(err, "cannot execute template")
		}

		return buf.String(), nil
	}

	err := t.Funcs(funcMap).ExecuteTemplate(buffer, name, data)
	if err != nil {
		return fmt.Errorf("cannot execute template %s: %w", t.Name(), err)
	}

	return nil
}

func funcMap() template.FuncMap {
	//nolint:varnamelen
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	extra := template.FuncMap{
		"include":  func(string, interface{}) string { return "not implemented" },
		"tpl":      func(string, interface{}) interface{} { return "not implemented" },
		"required": func(string, interface{}) (interface{}, error) { return "not implemented", nil },
		"lookup": func(string, string, string, string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
		"isEnabled": isEnabled,
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

// isEnabled returns true if given value is true (bool), 1 (int), "1" (string) or "true" (string).
func isEnabled(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return false
	}

	switch g.Kind() { //nolint:exhaustive
	case reflect.String:
		return strings.EqualFold(g.String(), "true") || g.String() == "1"
	case reflect.Bool:
		return g.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 1
	default:
		return false
	}
}
