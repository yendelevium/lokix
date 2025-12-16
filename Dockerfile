# Build the application from source
FROM golang:1.24.3 AS build-stage

WORKDIR /go/src/lokix

# Dependency Caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# vet examines Go source code and reports suspicious constructs, such as Printf calls whose arguments do not align with the format string. 
# vet uses heuristics that do not guarantee all reports are genuine problems, but it can find errors not caught by the compilers. 
RUN go vet -v ./... 
# RUN go test -v

# CGO_ENABLED=0 as my project doesn't use cgo, and I'll be running the binary on a distroless base-image
# This creates a statically linked binary suitable for distroless
# IF a library uses CGO, the build will FAIL
RUN CGO_ENABLED=0 go build -o /go/bin/lokix cmd/lokix/main.go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploying the application binary into a lean image
# Using a distroless base-image to minimize the image size and maximize security
# Copying only the binary to this image from the build stage
FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /go/bin/lokix /lokix

USER nonroot:nonroot

ENTRYPOINT ["/lokix"]