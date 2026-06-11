package pkg

import (
	"math/rand"
	"os"
)

func SelfImage() string {
	version, ok := os.LookupEnv("HELM_VERSION")
	if !ok {
		version = "1.0.107"
	}
	baseImage, ok1 := os.LookupEnv("IMAGE_REPO")
	if !ok1 {
		baseImage = "ccr.ccs.tencentyun.com/afan-public/w7panel"
	}
	return baseImage + ":" + version
}

func ServiceAccountName() string {
	sa, ok := os.LookupEnv("SERVICE_ACCOUNT_NAME")
	if !ok {
		sa = "w7panel-offline"
	}
	return sa
}

func RandomString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = 'a' + byte(rand.Intn(26)) // Simple example, generates lowercase letters
	}
	return string(bytes)
}


