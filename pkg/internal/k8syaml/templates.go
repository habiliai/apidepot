package k8syaml

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"io/fs"
	"strings"
	"text/template"
)

//go:embed templates
var templatesFS embed.FS

func md5sum(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func ToLabel(val string) string {
	return strings.ReplaceAll(val, " ", "-")
}

func vapiMainFile() string {
	return vapiMainFileContents
}

func NewTemplate() (*template.Template, error) {
	root := template.New("").Funcs(sprig.TxtFuncMap()).Funcs(template.FuncMap{
		"md5sum":       md5sum,
		"toLabel":      ToLabel,
		"vapiMainFile": vapiMainFile,
	})

	if err := fs.WalkDir(templatesFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "failed to walk templates")
		}

		if entry.IsDir() {
			return nil
		}

		templateName := strings.Replace(path, "templates/", "", -1)

		body, err := templatesFS.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file")
		}

		if _, err := root.New(templateName).Parse(string(body)); err != nil {
			return errors.Wrapf(err, "failed to parse template")
		}

		logger.Info("load template", "name", templateName)
		return nil
	}); err != nil {
		return nil, err
	}

	return root, nil
}
