# KCheck

Check K files from hash

## 📄 Supported format

- KCheck
- metadata
- filepath

## 📝 KCheck format

### example
make `kcheck.list`
```
{
  "createdAt": 1619071970000,
  "files": [
    {
      "path": "/foo/bar/baz.png",
      "sha1": "1a2b3c4d5e1a2b3c4d5e1a2b3c4d5e1a2b3c4d5e",
      "size": 12345
    }
    ...
  ]
}
```
## ⤴️ Output
If some files got failed, a `failed.list` file will be generated automatically

## License

GPLv3

PR Welcome
