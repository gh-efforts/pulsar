package config

import (
	"io"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/filecoin-project/lotus/node/config"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
)

var log = logging.Logger("bony/config")

// Conf defines the daemon config. It should be compatible with Lotus config.
type Conf struct {
	config.Common
	Client     config.Client
	Chainstore config.Chainstore
	Storage    StorageConf
	Cache      CacheConf
	Cluster    ClusterConf
	Auth       AuthConf
}

type StorageConf struct {
	Fs    map[string]FsStorageConf
	Mongo map[string]MongoStorageConf
	S3    map[string]S3StorageConf
}

type FsStorageConf struct {
	Format      string
	Path        string
	OmitHeader  bool   // when true, don't write column headers to new output files
	FilePattern string // pattern to use for filenames written in the path specified
}

type S3StorageConf struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	Bucket    string
}

type MongoStorageConf struct {
	URLEnv      string
	URL         string
	PoolSize    uint64
	AllowUpsert bool
}

type CacheConf struct {
	Type  string         // type of cache to use, such as "redis", "memory", default is "memory"
	Redis RedisCacheConf // redis cache config if type is "redis"
}

// RedisCacheConf is the config for redis cache.
// there is 2 types of redis cache:
// 1. single redis server:
//    RedisCacheConf{
//		Addrs: []string{":6379"}
//   }
// 2. redis sentinel:
//    RedisCacheConf{
//		MasterName: "master-name",
//   	Addrs: []string{":26379", ":26389", ":26399"}
//   }
type RedisCacheConf struct {
	MasterName string   // master name of redis sentinel, If empty, it means single redis server
	Addrs      []string // addresses of the redis servers, for example [":6379", ":6380"]
	DB         int      // redis db number
	Username   string   // redis auth username
	Password   string   // redis auth password
}

type ClusterConf struct {
	Name     string
	Role     string // "master" or "worker"
	BindAddr string
	BindPort int
	Seeds    []string
}

type AuthConf struct {
	Type string // local, kv, mix
	DSN  string // when type is kv or mix, this is the kv dsn
}

func DefaultConf() *Conf {
	return &Conf{
		Common: config.Common{
			API: config.API{
				ListenAddress: "/ip4/127.0.0.1/tcp/1234/http",
				Timeout:       config.Duration(30 * time.Second),
			},
			Libp2p: config.Libp2p{
				ListenAddresses: []string{
					"/ip4/0.0.0.0/tcp/0",
					"/ip6/::/tcp/0",
				},
				AnnounceAddresses:   []string{},
				NoAnnounceAddresses: []string{},

				ConnMgrLow:   150,
				ConnMgrHigh:  180,
				ConnMgrGrace: config.Duration(20 * time.Second),
			},
			Pubsub: config.Pubsub{
				Bootstrapper: false,
				DirectPeers:  nil,
				RemoteTracer: "/dns4/pubsub-tracer.filecoin.io/tcp/4001/p2p/QmTd6UvR47vUidRNZ1ZKXHrAFhqTJAD27rKL9XYghEKgKX",
			},
		},
		Client: config.Client{
			SimultaneousTransfersForStorage:   config.DefaultSimultaneousTransfers,
			SimultaneousTransfersForRetrieval: config.DefaultSimultaneousTransfers,
		},
	}
}

// SampleConf is the example configuration that is written when bony is first started. All entries will be commented out.
func SampleConf() *Conf {
	def := DefaultConf()
	cfg := *def
	cfg.Storage = StorageConf{
		Fs: map[string]FsStorageConf{
			"CSV": {
				Format:      "CSV",
				Path:        "/tmp",
				OmitHeader:  false,
				FilePattern: "{table}.csv",
			},
		},
	}

	return &cfg
}

func EnsureExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	c, err := os.Create(path)
	if err != nil {
		return err
	}

	comm, err := config.ConfigComment(SampleConf())
	if err != nil {
		return xerrors.Errorf("comment: %w", err) //nolint
	}
	_, err = c.Write(comm)
	if err != nil {
		_ = c.Close()                                  // ignore error since we are recovering from a write error anyway
		return xerrors.Errorf("write config: %w", err) //nolint
	}

	if err := c.Close(); err != nil {
		return xerrors.Errorf("close config: %w", err) //nolint
	}
	return nil
}

// FromFile loads config from a specified file. If file does not exist or is empty defaults are assumed.
func FromFile(path string) (*Conf, error) {
	log.Infof("reading config from %s", path)
	file, err := os.Open(path)
	switch {
	case os.IsNotExist(err):
		log.Warnf("config does not exist at %s, falling back to defaults", path)
		return DefaultConf(), nil
	case err != nil:
		return nil, err
	}

	defer file.Close() // nolint: lll
	return FromReader(file, DefaultConf())
}

// FromReader loads config from a reader instance.
func FromReader(reader io.Reader, def *Conf) (*Conf, error) {
	cfg := *def
	_, err := toml.DecodeReader(reader, &cfg) //nolint
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
