# Auto update joplin notify in S3

Joplin notify auto updater written in golang  
Keep your todo list clear

- automatically aggregate notes with markdown todo lists
- prioritize tasks


If you're looking for joplin s3 sync or web clipper api integration, see [joplin_privider](/joplin_privider) package

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
3) pick notebook ID, note for auto update ID (from s3, API or backups)

## Install
Setup

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

run local env
```shell
make up
```

stop local env
```shell
make up
```