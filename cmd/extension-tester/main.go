// main contains a "bare bones" CLI that can be used to test the extension.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/snyk/cli-extension-cloud/cloud"
	"github.com/snyk/go-application-framework/pkg/app"
	localworkflows "github.com/snyk/go-application-framework/pkg/local_workflows"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/cobra"
)

func main() {
	logger := log.New(os.Stderr, "", 0)
	engine := app.CreateAppEngineWithLogger(logger)
	engine.AddExtensionInitializer(cloud.Init)
	if err := engine.Init(); err != nil {
		log.Fatal(err)
	}
	root := newNode("snyk")
	for _, w := range engine.GetWorkflows() {
		fullCmd := workflow.GetCommandFromWorkflowIdentifier(w)
		parts := strings.Fields(fullCmd)
		root.add(parts, w)
	}
	rootCmd := root.cmd(engine)
	globalConfig := workflow.GetGlobalConfiguration()
	globalFlags := workflow.FlagsetFromConfigurationOptions(globalConfig)
	rootCmd.PersistentFlags().AddFlagSet(globalFlags)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type cmdTree struct {
	name       string
	children   map[string]*cmdTree
	workflowID workflow.Identifier
}

func newNode(name string) *cmdTree {
	return &cmdTree{
		name:     name,
		children: map[string]*cmdTree{},
	}
}

func (n *cmdTree) add(parts []string, workflowID workflow.Identifier) {
	if len(parts) < 1 {
		n.workflowID = workflowID
		return
	}
	head := parts[0]
	child, ok := n.children[head]
	if !ok {
		child = newNode(head)
		n.children[head] = child
	}
	child.add(parts[1:], workflowID)
}

func (n *cmdTree) cmd(engine workflow.Engine) *cobra.Command {
	var cmd *cobra.Command
	if n.workflowID == nil {
		cmd = &cobra.Command{
			Use:                n.name,
			Hidden:             true,
			RunE:               runEmpty,
			DisableFlagParsing: true,
		}
	} else {
		w, _ := engine.GetWorkflow(n.workflowID)
		cmd = &cobra.Command{
			Use:    n.name,
			Hidden: !w.IsVisible(),
			RunE: func(cmd *cobra.Command, args []string) error {
				data, err := engine.Invoke(n.workflowID)
				if err != nil {
					return err
				}
				_, err = engine.InvokeWithInput(localworkflows.WORKFLOWID_OUTPUT_WORKFLOW, data)
				return err
			},
		}
		options := w.GetConfigurationOptions()
		flagset := workflow.FlagsetFromConfigurationOptions(options)
		if flagset != nil {
			cmd.Flags().AddFlagSet(flagset)
		}
	}
	for _, child := range n.children {
		cmd.AddCommand(child.cmd(engine))
	}
	return cmd
}

func runEmpty(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("no workflow for command")
}
