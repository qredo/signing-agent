package util

import (
	"fmt"
)

func URLClientInit(baseURL, basePath string) string {
	return fmt.Sprintf("https://%s%s/coreclient/init", baseURL, basePath)
}

func URLRegisterConfirm(baseURL, basePath string) string {
	return fmt.Sprintf("https://%s%s/coreclient/finish", baseURL, basePath)
}

func URLActionMessages(baseURL, basePath, actionID string) string {
	return fmt.Sprintf("https://%s%s/coreclient/action/%s", baseURL, basePath, actionID)
}

func URLActionApprove(baseURL, basePath, actionID string) string {
	return fmt.Sprintf("https://%s%s/coreclient/action/%s", baseURL, basePath, actionID)
}

func URLActionReject(baseURL, basePath, actionID string) string {
	return fmt.Sprintf("https://%s%s/coreclient/action/%s", baseURL, basePath, actionID)
}
