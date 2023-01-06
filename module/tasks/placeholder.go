package tasks

import (
	"errors"
	"os"
	"strings"
)

type PlaceHolder interface {
	SetPH(key, value string)
	AppendToPH(key, value string) bool
	SetIfNotExists(key, value string)
	GetPHExists(key string) (string, bool)
	GetPH(key string) string
	GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string))
	HandlePlaceHolder(line string) string
	HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string
	ClearAll()
	ExportVarToFile(variable string, filename string) error
}

type DefaultPhHandler struct {
	ph map[string]string
}

func NewDefaultPhHandler() *DefaultPhHandler {
	return &DefaultPhHandler{
		ph: make(map[string]string),
	}
}

func (d *DefaultPhHandler) SetPH(key, value string) {
	d.ph[key] = value
}

func (d *DefaultPhHandler) AppendToPH(key, value string) bool {
	if _, ok := d.ph[key]; ok {
		d.ph[key] += value
		return true
	}
	return false
}

func (d *DefaultPhHandler) SetIfNotExists(key, value string) {
	if _, ok := d.ph[key]; !ok {
		d.ph[key] = value
	}
}

func (d *DefaultPhHandler) GetPHExists(key string) (string, bool) {
	if value, ok := d.ph[key]; ok {
		return value, true
	}
	return "", false
}

func (d *DefaultPhHandler) GetPH(key string) string {
	return d.ph[key]
}

func (d *DefaultPhHandler) GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string)) {
	for key, value := range d.ph {
		inspectFunc(key, value)
	}
}

func (d *DefaultPhHandler) HandlePlaceHolder(line string) string {
	return d.HandlePlaceHolderWithScope(line, d.ph)
}

func (d *DefaultPhHandler) HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string {
	for key, value := range scopeVars {
		line = strings.ReplaceAll(line, key, value)
	}
	return line
}

func (d *DefaultPhHandler) ClearAll() {
	d.ph = make(map[string]string)
}

func (d *DefaultPhHandler) ExportVarToFile(variable string, filename string) error {
	strData := d.GetPH(variable)
	if strData == "" {
		return errors.New("variable " + variable + " can not be used for export to file. not exists or empty")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err2 := f.WriteString(d.HandlePlaceHolder(strData)); err != nil {
		return err2
	}

	return nil
}
