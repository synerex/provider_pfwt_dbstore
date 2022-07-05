module pfwt_dbstore

go 1.15

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/UCLabNU/proto_pflow v0.0.5
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/kr/text v0.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20220517141722-cf486979b281 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/shirou/gopsutil v3.20.12+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.22.6 // indirect
	github.com/synerex/proto_pcounter v0.0.6
	github.com/synerex/proto_storage v0.2.0
	github.com/synerex/synerex_api v0.5.1
	github.com/synerex/synerex_nodeapi v0.5.6 // indirect
	github.com/synerex/synerex_proto v0.1.12
	github.com/synerex/synerex_sxutil v0.7.0
	github.com/tklauser/numcpus v0.5.0 // indirect
	golang.org/x/net v0.0.0-20220630215102-69896b714898 // indirect
	golang.org/x/sys v0.0.0-20220704084225-05e143d24a9e // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20220630174209-ad1d48641aa7 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
)

replace github.com/synerex/proto_pcounter v0.0.6 => github.com/nagata-yoshiteru/proto_pcounter v0.0.10

replace github.com/synerex/synerex_proto v0.1.12 => github.com/nagata-yoshiteru/synerex_proto v0.1.16
