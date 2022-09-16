package util

import (
	"fmt"
)

func URLClientInit(scheme, baseURL, basePath string) string {
	return fmt.Sprintf("%s://%s%s/coreclient/init", scheme, baseURL, basePath)
}

func URLRegisterConfirm(scheme, baseURL, basePath string) string {
	return fmt.Sprintf("%s://%s%s/coreclient/finish", scheme, baseURL, basePath)
}

func URLActionMessages(scheme, baseURL, basePath, actionID string) string {
	return fmt.Sprintf("%s://%s%s/coreclient/action/%s", scheme, baseURL, basePath, actionID)
}

func URLActionApprove(scheme, baseURL, basePath, actionID string) string {
	return fmt.Sprintf("%s://%s%s/coreclient/action/%s", scheme, baseURL, basePath, actionID)
}

func URLActionReject(scheme, baseURL, basePath, actionID string) string {
	return fmt.Sprintf("%s://%s%s/coreclient/action/%s", scheme, baseURL, basePath, actionID)
}
