# KarotteSSH

Easy, customizable SSH server framework written in Go.

KarotteSSH provides the plumbing required to run an SSH server while allowing you to customize:

* Authentication
* Session request handling
* Shell behavior
* Command execution behavior
* Session state management

The library is built on top of `golang.org/x/crypto/ssh`.

## Features

* SSH server implementation
* Public key authentication
* Password authentication using bcrypt hashes
* Optional no-auth mode
* Custom SSH request handlers
* Access to per-session state
* Built-in handlers for:

  * `pty-req`
  * `window-change`
  * `exec`
  * `shell`

---

## Installation

```bash
go get github.com/karotte128/karottessh
```

---

## Initial Setup

KarotteSSH expects its authentication and host key files to be stored in a local `.ssh` directory.

Create the directory:

```bash
mkdir -p .ssh
```

The resulting structure should look like:

```text
.ssh/
├── key
├── authorized_keys
└── passwords
```

---

### Generate a Host Key

The host key identifies your SSH server to clients.

Generate an ED25519 host key:

```bash
ssh-keygen -t ed25519 -f .ssh/key
```

You will be asked for a passphrase. For unattended servers, leave it empty.

This creates:

```text
.ssh/key
.ssh/key.pub
```

KarotteSSH uses only the private key:

```text
.ssh/key
```

If you want to use a different path, set:

```go
config.PrivateKeyFile = "/path/to/private/key"
```

---

### Configure Public Key Authentication

Public key authentication is enabled by default.

Create an authorized keys file:

```bash
touch .ssh/authorized_keys
```

Add one or more public keys:

```text
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB... alice@example.com
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... bob@example.com
```

#### Generating a Client Key Pair

If you do not already have a key pair:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
```

Append the public key to the server's authorized keys file:

```bash
cat ~/.ssh/id_ed25519.pub >> .ssh/authorized_keys
```

#### Testing Public Key Login

```bash
ssh -p 2222 localhost
```

or explicitly:

```bash
ssh -i ~/.ssh/id_ed25519 -p 2222 localhost
```

---

### Configure Password Authentication

Password authentication is disabled by default.

Enable it:

```go
config.Authentication.EnablePasswordAuth = true
```

Create the password file:

```bash
touch .ssh/passwords
```

Each line contains:

```text
username:bcrypt_hash
```

Example:

```text
alice:$2a$10$XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
bob:$2a$10$YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
```

---

### Generating Password Hashes

Passwords must be stored as bcrypt hashes.

### Using Go

You can generate a hash programmatically:

```go
package main

import (
    "fmt"

    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword(
        []byte("my-secret-password"),
        bcrypt.DefaultCost,
    )

    fmt.Println(string(hash))
}
```

---

### Enable Both Authentication Methods

You can allow both public key and password authentication:

```go
config.Authentication.EnablePublicKeyAuth = true
config.Authentication.EnablePasswordAuth = true
```

SSH clients will be able to authenticate using either method.

---

### Disable Authentication (Development Only)

For local testing:

```go
config.Authentication.EnableNoAuth = true
```

This allows any client to connect without credentials.

Do not enable this on publicly accessible servers.

---

## Quick Start

```go
package main

import (
    "log"

    "github.com/karotte128/karottessh"
)

func main() {
    config := karottessh.NewConfig()

    log.Println("Starting SSH server...")
    karottessh.RunServer(config)
}
```

By default this starts an SSH server on:

```text
0.0.0.0:2222
```

using the host key:

```text
.ssh/key
```

and public-key authentication against:

```text
.ssh/authorized_keys
```

---

## Default Behavior

### Shell Requests

The built-in shell handler behaves as a simple echo server.

Anything typed by the client is written back unchanged.

### Exec Requests

The built-in exec handler does not execute commands.

Instead it returns:

```text
exec command: "<command>"
```

and exits successfully.

This handler is intended as an example implementation.

---

# Configuration

Create a configuration using:

```go
config := karottessh.NewConfig()
```

## Config

```go
type Config struct {
    Port            int
    PrivateKeyFile  string
    RequestHandlers RequestHandlers
    Authentication  Authentication
}
```

### Port

Listening port.

Default:

```go
2222
```

### PrivateKeyFile

Path to the SSH host private key.

Default:

```go
".ssh/key"
```

### RequestHandlers

Map of SSH request types to handler functions.

### Authentication

Authentication configuration.

---

# Authentication

## Authentication Structure

```go
type Authentication struct {
    Attributes map[string]string

    PublicKeyHandler func(
        conn ssh.ConnMetadata,
        key ssh.PublicKey,
        attributes map[string]string,
    ) (*ssh.Permissions, error)

    EnablePublicKeyAuth bool

    PasswordHandler func(
        conn ssh.ConnMetadata,
        password []byte,
        attributes map[string]string,
    ) (*ssh.Permissions, error)

    EnablePasswordAuth bool

    EnableNoAuth bool
}
```

---

## Public Key Authentication

Enabled by default.

Default configuration:

```go
config.Authentication.EnablePublicKeyAuth = true
```

The default implementation reads authorized keys from:

```text
.ssh/authorized_keys
```

The path is stored in:

```go
config.Authentication.Attributes["authorized_keys_file"]
```

### Custom Public Key Authentication

```go
config.Authentication.PublicKeyHandler = func(
    conn ssh.ConnMetadata,
    key ssh.PublicKey,
    attrs map[string]string,
) (*ssh.Permissions, error) {

    // validate key

    return &ssh.Permissions{}, nil
}
```

---

## Password Authentication

Disabled by default.

Enable it:

```go
config.Authentication.EnablePasswordAuth = true
```

The default password handler reads a password file whose location is stored in:

```go
config.Authentication.Attributes["password_file"]
```

Default:

```text
.ssh/passwords
```

### Password File Format

```text
alice:$2a$10$...
bob:$2a$10$...
```

Each password must be stored as a bcrypt hash.

### Custom Password Authentication

```go
config.Authentication.PasswordHandler = func(
    conn ssh.ConnMetadata,
    password []byte,
    attrs map[string]string,
) (*ssh.Permissions, error) {

    // validate password

    return &ssh.Permissions{}, nil
}
```

---

## No Authentication

For testing:

```go
config.Authentication.EnableNoAuth = true
```

This allows clients to connect without authentication.

Do not use this in production.

---

# Session Handling

Each SSH session is represented by:

```go
type Session struct {
    Channel ssh.Channel
    Storage map[string]any
}
```

## Closing a Session

When a handler has finished processing a request, it should close the session using an SSH exit status.

```go
state.Close(0)
```

The exit status is sent to the client before the channel is closed.

### Success

An exit status of `0` indicates successful execution:

```go
state.Channel.Write([]byte("Hello!\n"))
state.Close(0)
```

### Failure

Non-zero values indicate an error:

```go
state.Channel.Write([]byte("Invalid command\n"))
state.Close(1)
```

### Example

```go
config.RequestHandlers["exec"] = func(
    state *karottessh.Session,
    req ssh.Request,
) {
    req.Reply(true, nil)

    if err := doSomething(); err != nil {
        state.Channel.Write([]byte(err.Error() + "\n"))
        state.Close(1)
        return
    }

    state.Channel.Write([]byte("Success\n"))
    state.Close(0)
}
```

SSH clients can access this status just like the exit code of a normal process.

For example:

```bash
ssh -p 2222 localhost some-command
echo $?
```

The value printed by `echo $?` will be the exit status passed to `Session.Close()`.

## Session Storage

Handlers can share state through:

```go
state.Storage
```

The built-in handlers use it to store:

```go
state.Storage["term"]
state.Storage["width"]
state.Storage["height"]
```

after PTY requests.

---

# Custom Request Handlers

Request handlers are registered through:

```go
type RequestHandlers map[string]func(
    state *Session,
    req ssh.Request,
)
```

Example:

```go
config.RequestHandlers["shell"] = func(
    state *karottessh.Session,
    req ssh.Request,
) {
    req.Reply(true, nil)

    state.Channel.Write([]byte("Hello!\n"))
    state.Close(0)
}
```

---

## Example: Custom Exec Handler

```go
config.RequestHandlers["exec"] = func(
    state *karottessh.Session,
    req ssh.Request,
) {
    req.Reply(true, nil)

    state.Channel.Write([]byte(
        "Custom exec handler\n",
    ))

    state.Close(0)
}
```

---

## Supported Request Types

The SSH request type is used as the key in the handler map.

Examples:

```go
"pty-req"
"window-change"
"shell"
"exec"
```

Any request type not present in the map is rejected automatically.

---

# Complete Example

```go
package main

import (
    "log"

    "github.com/karotte128/karottessh"
    "golang.org/x/crypto/ssh"
)

func main() {
    config := karottessh.NewConfig()

    config.RequestHandlers["shell"] = func(state *karottessh.Session, req ssh.Request) {
        req.Reply(true, nil)

        state.Channel.Write([]byte(
            "Welcome to KarotteSSH!\n",
        ))

        state.Close(0)
    }

    log.Fatal(karottessh.RunServer(config))
}
```

---

# Directory Layout

Expected default files:

```text
.ssh/
├── key
├── authorized_keys
└── passwords
```

### key

SSH server private key.

### authorized_keys

Authorized public keys for public-key authentication.

### passwords

User-to-bcrypt-hash mappings for password authentication.

---

# License

MIT License.
