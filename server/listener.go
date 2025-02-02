package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync_server/share"

	"github.com/google/uuid"
)

type Message struct {
	Msg    []byte
	Sender *Peer
}
type Peer struct {
	conn    net.Conn
	msgChan chan Message
}
type UploadListener struct {
	ClientId uuid.UUID
	FileName string
	Port     int
}
type Listener struct {
	cfg            *share.Config
	ln             net.Listener
	peers          map[*Peer]bool
	peerChan       chan *Peer
	quitChan       chan struct{}
	msgChan        chan Message
	uploadListener *UploadListener
	listeningPort  int
}

func NewPeer(conn net.Conn, msgChan chan Message) *Peer {
	return &Peer{
		conn,
		msgChan,
	}
}

func (p *Peer) readLoop() error {
	buf := make([]byte, 1024)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			slog.Error("peer read error", "err", err)
			return err
		}
		msgBuff := make([]byte, n)
		copy(msgBuff, buf[:n])
		msg := Message{
			Msg:    msgBuff,
			Sender: p,
		}
		p.msgChan <- msg
	}
}

func (p *Peer) Send(msg []byte) error {
	_, err := p.conn.Write(msg)
	if err != nil {
		slog.Error("peer failed to send message", "err", err)
		return err
	}
	return nil
}

func NewListener(cfg *share.Config, clientListener *UploadListener) *Listener {
	return &Listener{
		cfg:            cfg,
		peers:          make(map[*Peer]bool),
		peerChan:       make(chan *Peer),
		quitChan:       make(chan struct{}),
		msgChan:        make(chan Message),
		uploadListener: clientListener,
	}
}

func (s *Listener) Listen() error {
	var port int
	if s.uploadListener != nil {
		port = s.uploadListener.Port
	} else {
		port, _ = strconv.Atoi(s.cfg.ServerListenAddr)
	}
	s.listeningPort = port
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.ln = ln
	go s.loop()
	slog.Info("server running", "listenAddr", port)
	return s.acceptLoop()
}
func (s *Listener) Stop() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Listener) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Listener) loop() {
	for {
		select {
		case msg := <-s.msgChan:
			err := s.handleMsg(msg)
			if err != nil {
				msg.Sender.Send([]byte(err.Error()))
			}
		case peer := <-s.peerChan:
			s.peers[peer] = true
		case <-s.quitChan:
			return

		}
	}
}
func (s *Listener) handleConn(conn net.Conn) {
	cfPort, _ := strconv.Atoi(s.cfg.ServerListenAddr)
	if s.listeningPort == cfPort {
		peer := NewPeer(conn, s.msgChan)
		s.peerChan <- peer
		if err := peer.readLoop(); err != nil {
			slog.Error("peer read loop err", "err", err)
		}
	}
	// upload case
	go s.handleUpload(conn)

}

func (s *Listener) handleUpload(conn net.Conn) {

	buf := new(bytes.Buffer)

	var size int64
	fmt.Println("Waiting to receive file size...")
	err := binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		log.Fatalf("Failed to read file size: %v", err)
	}
	fmt.Println("Size received:", size)
	_, err = io.CopyN(buf, conn, size)
	if err != nil {
		slog.Error("receive file error", "err", err.Error())
	}
	err = os.WriteFile(s.uploadListener.FileName, buf.Bytes(), 0655)
	if err != nil {
		slog.Error("WriteFile file error", "err", err.Error())
	}
	defer s.Stop()
}
func (s *Listener) handleMsg(msg Message) error {
	msg.Sender.Send(msg.Msg)
	return nil
}

func (s *Server) InitListener(cl *UploadListener) *Listener {
	ln := NewListener(s.cfg, cl)
	go ln.Listen()
	return ln
}
