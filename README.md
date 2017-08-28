# gozzmock
Mock server in go!

# Run tests
go test ./...

# validations
golint . && golint ./controller/... && golint ./handler/... && golint ./model/...

# run coverage
go test -coverprofile=controller_cover.out ./controller && go tool cover -html=controller_cover.out && go test -coverprofile=handler_cover.out ./handler && go tool cover -html=handler_cover.out