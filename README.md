# Reservoird

Reservoird is an opinionated light weight pluggable stream processing
framework written in golang

## Terminology

The following are the 4 plugin types

- ingester - how data is brought into reservoird
- digester - how data is transformed within reservoird
- expeller - how data is pushed out of reservoird
- queue - how data is passed through reservoird

## Design

How data flows through the system

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBpbjAoaW5wdXQwKSAtLT4gaWcwW2luZ2VzdGVyMF1cbiAgICBpbjEoaW5wdXQxKSAtLT4gaWcxW2luZ2VzdGVyMV1cbiAgICBpbm4oaW5wdXROKSAtLT4gaWduW2luZ2VzdGVyTl1cbiAgICBpZzAgLS0-IGRpMFtkaWdlc3RlcjBdXG4gICAgaWcxIC0tPiBkaTFbZGlnZXN0ZXIxXVxuICAgIGlnbiAtLT4gZGluW2RpZ2VzdGVyTl1cbiAgICBkaTAgLS0-IGV4W2V4cGVsbGVyXVxuICAgIGRpMSAtLT4gZXhbZXhwZWxsZXJdXG4gICAgZGluIC0tPiBleFtleHBlbGxlcl1cbiAgICBleCAtLT4gbyhvdXRwdXQpIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBpbjAoaW5wdXQwKSAtLT4gaWcwW2luZ2VzdGVyMF1cbiAgICBpbjEoaW5wdXQxKSAtLT4gaWcxW2luZ2VzdGVyMV1cbiAgICBpbm4oaW5wdXROKSAtLT4gaWduW2luZ2VzdGVyTl1cbiAgICBpZzAgLS0-IGRpMFtkaWdlc3RlcjBdXG4gICAgaWcxIC0tPiBkaTFbZGlnZXN0ZXIxXVxuICAgIGlnbiAtLT4gZGluW2RpZ2VzdGVyTl1cbiAgICBkaTAgLS0-IGV4W2V4cGVsbGVyXVxuICAgIGRpMSAtLT4gZXhbZXhwZWxsZXJdXG4gICAgZGluIC0tPiBleFtleHBlbGxlcl1cbiAgICBleCAtLT4gbyhvdXRwdXQpIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifX0)

## Best Practices

1. Ingesters should only bring in data
2. Expellers should only push out data
3. Digesters should only transform/filter/annotate data
4. Queues can be either blocking or non-blocking. If non-blocking end with nb.
5. All names should be unique. If your plugin is located at github.com/me/xyz
then following is the recommended name: com.github.me.xyz
