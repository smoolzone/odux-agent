package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagPolicyName     string
	flagPolicySelector string
	flagPolicySpec     string
)

// policiesCmd manages versioned desired-state bundles (four-eyes approval).
var policiesCmd = &cobra.Command{
	Use:   "policies",
	Short: "Manage desired-state policies (versioned + four-eyes approval)",
}

var policiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		out, err := c.ListPolicies(at)
		if err != nil {
			return err
		}
		if flagOutput == "json" {
			printJSON(out)
			return nil
		}
		if len(out) == 0 {
			fmt.Println("(no policies)")
			return nil
		}
		for _, p := range out {
			fmt.Printf("%s %s v%d [%s] %s\n", emoji(shield), p.Name, p.Version, p.Status, p.Selector)
		}
		return nil
	},
}

var policiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a policy (draft v1)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		if flagPolicyName == "" {
			return fmt.Errorf("--name is required")
		}
		spec := map[string]any{}
		if flagPolicySpec != "" {
			if err := json.Unmarshal([]byte(flagPolicySpec), &spec); err != nil {
				return fmt.Errorf("--spec is not valid JSON: %w", err)
			}
		}
		p, err := c.CreatePolicy(at, flagPolicyName, flagPolicySelector, spec)
		if err != nil {
			return err
		}
		fmt.Printf("%s created %q v%d\n", emoji(check), p.Name, p.Version)
		return nil
	},
}

var policiesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show a policy's spec",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		p, err := c.GetPolicy(at, args[0])
		if err != nil {
			return err
		}
		if flagOutput == "json" {
			printJSON(p)
			return nil
		}
		b, _ := json.MarshalIndent(p.Spec, "", "  ")
		fmt.Printf("%s v%d [%s]\n%s\n", p.Name, p.Version, p.Status, string(b))
		return nil
	},
}

var policiesTargetsCmd = &cobra.Command{
	Use:   "targets <id>",
	Short: "Resolve the chambers a policy targets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		out, err := c.PolicyTargets(at, args[0])
		if err != nil {
			return err
		}
		if len(out) == 0 {
			fmt.Println("(no matching chambers)")
			return nil
		}
		for _, s := range out {
			fmt.Printf("%s %s (%s) role=%s\n", emoji(shield), s.Name, s.Host, s.Role)
		}
		return nil
	},
}

var policiesApplyCmd = &cobra.Command{
	Use:   "apply <id>",
	Short: "Apply an approved/active policy to its target chambers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		res, err := c.PolicyApply(at, args[0])
		if err != nil {
			return err
		}
		if flagOutput == "json" {
			printJSON(res)
			return nil
		}
		fmt.Printf("%s enqueued %d deploy command(s)\n", emoji(check), len(res.Enqueued))
		for _, e := range res.Enqueued {
			fmt.Printf("   -> %s (command %d)\n", e.Name, e.CommandID)
		}
		for _, s := range res.Skipped {
			fmt.Printf("   %s skipped %s (%s)\n", emoji(cross), s.Name, s.Reason)
		}
		return nil
	},
}

func policyActionCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s <id>", use),
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, at, err := session(cmd)
			if err != nil {
				return err
			}
			if err := c.PolicyAction(at, args[0], use); err != nil {
				return err
			}
			fmt.Printf("%s policy %s %sd\n", emoji(check), args[0], use)
			return nil
		},
	}
}

func init() {
	policiesCreateCmd.Flags().StringVar(&flagPolicyName, "name", "", "policy name (required)")
	policiesCreateCmd.Flags().StringVar(&flagPolicySelector, "selector", "", "label selector (e.g. role=db)")
	policiesCreateCmd.Flags().StringVar(&flagPolicySpec, "spec", "", "JSON spec bundle")

	policiesCmd.AddCommand(
		policiesListCmd,
		policiesCreateCmd,
		policiesGetCmd,
		policiesTargetsCmd,
		policiesApplyCmd,
		policyActionCmd("submit", "Submit a draft for approval"),
		policyActionCmd("approve", "Approve a pending policy (four-eyes)"),
		policyActionCmd("reject", "Reject a pending policy"),
		policyActionCmd("activate", "Activate an approved policy"),
		policyActionCmd("archive", "Archive an active/approved policy"),
	)
	rootCmd.AddCommand(policiesCmd)
}
