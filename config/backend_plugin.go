package config

import (
	"plugin"

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

type PluginConnection interface {
	load() error
	Connect() error
	NewClientManager() (client.Manager, error)
	NewGroupManager() (group.Manager, error)
	NewJWKManager() (jwk.Manager, error)
	NewOAuth2Manager(client.Manager) (pkg.FositeStorer, error)
	NewPolicyManager() (ladon.Manager, error)
}

type GoPluginConnection struct {
	Config     *Config
	plugin     *plugin.Plugin
	didConnect bool
	Logger     logrus.FieldLogger
	db         *sqlx.DB
}

func (c *GoPluginConnection) load() error {
	if c.plugin != nil {
		return nil
	}

	cf := c.Config
	p, err := plugin.Open(cf.DatabasePlugin)
	if err != nil {
		return errors.WithStack(err)
	}

	c.plugin = p
	return nil
}

func (c *GoPluginConnection) Connect() error {
	cf := c.Config
	if c.didConnect {
		return nil
	}

	if err := c.load(); err != nil {
		return errors.WithStack(err)
	}

	if l, err := c.plugin.Lookup("Connect"); err != nil {
		return errors.Wrap(err, "Unable to look up `Connect`")
	} else if con, ok := l.(func(url string) (*sqlx.DB, error)); !ok {
		return errors.New("Unable to type assert `Connect`")
	} else {
		if db, err := con(cf.DatabaseURL); err != nil {
			return errors.Wrap(err, "Could not connect to database")
		} else {
			cf.GetLogger().Info("Successfully connected through database plugin")
			c.db = db
			cf.GetLogger().Debugf("Address of database plugin is: %s", c.db)
			if err := db.Ping(); err != nil {
				cf.GetLogger().WithError(err).Fatal("Could not ping database connection from plugin")
			}
		}
	}
	return nil
}

func (c *GoPluginConnection) NewClientManager() (client.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	ctx := c.Config.Context()
	if l, err := c.plugin.Lookup("NewClientManager"); err != nil {
		return nil, errors.Wrap(err, "Unable to look up `NewClientManager`")
	} else if m, ok := l.(func(*sqlx.DB, fosite.Hasher) client.Manager); !ok {
		return nil, errors.New("Unable to type assert `NewClientManager`")
	} else {
		return m(c.db, ctx.Hasher), nil
	}
}

func (c *GoPluginConnection) NewGroupManager() (group.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	if l, err := c.plugin.Lookup("NewGroupManager"); err != nil {
		return nil, errors.Wrap(err, "Unable to look up `NewGroupManager`")
	} else if m, ok := l.(func(*sqlx.DB) group.Manager); !ok {
		return nil, errors.New("Unable to type assert `NewGroupManager`")
	} else {
		return m(c.db), nil
	}
}

func (c *GoPluginConnection) NewJWKManager() (jwk.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	if l, err := c.plugin.Lookup("NewJWKManager"); err != nil {
		return nil, errors.Wrap(err, "Unable to look up `NewJWKManager`")
	} else if m, ok := l.(func(*sqlx.DB, *jwk.AEAD) jwk.Manager); !ok {
		return nil, errors.New("Unable to type assert `NewJWKManager`")
	} else {
		return m(c.db, &jwk.AEAD{
			Key: c.Config.GetSystemSecret(),
		}), nil
	}
}

func (c *GoPluginConnection) NewOAuth2Manager(clientManager client.Manager) (pkg.FositeStorer, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	if l, err := c.plugin.Lookup("NewOAuth2Manager"); err != nil {
		return nil, errors.Wrap(err, "Unable to look up `NewOAuth2Manager`")
	} else if m, ok := l.(func(*sqlx.DB, client.Manager, logrus.FieldLogger) pkg.FositeStorer); !ok {
		return nil, errors.New("Unable to type assert `NewOAuth2Manager`")
	} else {
		return m(c.db, clientManager, c.Config.GetLogger()), nil
	}
}

func (c *GoPluginConnection) NewPolicyManager() (ladon.Manager, error) {
	if err := c.load(); err != nil {
		return nil, errors.WithStack(err)
	}

	if l, err := c.plugin.Lookup("NewPolicyManager"); err != nil {
		return nil, errors.Wrap(err, "Unable to look up `NewPolicyManager`")
	} else if m, ok := l.(func(*sqlx.DB) ladon.Manager); !ok {
		return nil, errors.Errorf("Unable to type assert `NewPolicyManager`, got %v", l)
	} else {
		return m(c.db), nil
	}
}
