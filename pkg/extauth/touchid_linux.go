// +build linux

package extauth

import "errors"

func TouchIDAuth(reason string) (bool, error) {
	return false, errors.New("touch id auth not support linux")
}
