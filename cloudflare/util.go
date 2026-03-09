package cloudflare

import (
	"errors"
	"fmt"
	"log"

	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/util"
	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

func FindDevice(devices []BoundDevice, deviceId string) (*BoundDevice, error) {
	for _, device := range devices {
		if device.Id == deviceId {
			return &device, nil
		}
	}
	return nil, errors.New("device not found in list")
}

func PrintAccountDetails(account *Account, boundDevices []BoundDevice) {
	log.Println("Printing account details:")
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("Account")
	fmt.Println("================================================================")
	fmt.Printf("%-12s : %s\n", "Id", account.Id)
	fmt.Printf("%-12s : %s\n", "Account type", account.AccountType)
	fmt.Printf("%-12s : %s\n", "Created", account.Created)
	fmt.Printf("%-12s : %s\n", "Updated", account.Updated)
	fmt.Printf("%-12s : %s\n", "Premium data", humanize.Bytes(uint64(account.PremiumData)))
	fmt.Printf("%-12s : %s\n", "Quota", humanize.Bytes(uint64(account.Quota)))
	fmt.Printf("%-12s : %s\n", "Role", account.Role)
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("Devices")
	fmt.Println("================================================================")
	for _, device := range boundDevices {
		name := "N/A"
		if device.Name != nil {
			name = *device.Name
		}
		id := device.Id
		if device.Id == viper.GetString(config.DeviceId) {
			id += " (current)"
		}
		fmt.Printf("%-9s : %s\n", "Id", id)
		fmt.Printf("%-9s : %s\n", "Type", device.Type)
		fmt.Printf("%-9s : %s\n", "Model", device.Model)
		fmt.Printf("%-9s : %s\n", "Name", name)
		fmt.Printf("%-9s : %t\n", "Active", device.Active)
		fmt.Printf("%-9s : %s\n", "Created", device.Created)
		fmt.Printf("%-9s : %s\n", "Activated", device.Activated)
		fmt.Printf("%-9s : %s\n", "Role", device.Role)
		fmt.Println()
	}
}

func SetDeviceName(ctx *config.Context, deviceName string) (*BoundDevice, error) {
	if deviceName == "" {
		deviceName += util.RandomHexString(3)
	}
	device, err := UpdateSourceBoundDeviceName(ctx, ctx.DeviceId, deviceName)
	if err == nil {
		if device.Name == nil || *device.Name != deviceName {
			return nil, errors.New("could not update device name")
		}
	} else if util.IsHttp500Error(err) {
		// server-side issue, but the operation still works
	} else {
		return nil, fmt.Errorf("UpdateSourceBoundDeviceName: %v", err)
	}

	return device, nil
}
