# Reservoird

Reservoird is an opinionated light weight pluggable streamm processing
framework written in golang

## Terminology

The following are the 4 plugin types

- ingester - how data is brought into reservoird
- digester - how data is transformed within reservoird
- expeller - how data is pushed out of reservoird
- queue - how data is passed through reservoird

## Data Flow

How data flows through the system

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBpbjAoaW5wdXQwKSAtLT4gaWcwW2luZ2VzdGVyMF1cbiAgICBpbjEoaW5wdXQxKSAtLT4gaWcxW2luZ2VzdGVyMV1cbiAgICBpbm4oaW5wdXROKSAtLT4gaWduW2luZ2VzdGVyTl1cbiAgICBpZzAgLS0-IGRpMFtkaWdlc3RlcjBdXG4gICAgaWcxIC0tPiBkaTFbZGlnZXN0ZXIxXVxuICAgIGlnbiAtLT4gZGluW2RpZ2VzdGVyTl1cbiAgICBkaTAgLS0-IGV4W2V4cGVsbGVyXVxuICAgIGRpMSAtLT4gZXhbZXhwZWxsZXJdXG4gICAgZGluIC0tPiBleFtleHBlbGxlcl1cbiAgICBleCAtLT4gbyhvdXRwdXQpIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBpbjAoaW5wdXQwKSAtLT4gaWcwW2luZ2VzdGVyMF1cbiAgICBpbjEoaW5wdXQxKSAtLT4gaWcxW2luZ2VzdGVyMV1cbiAgICBpbm4oaW5wdXROKSAtLT4gaWduW2luZ2VzdGVyTl1cbiAgICBpZzAgLS0-IGRpMFtkaWdlc3RlcjBdXG4gICAgaWcxIC0tPiBkaTFbZGlnZXN0ZXIxXVxuICAgIGlnbiAtLT4gZGluW2RpZ2VzdGVyTl1cbiAgICBkaTAgLS0-IGV4W2V4cGVsbGVyXVxuICAgIGRpMSAtLT4gZXhbZXhwZWxsZXJdXG4gICAgZGluIC0tPiBleFtleHBlbGxlcl1cbiAgICBleCAtLT4gbyhvdXRwdXQpIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifX0)

## Best Practices

1. Ingesters should only capture data. All filtering/annotating should be
completed in digesters.
2. Queues can be either blocking or non-blocking. If non-blocking end with nb.
3. All names should be unique. If your plugin is located at github.com/me/xyz
then following are the recommended names for each plugin type:
  - com.github.me.xyz.ingest (ingester)
  - com.github.me.xyz.digest (digester)
  - com.github.me.xyz.expel (expeller)
  - com.github.me.xyz.queue (blocking queue)
  - com.github.me.xyznb.queue (non-blocking queue)
