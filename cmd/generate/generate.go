package generate

import (
	"errors"
	"fmt"
	"log"

	"github.com/carmel/warp-cli/cloudflare"
	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profileFile string
var shortMsg = "Generates a WireGuard profile from the current Cloudflare Warp account"

var Cmd = &cobra.Command{
	Use:   "generate",
	Short: shortMsg,
	Long:  util.FormatMessage(shortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(generateProfile)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&profileFile, "profile", "p", "warp-cli-profile.conf", "WireGuard profile file")
}

func generateProfile() error {

	if !config.IsAccountValid() {
		return errors.New("no valid account found.")
	}

	ctx := config.CreateContext()
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return fmt.Errorf("GetSourceDevice: %v", err)
	}

	profile, err := util.NewProfile(&util.ProfileData{
		PrivateKey: viper.GetString(config.PrivateKey),
		Address1:   thisDevice.Config.Interface.Addresses.V4,
		Address2:   thisDevice.Config.Interface.Addresses.V6,
		PublicKey:  thisDevice.Config.Peers[0].PublicKey,
		Endpoint:   thisDevice.Config.Peers[0].Endpoint.Host,
	})
	if err != nil {
		return fmt.Errorf(": %v", err)
	}
	if err := profile.Save(profileFile); err != nil {
		return fmt.Errorf("profile save: %v", err)
	}

	log.Println("Successfully generated WireGuard profile:", profileFile)
	return nil
}
