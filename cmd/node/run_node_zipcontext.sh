# git clone https://github.com/meverselabs/meverse
# cd ./meverse/cmd/node
wget https://rpc.meversemainnet.io/zipcontext
unzip zipcontext
mv ./data ./ndata
sed -e "s/InitGenesisHash = \"\"/$(sed '1!d' _config.toml)/" config.toml > config.toml_ && mv config.toml_ config.toml
sed -e "s/InitHeight = 0/$(sed '2!d' _config.toml)/" config.toml > config.toml_ && mv config.toml_ config.toml
sed -e "s/InitHash = \"\"/$(sed '3!d' _config.toml)/" config.toml > config.toml_ && mv config.toml_ config.toml
sed -e "s/InitTimestamp = 0/$(sed '4!d' _config.toml)/" config.toml > config.toml_ && mv config.toml_ config.toml
# sed -e "s/RPCPort = 8541/RPCPort = 8541/" config.toml > config.toml_ && mv config.toml_ config.toml
rm -r _config.toml
rm -r zipcontext
go build
./node
