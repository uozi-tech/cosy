# for s in $(go list ./...); do if ! go test -failfast -v -p 1 $s; then break; fi; done
go test -coverprofile=coverage.out -count=1 ./...
# go tool cover -func=coverage.out
