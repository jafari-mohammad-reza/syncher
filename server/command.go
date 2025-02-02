package server

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync_server/share"

	"github.com/nats-io/nats.go"
)

type CommandHandler struct {
	server *Server
	cmap   map[string]func(cmd *share.ClientCommand) (*share.ServerReply, error)
}

func NewCommandHandler(client *Server) *CommandHandler {
	ch := &CommandHandler{
		server: client,
		cmap:   make(map[string]func(cmd *share.ClientCommand) (*share.ServerReply, error)),
	}
	ch.cmap["upload"] = ch.getUploadLink
	return ch
}
func (ch *CommandHandler) getUploadLink(cmd *share.ClientCommand) (*share.ServerReply, error) {
	// here we will create a listener that has an specific link and will be closed automatically after few minutes or after upload completed
	port := rand.IntN(3999-3000) + 3000
	ch.server.InitListener(&UploadListener{Port: port, ClientId: cmd.ClientId, FileName: fmt.Sprintf("%s_%s", cmd.ClientId, cmd.Args["fileName"])})

	repl := share.ServerReply{
		Msg:      strconv.Itoa(port),
		ClientId: cmd.ClientId,
	}
	return &repl, nil
}
func (ch *CommandHandler) parseCommand(msg *nats.Msg) (*share.ServerReply, error) {
	cmd, err := share.ParseClientCommand(msg.Data)

	names := strings.Split(msg.Subject, "-")
	if err != nil {
		return nil, err
	}
	name := names[len(names)-1]
	if handler, exists := ch.cmap[name]; exists {
		return handler(cmd)
	} else {
		return nil, fmt.Errorf("unknown command '%s'", name)

	}
}
