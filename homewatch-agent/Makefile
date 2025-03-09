Version=0.2-$(shell git rev-parse --short HEAD)
Program=homewatch-agent
CameraManagerUser?=cameras
CameraManagerHost?=CameraManager
Output=_dist
build:
	go build \
		-ldflags "-X main.softwareVersion=$(Version)" \
		-o ${Output}/${Program} \
		event_handler.go \
		file_event_handler.go \
		index.go \
		main.go \
		message_handler.go \
		messages.go \
		publishers.go \
		s3_uploader.go \
		syslog.go \
		video_event_handler.go 

build-pi:
	GOOS=linux GOARCH=arm go build -o ${Output}/${Program}-linux-arm64 \
		event_handler.go \
		file_event_handler.go \
		index.go \
		main.go \
		message_handler.go \
		messages.go \
		publishers.go \
		s3_uploader.go \
		syslog.go \
		video_event_handler.go 

build-linux:
	GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.softwareVersion=$(Version)" \
		-o ${Output}/${Program}-linux-amd64 \
		event_handler.go \
		file_event_handler.go \
		index.go \
		main.go \
		message_handler.go \
		messages.go \
		publishers.go \
		s3_uploader.go \
		syslog.go \
		video_event_handler.go 

deploy-scp: 
	scp ${Program} ${CameraManagerUser}@${CameraManager}:/opt/homewatch/${Program}

deploy-sftp: 
	sftp ${CameraManagerUser}@${CameraManager} <<< "put ${Program} ${Program}"
