package plugin

import (
	"github.com/jmoiron/sqlx"
	"github.com/ory/fosite"
	"github.com/ory/hydra/client"
	"github.com/ory/hydra/jwk"
	"github.com/ory/hydra/pkg"
	"github.com/ory/hydra/warden/group"
	"github.com/ory/ladon"
	"github.com/sirupsen/logrus"
	_ "gopkg.in/rana/ora.v4"
)

/**
 * implements hydra/config.VendoredPlugin
 */
type OracleVendoredPlugin struct {
}

func (p OracleVendoredPlugin) Connect(url string) (db *sqlx.DB, err error) {
	return Connect(url)
}

func (p OracleVendoredPlugin) NewClientManager(db *sqlx.DB, hasher fosite.Hasher) client.Manager {
	return NewClientManager(db, hasher)
}

func (p OracleVendoredPlugin) NewGroupManager(db *sqlx.DB) group.Manager {
	return NewGroupManager(db)
}

func (p OracleVendoredPlugin) NewJWKManager(db *sqlx.DB, aead *jwk.AEAD) jwk.Manager {
	return NewJWKManager(db, aead)
}

func (p OracleVendoredPlugin) NewOAuth2Manager(db *sqlx.DB, cm client.Manager, logger logrus.FieldLogger) pkg.FositeStorer {
	return NewOAuth2Manager(db, cm, logger)
}

func (p OracleVendoredPlugin) NewPolicyManager(db *sqlx.DB) ladon.Manager {
	return NewPolicyManager(db)
}
