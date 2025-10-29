# Common FTP Command Reference Table

| Command | Full Name / Meaning       | Description                                                                |
|---------|---------------------------|----------------------------------------------------------------------------|
| `USER`  | User login                | Specifies username for authentication.                                     |
| `PASS`  | Password                  | Provides password for the given user.                                      |
| `TYPE`  | Transfer Type             | Sets file transfer mode (ASCII or Binary). Usually `TYPE I` for binary.    |
| `PASV`  | Passive Mode              | Server opens a data port and tells client how to connect.                  |
| `PORT`  | Active Mode               | Client tells server which port it listens on for data connection.          |
| `LIST`  | List Directory            | Lists files and directories in current folder.                             |
| `CWD`   | Change Working Directory  | Changes current directory on the server.                                   |
| `PWD`   | Print Working Directory   | Returns the current directory path on the server.                          |
| `MKD`   | Make Directory            | Creates a new directory.                                                   |
| `RMD`   | Remove Directory          | Deletes an existing directory.                                             |
| `DELE`  | Delete File               | Removes a specific file.                                                   |
| `RNFR`  | Rename From               | Indicates the source file to rename or move. Must be followed by `RNTO`.   |
| `RNTO`  | Rename To                 | Specifies the destination name/path for `RNFR`. Together they act as “mv”. |
| `RETR`  | Retrieve File             | Downloads a file from the server.                                          |
| `STOR`  | Store File                | Uploads a file to the server.                                              |
| `QUIT`  | Quit Session              | Ends the FTP session gracefully.                                           |
| `SYST`  | System Type               | Returns system information (e.g., UNIX).                                   |
| `NOOP`  | No Operation              | Keeps connection alive; server replies `200 OK`.                           |
