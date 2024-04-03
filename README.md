# Messages
Messages is a simple go library/command that helps with the translation of messages and errors in your application.

It converts simple translations files like this:
```yaml
NotFound:
    default: User %(ID)d not found
    nl: Gebruiker %(ID)d niet gevonden
```

Into go code like this:
```go
func UserNotFound(ctx context.Context, ID int) string {
    switch messages.GetLocale(ctx) {
    case "nl":
        return fmt.Sprintf("Gebruiker %(ID)d niet gevonden", ID)
    }

    return fmt.Sprintf("User %d not found.", ID)
}
```

This means that using this in your code is type safe.
```go
func GetUser(ctx context.Context, ID int64) (*User, error) {
    user := db.GetUser(ID)
    if user == nil {
        return nil, errors.New(translations.UserNotFound(ctx, ID))
    }

    return user, nil
}
```

# Usage
Install the latest version of msggen.
```bash
$ go install github.com/wvell/messages/cmd/msggen@latest

# Test your installation.
$ msggen --help

# Write a sample file (in a new empty directory).
$ cat > sample.yml << EOF
NotFound:
  default: User %(ID)d not found
  nl: Gebruiker %(ID)d niet gevonden
EOF

# Generate the template file inside the same directory.
$ msggen -pkg translations
```

## Idea
The idea for this package comes from [this talk](https://youtu.be/RpmYXh0ppRo?t=1830) by Alan Shreve.