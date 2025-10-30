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

* **Run Server:**
```bash
# Tạo thư mục gốc và file mẫu
mkdir -p ftp_root && echo Hello > ftp_root/readme.txt

# Chạy server bắt buộc user/password
./fgofile --server --port 2121 --root ./ftp_root --suser user --spass pass

# Hoặc server không đặt tài khoản → cho phép anonymous login
./fgofile --server --port 2121 --root ./ftp_root
```

* **Run Client:**
```bash
# Kết nối tới server (truyền user/pass trực tiếp)
./fgofile 127.0.0.1 --port 2121 --cuser user --cpass pass

# Hoặc không truyền → chương trình sẽ hỏi nhập
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

# FTP Command table
* [**FTP Command**](FTPCommand.md)

# LICENCE
* [**MIT LICENCE**](LICENCE)
