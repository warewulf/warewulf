// SPDX-License-Identifier: BSD-3-Clause
// Copyright holders are listed in LICENSE.md.

// Package wwtemplate renders Warewulf overlay and provisioning templates.
package wwtemplate

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// New configures a named template for either parsing or rendering. Callers
// supply any context-specific functions. Sprig takes precedence on overlap.
func New(fileName string, funcs template.FuncMap) *template.Template {
	return template.New(filepath.Base(fileName)).Option("missingkey=default").Funcs(funcs).Funcs(sprig.TxtFuncMap())
}

// Render parses source and executes it against data. The caller owns the
// output and may receive partial writes if execution fails.
func Render(w io.Writer, fileName string, source []byte, data any, funcs template.FuncMap) error {
	tmpl, err := New(fileName, funcs).Parse(string(source))
	if err != nil {
		return fmt.Errorf("could not parse template %s: %w", fileName, err)
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}
	return nil
}
