# TriasCode
通过实现ABCI来完成与tendermint进行结合。
主要有三类交易，一类是合约的部署、执行、查询相关交易。一类是utxo帐户体系的实现。再一类是纯byte类型的交易，这种是tendermint默认的交易。
目前版本与合约、UTXO的通信为http，这个在下一个版本会修改为grpc。
# 使用
直接到cmd下go build二进制之后执行即可。其启动后terdermint通过tcp:46658端口连续abci server

