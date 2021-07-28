NAME=ssh-forward
BUILDDIR=build


GOBUILD=CGO_ENABLED=0 go build -trimpath

PLATFORM_LIST = \
	darwin-amd64 \
	linux-amd64 \
	linux-arm64

all-arch: $(PLATFORM_LIST)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@ .

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@ .

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/$(NAME)-$@ .

.PHONY: clean
clean:
	-rm -rf $(BUILDDIR)
