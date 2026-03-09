package config

import (
	"github.com/spf13/viper"
)

const (
	DeviceId    = "device_id"
	AccessToken = "access_token"
	PrivateKey  = "private_key"
	LicenseKey  = "license_key"
)

type Context struct {
	DeviceId    string
	AccessToken string
	PrivateKey  string
	LicenseKey  string
}

func IsAccountValid() bool {
	return viper.GetString(DeviceId) != "" &&
		viper.GetString(AccessToken) != "" &&
		viper.GetString(PrivateKey) != ""
}

// func EnsureConfigValidAccount() error {
// 	if isConfigValidAccount() {
// 		return nil
// 	} else {
// 		return ErrNoAccount
// 	}
// }

// func EnsureNoExistingAccount() error {
// 	if isConfigValidAccount() {
// 		return ErrExistingAccount
// 	} else {
// 		return nil
// 	}
// }

func CreateContext() *Context {
	ctx := Context{
		DeviceId:    viper.GetString(DeviceId),
		AccessToken: viper.GetString(AccessToken),
		LicenseKey:  viper.GetString(LicenseKey),
	}
	return &ctx
}
