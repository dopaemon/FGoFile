# FGoFile

## Install
* **Request Debian Base Distro**
* Add Key GPG:
```bash
curl -fsSL https://dopaemon.github.io/PPA/KEY.gpg | sudo gpg --dearmor -o /usr/share/keyrings/dopaemon.gpg
```
* Add Source PPA:
```bash
echo "deb [signed-by=/usr/share/keyrings/dopaemon.gpg] https://dopaemon.github.io/PPA ./" | sudo tee /etc/apt/sources.list.d/dopaemon.list
```
* Update Package PPA:
```bash
sudo apt update
```
* Install fgofile:
```bash
sudo apt install fgofile -y
```

## Build
* **Request GoLang 1.25**
* Build binary:
```bash
go build -v
```
* Build debian package:
```bash
make build
```

## Get started quickly
* **Run Server:**
```bash
mkdir -p ftp_root && echo Hello > ftp_root/readme.txt

./fgofile --server --port 2121 --root ./ftp_root --suser user --spass pass

./fgofile --server --port 2121 --root ./ftp_root
```

* **Run Client:**
```bash
./fgofile 127.0.0.1 --port 2121 --cuser user --cpass pass

./fgofile 127.0.0.1 --port 2121
Username (Enter for anonymous): user
Password: pass
```

* **Support CommandLine:**
```bash
ftp> ls
ftp> mkdir newfolder
ftp> mv old.txt new.txt
ftp> cd newfolder
ftp> put example.txt
ftp> get readme.txt
ftp> quit
```

# Docker Compose
* **Run Container With Docker Compose + h5ai Web Server:**
```bash
docker compose up -d
```

# FTP Command table
* [**FTP Command**](FTPCommand.md)

# LICENCE
* [**MIT LICENCE**](LICENCE)
