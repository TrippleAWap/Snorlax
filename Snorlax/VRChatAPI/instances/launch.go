package instances

import (
	"Snorlax/VRChatAPI/worlds"
	"fmt"
	"os/exec"
	"runtime"
)

func openVRChatURL(url string) error {
	var cmd *exec.Cmd

	// Determine the OS and set the command accordingly
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	// Run the command
	return cmd.Start()
}
func Launch(instance worlds.Instance) error {
	err := openVRChatURL(fmt.Sprintf("vrchat://launch?ref=vrchat.com&id=%s:%s&shortName=%s", instance.WorldId, instance.InstanceId, instance.ShortName))
	if err != nil {
		return err
	}
	return nil
}
