	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true:. \
	       $(find api -name *.proto | grep -v errors)

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