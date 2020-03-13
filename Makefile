githash = $(shell git rev-parse --short HEAD)
ldflags = -X main.BuildTag=$(githash)
current_dir = $(shell pwd)
DOCKER_REGISTRY = 
DOCKER_IMAGE = truongminh/gonpm

build: clean builddocker

builddocker:
	docker build -f Dockerfile -t $(DOCKER_IMAGE) .

clean:
	rm -rf ./dist

up:
	docker run -p 8999:8080 -v $(current_dir)/data:/data $(DOCKER_IMAGE)

debug:
	docker run -p 8999:8080 -v $(current_dir)/data:/data -it --entrypoint=/bin/sh $(DOCKER_IMAGE)

pushdev:	
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)$(DOCKER_IMAGE):dev
	docker push $(DOCKER_REGISTRY)$(DOCKER_IMAGE):dev

pushstage:
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)$(DOCKER_IMAGE):stage
	docker push $(DOCKER_REGISTRY)$(DOCKER_IMAGE):stage

pushprod:
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)$(DOCKER_NAME):prod
	docker push $(DOCKER_REGISTRY)$(DOCKER_NAME):prod
