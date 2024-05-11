# Define the default make action
.PHONY: all
all: build push

# Variables
REPO=026450499422.dkr.ecr.us-east-1.amazonaws.com/options
TAG=$(shell git rev-parse --short HEAD)  # Using git commit hash as image tag, ensure your directory is a git repository
IMAGE_NAME=$(REPO):$(TAG)

# Build the Go binary
.PHONY: build-go
build-go:
	@echo "Building Go binary..."
	go build -o main .

# Build the Docker image
.PHONY: build
build: build-go
	@echo "Building Docker image $(IMAGE_NAME)..."
	docker build -t $(IMAGE_NAME) .

# Log in to AWS ECR
.PHONY: login
login:
	@echo "Logging in to AWS ECR..."
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(REPO)

# Push the Docker image to AWS ECR
.PHONY: push
push: build login
	@echo "Pushing Docker image $(IMAGE_NAME)..."
	docker push $(IMAGE_NAME)

# Clean up the binaries
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f main
