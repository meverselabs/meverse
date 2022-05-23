module github.com/meverselabs/meverse

go 1.18

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/bluele/gcache v0.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.15
	github.com/gorilla/websocket v1.4.2
	github.com/labstack/echo v3.3.10+incompatible
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.18.1
	github.com/tidwall/btree v0.0.0-20170113224114-9876f1454cf0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.1.5
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/pkg/errors v0.9.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.6 // indirect
)

replace github.com/pkg/errors v0.9.1 => ./extern/errors
