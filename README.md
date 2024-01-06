# Auto update joplin notify

Joplin notify auto updater written in golang  
Supported protocols: S3, web clipper API

- automatically aggregate notes with markdown todo lists
- prioritize tasks

If you're looking for joplin S3 sync or web clipper api integration, see [joplin_provider](/joplin_provider) package

## Note format

```markdown
- [x] format like this. This note not auto listed
- [ ] joplin creates it with default editor. This note still not listed
- [ ] !this note will be listed because ! mark
- [ ] !!this note will be higher
- [ ] !!!this note will be on top
```

## Requirements

- joplin sync with S3
- or web clipper API

## Prepare

1) create joplin notebooks, add some notes in markdown todo format
2) create joplin note for auto update
3) pick notebook ID, note for auto update ID (from S3, API or backups)

## Install

configure

```shell
cp internal/config/config.template.yaml config.yaml
```

run

``` shell
go run .
```

``` shell
go run . --provider=web_clipper
```

## Local env

see [docker](/docker)

run local env

```shell
make up
```

stop local env

```shell
make up
```
