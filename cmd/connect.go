package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/api"
)

var (
	flagConnectName    string
	flagConnectHost    string
	flagConnectRole    string
	flagConnectGroup   string
	flagConnectTags    string
	flagConnectSSHUser string
	flagConnectSSHPort int
	flagConnectSSHKey  int
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Enroll a server as a chamber in your castle",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		s, err := c.CreateServer(at, api.ServerInput{
			Name:     flagConnectName,
			Host:     flagConnectHost,
			Role:     flagConnectRole,
			Group:    flagConnectGroup,
			Tags:     flagConnectTags,
			SSHUser:  flagConnectSSHUser,
			SSHPort:  flagConnectSSHPort,
			SSHKeyID: flagConnectSSHKey,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s chamber %q enrolled (id %d, %s)\n", emoji(castle), s.Name, s.ID, s.Host)
		return nil
	},
}

func init() {
	connectCmd.Flags().StringVar(&flagConnectName, "name", "", "chamber name (required)")
	connectCmd.Flags().StringVar(&flagConnectHost, "host", "", "host or IP (required)")
	connectCmd.Flags().StringVar(&flagConnectRole, "role", "", "role label (web|db|pos|app|backup)")
	connectCmd.Flags().StringVar(&flagConnectGroup, "group", "", "group label")
	connectCmd.Flags().StringVar(&flagConnectTags, "tags", "", "comma-separated tags")
	connectCmd.Flags().StringVar(&flagConnectSSHUser, "ssh-user", "root", "SSH user")
	connectCmd.Flags().IntVar(&flagConnectSSHPort, "ssh-port", 22, "SSH port")
	connectCmd.Flags().IntVar(&flagConnectSSHKey, "ssh-key", 0, "SSH key id (referenced, never pasted)")
	connectCmd.MarkFlagRequired("name")
	connectCmd.MarkFlagRequired("host")
	rootCmd.AddCommand(connectCmd)
}
