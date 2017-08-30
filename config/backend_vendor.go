package config

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ory/fosite"
	"github.com/ory/hydra/client"
	"github.com/ory/hydra/jwk"
	"github.com/ory/hydra/pkg"
	"github.com/ory/hydra/warden/group"
	"github.com/ory/ladon"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type VendoredPlugin interface {
	Connect(string) (*sqlx.DB, error)
	NewClientManager(*sqlx.DB, fosite.Hasher) client.Manager
	NewGroupManager(*sqlx.DB) group.Manager
	NewJWKManager(*sqlx.DB, *jwk.AEAD) jwk.Manager
	NewOAuth2Manager(*sqlx.DB, client.Manager, logrus.FieldLogger) pkg.FositeStorer
	NewPolicyManager(*sqlx.DB) ladon.Manager
}

type registeredVendoredPlugin struct {
	name   string
	plugin VendoredPlugin
	err    error
}

var vendoredPluginRegistry = map[string]registeredVendoredPlugin{}

var registryLogger = logrus.New()

func RegisterVendoredPlugin(name string, plugin VendoredPlugin, err error) error {
	//TODO sync on registry?
	registryLogger.Debugf("Registering %s %#v", name, plugin)
	if name == "" || plugin == nil {
		return fmt.Errorf("could not register name:%s, plugin:%v", name, plugin)
	}
	registered, ok := vendoredPluginRegistry[name]
	if ok {
		return fmt.Errorf("vendor plugin %s already registered", name)
	}
	registered.name = name
	registered.plugin = plugin
	registered.err = err
	vendoredPluginRegistry[name] = registered
	return nil
}

func Get(name string) (VendoredPlugin, error) {
	//TODO sync on registry?
	registered, ok := vendoredPluginRegistry[name]
	if !ok {
		return nil, fmt.Errorf("vendor plugin %s not registered", name)
	} else if registered.err != nil {
		return nil, errors.Wrap(registered.err, "requested plugin failed initialization")
	}
	return registered.plugin, nil
}

type VendoredPluginConnection struct {
	Config     *Config
	didConnect bool
	plugin     VendoredPlugin
	Logger     logrus.FieldLogger
	db         *sqlx.DB
}

func (c *VendoredPluginConnection) load() error {
	if c.plugin != nil {
		return nil
	}

	cf := c.Config
	p, err := Get(cf.DatabaseVendorPlugin)
	if err != nil {
		return errors.WithStack(err)
	}

	c.plugin = p
	return nil
}

func (c *VendoredPluginConnection) Connect() error {
	cf := c.Config
	if c.didConnect {
		return nil
	}

	if err := c.load(); err != nil {
		return errors.WithStack(err)
	}

	if db, err := c.plugin.Connect(cf.DatabaseURL); err != nil {
		return errors.Wrap(err, "Could not connect to database")
	} else {
		cf.GetLogger().Info("Successfully connected through database plugin")
		c.db = db
		cf.GetLogger().Debugf("Address of database plugin is: %s", c.db)
		if err := db.Ping(); err != nil {
			cf.GetLogger().WithError(err).Fatal("Could not ping database connection from plugin")
		}
	}
	return nil
}

func (c *VendoredPluginConnection) NewClientManager() (client.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	ctx := c.Config.Context()
	return c.plugin.NewClientManager(c.db, ctx.Hasher), nil
}

func (c *VendoredPluginConnection) NewGroupManager() (group.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	return c.plugin.NewGroupManager(c.db), nil
}

func (c *VendoredPluginConnection) NewJWKManager() (jwk.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}
	return c.plugin.NewJWKManager(c.db, &jwk.AEAD{
		Key: c.Config.GetSystemSecret()}), nil
}

func (c *VendoredPluginConnection) NewOAuth2Manager(clientManager client.Manager) (pkg.FositeStorer, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	return c.plugin.NewOAuth2Manager(c.db, clientManager, c.Config.GetLogger()), nil
}

func (c *VendoredPluginConnection) NewPolicyManager() (ladon.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}
	return c.plugin.NewPolicyManager(c.db), nil
}
