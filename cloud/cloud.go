package cloud

import (
	"github.com/snyk/cli-extension-cloud/internal/scaffold"
	"github.com/snyk/go-application-framework/pkg/workflow"

	"github.com/snyk/cli-extension-cloud/internal/push"
	"github.com/snyk/cli-extension-cloud/internal/spec"
)

func Init(e workflow.Engine) error {
	if err := scaffold.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := spec.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := push.RegisterWorkflows(e); err != nil {
		return err
	}
	return nil
}
