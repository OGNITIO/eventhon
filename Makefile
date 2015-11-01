.PHONY: build push

IMAGE = eventos
TAG = latest

build:
	docker build -t mesos-master-1.ognitio.net:5000/containers/$(IMAGE):$(TAG) .

push:
	docker push mesos-master-1.ognitio.net:5000/containers/$(IMAGE):$(TAG)
