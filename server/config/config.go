package config

import (
    "crypto/tls"
    "os"
    "time"
    "gopkg.in/yaml.v3" // Example for YAML, add to go.mod
)

type Config struct {
    Server          ServerConfig          `yaml:"server"`
    SPA             SPAConfig             `yaml:"spa"`
    PublicServices  PublicServicesConfig  `yaml:"public_services"`
    TunnelService   TunnelServiceConfig   `yaml:"tunnel_service"`
    TLS             TLSConfig             `yaml:"tls"`
    InternalBackend InternalBackendConfig `yaml:"internal_backend"`
}

type ServerConfig struct {
    ListenAddress string `yaml:"listen_address"`
    ListenPort    string `yaml:"listen_port"` // e.g., "443"
    ReadTimeout   time.Duration `yaml:"read_timeout"` // For initial SPA read
}

type SPAConfig struct {
    Enabled          bool          `yaml:"enabled"`
    EncryptionKey    string        `yaml:"encryption_key"` // Path to key file or the key itself (less secure)
    HMACKey          string        `yaml:"hmac_key"`       // Path to key file or the key itself
    TimestampWindow  time.Duration `yaml:"timestamp_window"`
    NonceStoreTTL    time.Duration `yaml:"nonce_store_ttl"`
    MaxSPAPacketSize int           `yaml:"max_spa_packet_size"`
}

type PublicServicesConfig struct {
    EnableHTTPS bool   `yaml:"enable_https"`
    EnableWSS   bool   `yaml:"enable_wss"`
    DecoyWebDir string `yaml:"decoy_web_dir"` // For serving static files
    // Add backend proxy targets if needed
}

type TunnelServiceConfig struct {
    WSSPath            string `yaml:"wss_path"` // e.g., "/_sEcReT_TuNnEl_"
}

type InternalBackendConfig struct {
    TargetHost string `yaml:"target_host"` // e.g., "localhost"
    TargetPort string `yaml:"target_port"` // e.g., "22" for SSH
}

type TLSConfig struct {
    CertFile string `yaml:"cert_file"`
    KeyFile  string `yaml:"key_file"`
    // Parsed TLS configurations for public and tunnel services
    PublicTLSConfig *tls.Config `yaml:"-"`
    TunnelTLSConfig *tls.Config `yaml:"-"`
}


func LoadConfig(filePath string) (*Config, error) {
    cfg := &Config{ // Set defaults
        Server: ServerConfig{
            ListenAddress: "0.0.0.0",
            ListenPort:    "443",
            ReadTimeout:   2 * time.Second,
        },
        SPA: SPAConfig{
            Enabled:          true,
            TimestampWindow:  60 * time.Second,
            NonceStoreTTL:    5 * time.Minute,
            MaxSPAPacketSize: 256,
        },
        TunnelService: TunnelServiceConfig{
            WSSPath: "/_sEcReT_TuNnEl_", // Make this configurable and non-obvious
        },
        InternalBackend: InternalBackendConfig{
            TargetHost: "localhost",
            TargetPort: "22",
        },
    }

    data, err := os.ReadFile(filePath)
    if err != nil {
        // If config file not found, could proceed with defaults or partial config
        // For now, let's return error if specified file not found for clarity
        return nil, err 
    }

    err = yaml.Unmarshal(data, cfg)
    if err != nil {
        return nil, err
    }

    // Placeholder: Actual TLS config loading would happen here or in main
    // Example:
    // cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
    // if err != nil { return nil, err }
    // cfg.TLS.PublicTLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
    // cfg.TLS.TunnelTLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}} // Could be same or different

    return cfg, nil
}
