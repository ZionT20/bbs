package state

import (
	"github.com/skycoin/bbs/src/misc/boo"
	"github.com/skycoin/bbs/src/misc/inform"
	"github.com/skycoin/bbs/src/store/obj"
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/cxo/node/gnet"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/skycoin/src/cipher"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	cxoLogPrefix         = "CXO"
	cxoSubDir            = "/cxo"
	cxoFileExtension     = ".db"
	cxoConnectionTimeout = time.Second * 10
)

var (
	ErrCXONotOpened     = boo.New(boo.NotAuthorised, "cxo not opened")
	ErrCXOAlreadyOpened = boo.New(boo.NotAuthorised, "cxo already opened")
)

// CXO updates events from cxo.
type CXO struct {
	mux  sync.Mutex
	c    *SessionConfig
	l    *log.Logger
	node *node.Node

	updater func(root *node.Root)
}

// NewCXO creates a new CXO.
func NewCXO(config *SessionConfig, updater func(root *node.Root)) *CXO {
	return &CXO{
		c:       config,
		l:       inform.NewLogger(true, os.Stdout, cxoLogPrefix),
		updater: updater,
	}
}

// Sets up CXO node.
func (c *CXO) Open(alias string, in *RetryIO) (*RetryIO, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.node != nil {
		return nil, ErrCXOAlreadyOpened
	}
	// Setup registry.
	r := skyobject.NewRegistry()
	r.Register("Board", obj.Board{})
	r.Register("Thread", obj.Thread{})
	r.Register("Post", obj.Post{})
	r.Register("ThreadPage", obj.ThreadPage{})
	r.Register("BoardPage", obj.BoardPage{})
	r.Register("ExternalRoot", obj.ExternalRoot{})
	r.Done()

	// Setup CXO Configurations.
	nc := node.NewConfig()
	nc.MaxMessageSize = 0 /* TODO: Adjust. */
	nc.InMemoryDB = *c.c.MemoryMode
	nc.DataDir = filepath.Join(*c.c.ConfigDir, cxoSubDir)
	nc.DBPath = filepath.Join(nc.DataDir, alias+cxoFileExtension)

	nc.EnableListener = true
	nc.Listen = "[::]:" + strconv.Itoa(*c.c.CXOPort)
	nc.EnableRPC = *c.c.CXORPCEnable
	nc.RPCAddress = "[::]:" + strconv.Itoa(*c.c.CXORPCPort)
	nc.GCInterval = 5 * time.Second

	// Setup CXO Callbacks.
	nc.OnRootFilled = c.updater

	// Attempt to setup CXO Node.
	var e error
	c.node, e = node.NewNodeReg(nc, r)
	if e != nil {
		return nil, e
	}

	// Initialize.
	failed := c.initialize(in)

	// Return.
	return failed, nil
}

func (c *CXO) lock() func() {
	c.mux.Lock()
	if c.node != nil {
		c.node.Container().LockGC()
	}

	return func() {
		if c.node != nil {
			c.node.Container().UnlockGC()
		}
		c.mux.Unlock()
	}
}

// Initialize attempts to connect/subscribe to public keys/addresses provided.
// Failed attempts are outputted.
func (c *CXO) Initialize(in *RetryIO) *RetryIO {
	defer c.lock()()
	if c.node == nil {
		return &RetryIO{}
	}
	return c.initialize(in)
}

func (c *CXO) initialize(in *RetryIO) *RetryIO {
	failed := new(RetryIO)
	// Connect to all addresses and record those that fail.
	for _, address := range in.addresses {
		if e := c.connect(address); e != nil {
			c.l.Printf("Failed to connect to '%s'. Error: '%v'.", address, e)
			failed.addresses = append(failed.addresses, address)
		} else {
			c.l.Printf("Connected to '%s'.", address)
		}
	}
	// Loop through addresses and generate (public key) -> (addresses).
	pkMap := make(map[cipher.PubKey][]string)
	for _, connection := range c.node.Pool().Connections() {
		address := connection.Address()
		gotPKs, e := c.node.ListOfFeedsResponseTimeout(connection, cxoConnectionTimeout)
		if e != nil {
			c.l.Printf("Failed to get list of feeds from '%s'. Error: '%v'.",
				address, e)
			continue
		}
		for _, gotPK := range gotPKs {
			pkMap[gotPK] = append(pkMap[gotPK], address)
		}
	}
	// Subscribe to all public keys.
	for _, pk := range in.pks {
		c.l.Printf("(START) SUBSCRIBE TO '%s'.", pk.Hex())
		passed := false
		// Self subscribe.
		if e := c.subscribe("", pk); e != nil {
			c.l.Printf("\t-Self subscribe failed.")
		}
		for _, address := range pkMap[pk] {
			if e := c.subscribe(address, pk); e != nil {
				c.l.Printf("\t-Failed for '%s'.", address)
				continue
			}
			c.l.Printf("\t-Success for '%s'.", address)
			passed = true
		}
		if !passed {
			failed.pks = append(failed.pks, pk)
		}
		c.l.Printf("(  END) SUBSCRIBE TO '%s' (result '%v').", pk.Hex(), passed)
	}
	return failed
}

// Close closes the BBS container.
func (c *CXO) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.node == nil {
		return ErrCXONotOpened
	}
	if e := c.node.Close(); e != nil {
		return e
	}
	<-c.node.Quiting()
	c.node = nil
	return nil
}

// GetRoot obtains a root.
func (c *CXO) GetRoot(pk cipher.PubKey) (*node.Root, error) {
	defer c.lock()()
	if c.node == nil {
		return nil, ErrCXONotOpened
	}
	return c.node.Container().LastFullRoot(pk), nil
}

// NewRoot creates a new root.
func (c *CXO) NewRoot(seed []byte, modifier RootModifier) (cipher.PubKey, cipher.SecKey, error) {
	pk, sk := cipher.GenerateDeterministicKeyPair(seed)
	defer c.lock()()
	if c.node == nil {
		return pk, sk, ErrCXONotOpened
	}
	root, e := c.node.Container().NewRoot(pk, sk)
	if e != nil {
		return pk, sk, boo.WrapType(e, boo.Internal, "failed to create root")
	}
	return pk, sk, modifier(root)
}

// RootModifier modifies the root.
type RootModifier func(r *node.Root) error

// GetMasterWalker obtains walker of root that node can edit.
func (c *CXO) ModifyRoot(pk cipher.PubKey, sk cipher.SecKey, modifier RootModifier) error {
	defer c.lock()()
	if c.node == nil {
		return ErrCXONotOpened
	}
	return modifier(c.node.Container().LastRootSk(pk, sk))
}

// IsMaster returns whether master or not.
func (c *CXO) IsMaster() bool {
	return *c.c.Master
}

// Feeds returns a list of all feeds we are subscribed to.
func (c *CXO) Feeds() []cipher.PubKey {
	defer c.lock()()
	if c.node == nil {
		return nil
	}
	return c.node.Feeds()
}

// GetAddress gets the address of the node.
func (c *CXO) GetAddress() (string, error) {
	defer c.lock()()
	if c.node == nil {
		return "", ErrCXONotOpened
	}
	return c.node.Pool().Address(), nil
}

// Subscribe subscribes to a cxo feed.
func (c *CXO) Subscribe(address string, pk cipher.PubKey) error {
	defer c.lock()()
	if c.node == nil {
		return ErrCXONotOpened
	}
	return c.subscribe(address, pk)
}

func (c *CXO) subscribe(address string, pk cipher.PubKey) error {
	if address == "" {
		c.node.Subscribe(nil, pk)
		return nil
	}
	conn, e := c.node.Pool().Dial(address)
	if e != nil {
		switch e {
		case gnet.ErrAlreadyListen:
		default:
			return e
		}
	}
	return c.node.SubscribeResponse(conn, pk)
}

// Unsubscribe unsubscribes from a cxo feed.
func (c *CXO) Unsubscribe(address string, pk cipher.PubKey) error {
	defer c.lock()()
	if c.node == nil {
		return ErrCXONotOpened
	}
	return c.unsubscribe(address, pk)
}

func (c *CXO) unsubscribe(address string, pk cipher.PubKey) error {
	if address == "" {
		c.node.Unsubscribe(nil, pk)
		return nil
	}
	conn := c.node.Pool().Connection(address)
	if conn == nil {
		c.node.Unsubscribe(nil, pk)
		return nil
	}
	c.node.Unsubscribe(conn, pk)
	return nil
}

// GetConnections gets connections.
func (c *CXO) GetConnections() ([]string, error) {
	defer c.lock()()
	if c.node == nil {
		return nil, ErrCXONotOpened
	}
	conns := c.node.Pool().Connections()
	addresses := make([]string, len(conns))
	for i, conn := range conns {
		addresses[i] = conn.Address()
	}
	return addresses, nil
}

// Connect adds a connection.
func (c *CXO) Connect(address string) error {
	defer c.lock()()
	if c.node == nil {
		return ErrCXONotOpened
	}
	return c.connect(address)
}

func (c *CXO) connect(address string) error {
	if _, e := c.node.Pool().Dial(address); e != nil {
		return e
	}
	return nil
}

// Disconnect removes a connection.
func (c *CXO) Disconnect(address string) error {
	defer c.lock()()
	if c.node == nil {
		return ErrCXONotOpened
	}
	return c.disconnect(address)
}

func (c *CXO) disconnect(address string) error {
	conn := c.node.Pool().Connection(address)
	if conn == nil {
		return nil
	}
	return conn.Close()
}