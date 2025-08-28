export CGO_ENABLED=0

export GOOS=linux
export GOARCH=amd64
export COMMIT=`git rev-parse --short HEAD`
mkdir ./dist &> /dev/null
go build -o ./dist/bepusdt_${GOOS}_${GOARCH} --trimpath --ldflags="-X main.COMMIT=${COMMIT} -s -w" ./main