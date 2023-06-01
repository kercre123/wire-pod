// +build vicos

package ipc

// GetSocketPath returns a platform-appropriate path for the given socket name
func GetSocketPath(socketName string) string {
	return "/dev/socket/" + socketName
}
