package project

import (
	"bytes"
	"errors"
	"path/filepath"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
	"github.com/spf13/afero"
)

var ErrFailedToParseRegoFile = errors.New("failed to parse rego file")

type libDir struct {
	*Dir
	relations *relationsFile
}

func (l *libDir) WriteChanges(fsys afero.Fs) error {
	if err := l.Dir.WriteChanges(fsys); err != nil {
		return err
	}
	if err := l.relations.WriteChanges(fsys); err != nil {
		return err
	}

	return nil
}

func (l *libDir) addRelation(contents string) (string, error) {
	return l.relations.addRelation(contents)
}

func libFromDir(fsys afero.Fs, root string) (*libDir, error) {
	path := filepath.Join(root, "lib")
	dir, err := DirFromPath(fsys, path)
	if err != nil {
		return nil, err
	}
	relations, err := relationsFileFromDir(fsys, dir.Path())
	if err != nil {
		return nil, err
	}
	l := &libDir{
		Dir:       dir,
		relations: relations,
	}
	return l, nil
}

type relationsFile struct {
	*File
	module *ast.Module
	lines  int
}

const relationsStub = `package relations

import data.snyk
`

func newRelationsFile(file *File) *relationsFile {
	module, err := ast.ParseModule(file.Path(), relationsStub)
	if err != nil {
		// Our tests should eliminate this possibility
		panic(err)
	}
	r := &relationsFile{
		File:   file,
		module: module,
	}
	if err := r.UpdateContents(); err != nil {
		panic(err)
	}
	return r
}

func relationsFileFromDir(fsys afero.Fs, parent string) (*relationsFile, error) {
	path := filepath.Join(parent, "relations.rego")
	file, err := FileFromPath(fsys, path)
	if err != nil {
		return nil, err
	}
	if !file.Exists() {
		return newRelationsFile(file), nil
	}
	contents, err := afero.ReadFile(fsys, file.Path())
	if err != nil {
		return nil, readPathError(file.Path(), err)
	}
	module, err := ast.ParseModule(file.Path(), string(contents))
	if err != nil {
		return nil, pathError(file.Path(), ErrFailedToParseRegoFile, err)
	}
	r := &relationsFile{
		File:   file,
		module: module,
		lines:  bytes.Count(contents, []byte{'\n'}),
	}
	return r, nil
}

func (r *relationsFile) addRelation(contents string) (string, error) {
	rule, err := ast.ParseRule(contents)
	if err != nil {
		return "", err
	}
	rule.Location.Row = r.lines + 1
	r.module.Rules = append(r.module.Rules, rule)
	if err := r.UpdateContents(); err != nil {
		return "", err
	}
	return r.Path(), nil
}

func (r *relationsFile) UpdateContents() error {
	formatted, err := format.Ast(r.module)
	if err != nil {
		return err
	}
	r.File.UpdateContents(formatted)
	r.lines = bytes.Count(formatted, []byte{'\n'})
	return nil
}
