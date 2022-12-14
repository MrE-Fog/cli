package create

import (
	"fmt"
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc/v2"
	awsCreate "github.com/OctopusDeploy/cli/pkg/cmd/account/aws/create"
	azureCreate "github.com/OctopusDeploy/cli/pkg/cmd/account/azure/create"
	gcpCreate "github.com/OctopusDeploy/cli/pkg/cmd/account/gcp/create"
	sshCreate "github.com/OctopusDeploy/cli/pkg/cmd/account/ssh/create"
	tokenCreate "github.com/OctopusDeploy/cli/pkg/cmd/account/token/create"
	usernameCreate "github.com/OctopusDeploy/cli/pkg/cmd/account/username/create"
	"github.com/OctopusDeploy/cli/pkg/constants"
	"github.com/OctopusDeploy/cli/pkg/factory"
	"github.com/spf13/cobra"
)

func NewCmdCreate(f factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates an account in an instance of Octopus Deploy",
		Long:  "Creates an account in an instance of Octopus Deploy.",
		Example: fmt.Sprintf(heredoc.Doc(`
			$ %s account create"
		`), constants.ExecutableName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createRun(f, cmd.OutOrStdout())
		},
	}

	return cmd
}

func createRun(f factory.Factory, w io.Writer) error {
	client, err := f.GetSpacedClient()
	if err != nil {
		return err
	}

	accountTypes := []string{
		"AWS Account",
		"Azure Account",
		"Google Cloud Account",
		"SSH Key Pair",
		"Username/Password",
		"Token",
	}

	var accountType string
	err = f.Ask(&survey.Select{
		Help:    "The type of account being created.",
		Message: "Account Type",
		Options: accountTypes,
	}, &accountType)
	if err != nil {
		return err
	}

	switch accountType {
	case "AWS Account":
		opts := &awsCreate.CreateOptions{
			Writer:      w,
			Octopus:     client,
			Ask:         f.Ask,
			Space:       f.GetCurrentSpace().GetID(),
			CreateFlags: awsCreate.NewCreateFlags(),
			CmdPath:     "octopus account aws create",
			Host:        f.GetCurrentHost(),
		}
		if err := awsCreate.CreateRun(opts); err != nil {
			return err
		}
	case "Azure Account":
		opts := &azureCreate.CreateOptions{
			Writer:      w,
			Octopus:     client,
			Ask:         f.Ask,
			Space:       f.GetCurrentSpace().GetID(),
			CreateFlags: azureCreate.NewCreateFlags(),
			CmdPath:     "octopus account azure create",
			Host:        f.GetCurrentHost(),
		}
		if err := azureCreate.CreateRun(opts); err != nil {
			return err
		}
	case "Google Cloud Account":
		opts := &gcpCreate.CreateOptions{
			Writer:      w,
			Octopus:     client,
			Ask:         f.Ask,
			Space:       f.GetCurrentSpace().GetID(),
			CreateFlags: gcpCreate.NewCreateFlags(),
			CmdPath:     "octopus account gcp create",
			Host:        f.GetCurrentHost(),
		}
		if err := gcpCreate.CreateRun(opts); err != nil {
			return err
		}
	case "SSH Key Pair":
		opts := &sshCreate.CreateOptions{
			Writer:      w,
			Octopus:     client,
			Ask:         f.Ask,
			Space:       f.GetCurrentSpace().GetID(),
			CreateFlags: sshCreate.NewCreateFlags(),
			CmdPath:     "octopus account ssh create",
			Host:        f.GetCurrentHost(),
		}
		if err := sshCreate.CreateRun(opts); err != nil {
			return err
		}
	case "Token":
		opts := &tokenCreate.CreateOptions{
			Writer:      w,
			Octopus:     client,
			Ask:         f.Ask,
			Space:       f.GetCurrentSpace().GetID(),
			CreateFlags: tokenCreate.NewCreateFlags(),
			CmdPath:     "octopus account token create",
			Host:        f.GetCurrentHost(),
		}
		if err := tokenCreate.CreateRun(opts); err != nil {
			return err
		}
	case "Username/Password":
		opts := &usernameCreate.CreateOptions{
			Writer:      w,
			Octopus:     client,
			Ask:         f.Ask,
			Space:       f.GetCurrentSpace().GetID(),
			CreateFlags: usernameCreate.NewCreateFlags(),
			CmdPath:     "octopus account username create",
			Host:        f.GetCurrentHost(),
		}
		if err := usernameCreate.CreateRun(opts); err != nil {
			return err
		}
	}

	return nil
}
