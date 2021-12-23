build:
	@gxt build -system "linux" -name "project-ci"
publish: build
	scp bin/linux_amd64/project-ci root@127.0.0.1:/root/services/project-ci