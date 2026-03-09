package generate

import (
	"fmt"
	"log"

	"github.com/carmel/warp-cli/cloudflare"
	. "github.com/carmel/warp-cli/cmd/shared"
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
	Long:  FormatMessage(shortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		RunCommandFatal(generateProfile)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&profileFile, "profile", "p", "warp-cli-profile.conf", "WireGuard profile file")
}

func generateProfile() error {
	if err := EnsureConfigValidAccount(); err != nil {
		return fmt.Errorf("EnsureConfigValidAccount :%s", err)
	}

	ctx := CreateContext()
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return fmt.Errorf("GetSourceDevice :%s", err)
	}

	profile, err := util.NewProfile(&util.ProfileData{
		PrivateKey: viper.GetString(config.PrivateKey),
		Address1:   thisDevice.Config.Interface.Addresses.V4,
		Address2:   thisDevice.Config.Interface.Addresses.V6,
		PublicKey:  thisDevice.Config.Peers[0].PublicKey,
		Endpoint:   thisDevice.Config.Peers[0].Endpoint.Host,
	})
	if err != nil {
		return fmt.Errorf(" :%s", err)
	}
	if err := profile.Save(profileFile); err != nil {
		return fmt.Errorf("profile save :%s", err)
	}

	log.Println("Successfully generated WireGuard profile:", profileFile)
	return nil
}
