export GOOS=linux
export GOARCH=amd64
mkdir ./dist &> /dev/null
go build -o ./dist/epusdt_${GOOS}_${GOARCH} --trimpath --ldflags="-w -s" ./main