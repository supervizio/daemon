# Logging - Gestion des Logs

Capture et formatage des sorties des processus supervisés.

## Rôle

Capturer stdout/stderr des processus, ajouter timestamps, écrire vers fichiers avec rotation.

## Composants

| Type | Fichier | Rôle |
|------|---------|------|
| `Capture` | `capture.go` | Coordonne stdout/stderr |
| `LineWriter` | `linewriter.go` | Buffer ligne par ligne |
| `MultiWriter` | `multiwriter.go` | Écrit vers plusieurs destinations |
| `TimestampWriter` | `timestamp.go` | Ajoute préfixe horodatage |
| `Writer` | `writer.go` | Writer de base vers fichier |
| `FileOpener` | `fileopener.go` | Ouvre fichiers (prêt rotation) |
| `NopCloser` | `nopcloser.go` | Wrapper sans Close() |

## Usage

```go
// Créer un writer avec timestamp
tw := logging.NewTimestampWriter(file, "2006-01-02 15:04:05")
lw := logging.NewLineWriter(tw)

// Attacher au processus
cmd.Stdout = lw
cmd.Stderr = lw
```

## Chaîne de Writers

```
Process stdout/stderr
         │
         ▼
    LineWriter       ← Buffer jusqu'à \n
         │
         ▼
  TimestampWriter    ← Ajoute [2024-01-15 10:30:45]
         │
         ▼
    MultiWriter      ← Fichier + Console
         │
         ▼
      Writer         ← Fichier avec rotation
```

## Configuration YAML

```yaml
logging:
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: 100MB
      max_files: 10
```

## Constructeurs

```go
NewCapture(stdout, stderr io.Writer) *Capture
NewLineWriter(w io.Writer) *LineWriter
NewTimestampWriter(w io.Writer, format string) *TimestampWriter
NewMultiWriter(writers ...io.Writer) *MultiWriter
```
