package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	codex "github.com/evertonmj/codex/app"
	"github.com/evertonmj/codex/pkg/resp"
)

// RESPServer implements the RESP protocol server
type RESPServer struct {
	store          *codex.Store
	listener       net.Listener
	config         *Config
	shutdownCh     chan struct{}
	wg             sync.WaitGroup
	authenticator  *Authenticator
	readTimeout    time.Duration
	writeTimeout   time.Duration
	connIDCounter  uint64
	connIDMutex    sync.Mutex
}

// Connection represents a client connection
type connection struct {
	id       uint64
	conn     net.Conn
	reader   *bufio.Reader
	writer   *resp.Writer
	server   *RESPServer
	authenticated bool
}

// NewRESPServer creates a new RESP protocol server
func NewRESPServer(store *codex.Store, config *Config, auth *Authenticator) *RESPServer {
	return &RESPServer{
		store:         store,
		config:        config,
		shutdownCh:    make(chan struct{}),
		authenticator: auth,
		readTimeout:   5 * time.Second,
		writeTimeout:  5 * time.Second,
	}
}

// Start starts the RESP server
func (s *RESPServer) Start() error {
	// Determine if API keys are required
	requiresAuth := len(s.config.APIKeys) > 0

	// Listen on the RESP port
	listener, err := net.Listen("tcp", net.JoinHostPort(s.config.Host, s.config.RESPPort))
	if err != nil {
		return fmt.Errorf("failed to listen on %s:%s: %w", s.config.Host, s.config.RESPPort, err)
	}

	s.listener = listener
	log.Printf("RESP server listening on %s:%s (auth required: %v)", s.config.Host, s.config.RESPPort, requiresAuth)

	// Accept connections in a goroutine
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// acceptConnections accepts incoming connections
func (s *RESPServer) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdownCh:
			return
		default:
		}

		// Set accept timeout
		s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := s.listener.Accept()
		if err != nil {
			// Check if it's a timeout (normal during shutdown)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			// Check if shutdown
			select {
			case <-s.shutdownCh:
				return
			default:
			}
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("Accept error: %v", err)
			}
			return
		}

		// Handle connection in goroutine
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (s *RESPServer) handleConnection(netConn net.Conn) {
	defer s.wg.Done()
	defer netConn.Close()

	// Create connection wrapper
	s.connIDMutex.Lock()
	s.connIDCounter++
	connID := s.connIDCounter
	s.connIDMutex.Unlock()

	conn := &connection{
		id:     connID,
		conn:   netConn,
		reader: bufio.NewReader(netConn),
		writer: resp.NewWriter(netConn),
		server: s,
		authenticated: len(s.config.APIKeys) == 0, // No auth needed if no keys configured
	}

	log.Printf("[conn:%d] Client connected from %s", conn.id, netConn.RemoteAddr())

	// Handle commands
	for {
		// Set read deadline
		conn.conn.SetReadDeadline(time.Now().Add(s.readTimeout))

		// Parse incoming command
		parser := resp.NewParser(conn.reader)
		val, err := parser.Parse()
		if err != nil {
			// Check if it's a timeout or closed connection
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				conn.writeError("ERR read timeout")
				continue
			}
			if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "closed") {
				break
			}
			log.Printf("[conn:%d] Parse error: %v", conn.id, err)
			conn.writeError(fmt.Sprintf("ERR %v", err))
			continue
		}

		if val == nil {
			continue
		}

		// Parse command
		cmd, args, err := resp.ParseCommand(val)
		if err != nil {
			log.Printf("[conn:%d] Command parse error: %v", conn.id, err)
			conn.writeError(fmt.Sprintf("ERR %v", err))
			continue
		}

		// Set write deadline
		conn.conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))

		// Execute command
		s.executeCommand(conn, cmd, args)
	}

	log.Printf("[conn:%d] Client disconnected", conn.id)
}

// executeCommand executes a RESP command
func (s *RESPServer) executeCommand(conn *connection, cmd string, args []string) {
	// Check authentication first
	if !conn.authenticated {
		if cmd != "CDX.AUTH" && cmd != "CDX.PING" {
			conn.writeError("NOAUTH authentication required")
			return
		}
	}

	switch cmd {
	case "CDX.SET":
		handleSet(conn, args)
	case "CDX.GET":
		handleGet(conn, args)
	case "CDX.DELETE":
		handleDelete(conn, args)
	case "CDX.HAS":
		handleHas(conn, args)
	case "CDX.KEYS":
		handleKeys(conn, args)
	case "CDX.CLEAR":
		handleClear(conn, args)
	case "CDX.PING":
		handlePing(conn, args)
	case "CDX.INFO":
		handleInfo(conn, args)
	case "CDX.AUTH":
		handleAuth(conn, args)
	default:
		conn.writeError(fmt.Sprintf("ERR unknown command '%s'", cmd))
	}
}

// Shutdown gracefully shuts down the server
func (s *RESPServer) Shutdown() error {
	log.Print("Shutting down RESP server...")
	close(s.shutdownCh)

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all connections to close with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Print("RESP server shut down gracefully")
		return nil
	case <-time.After(s.config.ShutdownTimeout):
		log.Print("RESP server shutdown timeout")
		return fmt.Errorf("RESP server shutdown timeout")
	}
}

// Connection helper methods

func (c *connection) writeError(msg string) error {
	return c.writer.WriteError(msg)
}

func (c *connection) writeSimpleString(s string) error {
	return c.writer.WriteSimpleString(s)
}

func (c *connection) writeInteger(n int64) error {
	return c.writer.WriteInteger(n)
}

func (c *connection) writeBulkString(b []byte) error {
	return c.writer.WriteBulkString(b)
}

func (c *connection) writeBulkStringStr(s string) error {
	return c.writer.WriteBulkStringStr(s)
}

func (c *connection) writeNull() error {
	return c.writer.WriteNull()
}

func (c *connection) writeArray(values []resp.Value) error {
	return c.writer.WriteArray(values)
}
