package cloudcmd

import (
	"os"
	"strings"

	"github.com/mach-composer/mcc-sdk-go/mccsdk"
	"github.com/spf13/cobra"
)

var listApiClientCmd = &cobra.Command{
	Use:   "list-api-clients",
	Short: "Manage your components",
	RunE: func(cmd *cobra.Command, args []string) error {
		organization := MustGetString(cmd, "organization")
		project := MustGetString(cmd, "project")

		client, ctx := getClient(cmd)
		paginator, _, err := (client.
			APIClientsApi.
			ApiClientQuery(ctx, organization, project).
			Execute())

		if err != nil {
			return handleError(err)
		}

		data := make([][]string, len(paginator.Results))
		for i, record := range paginator.Results {
			lastUsed := "never"
			if record.LastUsedAt != nil {
				lastUsed = record.GetLastUsedAt().Format("2006-01-02 15:04:05")
			}

			data[i] = []string{
				record.CreatedAt.Local().Format("2006-01-02 15:04:05"),
				record.ClientId,
				record.ClientSecret,
				lastUsed,
				record.GetDescription(),
				strings.Join(record.Scope, " "),
			}
		}

		writeTable(os.Stdout,
			[]string{"Created At", "Client ID", "Client Secret", "Last Used", "Description", "Scopes"},
			data,
		)

		return nil
	},
}

var createApiClientCmd = &cobra.Command{
	Use:   "create-api-client",
	Short: "Manage your components",
	RunE: func(cmd *cobra.Command, args []string) error {
		organization := MustGetString(cmd, "organization")
		project := MustGetString(cmd, "project")

		client, ctx := getClient(cmd)
		resource, _, err := (client.
			APIClientsApi.
			ApiClientCreate(ctx, organization, project).
			ApiClientDraft(mccsdk.ApiClientDraft{
				Scope: []string{"project:manage"},
			}).
			Execute())

		if err != nil {
			return handleError(err)
		}

		cmd.Printf("Client ID: %s\n", resource.ClientId)
		cmd.Printf("Client Secret: %s\n", resource.ClientSecret)
		cmd.Printf("Scopes: %s\n", strings.Join(resource.Scope, " "))

		return nil
	},
}

func init() {
	CloudCmd.AddCommand(listApiClientCmd)
	registerContextFlags(listApiClientCmd)

	CloudCmd.AddCommand(createApiClientCmd)
	registerContextFlags(createApiClientCmd)
}
