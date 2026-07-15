package main

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/i18n"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newLendingCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lending",
		Short: i18n.T("lending.short"),
	}

	cmd.AddCommand(&cobra.Command{
		Use:         "expected",
		Short:       i18n.T("lending.expected.short"),
		Long:        i18n.T("lending.expected.long"),
		Annotations: map[string]string{"source": "wts"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			data, err := app.client.GetLendingExpected(cmd.Context())
			if err != nil {
				return userFacingCommandError(err)
			}
			return output.WriteLendingExpected(cmd.OutOrStdout(), app.format, data)
		},
	})

	return cmd
}
