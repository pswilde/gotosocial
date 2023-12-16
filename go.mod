module github.com/superseriousbusiness/gotosocial

go 1.21

toolchain go1.21.3

require (
	codeberg.org/gruf/go-bytesize v1.0.2
	codeberg.org/gruf/go-byteutil v1.2.0
	codeberg.org/gruf/go-cache/v3 v3.5.7
	codeberg.org/gruf/go-debug v1.3.0
	codeberg.org/gruf/go-errors/v2 v2.3.1
	codeberg.org/gruf/go-fastcopy v1.1.2
	codeberg.org/gruf/go-iotools v0.0.0-20230811115124-5d4223615a7f
	codeberg.org/gruf/go-kv v1.6.4
	codeberg.org/gruf/go-logger/v2 v2.2.1
	codeberg.org/gruf/go-mutexes v1.3.1
	codeberg.org/gruf/go-runners v1.6.2
	codeberg.org/gruf/go-sched v1.2.3
	codeberg.org/gruf/go-store/v2 v2.2.4
	github.com/DmitriyVTitov/size v1.5.0
	github.com/KimMachineGun/automemlimit v0.4.0
	github.com/abema/go-mp4 v1.1.1
	github.com/buckket/go-blurhash v1.1.0
	github.com/coreos/go-oidc/v3 v3.9.0
	github.com/disintegration/imaging v1.6.2
	github.com/gin-contrib/cors v1.5.0
	github.com/gin-contrib/gzip v0.0.6
	github.com/gin-contrib/sessions v0.0.5
	github.com/gin-gonic/gin v1.9.1
	github.com/go-fed/httpsig v1.1.0
	github.com/go-playground/form/v4 v4.2.1
	github.com/google/uuid v1.4.0
	github.com/gorilla/feeds v1.1.2
	github.com/gorilla/websocket v1.5.1
	github.com/h2non/filetype v1.1.3
	github.com/jackc/pgx/v5 v5.5.0
	github.com/microcosm-cc/bluemonday v1.0.26
	github.com/miekg/dns v1.1.57
	github.com/minio/minio-go/v7 v7.0.65
	github.com/mitchellh/mapstructure v1.5.0
	github.com/oklog/ulid v1.3.1
	github.com/prometheus/client_golang v1.17.0
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.4
	github.com/superseriousbusiness/activity v1.4.0-gts
	github.com/superseriousbusiness/exif-terminator v0.6.0
	github.com/superseriousbusiness/oauth2/v4 v4.3.2-SSB.0.20230227143000-f4900831d6c8
	github.com/tdewolff/minify/v2 v2.20.9
	github.com/technologize/otel-go-contrib v1.1.0
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/ulule/limiter/v3 v3.11.2
	github.com/uptrace/bun v1.1.16
	github.com/uptrace/bun/dialect/pgdialect v1.1.16
	github.com/uptrace/bun/dialect/sqlitedialect v1.1.16
	github.com/uptrace/bun/extra/bunotel v1.1.16
	github.com/wagslane/go-password-validator v0.3.0
	github.com/yuin/goldmark v1.6.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0
	go.opentelemetry.io/otel/exporters/prometheus v0.44.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/sdk/metric v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.uber.org/automaxprocs v1.5.3
	golang.org/x/crypto v0.16.0
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
	golang.org/x/image v0.14.0
	golang.org/x/net v0.19.0
	golang.org/x/oauth2 v0.15.0
	golang.org/x/text v0.14.0
	gopkg.in/mcuadros/go-syslog.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.27.0
	mvdan.cc/xurls/v2 v2.5.0
)

require (
	codeberg.org/gruf/go-atomics v1.1.0 // indirect
	codeberg.org/gruf/go-bitutil v1.1.0 // indirect
	codeberg.org/gruf/go-bytes v1.0.2 // indirect
	codeberg.org/gruf/go-fastpath/v2 v2.0.0 // indirect
	codeberg.org/gruf/go-mangler v1.2.3 // indirect
	codeberg.org/gruf/go-maps v1.0.3 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.10.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/containerd/cgroups/v3 v3.0.1 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cornelk/hashmap v1.0.8 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dsoprea/go-exif/v3 v3.0.0-20210625224831-a6301f85c82b // indirect
	github.com/dsoprea/go-iptc v0.0.0-20200610044640-bc9ca208b413 // indirect
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd // indirect
	github.com/dsoprea/go-photoshop-info-format v0.0.0-20200610045659-121dd752914d // indirect
	github.com/dsoprea/go-utility/v2 v2.0.0-20200717064901-2fccff4aa15e // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-errors/errors v1.4.1 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.15.5 // indirect
	github.com/go-xmlfmt/xmlfmt v0.0.0-20211206191508-7fd73a941850 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/quasoft/memstore v0.0.0-20191010062613-2bce066d2b0b // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/superseriousbusiness/go-jpeg-image-structure/v2 v2.0.0-20220321154430-d89a106fdabe // indirect
	github.com/superseriousbusiness/go-png-image-structure/v2 v2.0.1-SSB // indirect
	github.com/tdewolff/parse/v2 v2.7.6 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.2.2 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/arch v0.5.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.29.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.7.2 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
)
