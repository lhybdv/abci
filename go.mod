module github.com/lhybdv/abci

go 1.12

require (
	github.com/gogo/protobuf v1.2.0
	github.com/golang/protobuf v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/spf13/cobra v0.0.3
	github.com/tendermint/abci v0.0.0-00010101000000-000000000000
	github.com/tendermint/go-crypto v0.2.1
	github.com/tendermint/go-wire v0.14.1
	github.com/tendermint/iavl v0.12.0
	github.com/tendermint/merkleeyes v0.2.4
	github.com/tendermint/tmlibs v1.0.5
	github.com/urfave/cli v1.20.0
	golang.org/x/net v0.0.0-20181220203305-927f97764cc3
	google.golang.org/grpc v1.17.0
)

replace github.com/tendermint/abci => ./

replace github.com/tendermint/tmlibs => github.com/lhybdv/tmlibs v1.0.5

replace github.com/tendermint/go-wire => github.com/lhybdv/go-wire v0.7.2
