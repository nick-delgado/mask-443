package network

import (
    "bufio"
    "crypto/tls"
    "mask-443/config"
    "mask-443/logger"
    "mask-443/protocolhandler"
    "mask-443/spa"
    "net"
    "time"
)

type Server struct {
    config *config.Config
    // Potentially add listener net.Listener here for graceful shutdown
}

func NewServer(cfg *config.Config) (*Server, error) {
    // Load actual TLS configurations
    // This is a simplified placeholder; robust loading and error handling needed
    if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
        cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
        if err != nil {
            logger.Error.Printf("Failed to load TLS certificate/key: %v", err)
            // Decide if this is fatal or if server can run without TLS for some services
            // For this project, TLS is crucial.
            return nil, err
        }
        tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
        cfg.TLS.PublicTLSConfig = tlsCfg
        cfg.TLS.TunnelTLSConfig = tlsCfg // Can be same or different; here same for simplicity
    } else {
        logger.Warning.Println("TLS CertFile or KeyFile not specified. TLS will not be enabled.")
        // For this project, this should probably be a fatal error unless specifically allowing non-TLS mode.
    }


    return &Server{config: cfg}, nil
}

func (s *Server) Start() error {
    listenAddr := s.config.Server.ListenAddress + ":" + s.config.Server.ListenPort
    listener, err := net.Listen("tcp", listenAddr)
    if err != nil {
        return err
    }
    defer listener.Close()

    logger.Info.Printf("Listening on %s", listenAddr)

    for {
        conn, err := listener.Accept()
        if err != nil {
            logger.Error.Printf("Failed to accept connection: %v", err)
            continue // Or handle more gracefully, e.g., check for listener closed error
        }
        go s.handleConnection(conn)
    }
}

func (s *Server) handleConnection(conn net.Conn) {
    defer conn.Close()
    logger.Info.Printf("Accepted connection from %s", conn.RemoteAddr().String())

    // Use a buffered reader to allow peeking for SPA without consuming
    // from the main stream if it's not an SPA packet.
    br := bufio.NewReader(conn)

    // 1. Attempt SPA Knock Detection (on raw TCP connection)
    // Peeking allows us to check for SPA without consuming bytes if it's not an SPA packet.
    // The SPA packet itself would be the first thing sent by a SPA client.
    isSPARequest := false
    var spaClientInfo spa.ClientInfo // To store authorized client details

    if s.config.SPA.Enabled {
        // Peek initial bytes for SPA
        // MaxSPAPacketSize is important here for the peek buffer size
        peekedBytes, err := br.Peek(s.config.SPA.MaxSPAPacketSize)
        if err != nil && err != bufio.ErrBufferFull && err.Error() != "EOF" { // EOF might mean client sent less than MaxSPAPacketSize
            logger.Warning.Printf("Error peeking for SPA from %s: %v", conn.RemoteAddr().String(), err)
            // Decide how to handle this, maybe proceed as non-SPA
        }
        
        // If peekedBytes is shorter than a minimal SPA packet, it's not SPA
        // The actual ProcessKnock would need to handle potentially short peekedBytes
        // by trying to parse and only consuming if valid.

        // spa.ProcessKnock needs to be sophisticated:
        // - Try to parse peekedBytes as SPA.
        // - If it looks like a valid SPA structure and MAC checks out, consume those bytes from br.
        // - Return success and authorized info.
        // - If not, return failure, and the peekedBytes remain unconsumed for TLS handshake.

        // This is a simplified call. The actual spa.ProcessKnock would take br *bufio.Reader
        // or the peekedBytes and then conditionally consume from br.
        spaSuccess, clientInfo, err := spa.ProcessKnock(peekedBytes, conn.RemoteAddr(), s.config.SPA)
        if err == nil && spaSuccess {
            isSPARequest = true
            spaClientInfo = clientInfo
            logger.Info.Printf("SPA successful for %s. Transitioning to WSS tunnel.", conn.RemoteAddr().String())
            // Consume the bytes that formed the SPA packet from the buffered reader
            // This consumption length must be determined by ProcessKnock.
            // For this skeleton, let's assume ProcessKnock tells us how many bytes it was.
            // _, err = br.Discard(spaPacketActualLength) // spaPacketActualLength comes from ProcessKnock
            // if err != nil {
            //    logger.Error.Printf("Error discarding SPA bytes for %s: %v", conn.RemoteAddr().String(), err)
            //    return
            // }
        } else if err != nil {
            logger.Info.Printf("SPA processing error for %s: %v. Treating as public.", conn.RemoteAddr().String(), err)
        } else {
            logger.Info.Printf("No SPA or invalid SPA from %s. Treating as public.", conn.RemoteAddr().String())
        }
    }


    // 2. Handle based on SPA result
    var tlsConn *tls.Conn
    var tlsConfigToUse *tls.Config

    if isSPARequest {
        if s.config.TLS.TunnelTLSConfig == nil {
            logger.Error.Printf("Tunnel TLS config not loaded. Cannot proceed with SPA tunnel for %s.", conn.RemoteAddr().String())
            return
        }
        tlsConfigToUse = s.config.TLS.TunnelTLSConfig
        // The original conn (via br) is now wrapped with TLS for the tunnel
        // The client, after sending SPA, must immediately initiate a TLS handshake on the same conn.
        // Here, br (buffered reader on top of conn) is passed to tls.Server
        // The TLS handshake will read from 'br'.
        tlsConn = tls.Server(NewBufferedConn(conn, br), tlsConfigToUse)
        err := tlsConn.Handshake()
        if err != nil {
            logger.Error.Printf("TLS handshake failed for SPA tunnel for %s: %v", conn.RemoteAddr().String(), err)
            return
        }
        logger.Info.Printf("TLS handshake successful for SPA tunnel from %s.", conn.RemoteAddr().String())
        protocolhandler.HandleAuthenticatedConnection(tlsConn, s.config, spaClientInfo) // spaClientInfo might be used by the handler
    } else {
        // Not an SPA request, or SPA disabled/failed, treat as public
        if s.config.TLS.PublicTLSConfig == nil {
            logger.Error.Printf("Public TLS config not loaded. Cannot serve public content for %s.", conn.RemoteAddr().String())
            // If you want to support non-TLS on 443 (not recommended), handle raw conn here.
            return
        }
        tlsConfigToUse = s.config.TLS.PublicTLSConfig
        // Wrap the original conn (via br) with TLS for public services
        // The TLS handshake will read from 'br'. Any bytes peeked for SPA and not consumed
        // will be available for the TLS handshake here.
        tlsConn = tls.Server(NewBufferedConn(conn, br), tlsConfigToUse)
        err := tlsConn.Handshake()
        if err != nil {
            logger.Error.Printf("Public TLS handshake failed for %s: %v", conn.RemoteAddr().String(), err)
            return
        }
        logger.Info.Printf("Public TLS handshake successful for %s.", conn.RemoteAddr().String())
        protocolhandler.HandlePublicConnection(tlsConn, s.config)
    }
}

// BufferedConn wraps a net.Conn and a *bufio.Reader to ensure that
// bytes peeked by the bufio.Reader are available to subsequent reads on the Conn,
// particularly for the TLS handshake after an SPA check.
type BufferedConn struct {
    net.Conn
    r *bufio.Reader
}

func NewBufferedConn(conn net.Conn, r *bufio.Reader) *BufferedConn {
    return &BufferedConn{Conn: conn, r: r}
}

// Read from the buffered reader first, then from the underlying connection if empty.
// This ensures that peeked bytes are consumed before falling back to the conn.
func (bc *BufferedConn) Read(b []byte) (int, error) {
    if bc.r.Buffered() > 0 {
        return bc.r.Read(b)
    }
    return bc.Conn.Read(b)
}

// Write, Close, LocalAddr, RemoteAddr, SetDeadline, SetReadDeadline, SetWriteDeadline
// are passed through to the underlying net.Conn.
func (bc *BufferedConn) Write(b []byte) (int, error) {
    return bc.Conn.Write(b)
}

func (bc *BufferedConn) Close() error {
    return bc.Conn.Close()
}

func (bc *BufferedConn) LocalAddr() net.Addr {
    return bc.Conn.LocalAddr()
}

func (bc *BufferedConn) RemoteAddr() net.Addr {
    return bc.Conn.RemoteAddr()
}

func (bc *BufferedConn) SetDeadline(t time.Time) error {
    return bc.Conn.SetDeadline(t)
}

func (bc *BufferedConn) SetReadDeadline(t time.Time) error {
    return bc.Conn.SetReadDeadline(t)
}

func (bc *BufferedConn) SetWriteDeadline(t time.Time) error {
    return bc.Conn.SetWriteDeadline(t)
}

