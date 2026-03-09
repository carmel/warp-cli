package status

import (
	"fmt"

	"github.com/carmel/warp-cli/cloudflare"
	. "github.com/carmel/warp-cli/cmd/shared"
	"github.com/spf13/cobra"
)

var shortMsg = "Prints the status of the current Cloudflare Warp device"

var Cmd = &cobra.Command{
	Use:   "status",
	Short: shortMsg,
	Long:  FormatMessage(shortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		RunCommandFatal(status)
	},
}

func init() {
}

func status() error {
	if err := EnsureConfigValidAccount(); err != nil {
		return fmt.Errorf("EnsureConfigValidAccount: %s", err)
	}

	ctx := CreateContext()

	account, err := cloudflare.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount: %s", err)
	}
	boundDevices, err := cloudflare.GetBoundDevices(ctx)
	if err != nil {
		return fmt.Errorf("GetBoundDevices: %s", err)
	}

	PrintAccountDetails(account, boundDevices)
	return nil
}
