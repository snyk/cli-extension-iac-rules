package scaffold

import "github.com/snyk/go-application-framework/pkg/workflow"

func RegisterWorkflows(e workflow.Engine) error {
	if err := RegisterProjectWorkflow(e); err != nil {
		return err
	}
	return nil
}
