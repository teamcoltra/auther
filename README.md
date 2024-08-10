# Authinator

Authinator is a command-line interface (CLI) tool written in Go for managing Time-based One-Time Passwords (TOTP). It allows users to create, retrieve, and manage TOTP codes for two-factor authentication (2FA). Additionally, Authinator includes an HTTP server mode for managing TOTP entries via a RESTful API.

## Features

- **Create TOTP Entries:** Easily add new TOTP entries by specifying a name and a secret key.
- **List TOTP Entries:** View all stored TOTP entries along with their current codes and the time remaining until the next code.
- **Retrieve TOTP Codes:** Get the current TOTP code for a specified entry, with the code automatically copied to your clipboard.
- **Remove TOTP Entries:** Delete a specific TOTP entry by name.
- **Serve via HTTP:** Start an HTTP server to interact with your TOTP entries through REST API commands.

## Installation

To install Authinator, ensure you have Go 1.21 or later installed, and then run:

```bash
go install github.com/teamcoltra/authinator@latest
```

## Usage

After installing, you can use the `authinator` command followed by the desired command and arguments:

```bash
authinator [command] [arguments...]
```

### Commands

- **`create [name] [secret]`**  
  Create a new TOTP entry with the given name and secret.  
  Example:  
  ```bash
  authinator create my_account JBSWY3DPEHPK3PXP
  ```

- **`list`**  
  List all stored TOTP entries with their current codes and time remaining.  
  Example:  
  ```bash
  authinator list
  ```

- **`[name]`**  
  Get the current TOTP code for the entry with the specified name. The code will also be copied to your clipboard automatically.  
  Example:  
  ```bash
  authinator my_account
  ```

- **`remove [name]`**  
  Remove the TOTP entry with the specified name.  
  Example:  
  ```bash
  authinator remove my_account
  ```

- **`serve`**  
  Start an HTTP server on port 8055 to manage TOTP entries via REST API.  
  Example:  
  ```bash
  authinator serve
  ```

- **`help`**  
  Display the help guide with detailed information on how to use each command.  
  Example:  
  ```bash
  authinator help
  ```

## HTTP Server

When running in server mode with the `serve` command, Authinator listens on port 8055 by default and exposes the following endpoints:

- **`GET /totps`**  
  List all TOTP entries.

- **`GET /totps/{name}`**  
  Get the current TOTP code for the specified entry.

- **`POST /totps`**  
  Create a new TOTP entry by sending a JSON payload.  
  Example payload:  
  ```json
  {
    "name": "example",
    "secret": "SECRETKEY"
  }
  ```

- **`DELETE /totps/{name}`**  
  Delete a TOTP entry.

## Example HTTP Requests

Using `curl`, you can interact with the HTTP server as follows:

- List all entries:
  ```bash
  curl -X GET http://localhost:8055/totps
  ```

- Create a new entry:
  ```bash
  curl -X POST -H "Content-Type: application/json" -d '{"name":"example","secret":"SECRETKEY"}' http://localhost:8055/totps
  ```

- Get the TOTP code for an entry:
  ```bash
  curl -X GET http://localhost:8055/totps/example
  ```

- Delete an entry:
  ```bash
  curl -X DELETE http://localhost:8055/totps/example
  ```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## Acknowledgements

- [pquerna/otp](https://github.com/pquerna/otp) - The Go library used for generating TOTP codes.
- [atotto/clipboard](https://github.com/atotto/clipboard) - Clipboard package for copying the generated TOTP codes.

## Author

[teamcoltra](https://github.com/teamcoltra)
