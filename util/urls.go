package util

import "fmt"

func URLRegisterConfirm(baseURL, clientID string) string {
	return fmt.Sprintf("%s/api/v1/p/coreclient/%s/finish", baseURL, clientID)
}
