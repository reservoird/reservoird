# Reservoird

Reservoird is an opinionated light weight pluggable stream processing
framework written in golang

## Terminology

The following are the 4 types of plugins used within reservoird

- Ingester - how data is brought into reservoird
- Digester - how data is transformed within reservoird
- Expeller - how data is pushed out of reservoird
- Queue - how data is passed through reservoird

## Design

How data flows through the system

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBpbjAoaW5wdXQwKSAtLT4gaWcwW2luZ2VzdGVyMF1cbiAgICBpbjEoaW5wdXQxKSAtLT4gaWcxW2luZ2VzdGVyMV1cbiAgICBpbm4oaW5wdXROKSAtLT4gaWduW2luZ2VzdGVyTl1cbiAgICBzdWJncmFwaCByZXNlcnZvaXJkXG4gICAgaWcwIC0tPiBkaTBbZGlnZXN0ZXIwXVxuICAgIGlnMSAtLT4gZGkxW2RpZ2VzdGVyMV1cbiAgICBpZ24gLS0-IGRpbltkaWdlc3Rlck5dXG4gICAgZGkwIC0tPiBleFtleHBlbGxlcl1cbiAgICBkaTEgLS0-IGV4W2V4cGVsbGVyXVxuICAgIGRpbiAtLT4gZXhbZXhwZWxsZXJdXG4gICAgZW5kXG4gICAgZXggLS0-IG8ob3V0cHV0KSIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In19)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBpbjAoaW5wdXQwKSAtLT4gaWcwW2luZ2VzdGVyMF1cbiAgICBpbjEoaW5wdXQxKSAtLT4gaWcxW2luZ2VzdGVyMV1cbiAgICBpbm4oaW5wdXROKSAtLT4gaWduW2luZ2VzdGVyTl1cbiAgICBzdWJncmFwaCByZXNlcnZvaXJkXG4gICAgaWcwIC0tPiBkaTBbZGlnZXN0ZXIwXVxuICAgIGlnMSAtLT4gZGkxW2RpZ2VzdGVyMV1cbiAgICBpZ24gLS0-IGRpbltkaWdlc3Rlck5dXG4gICAgZGkwIC0tPiBleFtleHBlbGxlcl1cbiAgICBkaTEgLS0-IGV4W2V4cGVsbGVyXVxuICAgIGRpbiAtLT4gZXhbZXhwZWxsZXJdXG4gICAgZW5kXG4gICAgZXggLS0-IG8ob3V0cHV0KSIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In19)

## Best Practices

1. Ingesters should only bring in data
2. Expellers should only push out data
3. Digesters should only transform/filter/annotate data
4. Queues can be either blocking or non-blocking. If non-blocking end name/repo
with nb.
5. All names should be unique. If your plugin is located at
github.com/reservoird/fwd then the recommended name is
com.github.reservoird.fwd
