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
* Clone Source:
```bash
git clone -b main --single-branch --recurse-submodules https://github.com/dopaemon/FGoFile.git
```
* Build binary:
```bash
make build-binary
```
* Build debian package:
```bash
make build-deb -j$(nproc --all) | tee log.txt
```

## Get started quickly
* **Run Server:**
```bash
mkdir -p ftp_root && echo Hello > ftp_root/readme.txt

# With user / password (Administrator Mode)
./fgofile --server --port 2121 --root ./ftp_root --suser user --spass pass
./fgofile --server --port 2121 --root ./ftp_root --suser user --spass pass

# Or without user / password (Anon Mode)
./fgofile --server --port 2121 --root ./ftp_root
```

* **Run Client:**
```bash
# With Administrator mode
./fgofile --host 127.0.0.1 --port 2121 --cuser user --cpass pass
# Username (Enter for anonymous): user
# Password: pass

# Or with Anon mode
./fgofile --host 127.0.0.1 --port 2121
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
* Read more using `help` command.

# Docker Compose
* Default user and password when you run `FGoFile` with Docker compose is `user` and `pass`
* **Build and Run Container With Docker Compose + h5ai Web Server:**
```bash
make up
```
* **Stop docker compose:**
```bash
make down
```
* File Storage Folder is `data`, edit in [**docker-compose.yml**](docker-compose.yml) file.
* The `data` folder must have read and write permissions so that fgofile can write data.
```bash
chmod -R 777 ./data
```

# FTP Command table
* [**FTP Command**](FTPCommand.md)

# LICENCE
* [**MIT LICENCE**](LICENCE)
