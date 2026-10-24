package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
)

var (
	flagDeployServer   int
	flagDeployRepo     string
	flagDeployBranch   string
	flagDeployBuild    string
	flagDeployRollback string
)

func deployAccess(cmd *cobra.Command) (*api.Client, string, error) {
	return session(cmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Stage, approve, and run a deployment (four-eyes)",
}

var deploySubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Stage a deployment (requires a second approver to run)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := deployAccess(cmd)
		if err != nil {
			return err
		}
		id, err := c.CreateDeployment(at, api.DeploymentInput{
			ServerID:        flagDeployServer,
			Repo:            flagDeployRepo,
			Branch:          flagDeployBranch,
			BuildCommand:    flagDeployBuild,
			RollbackCommand: flagDeployRollback,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s deployment %d staged — a different approver must approve it before it runs\n", emoji(check), id)
		return nil
	},
}

func deployAction(use, short, path string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, at, err := deployAccess(cmd)
			if err != nil {
				return err
			}
			var e error
			switch path {
			case "approve":
				e = c.ApproveDeployment(at, args[0])
			case "run":
				e = c.RunDeployment(at, args[0])
			case "rollback":
				e = c.RollbackDeployment(at, args[0])
			}
			if e != nil {
				return e
			}
			fmt.Printf("%s deployment %s %sed\n", emoji(check), args[0], path)
			return nil
		},
	}
}

func init() {
	deploySubmitCmd.Flags().IntVar(&flagDeployServer, "server", 0, "chamber id to deploy to")
	deploySubmitCmd.Flags().StringVar(&flagDeployRepo, "repo", "", "git repo URL")
	deploySubmitCmd.Flags().StringVar(&flagDeployBranch, "branch", "main", "git branch")
	deploySubmitCmd.Flags().StringVar(&flagDeployBuild, "build", "", "build command")
	deploySubmitCmd.Flags().StringVar(&flagDeployRollback, "rollback", "", "rollback command")
	deploySubmitCmd.MarkFlagRequired("server")
	deploySubmitCmd.MarkFlagRequired("repo")
	deploySubmitCmd.MarkFlagRequired("build")

	deployCmd.AddCommand(deploySubmitCmd)
	deployCmd.AddCommand(deployAction("approve", "Approve a staged deployment", "approve"))
	deployCmd.AddCommand(deployAction("run", "Run an approved deployment", "run"))
	deployCmd.AddCommand(deployAction("rollback", "Roll back a deployment", "rollback"))
	rootCmd.AddCommand(deployCmd)
}
