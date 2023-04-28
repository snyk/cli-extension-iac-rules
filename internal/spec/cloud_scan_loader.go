package spec

import (
	"encoding/json"

	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/models"
)

// cloudScanDetector is a simple Detector to load cloud scan files.  These
// are usually generated using `snyk iac scaffold spec`.
type cloudScanDetector struct {
}

type cloudScan struct {
	path  string
	state models.State
}

func (cloudScanDetector) DetectFile(file *input.File, opts input.DetectOptions) (input.IACConfiguration, error) {
	if file.Ext() != ".json" {
		return nil, input.UnrecognizedFileExtension
	}

	contents, err := file.Contents()
	if err != nil {
		return nil, err
	}

	var state models.State
	if err := json.Unmarshal(contents, &state); err != nil {
		return nil, input.FailedToParseInput
	}

	if state.InputType != "cloud_scan" {
		return nil, input.FailedToParseInput
	}

	return cloudScan{path: file.Path, state: state}, nil
}

func (cloudScanDetector) DetectDirectory(*input.Directory, input.DetectOptions) (input.IACConfiguration, error) {
	return nil, nil
}

func (c cloudScan) Type() *input.Type {
	return input.CloudScan
}

func (c cloudScan) ToState() models.State {
	return c.state
}

func (c cloudScan) LoadedFiles() []string {
	return []string{c.path}
}

func (cloudScan) Errors() []error {
	return nil
}

func (cloudScan) Location([]interface{}) ([]input.Location, error) {
	return nil, nil
}
