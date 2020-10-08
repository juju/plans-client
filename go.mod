module github.com/juju/plans-client

go 1.13

require (
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/canonical/candid v1.4.3
	github.com/gosuri/uitable v0.0.4
	github.com/juju/charm/v8 v8.0.0-20200925053015-07d39c0154ac
	github.com/juju/charmrepo/v6 v6.0.0-20200817155725-120bd7a8b1ed
	github.com/juju/cmd v0.0.0-20200108104440-8e43f3faa5c9
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d
	github.com/juju/juju v0.0.0-20201007080928-1f35f6a20b57
	github.com/juju/names v0.0.0-20180129205841-f9b5b8b7614d
	github.com/juju/names/v4 v4.0.0-20200923012352-008effd8611b
	github.com/juju/persistent-cookiejar v0.0.0-20171026135701-d5e5a8405ef9
	github.com/juju/testing v0.0.0-20200923013621-75df6121fbb0
	github.com/juju/utils v0.0.0-20200604140309-9d78121a29e0
	golang.org/x/net v0.0.0-20200904194848-62affa334b73
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b
	gopkg.in/juju/charmstore.v5 v5.10.0 // indirect
	gopkg.in/juju/environschema.v1 v1.0.0
	gopkg.in/macaroon-bakery.v2 v2.2.0
	gopkg.in/macaroon.v1 v1.0.0
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/altoros/gosigma => github.com/juju/gosigma v0.0.0-20200420012028-063911838a9e

replace gopkg.in/mgo.v2 => github.com/juju/mgo v2.0.0-20190418114320-e9d4866cb7fc+incompatible

replace github.com/hashicorp/raft => github.com/juju/raft v2.0.0-20200420012049-88ad3b3f0a54+incompatible
