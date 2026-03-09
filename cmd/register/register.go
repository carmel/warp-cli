package register

import (
	"fmt"
	"log"

	"github.com/carmel/warp-cli/cloudflare"
	. "github.com/carmel/warp-cli/cmd/shared"
	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deviceName string
var deviceModel string
var existingKey string
var acceptedTOS = false
var shortMsg = "Registers a new Cloudflare Warp device and creates a new account, preparing it for connection"

var Cmd = &cobra.Command{
	Use:   "register",
	Short: shortMsg,
	Long:  FormatMessage(shortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		RunCommandFatal(registerAccount)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&deviceName, "name", "n", "", "Device name displayed under the 1.1.1.1 app (defaults to random)")
	Cmd.PersistentFlags().StringVarP(&deviceModel, "model", "m", "PC", "Device model displayed under the 1.1.1.1 app")
	Cmd.PersistentFlags().StringVarP(&existingKey, "key", "k", "", "Base64 private key used to authenticate your device over WireGuard (defaults to random)")
	Cmd.PersistentFlags().BoolVar(&acceptedTOS, "accept-tos", false, "Accept Cloudflare's Terms of Service non-interactively")
}

func registerAccount() error {
	if err := EnsureNoExistingAccount(); err != nil {
		return fmt.Errorf("EnsureNoExistingAccount :%s", err)
	}

	if err := checkTOS(); err != nil {
		return fmt.Errorf("checkTOS :%s", err)
	}

	var privateKey *util.Key
	var err error

	if existingKey != "" {
		privateKey, err = util.NewKey(existingKey)
	} else {
		privateKey, err = util.NewPrivateKey()
	}
	if err != nil {
		return fmt.Errorf("NewKey :%s", err)
	}

	device, err := cloudflare.Register(privateKey.Public(), deviceModel)
	if err != nil {
		return fmt.Errorf("Register :%s", err)
	}

	viper.Set(config.PrivateKey, privateKey.String())
	viper.Set(config.DeviceId, device.Id)
	viper.Set(config.AccessToken, device.Token)
	viper.Set(config.LicenseKey, device.Account.License)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("WriteConfig :%s", err)
	}

	ctx := CreateContext()
	if _, err := SetDeviceName(ctx, deviceName); err != nil {
		return fmt.Errorf("SetDeviceName :%s", err)
	}

	account, err := cloudflare.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount :%s", err)
	}
	boundDevices, err := cloudflare.GetBoundDevices(ctx)
	if err != nil {
		return fmt.Errorf("GetBoundDevices :%s", err)
	}

	PrintAccountDetails(account, boundDevices)
	log.Println("Successfully created Cloudflare Warp account")
	return nil
}

func checkTOS() error {
	if !acceptedTOS {
		fmt.Println("This project is in no way affiliated with Cloudflare")
		fmt.Println("Cloudflare's Terms of Service: https://www.cloudflare.com/application/terms/")
		prompt := promptui.Select{
			Label: "Do you agree?",
			Items: []string{"Yes", "No"},
		}
		if _, result, err := prompt.Run(); err != nil {
			return fmt.Errorf("prompt run :%s", err)
		} else if result != "Yes" {
			return ErrTOSNotAccepted
		}
	}
	return nil
}
