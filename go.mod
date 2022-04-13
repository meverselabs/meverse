module github.com/fletaio/fleta_v2

go 1.17

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/bluele/gcache v0.0.2
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.15
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.2
	github.com/labstack/echo v3.3.10+incompatible
	github.com/tidwall/btree v0.0.0-20170113224114-9876f1454cf0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/allegro/bigcache/v3 v3.0.1 // indirect
	github.com/cupcake/rdb v0.0.0-20161107195141-43ba34106c76 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/siddontang/go v0.0.0-20170517070808-cb568a3e5cc0 // indirect
	github.com/siddontang/rdb v0.0.0-20150307021120-fc89ed2e418d // indirect
)

require (
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.16.0+incompatible // indirect
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883 // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20200923215132-ac86123a3f01 // indirect
	github.com/benbjohnson/immutable v0.2.1 // indirect
	github.com/benbjohnson/tmpl v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.1.5 // indirect
	github.com/huandu/xstrings v1.0.0 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/influxdata/flux v0.131.0 // indirect
	github.com/influxdata/influxdb v1.9.5 // indirect
	github.com/influxdata/influxql v1.1.1-0.20210223160523-b6ab99450c93 // indirect
	github.com/influxdata/pkg-config v0.2.8 // indirect
	github.com/joho/godotenv v1.4.0
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/ledisdb/ledisdb v0.0.0-20200510135210-d35789ec47e6
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/uber/jaeger-client-go v2.28.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/xlab/treeprint v0.0.0-20180616005107-d6fb6747feb6 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/tools v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/pkg/errors v0.9.1 => ./extern/errors
