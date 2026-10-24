	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./api \
	       $(find api -name *.proto | grep -v errors)

# 用户模块的 pb 写到 ../user/api，不回写 server。
protoc --proto_path=../user/api \
       --proto_path=./third_party \
       --go_out=paths=source_relative:../user/api \
       --go-http_out=paths=source_relative:../user/api \
       --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../user/api \
       user/user.proto

# openapi 合并商城接口和用户接口。用户 proto 的 base 依赖走 server/api。
protoc --proto_path=./api \
       --proto_path=./third_party \
       --openapi_out=fq_schema_naming=true:. \
       $(find api -name *.proto | grep -v errors) \
       ../user/api/user/user.proto

#	protoc --proto_path=./api \
#	       --proto_path=./third_party \
#	       --connect-go_out=paths=source_relative:./api/ \
#	       $(find api -name *.proto | grep -v errors)



#protoc --proto_path=. \
#         --proto_path=./third_party \
#         --go_out=paths=source_relative:. \
#         --go-errors_out=paths=source_relative:. \
#         api/base/errors.proto


exit 0