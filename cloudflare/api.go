package cloudflare

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/openapi"
	"github.com/carmel/warp-cli/util"
)

const (
	ApiUrl     = "https://api.cloudflareclient.com"
	ApiVersion = "v0a1922"
)

var (
	DefaultHeaders = map[string]string{
		"User-Agent":        "okhttp/3.12.1",
		"CF-Client-Version": "a-6.3-1922",
	}
	DefaultTransport = &http.Transport{
		// Match app's TLS config or API will reject us with code 403 error 1020
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12},
		ForceAttemptHTTP2: false,
		// From http.DefaultTransport
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

var apiClient = MakeApiClient(nil)
var apiClientAuth *openapi.APIClient

func MakeApiClient(authToken *string) *openapi.APIClient {
	httpClient := http.Client{Transport: DefaultTransport}
	apiClient := openapi.NewAPIClient(&openapi.Configuration{
		DefaultHeader: DefaultHeaders,
		UserAgent:     DefaultHeaders["User-Agent"],
		Debug:         false,
		Servers: []openapi.ServerConfiguration{
			{URL: ApiUrl},
		},
		HTTPClient: &httpClient,
	})
	if authToken != nil {
		apiClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + *authToken
	}
	return apiClient
}

func Register(publicKey *util.Key, deviceModel string) (*openapi.Register200Response, error) {
	timestamp := util.GetTimestamp()
	result, _, err := apiClient.DefaultAPI.
		Register(context.TODO(), ApiVersion).
		RegisterRequest(openapi.RegisterRequest{
			FcmToken:  "", // not empty on actual client
			InstallId: "", // not empty on actual client
			Key:       publicKey.String(),
			Locale:    "en_US",
			Model:     deviceModel,
			Tos:       timestamp,
			Type:      "Android",
		}).Execute()
	return result, fmt.Errorf("Register: %s", err)
}

type SourceDevice openapi.GetSourceDevice200Response

func GetSourceDevice(ctx *config.Context) (*SourceDevice, error) {
	result, _, err := globalClientAuth(ctx.AccessToken).DefaultAPI.
		GetSourceDevice(context.TODO(), ApiVersion, ctx.DeviceId).
		Execute()
	return (*SourceDevice)(result), fmt.Errorf("GetSourceDevice: %s", err)
}

func globalClientAuth(authToken string) *openapi.APIClient {
	if apiClientAuth == nil {
		apiClientAuth = MakeApiClient(&authToken)
	}
	return apiClientAuth
}

type Account openapi.Account

func GetAccount(ctx *config.Context) (*Account, error) {
	result, _, err := globalClientAuth(ctx.AccessToken).DefaultAPI.
		GetAccount(context.TODO(), ctx.DeviceId, ApiVersion).
		Execute()
	castResult := (*Account)(result)
	return castResult, fmt.Errorf("GetAccount: %s", err)
}

func UpdateLicenseKey(ctx *config.Context) (*openapi.UpdateAccount200Response, error) {
	result, _, err := globalClientAuth(ctx.AccessToken).DefaultAPI.
		UpdateAccount(context.TODO(), ctx.DeviceId, ApiVersion).
		UpdateAccountRequest(openapi.UpdateAccountRequest{License: ctx.LicenseKey}).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("UpdateLicenseKey: %s", err)
	}

	return result, nil
}

type BoundDevice openapi.BoundDevice

func GetBoundDevices(ctx *config.Context) ([]BoundDevice, error) {
	result, _, err := globalClientAuth(ctx.AccessToken).DefaultAPI.
		GetBoundDevices(context.TODO(), ctx.DeviceId, ApiVersion).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("GetBoundDevices: %s", err)
	}
	var castResult []BoundDevice
	for _, device := range result {
		castResult = append(castResult, BoundDevice(device))
	}
	return castResult, nil
}

func GetSourceBoundDevice(ctx *config.Context) (*BoundDevice, error) {
	result, err := GetBoundDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetSourceBoundDevice: %s", err)
	}
	return FindDevice(result, ctx.DeviceId)
}

func UpdateSourceBoundDeviceName(ctx *config.Context, targetDeviceId string, newName string) (*BoundDevice, error) {
	return updateSourceBoundDevice(ctx, targetDeviceId, openapi.UpdateBoundDeviceRequest{
		Name: &newName,
	})
}

func UpdateSourceBoundDeviceActive(ctx *config.Context, targetDeviceId string, active bool) (*BoundDevice, error) {
	return updateSourceBoundDevice(ctx, targetDeviceId, openapi.UpdateBoundDeviceRequest{
		Active: &active,
	})
}

func updateSourceBoundDevice(ctx *config.Context, targetDeviceId string, data openapi.UpdateBoundDeviceRequest) (*BoundDevice, error) {
	result, _, err := globalClientAuth(ctx.AccessToken).DefaultAPI.
		UpdateBoundDevice(context.TODO(), ctx.DeviceId, ApiVersion, targetDeviceId).
		UpdateBoundDeviceRequest(data).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("updateSourceBoundDevice: %s", err)
	}
	var castResult []BoundDevice
	for _, device := range result {
		castResult = append(castResult, BoundDevice(device))
	}
	return FindDevice(castResult, ctx.DeviceId)
}

func DeleteBoundDevice(ctx *config.Context, targetDeviceId string) error {
	if _, err := globalClientAuth(ctx.AccessToken).DefaultAPI.
		DeleteBoundDevice(context.TODO(), ctx.DeviceId, ApiVersion, targetDeviceId).
		Execute(); err != nil {
		return err
	}
	return nil
}
