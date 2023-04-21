package cloud

import (
	"github.com/snyk/cli-extension-cloud/internal/scaffold"
	"github.com/snyk/go-application-framework/pkg/workflow"
)

func Init(e workflow.Engine) error {
	if err := scaffold.RegisterWorkflows(e); err != nil {
		return err
	}

	return nil
}
