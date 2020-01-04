// +build !windows

package wish

import (
	"os/exec"
)

func BrowserUrl(url string) error {
	cmd := exec.Command("xdg-open", url)
	return cmd.Run()
}
