// +build windows

package wish

import (
	"os/exec"
)

func BrowserUrl(url string) error {
	cmd := exec.Command("cmd.exe", "/C", "start "+url)
	return cmd.Run()
}
