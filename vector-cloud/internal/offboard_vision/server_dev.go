// +build !shipping

package offboard_vision

import (
	"context"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
)

// Run starts the offboard vision service
func Run(ctx context.Context) {
	serv, err := ipc.NewUnixgramServer(ipc.GetSocketPath("offboard_vision_server"))
	if err != nil {
		log.Println("Error creating offboard vision server:", err)
		return
	}

	if done := ctx.Done(); done != nil {
		go func() {
			<-done
			serv.Close()
		}()
	}

	log.Println("Elemental offboard vision server is running")

	for c := range serv.NewConns() {
		cl := client{Conn: c}
		go cl.handleConn(ctx)
	}
}
