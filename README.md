# Auto update joplin notify in S3

Joplin notify auto updater written in golang  
Keep your todo list clear

- automatically aggregate notes with markdown todo lists
- prioritize tasks


If you're looking just for joplin s3 data integration, see [joplin_privider](/joplin_privider) package

## Note format
```markdown
- [x] format like this. This note not auto listed
- [ ] joplin creates it with default editor. This note still not listed
- [ ] !this note will be listed because ! mark
- [ ] !!this note will be higher
- [ ] !!!this note will be on top
```

## Requirements
- joplin sync with s3

## Prepare
1) create few joplin notebook, add few notes
2) create joplin note for auto update
3) pick notebook ID, note for auto update ID (from s3 or joplin backups)

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