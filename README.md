# Reservoird

Reservoird is an opinionated light weight pluggable streamm processing
framework written in golang

## Terminology

The following are the 4 plugin types

- ingester - how data is brought into reservoird
- digester - how data is transformed within reservoird
- expeller - how data is pushed out of reservoird
- queue - how data is passed through reservoird

## Data flow

```mermaid
graph LR
input0 --> ingester0
input1 --> ingester1
inputN --> ingesterN
ingester0 --> digester0
ingester1 --> digester1
ingesterN --> digesterN
digester0 --> expeller
digester1 --> expeller
digesterN --> expeller
expeller --> output
```
