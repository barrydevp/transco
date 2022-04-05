package transco

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	ErrNotLeader = errors.New("not leader")
)

// Scheme constants
const (
	SchemeHttp       = "http"
	SchemeHttps      = "https"
	DefaultPort      = "8000"
	SysPrefixApiPath = "api/sys"
	MaxRetry         = 10
)

type RequestFunc func(req *RestRequest) (*resty.Response, error)

type node struct {
	conf        *nodeConfiguration
	IP          *net.IP
	BaseURL     string
	rest        *Rest
	isAvailable bool
}

type nodeConfiguration struct {
	ID   string
	Host string
}

type rsConfiguration struct {
	RsName string
	Nodes  []*nodeConfiguration
	Leader *nodeConfiguration
	// Current *node // current node is the node that own this rsconf
}

type connString struct {
	uri   string
	url   *url.URL
	hosts []string
}

type Connection struct {
	connStr *connString
	rsconf  *rsConfiguration
	nodes   []*node
	leader  *node
}

func newConnStr(uri string) (*connString, error) {
	if uri == "" {
		uri = "http://localhost" + DefaultPort
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != SchemeHttp && u.Scheme != SchemeHttps {
		return nil, fmt.Errorf("scheme must be \"http\" or \"https\"")
	}

	cs := &connString{}
	cs.uri = uri
	cs.url = u

	rawHosts := strings.Split(u.Host, ",")
	if len(rawHosts) == 0 {
		return nil, fmt.Errorf("empty host")
	}
	cs.hosts = make([]string, len(rawHosts))
	for i, host := range rawHosts {
		vHost := host
		if !strings.Contains(host, ":") {
			vHost += ":" + DefaultPort
		}
		cs.hosts[i] = vHost
	}

	return cs, nil
}

func NewConn(uri string) (*Connection, error) {
	cs, err := newConnStr(uri)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		connStr: cs,
	}

	return conn, nil
}

func (cs *connString) getBaseURL(host string) string {
	return cs.url.Scheme + "://" + host
}

func (c *Connection) loadNodes() error {
	cs := c.connStr
	nodes := make([]*node, len(cs.hosts))
	for i, host := range cs.hosts {
		baseURL := cs.getBaseURL(host)
		rest := NewRest()
		rest.SetBaseURL(baseURL)
		n := &node{
			BaseURL: baseURL,
			rest:    rest,
		}
		if err := n.init(); err != nil {
			return fmt.Errorf("init node failed: %w", err)
		}
		nodes[i] = n
	}

	c.nodes = nodes

	return nil
}

func (c *Connection) loadCluster() error {
	if len(c.nodes) == 0 {
		return fmt.Errorf("empty nodes")
	}

	// reset leader
	// @TODO: separate into individual method, using mutex lock to manage concurrent access
	c.rsconf = nil
	c.leader = nil
	var err error

	// fetch rsconf
	for _, node := range c.nodes {
		// firstNode := c.nodes[0]
		c.rsconf, err = node.rsconf()
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	// populate leader
	leaderConf := c.rsconf.Leader
	if leaderConf == nil {
		return fmt.Errorf("no leader")
	}

	var leader *node
	for _, n := range c.nodes {
		if n.conf.Host == leaderConf.Host && n.conf.ID == leaderConf.ID {
			leader = n
		}
	}

	if leader == nil {
		return fmt.Errorf("cannot found leader in uri")
	}

	c.leader = leader

	return nil

}

func (n *node) init() error {
	nconf, err := n.nconf()
	if err != nil {
		return err
	}
	n.conf = nconf

	return nil
}

func (n *node) request(fn RequestFunc) (resp *resty.Response, err error) {
	resp, err = fn(n.rest.requester())
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() != 200 {
		errBody, ok := resp.Error().(*ErrResponse)
		if ok {
			if strings.Contains(errBody.Err, "node is not the leader") {
				err = ErrNotLeader
			} else {
				err = fmt.Errorf("non 200 status: %v, msg: %v, err: %v", resp.StatusCode(), errBody.Msg, errBody.Err)
			}
		} else {
			err = fmt.Errorf("non 200 status: %v", resp.StatusCode())
		}
		return
	}

	return
}

func (n *node) Ping() error {
	if _, err := n.request(func(req *RestRequest) (*resty.Response, error) {
		return req.Get(SysPrefixApiPath + "/ping")
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) nconf() (*nodeConfiguration, error) {
	nconf := &nodeConfiguration{}
	if _, err := n.request(func(req *RestRequest) (*resty.Response, error) {
		return req.SetResult(nconf).Get(SysPrefixApiPath + "/nconf")
	}); err != nil {
		return nil, err
	}

	return nconf, nil
}

func (n *node) rsconf() (*rsConfiguration, error) {
	rsconf := &rsConfiguration{}
	if _, err := n.request(func(req *RestRequest) (*resty.Response, error) {
		return req.SetResult(rsconf).Get(SysPrefixApiPath + "/rsconf")
	}); err != nil {
		return nil, err
	}

	return rsconf, nil
}

func (c *Connection) Connect() error {
	if err := c.loadNodes(); err != nil {
		return err
	}

	if err := c.loadCluster(); err != nil {
		return err
	}

	return nil
}

func (c *Connection) request(fn RequestFunc) (resp *resty.Response, err error) {
	exec := func() {
		resp, err = c.leader.request(fn)
	}

	exec()
	retries := 0
	for retries < MaxRetry && errors.Is(err, ErrNotLeader) {
		// handle change leader by reload cluster
		if err = c.loadCluster(); err != nil {
			return
		}
		err = nil
		resp = nil
		exec()
		retries++
		delay := time.Duration(math.Pow(200, float64(retries))) * time.Millisecond
		time.Sleep(delay)
	}

	return
}

func (c *Connection) GetRsConf() *rsConfiguration {
	return c.rsconf
}

func (c *Connection) GetNodes() []*node {
	return c.nodes
}

func (c *Connection) Leader() *node {
	return c.leader
}
