# Devcontainer quickstart

1. Open repo in VS Code and choose "Reopen in Container".
2. In the container:
   - `cp config.toml.example config.toml`
   - Edit `config.toml` and set `data_dir = "/workspaces/podsync/data"`
   - `make build`
   - `./bin/podsync --config config.toml`
3. Open `http://localhost:8080`.
