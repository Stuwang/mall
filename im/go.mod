module mall/im

go 1.23

require (
	github.com/coder/websocket v1.8.12
	github.com/glebarez/sqlite v1.11.0
	github.com/go-kratos/kratos/v2 v2.8.2
	github.com/google/uuid v1.6.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240102182953-50ed04b92917
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.36.4
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/gorm v1.31.2
	mall/file v0.0.0
	mall/kit v0.0.0
	mall/user v0.0.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-playground/form/v4 v4.2.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/silentred/gid v1.0.3 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20231212172506-995d672761c0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
)

replace mall/file => ../modules/file

replace mall/kit => ../modules/kit

replace mall/user => ../modules/user
