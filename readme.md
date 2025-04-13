# tex

`tex` is a code context manager for getting code into AI (right now local ollama)

# Install
```bash
go install github.com/nanvenomous/tex@latest
```

# Usage
Just chat with a model
```bash
tex "how are you today?"
```

give tex a file (or multiple) as context
```bash
tex --file my/awesome/code.go "what does this awesome code do?"
```

see all the context in your editor (defined by `EDITOR` environment variable) before sending to the model
```bash
tex --file my/awesome/code.go --file my/other/code.go
```

# Config

as of now we only connect to ollama see an example config file below

[example/tex.yml](https://github.com/nanvenomous/tex/blob/5c9776328650f016acb8b5545d3b90ee21b3758a/example/tex.yml)
https://github.com/nanvenomous/tex/blob/1c10bfa1195658e745923c58acc64b7dcbc20624/example/tex.yml#L1-L5

# Vision
- [ ] somehow find a way to stream to the editor (mine is neovim)
- [ ] find a way to leverage a single lsp (probably `gopls` to start) and add a `--code` flag which can reference `functions`, `structs` ...
    - thought about doing it with `go/ast` but language server protocol would be more general
- [ ] power code search with a local vector store
    - possibly manual indexing or on file save
    - [milvus](https://github.com/milvus-io/milvus)?
