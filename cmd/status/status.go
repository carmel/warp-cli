package status

import (
	"errors"
	"fmt"

	"github.com/carmel/warp-cli/cloudflare"
	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
)

var shortMsg = "Prints the status of the current Cloudflare Warp device"

var Cmd = &cobra.Command{
	Use:   "status",
	Short: shortMsg,
	Long:  util.FormatMessage(shortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(status)
	},
}

func init() {
}

func status() error {

	if !config.IsAccountValid() {
		return errors.New("no valid account found.")
	}

	ctx := config.CreateContext()

	account, err := cloudflare.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount: %v", err)
	}
	boundDevices, err := cloudflare.GetBoundDevices(ctx)
	if err != nil {
		return fmt.Errorf("GetBoundDevices: %v", err)
	}

	cloudflare.PrintAccountDetails(account, boundDevices)
	return nil
}
