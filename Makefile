build:
	go build -o dist/ffdownload .

dockerbuild:
	docker build -t joyme/ffdownload:1.0 .