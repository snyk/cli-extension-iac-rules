package forms

import "github.com/snyk/policy-engine/pkg/input"

type Form interface {
	Run() error
}

func inputTypes() []string {
	return []string{
		input.Terraform.Name,
		input.CloudScan.Name,
		input.Kubernetes.Name,
		input.CloudFormation.Name,
		input.Arm.Name,
	}
}
