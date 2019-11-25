module github.com/ovh/utask

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/Masterminds/squirrel v1.1.0
	github.com/SSSaaS/sssa-golang v0.0.0-20170502204618-d37d7782d752 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/facebookgo/ensure v0.0.0-20160127193407-b4ab57deab51 // indirect
	github.com/facebookgo/freeport v0.0.0-20150612182905-d4adf43b75b9 // indirect
	github.com/facebookgo/httpcontrol v0.0.0-20150708234001-ccde4420e1fe // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20150612182917-8dac2c3c4870 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.4.0
	github.com/go-gorp/gorp v2.0.0+incompatible
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jinzhu/now v1.0.1 // indirect
	github.com/juju/errors v0.0.0-20190207033735-e65537c515d7
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lib/pq v1.2.0
	github.com/loopfz/gadgeto v0.9.0
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/miscreant/miscreant-go v0.0.0-20190615163012-4f5dc8c406f6 // indirect
	github.com/miscreant/miscreant.go v0.0.0-20190615163012-4f5dc8c406f6 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/ovh/configstore v0.3.2
	github.com/ovh/go-ovh v0.0.0-20181109152953-ba5adb4cf014
	github.com/ovh/symmecrypt v0.3.0
	github.com/ovh/tat v5.2.5+incompatible
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/santhosh-tekuri/jsonschema v1.2.4
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/sparrc/go-ping v0.0.0-20190613174326-4e5b6552494c
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/tjarratt/babble v0.0.0-20140317234543-2cf06e8d98b0 // indirect
	github.com/ugorji/go v1.1.7 // indirect
	github.com/wI2L/fizz v0.0.0-20190425144348-6274bc96d962
	github.com/ziutek/mymysql v1.5.4 // indirect
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392
	golang.org/x/net v0.0.0-20190921015927-1a5e07d1ff72 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/appengine v1.6.3 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.46.0 // indirect
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

// Until https://github.com/tjarratt/babble/pull/6 is merged
replace github.com/tjarratt/babble => github.com/codeactual/babble v0.0.0-20190902213713-06cd230ffb31
